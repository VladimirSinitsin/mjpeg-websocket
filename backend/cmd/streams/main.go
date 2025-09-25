package main

import (
	"context"
	"os"
	"stream-server/logger"

	"stream-server/config"

	kzerolog "github.com/go-kratos/kratos/contrib/log/zerolog/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name = "stream.service"
	// Version is the version of the compiled software.
	Version string

	id, _ = os.Hostname()
)

func main() {
	cfg, err := conf.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	l := logger.New(cfg.Level)
	klog := log.With(kzerolog.NewLogger(l),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
	logHelper := log.NewHelper(klog)

	app, cancel, err := initApp(context.Background(), cfg, logHelper)
	if err != nil {
		logHelper.Fatalf("init app error: %s", err)
		panic(err)
	}
	defer cancel()

	// start and wait for stop signal
	if err = app.Run(); err != nil {
		logHelper.Fatalf("app run error: %s", err)
		panic(err)
	}
}
