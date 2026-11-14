package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"redirect_service/internal/config"
	"time"

	urlv1 "github.com/milwad-dev/url-shortener/contracts/url/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg := config.LoadConfig()

	mux := http.NewServeMux()

	server := &http.Server{
		Addr:              ":8081",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Create connection
	conn, err := grpc.NewClient(
		cfg.URL.GrpcAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Create gRPC client
	client := urlv1.NewUrlServiceClient(conn)

	// Set request timeout
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	// Call gRPC method
	response, err := client.FindByCode(ctx, &urlv1.FindByCodeRequest{Code: "iwXj-Z-r"})
	if err != nil {
		log.Fatal(err)
	}

	log.Println(response)

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
