package config

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Service  ServiceSettings
	Postgres PostgresSettings
}

type ServiceSettings struct {
	Port        int    `env:"SERVICE_PORT" envDefault:"8080"`
	Environment string `env:"SERVICE_ENV" envDefault:"development"`
	BaseURL     string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	Storage     string `env:"STORAGE" envDefault:"postgres"`
}

type PostgresSettings struct {
	DSN           string `env:"DB_DSN"`
	Host          string `env:"DB_HOST" envDefault:"localhost"`
	Port          string `env:"DB_PORT" envDefault:"5432"`
	User          string `env:"DB_USER" envDefault:"postgres"`
	Password      string `env:"DB_PASSWORD" envDefault:"postgres"`
	Name          string `env:"DB_NAME" envDefault:"urls"`
	SSLMode       string `env:"DB_SSLMODE" envDefault:"disable"`
	MigrateEnable bool   `env:"MIGRATE_ENABLE" envDefault:"true"`
	MigrateFolder string `env:"MIGRATE_FOLDER" envDefault:"migrations"`
}

func (p PostgresSettings) ResolvedDSN() (string, error) {

	if strings.TrimSpace(p.DSN) != "" {
		return p.DSN, nil
	}

	if strings.TrimSpace(p.Host) == "" || strings.TrimSpace(p.User) == "" || strings.TrimSpace(p.Name) == "" {
		return "", fmt.Errorf("postgres: lost DB_DSN or DB_HOST or DB_USER or DB_NAME")
	}

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.Name, p.SSLMode,
	), nil
}

func Load() (*Config, error) {

	var svc ServiceSettings

	if err := env.Parse(&svc); err != nil {
		return nil, fmt.Errorf("config: service env: %w", err)
	}
	var pg PostgresSettings
	if err := env.Parse(&pg); err != nil {
		return nil, fmt.Errorf("config: postgres env: %w", err)
	}
	cfg := &Config{Service: svc, Postgres: pg}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Validate() error {

	switch strings.ToLower(strings.TrimSpace(c.Service.Storage)) {
	case "memory", "postgres":
	default:
		return fmt.Errorf("config: storage must be memory or postgres, got %q", c.Service.Storage)
	}
	if strings.ToLower(strings.TrimSpace(c.Service.Storage)) == "postgres" {
		if _, err := c.Postgres.ResolvedDSN(); err != nil {
			return err
		}
	}
	return nil
}
