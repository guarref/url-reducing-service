package web

import (
	"context"
	"net/http"
	"strings"

	"github.com/guarref/url-reducing-service/internal/apperrors"
	"github.com/labstack/echo/v4"
)

type LinkService interface {
	ReduceURL(ctx context.Context, originalURL string) (shortCode string, created bool, err error)
	GetOriginalURL(ctx context.Context, shortCode string) (originalURL string, err error)
}

type LinkHandler struct {
	service LinkService
	baseURL string
}

func NewLinkHandler(svc LinkService, baseURL string) *LinkHandler {
	return &LinkHandler{
		service: svc,
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

func RegisterRoutes(e *echo.Echo, svc LinkService, baseURL string) {
	
	h := NewLinkHandler(svc, baseURL)

	api := e.Group("/api/v1/links")
	api.POST("/reduce", h.CreateShortURL)

	e.GET("/:code", h.GetOriginalURL)
}

func (h *LinkHandler) CreateShortURL(c echo.Context) error {

	var req reduceURLRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.BadRequest("failed to decode request body")
	}

	req.URL = strings.TrimSpace(req.URL)
	code, created, err := h.service.ReduceURL(c.Request().Context(), req.URL)
	if err != nil {
		return err
	}

	resp := reduceURLResponse{
		URL:       req.URL,
		ShortCode: code,
		ShortURL:  h.baseURL + "/" + code,
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}

	return c.JSON(status, resp)
}

func (h *LinkHandler) GetOriginalURL(c echo.Context) error {

	code := c.Param("code")

	originalURL, err := h.service.GetOriginalURL(c.Request().Context(), code)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, getOriginalURLResponse{
		URL: originalURL,
	})
}
