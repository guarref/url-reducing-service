package web_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/guarref/url-reducing-service/internal/apperrors"
	"github.com/guarref/url-reducing-service/internal/web"
)

type mockService struct {
	reduceURLFn      func(ctx context.Context, url string) (string, bool, error)
	getOriginalURLFn func(ctx context.Context, shortURL string) (string, error)
}

func (m *mockService) ReduceURL(ctx context.Context, url string) (string, bool, error) {
	return m.reduceURLFn(ctx, url)
}

func (m *mockService) GetOriginalURL(ctx context.Context, shortURL string) (string, error) {
	return m.getOriginalURLFn(ctx, shortURL)
}

func decodeJSON(t *testing.T, body string, v any) {
	t.Helper()
	if err := json.Unmarshal([]byte(body), v); err != nil {
		t.Fatalf("decode body: %v (body was: %s)", err, body)
	}
}

func doJSONRequest(e *echo.Echo, method, path, body string) *httptest.ResponseRecorder {

	var reader *strings.Reader

	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func assertStatus(t *testing.T, rec *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if rec.Code != expected {
		t.Fatalf("expected status %d, got %d", expected, rec.Code)
	}
}

const testBaseURL = "http://localhost:8080"

func newTestEcho(svc *mockService) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = apperrors.HttpErrorHandler("development")
	web.RegisterRoutes(e, svc, testBaseURL)

	return e
}

// POST /api/v1/links/reduce
func TestReduceURL_201Created(t *testing.T) {

	svc := &mockService{
		reduceURLFn: func(_ context.Context, _ string) (string, bool, error) {
			return "qwertyuiop", true, nil
		},
	}
	e := newTestEcho(svc)

	rec := doJSONRequest(e, http.MethodPost, "/api/v1/links/reduce", `{"url":"https://ozon.ru/test_link"}`)
	assertStatus(t, rec, http.StatusCreated)

	var resp map[string]string
	decodeJSON(t, rec.Body.String(), &resp)
	if resp["short_code"] != "qwertyuiop" {
		t.Errorf("expected short_code %q, got %q", "qwertyuiop", resp["short_code"])
	}
	if resp["short_url"] != testBaseURL+"/qwertyuiop" {
		t.Errorf("expected short_url %q, got %q", testBaseURL+"/qwertyuiop", resp["short_url"])
	}
}

func TestReduceURL_200OK_AlreadyExists(t *testing.T) {

	svc := &mockService{
		reduceURLFn: func(_ context.Context, _ string) (string, bool, error) {
			return "qwertyuiop", false, nil
		},
	}
	e := newTestEcho(svc)

	rec := doJSONRequest(e, http.MethodPost, "/api/v1/links/reduce", `{"url":"https://ozon.ru/test_link"}`)
	assertStatus(t, rec, http.StatusOK)

	var resp map[string]string
	decodeJSON(t, rec.Body.String(), &resp)
	if resp["short_code"] != "qwertyuiop" {
		t.Errorf("expected short_code %q, got %q", "qwertyuiop", resp["short_code"])
	}
}

func TestReduceURL_400_InvalidBody(t *testing.T) {

	e := newTestEcho(&mockService{})

	rec := doJSONRequest(e, http.MethodPost, "/api/v1/links/reduce", `not json at all`)
	assertStatus(t, rec, http.StatusBadRequest)
}

func TestReduceURL_400_InvalidURL(t *testing.T) {

	svc := &mockService{
		reduceURLFn: func(_ context.Context, _ string) (string, bool, error) {
			return "", false, apperrors.BadRequest("invalid original url")
		},
	}
	e := newTestEcho(svc)

	rec := doJSONRequest(e, http.MethodPost, "/api/v1/links/reduce", `{"url":"not-a-url"}`)
	assertStatus(t, rec, http.StatusBadRequest)
}

func TestReduceURL_500_ServiceError(t *testing.T) {

	svc := &mockService{
		reduceURLFn: func(_ context.Context, _ string) (string, bool, error) {
			return "", false, apperrors.InternalServerError("db fail")
		},
	}
	e := newTestEcho(svc)

	rec := doJSONRequest(e, http.MethodPost, "/api/v1/links/reduce", `{"url":"https://ozon.ru/test_link"}`)
	assertStatus(t, rec, http.StatusInternalServerError)
}

// GET /:code
func TestGetOriginalURL_200(t *testing.T) {

	svc := &mockService{
		getOriginalURLFn: func(_ context.Context, _ string) (string, error) {
			return "https://ozon.ru/test_link", nil
		},
	}
	e := newTestEcho(svc)

	rec := doJSONRequest(e, http.MethodGet, "/qwertyuiop", "")
	assertStatus(t, rec, http.StatusOK)

	var resp map[string]string
	decodeJSON(t, rec.Body.String(), &resp)
	if resp["url"] != "https://ozon.ru/test_link" {
		t.Errorf("expected %q, got %q", "https://ozon.ru/test_link", resp["url"])
	}
}

func TestGetOriginalURL_404(t *testing.T) {

	svc := &mockService{
		getOriginalURLFn: func(_ context.Context, _ string) (string, error) {
			return "", apperrors.NotFound("short code not found")
		},
	}
	e := newTestEcho(svc)

	rec := doJSONRequest(e, http.MethodGet, "/notfound11", "")
	assertStatus(t, rec, http.StatusNotFound)
}

func TestGetOriginalURL_500_ServiceError(t *testing.T) {

	svc := &mockService{
		getOriginalURLFn: func(_ context.Context, _ string) (string, error) {
			return "", apperrors.InternalServerError("db fail")
		},
	}
	e := newTestEcho(svc)

	rec := doJSONRequest(e, http.MethodGet, "/qwertyuiop", "")
	assertStatus(t, rec, http.StatusInternalServerError)
}
