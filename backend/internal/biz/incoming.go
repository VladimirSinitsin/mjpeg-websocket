package biz

import (
	"stream-server/config"
	"stream-server/internal/biz/session/store_pool"
	"stream-server/internal/interfaces"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StreamUsecase struct {
	repo interfaces.IRepo
	log  *log.Helper
	cfg  *conf.Config
}

func NewStreamUsecase(repo interfaces.IRepo, l *log.Helper, cfg *conf.Config) *StreamUsecase {
	return &StreamUsecase{
		repo: repo,
		log:  l,
		cfg:  cfg,
	}
}

func NewStreamPoolStore(cfg *conf.Config, db *pgxpool.Pool) *store_pool.ChunkStore {
	return store_pool.NewChunkStore(db, store_pool.Sizes, cfg.CacheCapBytes, cfg.ChunkFrames)
}
