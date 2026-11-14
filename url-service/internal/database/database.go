package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ConnectDB opens a connection pool to the database, verifies it and creates the tables.
func ConnectDB(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to create database pool: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to reach database: %w", err)
	}

	if err = createTables(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}

	fmt.Println("Connected to database")

	return pool, nil
}

// createTables creates the database tables
func createTables(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS urls (
    	id serial PRIMARY KEY,
    	original_url text UNIQUE NOT NULL,
    	code text UNIQUE NOT NULL,
    	created_at TIMESTAMP WITH TIME ZONE NOT NULL
	)`)
	if err != nil {
		return fmt.Errorf("unable to create tables: %w", err)
	}

	fmt.Println("The tables has been created")

	return nil
}
