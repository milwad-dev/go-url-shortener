package server

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	urlv1 "github.com/milwad-dev/url-shortener/contracts/url/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"url-service/internal/domain"
)

type URLHandler struct {
	urlv1.UnimplementedUrlServiceServer
	db *pgxpool.Pool
}

func NewURLHandler(db *pgxpool.Pool) *URLHandler {
	return &URLHandler{db: db}
}

func (h *URLHandler) FindByCode(ctx context.Context, req *urlv1.FindByCodeRequest) (*urlv1.FindByCodeResponse, error) {
	if req.GetCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "missing code")
	}

	var url domain.URL
	err := h.db.QueryRow(ctx,
		"SELECT id::text, original_url, code, created_at FROM urls WHERE code = $1",
		req.GetCode(),
	).Scan(&url.ID, &url.OriginalURL, &url.Code, &url.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "url not found")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &urlv1.FindByCodeResponse{
		Id:          url.ID,
		OriginalUrl: url.OriginalURL,
		Code:        url.Code,
		CreatedAt:   timestamppb.New(url.CreatedAt),
	}, nil
}
