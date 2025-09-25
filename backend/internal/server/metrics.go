package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"stream-server/config"
)

type MetricsServer struct {
	srv    *http.Server
	logger *log.Helper
}

func NewMetricsServer(cfg *conf.Config, logger *log.Helper) transport.Server {
	if !cfg.Metrics.Enabled {
		logger.Infof("[metrics] disabled")
		return nil
	}

	path := cfg.Metrics.MetricsPath
	if path == "" {
		path = "/metrics"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	mux := http.NewServeMux()
	mux.Handle(path, promhttp.Handler())

	addr := fmt.Sprintf(":%d", cfg.Metrics.ServerPort)
	return &MetricsServer{
		srv: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
		logger: log.NewHelper(log.With(logger.Logger(), "module", "metrics")),
	}
}

func (s *MetricsServer) Start(ctx context.Context) error {
	if s == nil || s.srv == nil {
		return nil
	}

	s.logger.Infof("starting metrics server on %s%s", s.srv.Addr, "/metrics")
	go func() {
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Errorf("metrics server error: %v", err)
		}
	}()
	return nil
}

func (s *MetricsServer) Stop(ctx context.Context) error {
	if s == nil || s.srv == nil {
		return nil
	}
	s.logger.Info("stopping metrics server")
	return s.srv.Shutdown(ctx)
}
