package main

import (
	"context"
	"log"
	"net/http"
	"url-service/internal/config"
	"url-service/internal/database"
	"url-service/internal/handler"
	server2 "url-service/internal/server"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Load config
	cfg := config.LoadConfig()

	// Connect to DB
	db, err := database.ConnectDB(context.Background(), cfg.DSN())
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	// Start gRPC server
	grpcServer := startGrpcServer(db)
	defer grpcServer.Stop()

	// SetUp handlers
	handlers := handler.CreateHandler(db)

	http.HandleFunc("POST /url/create", handlers.CreateURL)

	http.HandleFunc("/{code}", handlers.FindURLByCode)

	if err = http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func startGrpcServer(db *pgxpool.Pool) *server2.Server {
	server := server2.NewServer("50051", db)
	server.RegisterServices()

	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	return server
}
