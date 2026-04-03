package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {

	p, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("db: create pool: %w", err)
	}

	if err := p.Ping(ctx); err != nil {
		p.Close()
		return nil, fmt.Errorf("db: ping: %w", err)
	}
	return p, nil
}

func RunMigrations(dsn, migrationsDir string) error {

	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("db: open for migrations: %w", err)
	}
	defer sqlDB.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("db: goose dialect: %w", err)
	}
	if err := goose.Up(sqlDB, migrationsDir); err != nil {
		return fmt.Errorf("db: goose up: %w", err)
	}
	return nil
}
