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
