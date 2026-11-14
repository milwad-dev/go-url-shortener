package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

// ConnectDB connect to database
func ConnectDB(dsn string) (*pgx.Conn, error) {
	// Connect to database
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	// Ping the database
	if err = conn.Ping(context.Background()); err != nil {
		return nil, err
	}

	// Create database tables
	createTables(conn)

	fmt.Println("Connected to database")

	return conn, nil
}

// createTables create the database tables
func createTables(conn *pgx.Conn) {
	// URLs table
	conn.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS urls (
    	id serial PRIMARY KEY,
    	original_url text UNIQUE NOT NULL,
    	code text UNIQUE NOT NULL,
    	created_at TIMESTAMP WITH TIME ZONE NOT NULL
	)`)

	fmt.Println("The tables has been created")
}
