package conf

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type (
	Config struct {
		Metadata
		Http
		Log
		Database
		Metrics
		SocketPool
	}

	Metadata struct {
		Name      string `env:"APP_NAME,required"`
		Namespace string `env:"APP_NAMESPACE,required"`
		Env       string `env:"APP_ENV,required"`
	}

	Http struct {
		Network string `env:"HTTP_NETWORK"`
		Addr    string `env:"HTTP_ADDRESS,required"`
		Timeout int    `env:"HTTP_TIMEOUT"`
	}

	Log struct {
		Level string `env:"LOG_LEVEL,required"`
	}

	Database struct {
		Host              string `env:"DB_HOST,required"`
		Port              string `env:"DB_PORT,required"`
		User              string `env:"DB_USER,required"`
		Password          string `env:"DB_PASSWORD,required"`
		DbName            string `env:"DB_NAME,required"`
		SSLMode           string `env:"DB_SSL_MODE,required"`
		MaxOpenConns      int    `env:"DB_MAX_OPEN_CONNS" envDefault:"10"`
		MinOpenConns      int    `env:"DB_MIN_OPEN_CONNS" envDefault:"0"`
		MaxConnIdleTime   int    `env:"DB_MAX_IDLE_CONNS" envDefault:"30"`
		MaxConnLifetime   int    `env:"DB_CONN_MAX_LIFE_TIME" envDefault:"60"`
		HealthCheckPeriod int    `env:"DB_HEALTH_CHECK_PERIOD" envDefault:"1"`
		ConnTimeoutSec    int    `env:"DB_CONN_TIMEOUT_SEC" envDefault:"120"`
	}

	Metrics struct {
		Enabled     bool   `env:"PROMETHEUS_ENABLED,required"`
		Namespace   string `env:"PROMETHEUS_SPACE,required"`
		Name        string `env:"PROMETHEUS_NAME,required"`
		MetricsPath string `env:"PROMETHEUS_METRICS_PATH,required"`
		ServerPort  int    `env:"PROMETHEUS_SERVER_PORT,required"`
	}

	SocketPool struct {
		ChunkFrames   int64 `env:"CHUNK_FRAMES" envDefault:"256"`
		CacheCapBytes int64 `env:"CACHE_CAP_BYTES" envDefault:"536870912"` // 512 MB (512<<20)
	}
)

func NewConfig() (*Config, error) {
	cfg := Config{}
	opts := env.Options{Prefix: "STREAM_"}

	if err := env.ParseWithOptions(&cfg, opts); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	return &cfg, nil
}
