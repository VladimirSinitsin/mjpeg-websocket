package repo

import (
	"stream-server/config"
	"stream-server/internal/data"
	"stream-server/internal/data/repo"

	"github.com/go-kratos/kratos/v2/log"
)

type StreamRepo struct {
	queries *repo.Queries
	log     *log.Helper
	cfg     *conf.Config
	data    *data.Clients
}

func NewStreamRepo(queries *repo.Queries, logger *log.Helper, cfg *conf.Config, data *data.Clients) *StreamRepo {
	return &StreamRepo{
		queries: queries,
		log:     logger,
		cfg:     cfg,
		data:    data,
	}
}
