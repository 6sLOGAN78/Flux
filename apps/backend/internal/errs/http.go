package errs

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

// HTTPErrorResponse represents standardized HTTP JSON error payload.
type HTTPErrorResponse struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ToHTTPError converts internal domain errors to Echo HTTP Error responses.
func ToHTTPError(err error) *echo.HTTPError {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		switch {
		case errors.Is(appErr.Err, ErrNotFound):
			return echo.NewHTTPError(http.StatusNotFound, appErr.Message)
		case errors.Is(appErr.Err, ErrUnauthorized):
			return echo.NewHTTPError(http.StatusUnauthorized, appErr.Message)
		case errors.Is(appErr.Err, ErrForbidden):
			return echo.NewHTTPError(http.StatusForbidden, appErr.Message)
		case errors.Is(appErr.Err, ErrExpired):
			return echo.NewHTTPError(http.StatusGone, appErr.Message)
		case errors.Is(appErr.Err, ErrRateLimitExceeded):
			return echo.NewHTTPError(http.StatusTooManyRequests, appErr.Message)
		}
	}

	switch {
	case errors.Is(err, ErrNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "resource not found")
	case errors.Is(err, ErrUnauthorized):
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized access")
	case errors.Is(err, ErrForbidden):
		return echo.NewHTTPError(http.StatusForbidden, "forbidden request")
	case errors.Is(err, ErrExpired):
		return echo.NewHTTPError(http.StatusGone, "resource has expired")
	case errors.Is(err, ErrRateLimitExceeded):
		return echo.NewHTTPError(http.StatusTooManyRequests, "rate limit exceeded")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
}
