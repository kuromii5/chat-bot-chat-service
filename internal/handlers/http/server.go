package http

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
	router     http.Handler
}

func NewServer(host, port string, router http.Handler) *Server {
	addr := fmt.Sprintf("%s:%s", host, port)

	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		router: router,
	}
}

func (s *Server) Addr() string {
	return s.httpServer.Addr
}

func (s *Server) Start() error {
	if err := s.httpServer.ListenAndServe(); err != nil {
		return fmt.Errorf("listen and serve: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}
