// Package handler adapts validated requests and serves health and embedded documentation.
package handler

import (
	"errors"

	"github.com/6sLOGAN78/flux/internal/middleware"
	"github.com/6sLOGAN78/flux/internal/server"
	"github.com/6sLOGAN78/flux/internal/validation"
	"github.com/labstack/echo/v4"
)

// Handler provides base functionality for all handlers
type Handler struct {
	server *server.Server
}

// NewHandler creates a new base handler
func NewHandler(s *server.Server) Handler {
	return Handler{server: s}
}

// HandlerFunc represents a typed handler function that processes a request and returns a response
//
//nolint:revive // Preserve the established exported type name for existing callers.
type HandlerFunc[Req validation.Validatable, Res any] func(c echo.Context, req Req) (Res, error)

// HandlerFuncNoContent represents a typed handler function that processes a request without returning content
//
//nolint:revive // Preserve the established exported type name for existing callers.
type HandlerFuncNoContent[Req validation.Validatable] func(c echo.Context, req Req) error

// ResponseHandler defines the interface for handling different response types
type ResponseHandler interface {
	Handle(c echo.Context, result interface{}) error
}

// JSONResponseHandler handles JSON responses
type JSONResponseHandler struct {
	status int
}

// Handle serializes the response through Echo.
// Handle serializes the response through Echo.
func (h JSONResponseHandler) Handle(c echo.Context, result interface{}) error {
	return c.JSON(h.status, result)
}

// NoContentResponseHandler handles no-content responses
type NoContentResponseHandler struct {
	status int
}

// Handle sends the configured status with no response body.
func (h NoContentResponseHandler) Handle(c echo.Context, _ interface{}) error {
	return c.NoContent(h.status)
}

// FileResponseHandler handles file responses
type FileResponseHandler struct {
	filename    string
	contentType string
	status      int
}

// Handle writes byte responses as an attachment and rejects unexpected result types.
func (h FileResponseHandler) Handle(c echo.Context, result interface{}) error {
	data, ok := result.([]byte)
	if !ok {
		return errors.New("file handler response must contain bytes")
	}
	c.Response().Header().Set("Content-Disposition", "attachment; filename="+h.filename)
	return c.Blob(h.status, h.contentType, data)
}

// handleRequest is the unified handler function that eliminates code duplication
func handleRequest[Req validation.Validatable](
	c echo.Context,
	req Req,
	handler func(c echo.Context, req Req) (interface{}, error),
	responseHandler ResponseHandler,
) error {
	log := middleware.GetLogger(c)
	if err := validation.BindAndValidate(c, req); err != nil {
		log.Warn().Str("operation", "http.request").Str("error.category", "validation").Msg("http.request")
		return err
	}
	result, err := handler(c, req)
	if err != nil {
		log.Error().Str("operation", "http.request").Str("error.category", "unknown").Msg("http.request")
		return err
	}

	return responseHandler.Handle(c, result)
}

// Handle wraps a handler with validation, error handling, logging, metrics, and tracing.
// newRequest must return a fresh request value for every invocation.
func Handle[Req validation.Validatable, Res any](
	_ Handler,
	handler HandlerFunc[Req, Res],
	status int,
	newRequest func() Req,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		return handleRequest(c, newRequest(), func(c echo.Context, req Req) (interface{}, error) {
			return handler(c, req)
		}, JSONResponseHandler{status: status})
	}
}

// HandleFile wraps a file handler; newRequest must return a fresh request per invocation.
func HandleFile[Req validation.Validatable](
	_ Handler,
	handler HandlerFunc[Req, []byte],
	status int,
	newRequest func() Req,
	filename string,
	contentType string,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		return handleRequest(c, newRequest(), func(c echo.Context, req Req) (interface{}, error) {
			return handler(c, req)
		}, FileResponseHandler{
			status:      status,
			filename:    filename,
			contentType: contentType,
		})
	}
}

// HandleNoContent wraps a handler with validation, error handling, logging, metrics, and tracing for
// endpoints that don't return content
// newRequest must return a fresh request value for every invocation.
func HandleNoContent[Req validation.Validatable](
	_ Handler,
	handler HandlerFuncNoContent[Req],
	status int,
	newRequest func() Req,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		return handleRequest(c, newRequest(), func(c echo.Context, req Req) (interface{}, error) {
			err := handler(c, req)
			return nil, err
		}, NoContentResponseHandler{status: status})
	}
}
