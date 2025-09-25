package server

import (
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/http"
	otel "go.opentelemetry.io/otel/metric"

	v1 "stream-server/api/v1"
	"stream-server/config"
	"stream-server/internal/interfaces"
	utils "stream-server/internal/server/server_utils"
)

func NewHTTPStreamServer(cfg *conf.Config, service interfaces.IStreamService, healthService interfaces.IHealthService, meter otel.Meter, logger *log.Helper) *http.Server {
	srv := newHTTPServer(cfg, meter, logger)
	v1.RegisterStreamServiceHTTPServer(srv, service)

	// health
	v1.RegisterHealthServiceHTTPServer(srv, healthService)

	// Websocket
	srv.Handle("/v1/streams/{id}/ws", service.StreamWSHandler())

	return srv
}

func newHTTPServer(cfg *conf.Config, meter otel.Meter, logger *log.Helper) *http.Server {
	counter, err := metrics.DefaultRequestsCounter(meter, metrics.DefaultServerRequestsCounterName)
	if err != nil {
		return nil
	}

	seconds, err := metrics.DefaultSecondsHistogram(meter, metrics.DefaultServerSecondsHistogramName)
	if err != nil {
		return nil
	}

	var opts = []http.ServerOption{
		http.Filter(utils.CORS()),
		http.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			metrics.Server(metrics.WithRequests(counter), metrics.WithSeconds(seconds)),
			logging.Server(logger.Logger()),
		),
	}
	if cfg.Http.Network != "" {
		opts = append(opts, http.Network(cfg.Http.Network))
	}
	if cfg.Http.Addr != "" {
		opts = append(opts, http.Address(cfg.Http.Addr))
	}
	if cfg.Http.Timeout != 0 {
		opts = append(opts, http.Timeout(time.Duration(cfg.Http.Timeout)*time.Second))
	}
	srv := http.NewServer(opts...)

	return srv
}
