package service_test

import (
	"context"
	"testing"

	"github.com/guarref/url-reducing-service/internal/apperrors"
	"github.com/guarref/url-reducing-service/internal/model"
	"github.com/guarref/url-reducing-service/internal/service"
	"github.com/guarref/url-reducing-service/internal/storage"
)

type mockRepo struct {
	codeToLink map[string]*model.Link
	urlToCode  map[string]string
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		codeToLink: make(map[string]*model.Link),
		urlToCode:  make(map[string]string),
	}
}

func (m *mockRepo) GetByShortCode(_ context.Context, code string) (*model.Link, error) {
	
	link, ok := m.codeToLink[code]
	if !ok {
		return nil, apperrors.ErrNotFound
	}

	return link, nil
}

func (m *mockRepo) GetByOriginalURL(_ context.Context, originalURL string) (*model.Link, error) {
	
	code, ok := m.urlToCode[originalURL]
	if !ok {
		return nil, apperrors.ErrNotFound
	}

	return m.codeToLink[code], nil
}

func (m *mockRepo) Create(_ context.Context, link *model.Link) error {
	
	if _, exists := m.urlToCode[link.OriginalURL]; exists {
		return apperrors.ErrOriginalURLConflict
	}
	if _, exists := m.codeToLink[link.ShortCode]; exists {
		return apperrors.ErrShortCodeConflict
	}
	stored := &model.Link{ShortCode: link.ShortCode, OriginalURL: link.OriginalURL}
	m.codeToLink[link.ShortCode] = stored
	m.urlToCode[link.OriginalURL] = link.ShortCode

	return nil
}

type mockGenerator struct {
	codes []string
	idx   int
}

func (g *mockGenerator) Generate() (string, error) {
	code := g.codes[g.idx%len(g.codes)]
	g.idx++
	return code, nil
}

func TestReduceURL_Success(t *testing.T) {
	
	svc := service.NewServiceWithGenerator(newMockRepo(), &mockGenerator{codes: []string{"qwertyuiop"}})

	code, created, err := svc.ReduceURL(context.Background(), "https://ozon.ru/test_link")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created {
		t.Error("expected created=true on first insert")
	}
	if code != "qwertyuiop" {
		t.Errorf("expected %q, got %q", "qwertyuiop", code)
	}
}

func TestReduceURL_ExistingURL_ReturnsSameCode(t *testing.T) {
	
	svc := service.NewServiceWithGenerator(newMockRepo(), &mockGenerator{codes: []string{"qwertyuiop", "aaaaaaaaaa"}})

	_, _, err := svc.ReduceURL(context.Background(), "https://ozon.ru/test_link")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	code, created, err := svc.ReduceURL(context.Background(), "https://ozon.ru/test_link")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created {
		t.Error("expected created=false when URL already exists")
	}
	if code != "qwertyuiop" {
		t.Errorf("expected existing code %q, got %q", "qwertyuiop", code)
	}
}

func TestReduceURL_ShortCodeCollision_Retries(t *testing.T) {
	
	repo := newMockRepo()
	repo.codeToLink["conflict_1"] = &model.Link{ShortCode: "conflict_1", OriginalURL: "https://google.com"}
	repo.urlToCode["https://google.com"] = "conflict_1"

	svc := service.NewServiceWithGenerator(repo, &mockGenerator{codes: []string{"conflict_1", "all_succes"}})

	code, _, err := svc.ReduceURL(context.Background(), "https://ozon.ru/test_link")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != "all_succes" {
		t.Errorf("expected %q after retry, got %q", "all_succes", code)
	}
}

func TestGetOriginalURL_Success(t *testing.T) {
	
	repo := newMockRepo()
	svc := service.NewServiceWithGenerator(repo, &mockGenerator{codes: []string{"qwertyuiop"}})

	_, _, err := svc.ReduceURL(context.Background(), "https://ozon.ru/test_link")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	original, err := svc.GetOriginalURL(context.Background(), "qwertyuiop")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if original != "https://ozon.ru/test_link" {
		t.Errorf("expected %q, got %q", "https://ozon.ru/test_link", original)
	}
}

func TestGetOriginalURL_NotFound(t *testing.T) {
		svc := service.NewServiceWithGenerator(newMockRepo(), &mockGenerator{codes: []string{"qwertyuiop"}})

	_, err := svc.GetOriginalURL(context.Background(), "notexists0")
	if !apperrors.IsNotFound(err) {
		t.Errorf("expected IsNotFound, got %v", err)
	}
}

func TestReduceURL_InvalidURL(t *testing.T) {
	svc := service.NewServiceWithGenerator(newMockRepo(), &mockGenerator{codes: []string{"qwertyuiop"}})

	_, _, err := svc.ReduceURL(context.Background(), "not-a-url")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestReduceURL_EmptyURL(t *testing.T) {
	svc := service.NewServiceWithGenerator(newMockRepo(), &mockGenerator{codes: []string{"qwertyuiop"}})

	_, _, err := svc.ReduceURL(context.Background(), "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetOriginalURL_InvalidShortCode(t *testing.T) {
	svc := service.NewServiceWithGenerator(newMockRepo(), &mockGenerator{codes: []string{"qwertyuiop"}})

	_, err := svc.GetOriginalURL(context.Background(), "abc")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

var _ storage.Repository = (*mockRepo)(nil)
