package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
	"url-service/internal/domain"

	"github.com/jackc/pgx/v5"
)

var DB *pgx.Conn

func main() {
	// Connect to DB
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		os.Getenv("DATABASE_USER"),
		os.Getenv("DATABASE_PASSWORD"),
		os.Getenv("DATABASE_HOST"),
		os.Getenv("DATABASE_PORT"),
		os.Getenv("DATABASE_NAME"),
	)
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	if err = conn.Ping(context.Background()); err != nil {
		panic(err)
	}

	// Create DB tables
	createTables(conn)

	DB = conn

	fmt.Println("Connected to database")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("milwad"))
	})

	http.HandleFunc("POST /url/create", createURLHandler)

	http.HandleFunc("/{code}", findURLByCode)

	if err = http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func createURLHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the original URL already exists
	var count int
	err := DB.QueryRow(context.Background(), "SELECT COUNT(*) FROM urls WHERE original_url = $1", body.URL).Scan(&count)
	if err != nil {
		http.Error(w, "The URL already exists.", http.StatusBadRequest)
		return
	}
	if count > 0 {
		http.Error(w, "The URL already exists.", http.StatusBadRequest)
		return
	}

	// Generate random code
	code, err := generateCode()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert new short URL in DB
	_, err = DB.Exec(
		context.Background(),
		`INSERT INTO urls (original_url, code, created_at) 
			 VALUES ($1, $2, $3)`,
		body.URL,
		code,
		time.Now(),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(code)
}

func findURLByCode(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	var url domain.URL
	err := DB.
		QueryRow(context.Background(), "SELECT id, original_url, code, created_at FROM urls WHERE code = $1", code).
		Scan(&url.ID, &url.OriginalURL, &url.Code, &url.CreatedAt)
	if err != nil {
		http.Error(w, "Can't find the original url", http.StatusBadRequest)
		return
	}

	response := map[string]string{
		"id":           url.ID,
		"original_url": url.OriginalURL,
		"code":         url.Code,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func generateCode() (string, error) {
	b := make([]byte, 6)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

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
