package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v9"
)

type (
	Config struct {
		AdminUserIDs []int64 `env:"ADMIN_USER_ID,notEmpty" envSeparator:"," envDefault:"-1"`

		Level        string         `env:"LEVEL" envDefault:"info"`
		Env          string         `env:"ENV" envDefault:"dev"`
		Token        string         `env:"TOKEN,notEmpty"`
		DB           DatabaseConfig `env:"-"`
		PollInterval time.Duration  `env:"POLL_DURATION" envDefault:"1m"`

		SentryDSN     string        `env:"SENTRY_DSN"`
		SentryTimeout time.Duration `env:"SENTRY_TIMEOUT" envDefault:"5s"`

		ReleaseVersion string `env:"RELEASE_VERSION,notEmpty"`

		MetricsPort int `env:"METRICS_PORT" envDefault:"7777"`
		APIPort     int `env:"API_PORT" envDefault:"8080"`

		AllowedOrigins string `env:"ALLOWED_ORIGINS" envDefault:""`
	}

	DatabaseConfig struct {
		Host         string `env:"DB_HOST,notEmpty"`
		Port         string `env:"DB_PORT" envDefault:"5432"`
		Name         string `env:"DB_NAME,notEmpty"`
		User         string `env:"DB_USER,notEmpty"`
		Pwd          string `env:"DB_PWD,notEmpty"`
		SslMode      string `env:"DB_SSL_MODE" envDefault:"disable"`
		MigrationsUp bool   `env:"DB_MIGRATIONS_UP" envDefault:"true"`
	}
)

// New lookup for envs and fill default values if not found
func New() (Config, error) {
	cnf := Config{}
	if err := env.Parse(&cnf); err != nil {
		return Config{}, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := env.Parse(&cnf.DB); err != nil {
		return Config{}, fmt.Errorf("failed to parse db config: %w", err)
	}

	return cnf, nil
}

func (c *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Pwd, c.Host, c.Port, c.Name, c.SslMode,
	)
}
