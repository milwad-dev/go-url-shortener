package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
	"url-service/internal/config"
	"url-service/internal/database"
	"url-service/internal/handler"
	server2 "url-service/internal/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Load config
	cfg := config.LoadConfig()

	// Connect to DB
	db, err := database.ConnectDB(ctx, cfg.DSN())
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	// SetUp handlers
	handlers := handler.CreateHandler(db)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /url/create", handlers.CreateURL)
	mux.HandleFunc("GET /{code}", handlers.FindURLByCode)

	httpServer := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	grpcServer := server2.NewServer("50051", db)
	grpcServer.RegisterServices()

	// Run both servers; the first one to fail triggers shutdown
	errCh := make(chan error, 2)

	go func() {
		if err := grpcServer.Start(); err != nil {
			errCh <- err
		}
	}()

	go func() {
		log.Println("HTTP server listening on :8080")
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP shutdown: %v", err)
		exitCode = 1
	}
	grpcServer.Stop(shutdownCtx)

	// Close the DB pool last, after no handler can use it
	db.Close()
	log.Println("url-service stopped")

	if exitCode != 0 {
		log.Fatal("exiting with error")
	}
}
