package main

import (
	"errors"
	"log"
	"net/http"
	"redirect_service/internal/config"
	"redirect_service/internal/handler"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg := config.LoadConfig()

	// Create connection
	conn, err := grpc.NewClient(
		cfg.URL.GrpcAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// SetUp handlers
	handlers := handler.NewHandler(conn)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{code}", handlers.FindURLByCode)

	server := &http.Server{
		Addr:              ":8081",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
