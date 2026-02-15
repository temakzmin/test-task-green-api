package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"green-api/internal/config"
	"green-api/internal/greenapi"
	"green-api/internal/http/router"
	"green-api/internal/logging"
	"green-api/internal/service"
)

type Server struct {
	cfg    config.Config
	logger *zap.Logger
	http   *http.Server
}

func New(configPath string) (*Server, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}

	logger, err := logging.New(cfg.Logging)
	if err != nil {
		return nil, err
	}

	client := greenapi.NewClient(cfg.GreenAPI, logger)
	svc := service.New(client)
	engine := router.New(cfg, logger, svc)

	httpServer := &http.Server{
		Addr:         cfg.Server.Address(),
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout(),
		WriteTimeout: cfg.Server.WriteTimeout(),
	}

	return &Server{cfg: cfg, logger: logger, http: httpServer}, nil
}

func (s *Server) Run() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("server_started", zap.String("address", s.http.Addr))
		err := s.http.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("listen and serve: %w", err)
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case sig := <-stop:
		s.logger.Info("shutdown_signal_received", zap.String("signal", sig.String()))
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.Server.ShutdownTimeout())
	defer cancel()

	if err := s.http.Shutdown(ctx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	s.logger.Info("server_stopped")
	return nil
}
