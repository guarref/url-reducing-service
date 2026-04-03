package inmemory_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/guarref/url-reducing-service/internal/apperrors"
	"github.com/guarref/url-reducing-service/internal/model"
	"github.com/guarref/url-reducing-service/internal/storage/inmemory"
)

func mustCreate(t *testing.T, r *inmemory.Repository, link *model.Link) {
	t.Helper()
	if err := r.Create(context.Background(), link); err != nil {
		t.Fatalf("setup create failed: %v", err)
	}
}

func TestCreate_Success(t *testing.T) {

	r := inmemory.NewRepository()
	err := r.Create(context.Background(), &model.Link{ShortCode: "qazwsx123_", OriginalURL: "https://ozon.ru/test_link"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestCreate_ShortConflict(t *testing.T) {
	r := inmemory.NewRepository()
	ctx := context.Background()

	mustCreate(t, r, &model.Link{ShortCode: "qazwsx123_", OriginalURL: "https://ozon.ru/test_link"})

	err := r.Create(ctx, &model.Link{ShortCode: "qazwsx123_", OriginalURL: "https://google.com/test_link"})
	if !errors.Is(err, apperrors.ErrShortCodeConflict) {
		t.Errorf("expected ErrShortCodeConflict, got %v", err)
	}
}

func TestCreate_OriginalExists(t *testing.T) {

	r := inmemory.NewRepository()
	ctx := context.Background()

	mustCreate(t, r, &model.Link{ShortCode: "qazwsx123_", OriginalURL: "https://ozon.ru/test_link"})

	err := r.Create(ctx, &model.Link{ShortCode: "zaqxsw3210", OriginalURL: "https://ozon.ru/test_link"})
	if !errors.Is(err, apperrors.ErrOriginalURLConflict) {
		t.Errorf("expected ErrOriginalURLConflict, got %v", err)
	}
}

func TestGetByShortCode_Success(t *testing.T) {

	r := inmemory.NewRepository()
	ctx := context.Background()

	mustCreate(t, r, &model.Link{ShortCode: "qazwsx123_", OriginalURL: "https://ozon.ru/test_link"})

	link, err := r.GetByShortCode(ctx, "qazwsx123_")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link.OriginalURL != "https://ozon.ru/test_link" {
		t.Errorf("expected %q, got %q", "https://ozon.ru/test_link", link.OriginalURL)
	}
}

func TestGetByShortCode_NotFound(t *testing.T) {

	r := inmemory.NewRepository()

	_, err := r.GetByShortCode(context.Background(), "qwerty1234")
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetByOriginalURL_Success(t *testing.T) {

	r := inmemory.NewRepository()
	ctx := context.Background()

	mustCreate(t, r, &model.Link{ShortCode: "qazwsx123_", OriginalURL: "https://ozon.ru/test_link"})

	link, err := r.GetByOriginalURL(ctx, "https://ozon.ru/test_link")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link.ShortCode != "qazwsx123_" {
		t.Errorf("expected %q, got %q", "qazwsx123_", link.ShortCode)
	}
}

func TestGetByOriginalURL_NotFound(t *testing.T) {

	r := inmemory.NewRepository()

	_, err := r.GetByOriginalURL(context.Background(), "https://not-found.com")
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestConcurrentCreate(t *testing.T) {

	r := inmemory.NewRepository()
	ctx := context.Background()
	const goroutines = 10

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results []error
	)

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()
			err := r.Create(ctx, &model.Link{
				ShortCode:   fmt.Sprintf("short%06d", i),
				OriginalURL: "https://ozon.ru/test_link",
			})
			mu.Lock()
			results = append(results, err)
			mu.Unlock()
		}(i)
	}
	wg.Wait()

	successCount := 0
	for _, err := range results {
		if err == nil {
			successCount++
		} else if !errors.Is(err, apperrors.ErrOriginalURLConflict) {
			t.Errorf("unexpected error: %v", err)
		}
	}
	if successCount != 1 {
		t.Errorf("expected exactly 1 success, got %d", successCount)
	}
}
