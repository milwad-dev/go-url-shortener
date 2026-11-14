package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	urlv1 "github.com/milwad-dev/url-shortener/contracts/url/v1"
	"google.golang.org/grpc"
)

type Handler struct {
	GrpcConn *grpc.ClientConn
}

func NewHandler(grpcConn *grpc.ClientConn) *Handler {
	return &Handler{
		GrpcConn: grpcConn,
	}
}

// FindURLByCode find url via code from url-service
func (h *Handler) FindURLByCode(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	// Create gRPC client
	client := urlv1.NewUrlServiceClient(h.GrpcConn)

	// Set request timeout
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	// Call gRPC method
	clientResponse, err := client.FindByCode(ctx, &urlv1.FindByCodeRequest{Code: code})
	if err != nil {
		log.Fatal(err)
	}

	response := map[string]string{
		"id":           clientResponse.Id,
		"original_url": clientResponse.OriginalUrl,
		"code":         clientResponse.Code,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CreateURL create new url if not exists before
func (h *Handler) CreateURL(w http.ResponseWriter, r *http.Request) {
	var body struct {
		URL string `json:"url"`
	}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "URL is required.", http.StatusBadRequest)
		return
	}

	// Create gRPC client
	client := urlv1.NewUrlServiceClient(h.GrpcConn)

	// Set request timeout
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	clientResponse, err := client.CreateURL(ctx, &urlv1.CreateURLRequest{Url: body.URL})
	if err != nil {
		if err.Error() == "rpc error: code = AlreadyExists desc = url already exists" {
			http.Error(w, "URL already exists", http.StatusConflict)
			return
		}

		http.Error(w, "Something wrong on creating URL.", http.StatusBadRequest)
		return
	}

	response := map[string]string{
		"id":           clientResponse.Id,
		"original_url": clientResponse.OriginalUrl,
		"code":         clientResponse.Code,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
