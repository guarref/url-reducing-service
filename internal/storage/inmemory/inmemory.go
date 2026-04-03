package inmemory

import (
	"context"
	"sync"
	"time"

	"github.com/guarref/url-reducing-service/internal/apperrors"
	"github.com/guarref/url-reducing-service/internal/model"
)

type Repository struct {
	mu         sync.RWMutex
	nextID     int64
	codeToLink map[string]*model.Link
	urlToCode  map[string]string
}

func NewRepository() *Repository {
	return &Repository{
		codeToLink: make(map[string]*model.Link),
		urlToCode:  make(map[string]string),
	}
}

func (r *Repository) GetByShortCode(ctx context.Context, shortCode string) (*model.Link, error) {

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	link, exists := r.codeToLink[shortCode]
	if !exists {
		return nil, apperrors.ErrNotFound
	}

	copied := *link
	return &copied, nil
}

func (r *Repository) GetByOriginalURL(ctx context.Context, originalURL string) (*model.Link, error) {

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	code, exists := r.urlToCode[originalURL]
	if !exists {
		return nil, apperrors.ErrNotFound
	}

	link, exists := r.codeToLink[code]
	if !exists {
		return nil, apperrors.ErrNotFound
	}

	copied := *link
	return &copied, nil
}

func (r *Repository) Create(ctx context.Context, link *model.Link) error {

	if err := ctx.Err(); err != nil {
		return err
	}

	if link == nil {
		return apperrors.ErrNilLink
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}

	if _, exists := r.urlToCode[link.OriginalURL]; exists {
		return apperrors.ErrOriginalURLConflict
	}

	if _, exists := r.codeToLink[link.ShortCode]; exists {
		return apperrors.ErrShortCodeConflict
	}

	r.nextID++

	createdAt := link.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	stored := &model.Link{
		ID:          r.nextID,
		OriginalURL: link.OriginalURL,
		ShortCode:   link.ShortCode,
		CreatedAt:   createdAt,
	}

	r.codeToLink[stored.ShortCode] = stored
	r.urlToCode[stored.OriginalURL] = stored.ShortCode

	link.ID = stored.ID
	link.CreatedAt = stored.CreatedAt

	return nil
}
