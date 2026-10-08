package serve

import (
	"context"
	"net/http"

	"github.com/ksendel-coder/onboarding_croquant/platform/config"
	"github.com/rs/zerolog"
)

type (
	// ServerInterface defines interface of server of this project
	ServerInterface interface {
		Start() error
		Shutdown(ctx context.Context) error
	}

	// Server responsible for up HTTP server
	Server struct {
		logger     *zerolog.Logger
		httpServer *http.Server
	}
)

// New wraps an already-routed handler in the listen/shutdown machinery. Both
// services build their own router and share nothing but this: what they serve
// differs, how they are started and stopped does not.
func New(cfg *config.ServerConfig, logger *zerolog.Logger, handler http.Handler) *Server {
	return &Server{
		logger: logger,
		httpServer: &http.Server{
			Addr:    cfg.Address,
			Handler: handler,
		},
	}
}

// Handler exposes the routed handler so a test can drive the fully assembled
// server through httptest without opening a socket.
func (s *Server) Handler() http.Handler {
	return s.httpServer.Handler
}

func (s *Server) Start() error {
	s.logger.Info().Str("address", s.httpServer.Addr).Msg("Starting server...")
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Error().Err(err).Msg("Server failed to start")
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info().Msg("Shutting down server...")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error().Err(err).Msg("Server failed to shutdown")
		return err
	}
	return nil
}
