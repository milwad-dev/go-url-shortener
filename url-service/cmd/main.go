package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"
	"url-service/internal/config"
	"url-service/internal/database"
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

	grpcServer := server2.NewServer("50051", db)
	grpcServer.RegisterServices()

	errCh := make(chan error, 1)

	go func() {
		if err := grpcServer.Start(); err != nil {
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

	grpcServer.Stop(shutdownCtx)

	// Close the DB pool last, after no handler can use it
	db.Close()
	log.Println("url-service stopped")

	if exitCode != 0 {
		log.Fatal("exiting with error")
	}
}
