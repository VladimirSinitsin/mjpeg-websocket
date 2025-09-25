package data

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/jackc/pgx/v5/pgxpool"

	"stream-server/config"
)

type Clients struct {
	DBClientPool *pgxpool.Pool
}

func NewClients(ctx context.Context, cfg *conf.Database, l *log.Helper) (*Clients, func(), error) {
	c := context.WithoutCancel(ctx)

	// Make DSN address
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DbName,
		cfg.SSLMode,
	)

	// Make pool config
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		l.Errorf("failed to parse postgres config: %v", err)
		return nil, nil, err
	}

	// Update pool config
	poolConfig.MinConns = int32(cfg.MinOpenConns)
	if cfg.MaxOpenConns > 0 {
		poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	}
	if cfg.MaxConnLifetime > 0 {
		poolConfig.MaxConnLifetime = time.Duration(cfg.MaxConnLifetime) * time.Minute
	}
	if cfg.MaxConnIdleTime > 0 {
		poolConfig.MaxConnIdleTime = time.Duration(cfg.MaxConnIdleTime) * time.Minute
	}
	if cfg.HealthCheckPeriod > 0 {
		poolConfig.HealthCheckPeriod = time.Duration(cfg.HealthCheckPeriod) * time.Minute
	}
	if cfg.ConnTimeoutSec > 0 {
		poolConfig.ConnConfig.ConnectTimeout = time.Duration(cfg.ConnTimeoutSec) * time.Second
	}

	// Make pgx pool
	pool, err := pgxpool.NewWithConfig(c, poolConfig)
	if err != nil {
		l.Errorf("failed to create connection pool: %v", err)
		return nil, nil, err
	}

	// Ping connection by pool
	if err = pool.Ping(c); err != nil {
		l.Errorf("failed to ping database: %v", err)
		return nil, nil, err
	}

	d := &Clients{
		DBClientPool: pool,
	}

	return d, func() {
		l.Info("message", "closing the data resources")
		pool.Close()
	}, nil
}
