package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
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

func (s *Server) Start() error {
	go func() {
		logrus.Infof("server address: %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("Server failed to start")
		}
	}()
	return nil
}

func (s *Server) WaitAndShutdown(ctx context.Context) {
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.WithError(err).Fatal("Server forced to shutdown")
	}
}
