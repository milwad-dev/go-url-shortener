package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"redirect_service/internal/config"
	"redirect_service/internal/handler"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.LoadConfig()

	// Create connection
	conn, err := grpc.NewClient(
		cfg.URL.GrpcAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}

	// SetUp handlers
	handlers := handler.NewHandler(conn)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handlers.HealthZ)
	mux.HandleFunc("GET /readyz", handlers.ReadyZ)
	mux.HandleFunc("POST /create", handlers.CreateURL)
	mux.HandleFunc("GET /{code}", handlers.FindURLByCode)

	server := &http.Server{
		Addr:              ":8081",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Println("HTTP server listening on :8081")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	// Wait for a signal or a server failure
	exitCode := 0
	select {
	case <-ctx.Done():
		log.Println("Shutdown signal received")
	case err := <-errCh:
		log.Printf("Server error: %v", err)
		exitCode = 1
	}
	stop()

	// Stop accepting new requests and let in-flight ones finish
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP shutdown: %v", err)
		exitCode = 1
	}

	// Close the gRPC connection last, after no handler can use it
	if err := conn.Close(); err != nil {
		log.Printf("gRPC connection close: %v", err)
	}
	log.Println("redirect-service stopped")

	if exitCode != 0 {
		log.Fatal("exiting with error")
	}
}
