package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"

	"github.com/guarref/url-reducing-service/config"
	"github.com/guarref/url-reducing-service/internal/apperrors"
	"github.com/guarref/url-reducing-service/internal/db"
	"github.com/guarref/url-reducing-service/internal/service"
	"github.com/guarref/url-reducing-service/internal/storage"
	"github.com/guarref/url-reducing-service/internal/storage/inmemory"
	postgresStorage "github.com/guarref/url-reducing-service/internal/storage/postgres"
	"github.com/guarref/url-reducing-service/internal/web"

	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	cfg  *config.Config
	pool *pgxpool.Pool
	echo *echo.Echo
}

func New(ctx context.Context, cfg *config.Config) (*App, error) {

	var repo storage.Repository
	var pool *pgxpool.Pool

	switch strings.ToLower(strings.TrimSpace(cfg.Service.Storage)) {
	case "memory":
		repo = inmemory.NewRepository()
		log.Info().Msg("storage: in-memory")

	case "postgres":
		dsn, err := cfg.Postgres.ResolvedDSN()
		if err != nil {
			return nil, err
		}

		if cfg.Postgres.MigrateEnable {
			if err := db.RunMigrations(dsn, cfg.Postgres.MigrateFolder); err != nil {
				return nil, fmt.Errorf("app: migrations: %w", err)
			}
			log.Info().Str("dir", cfg.Postgres.MigrateFolder).Msg("db: migrations applied")
		}

		pool, err = db.NewPool(ctx, dsn)
		if err != nil {
			return nil, fmt.Errorf("app: postgres pool: %w", err)
		}
		repo = postgresStorage.NewRepository(pool)
		log.Info().Msg("storage: postgres")

	default:
		return nil, fmt.Errorf("app: unknown storage %q", cfg.Service.Storage)
	}

	svc := service.NewService(repo)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.HTTPErrorHandler = apperrors.HttpErrorHandler(cfg.Service.Environment)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	web.RegisterRoutes(e, svc, cfg.Service.BaseURL)

	return &App{cfg: cfg, pool: pool, echo: e}, nil
}

func (a *App) Run(ctx context.Context) error {

	addr := fmt.Sprintf(":%d", a.cfg.Service.Port)
	log.Info().Str("addr", addr).Msg("server starting")

	serverErr := make(chan error, 1)
	go func() {
		err := a.echo.Start(addr)
		if err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info().Msg("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := a.echo.Shutdown(shutdownCtx); err != nil {
			_ = a.echo.Close()
			a.closePool()
			return fmt.Errorf("shutdown echo: %w", err)
		}
		a.closePool()
		return nil

	case err := <-serverErr:
		a.closePool()
		return fmt.Errorf("server: %w", err)
	}
}

func (a *App) closePool() {
	if a.pool != nil {
		a.pool.Close()
	}
}
