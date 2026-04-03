package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/guarref/url-reducing-service/internal/apperrors"
	"github.com/guarref/url-reducing-service/internal/model"
	"github.com/guarref/url-reducing-service/internal/storage"
)

const (
	alphabet   = "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890_"
	codeLength = 10
	retryCount = 10
)

type codeGenerator interface {
	Generate() (string, error)
}

type randomCodeGenerator struct{}

type Service struct {
	repo      storage.Repository
	generator codeGenerator
}

func NewService(repo storage.Repository) *Service {
	return &Service{
		repo:      repo,
		generator: &randomCodeGenerator{},
	}
}

func NewServiceWithGenerator(repo storage.Repository, generator codeGenerator) *Service {
	return &Service{
		repo:      repo,
		generator: generator,
	}
}

func (s *Service) ReduceURL(ctx context.Context, originalURL string) (string, bool, error) {

	originalURL = strings.TrimSpace(originalURL)

	if originalURL == "" {
		return "", false, apperrors.BadRequest("original url is empty")
	}

	parsedURL, err := url.ParseRequestURI(originalURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return "", false, apperrors.BadRequest("invalid original url")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", false, apperrors.BadRequest("original url must use http or https")
	}

	link, err := s.repo.GetByOriginalURL(ctx, originalURL)
	if err == nil {
		return link.ShortCode, false, nil
	}

	if !errors.Is(err, apperrors.ErrNotFound) {
		return "", false, apperrors.Wrap(err, "check existing url")
	}

	for i := 0; i < retryCount; i++ {
		code, err := s.generator.Generate()
		if err != nil {
			return "", false, apperrors.Wrap(err, "generate code")
		}

		err = s.repo.Create(ctx, &model.Link{
			ShortCode:   code,
			OriginalURL: originalURL,
		})
		if err == nil {
			return code, true, nil
		}

		if errors.Is(err, apperrors.ErrShortCodeConflict) {
			continue
		}

		if errors.Is(err, apperrors.ErrOriginalURLConflict) {
			existing, getErr := s.repo.GetByOriginalURL(ctx, originalURL)
			if getErr != nil {
				return "", false, apperrors.Wrap(getErr, "get existing after race")
			}
			return existing.ShortCode, false, nil
		}

		return "", false, apperrors.Wrap(err, "create link")
	}

	return "", false, apperrors.InternalServerError("failed to generate unique short code after %d attempts", retryCount)
}

func (s *Service) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	
	shortCode = strings.TrimSpace(shortCode)

	if shortCode == "" {
		return "", apperrors.BadRequest("short code is required")
	}

	if !isValidShortCode(shortCode) {
		return "", apperrors.BadRequest("invalid short code")
	}

	link, err := s.repo.GetByShortCode(ctx, shortCode)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return "", apperrors.NotFound("short code %s not found", shortCode)
		}
		return "", apperrors.Wrap(err, "get by short code")
	}

	return link.OriginalURL, nil
}

func (g *randomCodeGenerator) Generate() (string, error) {

	result := make([]byte, codeLength)
	random := make([]byte, codeLength*2)

	for i := 0; i < codeLength; {
		if _, err := rand.Read(random); err != nil {
			return "", fmt.Errorf("read random bytes: %w", err)
		}

		for _, b := range random {
			result[i] = alphabet[int(b)%len(alphabet)]
			i++

			if i == codeLength {
				break
			}
		}
	}

	return string(result), nil
}

func isValidShortCode(code string) bool {
	
	if len(code) != codeLength {
		return false
	}

	for _, ch := range code {
		switch {
		case ch >= 'a' && ch <= 'z':
		case ch >= 'A' && ch <= 'Z':
		case ch >= '0' && ch <= '9':
		case ch == '_':
		default:
			return false
		}
	}

	return true
}
