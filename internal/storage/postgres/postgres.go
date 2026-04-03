package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/guarref/url-reducing-service/internal/apperrors"
	"github.com/guarref/url-reducing-service/internal/model"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetByShortCode(ctx context.Context, shortCode string) (*model.Link, error) {

	query := `SELECT id, original_url, short_code, created_at
				FROM links
				WHERE short_code = @short_code`

	rows, err := r.pool.Query(ctx, query, pgx.NamedArgs{"short_code": shortCode})
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to query link by short code")
	}

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.Link])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.Wrap(err, "failed to scan link")
	}

	return &result, nil
}

func (r *Repository) GetByOriginalURL(ctx context.Context, originalURL string) (*model.Link, error) {

	query := `SELECT id, original_url, short_code, created_at
				FROM links
				WHERE original_url = @original_url`

	rows, err := r.pool.Query(ctx, query, pgx.NamedArgs{"original_url": originalURL})
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to query link by original url")
	}

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.Link])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.Wrap(err, "failed to scan link")
	}

	return &result, nil
}

func (r *Repository) Create(ctx context.Context, link *model.Link) error {
	if link == nil {
		return apperrors.ErrNilLink
	}

	query := `INSERT INTO links (short_code, original_url)
		VALUES (@short_code, @original_url)
		ON CONFLICT DO NOTHING
		RETURNING id, short_code, original_url, created_at`

	args := pgx.NamedArgs{
		"short_code":   link.ShortCode,
		"original_url": link.OriginalURL,
	}

	rows, err := r.pool.Query(ctx, query, args)
	if err != nil {
		return apperrors.Wrap(err, "failed to create link")
	}

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.Link])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_, getErr := r.GetByOriginalURL(ctx, link.OriginalURL)
			switch {
			case getErr == nil:
				return apperrors.ErrOriginalURLConflict
			case errors.Is(getErr, apperrors.ErrNotFound):
				return apperrors.ErrShortCodeConflict
			default:
				return apperrors.Wrap(getErr, "failed to detect conflict reason")
			}
		}
		return apperrors.Wrap(err, "failed to scan created link")
	}

	link.ID = result.ID
	link.OriginalURL = result.OriginalURL
	link.ShortCode = result.ShortCode
	link.CreatedAt = result.CreatedAt

	return nil
}
