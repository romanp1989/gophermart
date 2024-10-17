package server

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/romanp1989/gophermart/internal/config"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type Server struct {
	server *http.Server
	log    *zap.Logger
}

func NewServer(router *chi.Mux, log *zap.Logger) *Server {
	log.Info("Running server on ", zap.String("port", config.Options.FlagServerAddress))

	return &Server{
		server: &http.Server{
			Addr:    config.Options.FlagServerAddress,
			Handler: router,
		},
		log: log,
	}
}

func (s *Server) RunServer() error {
	errChannel := make(chan error, 1)

	go func() {
		err := s.server.ListenAndServe()
		if err != nil {
			errChannel <- err
			return
		}

		close(errChannel)
	}()

	return <-errChannel
}

func (s *Server) Stop() {
	ctx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdown()

	_ = s.server.Shutdown(ctx)
}
