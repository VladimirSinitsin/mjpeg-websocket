package main

import (
	"context"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"

	_ "go.uber.org/automaxprocs"

	"stream-server/config"
	"stream-server/internal/biz"
	idata "stream-server/internal/data"
	queries "stream-server/internal/data/repo"
	"stream-server/internal/dep"
	"stream-server/internal/repo"
	"stream-server/internal/server"
	"stream-server/internal/service"
	"stream-server/internal/wrapper"
)

func initApp(ctx context.Context, conf *conf.Config, logger *log.Helper) (*kratos.App, func(), error) {
	meterProvider, err := dep.NewMeterProvider(conf)
	if err != nil {
		return nil, nil, err
	}
	meter, err := dep.NewMeter(&conf.Metadata, meterProvider)
	if err != nil {
		return nil, nil, err
	}

	// Repo
	dataClients, cleanup, err := idata.NewClients(ctx, &conf.Database, logger)
	if err != nil {
		return nil, nil, err
	}
	streamQueries := queries.New(dataClients.DBClientPool)
	streamRepo := repo.NewStreamRepo(streamQueries, logger, conf, dataClients)
	streamRepoWrapper := wrapper.NewStreamRepoWrapper(streamRepo)

	// Usecase
	streamUsecase := biz.NewStreamUsecase(streamRepoWrapper, logger, conf)
	streamUsecaseWrapper := wrapper.NewStreamUsecaseWrapper(streamUsecase)
	streamPoolStore := biz.NewStreamPoolStore(conf, dataClients.DBClientPool)

	// Services
	streamService := service.NewStreamService(streamUsecaseWrapper, logger, streamPoolStore)
	streamServiceWrapper := wrapper.NewStreamServiceWrapper(streamService)
	healthService := service.NewHealthService(dataClients.DBClientPool)

	streamServer := server.NewHTTPStreamServer(conf, streamServiceWrapper, healthService, meter, logger)
	metricsServer := server.NewMetricsServer(conf, logger)
	app := newApp(ctx, logger.Logger(), streamServer, metricsServer)

	return app, func() {
		cleanup()
	}, nil
}

func newApp(ctx context.Context, logger log.Logger, srv ...transport.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Context(ctx),
		kratos.Server(srv...),
	)
}
