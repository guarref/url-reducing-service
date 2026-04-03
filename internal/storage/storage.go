package storage

import (
	"context"

	"github.com/guarref/url-reducing-service/internal/model"
)

type Repository interface {
	GetByShortCode(ctx context.Context, shortCode string) (*model.Link, error)
	GetByOriginalURL(ctx context.Context, originalURL string) (*model.Link, error)
	Create(ctx context.Context, link *model.Link) error
}
