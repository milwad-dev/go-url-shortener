package server

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

// dbPingTimeout bounds how long a health check waits on PostgreSQL.
const dbPingTimeout = 2 * time.Second

type HealthHandler struct {
	healthpb.UnimplementedHealthServer
	db *pgxpool.Pool
}

func NewHealthHandler(db *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{db: db}
}

// Check reports SERVING for the whole server ("") and for "url.v1.UrlService".
func (h *HealthHandler) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	switch req.GetService() {
	case "", "url.v1.UrlService":
	default:
		return nil, status.Errorf(codes.NotFound, "unknown service %q", req.GetService())
	}

	ctx, cancel := context.WithTimeout(ctx, dbPingTimeout)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_NOT_SERVING}, nil
	}

	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}
