package errs

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

// HTTPErrorResponse represents standardized HTTP JSON error payload.
type HTTPErrorResponse struct {
	Status      int          `json:"status"`
	Code        string       `json:"code,omitempty"`
	Message     string       `json:"message"`
	FieldErrors []FieldError `json:"field_errors,omitempty"`
}

// ToHTTPError converts internal domain errors or HTTPError to Echo HTTP Error responses.
func ToHTTPError(err error) *echo.HTTPError {
	if err == nil {
		return nil
	}

	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		codeStr := ""
		if httpErr.Code != nil {
			codeStr = *httpErr.Code
		}
		return echo.NewHTTPError(httpErr.StatusCode, HTTPErrorResponse{
			Status:      httpErr.StatusCode,
			Code:        codeStr,
			Message:     httpErr.Message,
			FieldErrors: httpErr.FieldErrors,
		})
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
		case errors.Is(appErr.Err, ErrQuotaExceeded):
			return echo.NewHTTPError(http.StatusPaymentRequired, appErr.Message)
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
	case errors.Is(err, ErrQuotaExceeded):
		return echo.NewHTTPError(http.StatusPaymentRequired, "resource quota exceeded")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
}

// CustomHTTPErrorHandler is a centralized error handler for Echo.
func CustomHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	// Unpack echo.HTTPError if it's already an echo error (e.g. from middleware or router)
	var he *echo.HTTPError
	if errors.As(err, &he) {
		// Already an echo error, but let's see if we can structure it
		if he.Message == nil {
			he.Message = http.StatusText(he.Code)
		}
	} else {
		// Convert our domain errors to echo.HTTPError
		he = ToHTTPError(err)
	}

	// Fallback generic format if it's not our struct
	var response interface{}
	switch m := he.Message.(type) {
	case HTTPErrorResponse:
		response = m
	case string:
		response = HTTPErrorResponse{
			Status:  he.Code,
			Message: m,
		}
	default:
		response = HTTPErrorResponse{
			Status:  he.Code,
			Message: "An error occurred",
		}
	}

	if c.Request().Method == http.MethodHead {
		c.NoContent(he.Code)
	} else {
		c.JSON(he.Code, response)
	}
}
