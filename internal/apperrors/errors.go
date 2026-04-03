package apperrors

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"
)

var (
	ErrNotFound            = errors.New("not_found")
	ErrShortCodeConflict   = errors.New("short_code_conflict")
	ErrOriginalURLConflict = errors.New("original_url_conflict")
	ErrNilLink             = errors.New("nil_link")
)

type AppError struct {
	HTTPCode int    `json:"code"`
	Message  string `json:"message"`
	Details  string `json:"details,omitempty"`
	Caller   string `json:"-"`
	Err      error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func Wrap(err error, message string, args ...any) error {

	_, file, line, _ := runtime.Caller(1)
	caller := fmt.Sprintf("%s:%d", filepath.Base(file), line)
	details := fmt.Sprintf(message, args...)

	var appErr *AppError
	if errors.As(err, &appErr) {
		newDetails := details
		if appErr.Details != "" {
			newDetails += ": " + appErr.Details
		}

		return &AppError{
			HTTPCode: appErr.HTTPCode,
			Message:  appErr.Message,
			Details:  newDetails,
			Caller:   caller,
			Err:      fmt.Errorf("%s: %w", details, appErr.Err),
		}
	}

	return &AppError{
		HTTPCode: http.StatusInternalServerError,
		Message:  http.StatusText(http.StatusInternalServerError),
		Details:  details + ": " + err.Error(),
		Caller:   caller,
		Err:      fmt.Errorf("%s: %w", details, err),
	}
}

func newAppError(httpCode int, sentinel error, message, details string, args ...any) *AppError {

	_, file, line, _ := runtime.Caller(2)
	d := fmt.Sprintf(details, args...)

	var err error
	if sentinel != nil {
		err = fmt.Errorf("%s: %w", d, sentinel)
	} else {
		err = errors.New(d)
	}

	return &AppError{
		HTTPCode: httpCode,
		Message:  message,
		Details:  d,
		Caller:   fmt.Sprintf("%s:%d", filepath.Base(file), line),
		Err:      err,
	}
}

func NotFound(details string, args ...any) error {
	return newAppError(http.StatusNotFound, ErrNotFound, "Not Found", details, args...)
}

func BadRequest(details string, args ...any) error {
	return newAppError(http.StatusBadRequest, nil, "Bad Request", details, args...)
}

func InternalServerError(details string, args ...any) error {
	return newAppError(http.StatusInternalServerError, nil, "Internal Server Error", details, args...)
}

func OriginalURLConflict(details string, args ...any) error {
	return newAppError(http.StatusConflict, ErrOriginalURLConflict, "Conflict", details, args...)
}

func ShortCodeConflict(details string, args ...any) error {
	return newAppError(http.StatusConflict, ErrShortCodeConflict, "Conflict", details, args...)
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

func IsShortCodeConflict(err error) bool {
	return errors.Is(err, ErrShortCodeConflict)
}

func IsOriginalURLConflict(err error) bool {
	return errors.Is(err, ErrOriginalURLConflict)
}

func IsNilLink(err error) bool {
	return errors.Is(err, ErrNilLink)
}
