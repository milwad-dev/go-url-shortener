package server

import (
	"context"
	"fmt"
	"net"

	"github.com/jackc/pgx/v5/pgxpool"
	urlv1 "github.com/milwad-dev/url-shortener/contracts/url/v1"
	"google.golang.org/grpc"
)

type Server struct {
	server *grpc.Server
	port   string
	db     *pgxpool.Pool
}

func NewServer(port string, db *pgxpool.Pool) *Server {
	return &Server{
		server: grpc.NewServer(),
		port:   port,
		db:     db,
	}
}

func (s *Server) RegisterServices() {
	urlv1.RegisterUrlServiceServer(
		s.server,
		NewURLHandler(s.db),
	)
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	fmt.Printf("gRPC server listening on :%s\n", s.port)

	return s.server.Serve(listener)
}

// Stop waits for in-flight RPCs to finish.
func (s *Server) Stop(ctx context.Context) {
	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		s.server.Stop()
		<-done
	}
}
