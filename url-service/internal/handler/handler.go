package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"
	"url-service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	DB *pgxpool.Pool
}

func CreateHandler(db *pgxpool.Pool) *Handler {
	return &Handler{DB: db}
}

func (handler *Handler) CreateURL(w http.ResponseWriter, r *http.Request) {
	var body struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the original URL already exists
	var count int
	err := handler.DB.
		QueryRow(context.Background(), "SELECT COUNT(*) FROM urls WHERE original_url = $1", body.URL).
		Scan(&count)
	if err != nil {
		http.Error(w, "Error on finding url.", http.StatusBadRequest)
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
	_, err = handler.DB.Exec(
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

func (handler *Handler) FindURLByCode(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	var url domain.URL
	err := handler.DB.
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
