package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

// uniqueViolation is the Postgres error code for a UNIQUE constraint violation.
const uniqueViolation = "23505"

func (h *URLHandler) CreateURL(ctx context.Context, req *urlv1.CreateURLRequest) (*urlv1.CreateURLResponse, error) {
	originalURL := strings.TrimSpace(req.GetUrl())
	if originalURL == "" {
		return nil, status.Error(codes.InvalidArgument, "missing url")
	}
	parsed, err := url.ParseRequestURI(originalURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, status.Error(codes.InvalidArgument, "url must be an absolute http or https URL")
	}

	for range 5 {
		code, err := generateCode()
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		var id string
		err = h.db.QueryRow(ctx,
			`INSERT INTO urls (original_url, code, created_at)
			 VALUES ($1, $2, $3)
			 RETURNING id::text`,
			originalURL, code, time.Now(),
		).Scan(&id)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			if pgErr.ConstraintName == "urls_code_key" {
				continue // random code collided, try another one
			}
			return nil, status.Error(codes.AlreadyExists, "url already exists")
		}
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		return &urlv1.CreateURLResponse{
			Id:          id,
			OriginalUrl: originalURL,
			Code:        code,
		}, nil
	}

	return nil, status.Error(codes.Internal, "could not generate a unique code")
}

// generateCode returns a random 8-character URL-safe code.
func generateCode() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}
