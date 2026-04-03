package apperrors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func HttpErrorHandler(env string) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		code := http.StatusInternalServerError
		message := http.StatusText(http.StatusInternalServerError)
		var details any

		var appErr *AppError
		var httpErr *echo.HTTPError

		switch {
		case errors.As(err, &appErr):
			code = appErr.HTTPCode
			message = appErr.Message
			details = appErr.Details

		case errors.As(err, &httpErr):
			code = httpErr.Code
			message = fmt.Sprintf("%v", httpErr.Message)
			if httpErr.Internal != nil {
				details = httpErr.Internal.Error()
			}

		default:
			details = err.Error()
		}

		response := map[string]any{
			"message": message,
		}
		if env != "production" && details != nil && details != "" {
			response["details"] = details
		}

		_ = c.JSON(code, response)
	}
}
