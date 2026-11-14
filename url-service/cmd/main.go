package main

import (
	"context"
	"log"
	"net/http"
	"url-service/internal/config"
	"url-service/internal/database"
	"url-service/internal/handler"
	server2 "url-service/internal/server"

	"github.com/jackc/pgx/v5"
)

func main() {
	// Load config
	cfg := config.LoadConfig()

	// Connect to DB
	dsn := cfg.DSN()
	db, err := database.ConnectDB(dsn)
	defer db.Close(context.Background())

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

func startGrpcServer(db *pgx.Conn) *server2.Server {
	server := server2.NewServer("50051", db)
	server.RegisterServices()

	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	return server
}
