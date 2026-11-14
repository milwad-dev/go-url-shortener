package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	urlv1 "github.com/milwad-dev/url-shortener/contracts/url/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type Handler struct {
	GrpcConn *grpc.ClientConn
}

func NewHandler(grpcConn *grpc.ClientConn) *Handler {
	return &Handler{
		GrpcConn: grpcConn,
	}
}

// HealthZ handles liveness checks to verify that the application is running and able to respond to HTTP requests
func (h *Handler) HealthZ(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"status\": \"OK\"}"))
}

// ReadyZ handles readiness checks
func (h *Handler) ReadyZ(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	services := map[string]bool{
		"grpc": h.urlServiceReady(ctx),
	}

	ready := true
	for _, ok := range services {
		ready = ready && ok
	}

	statusCode, statusText := http.StatusOK, "OK"
	if !ready {
		statusCode, statusText = http.StatusServiceUnavailable, "NOT_READY"
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]any{
		"status":   statusText,
		"services": services,
	})
}

// urlServiceReady asks url-service for its health over the standard gRPC health protocol
func (h *Handler) urlServiceReady(ctx context.Context) bool {
	resp, err := healthpb.NewHealthClient(h.GrpcConn).Check(ctx, &healthpb.HealthCheckRequest{})
	if err != nil {
		log.Printf("readyz: url-service health check failed: %v", err)
		return false
	}

	return resp.GetStatus() == healthpb.HealthCheckResponse_SERVING
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
		writeGrpcError(w, err)
		return
	}

	http.Redirect(w, r, clientResponse.OriginalUrl, http.StatusFound)
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
		writeGrpcError(w, err)
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

// writeGrpcError translates a url-service gRPC error into an HTTP response
func writeGrpcError(w http.ResponseWriter, err error) {
	st := status.Convert(err)

	switch st.Code() {
	case codes.InvalidArgument:
		http.Error(w, st.Message(), http.StatusBadRequest)
	case codes.NotFound:
		http.Error(w, st.Message(), http.StatusNotFound)
	case codes.AlreadyExists:
		http.Error(w, st.Message(), http.StatusConflict)
	case codes.DeadlineExceeded:
		log.Printf("url-service timeout: %v", err)
		http.Error(w, "upstream timeout", http.StatusGatewayTimeout)
	case codes.Unavailable:
		log.Printf("url-service unavailable: %v", err)
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
	default:
		log.Printf("url-service error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
