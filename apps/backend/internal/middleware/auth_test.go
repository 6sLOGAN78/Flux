package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/middleware"
	"github.com/6sLOGAN78/flux/internal/server"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestSessionExactPreflight(t *testing.T) {
	for _, origin := range []string{"", "null", "https://foreign.test", "https://flux.test.attacker.test"} {
		t.Run(origin, func(t *testing.T) {
			e := echo.New()
			global := middleware.NewGlobalMiddlewares(&server.Server{Config: &config.Config{
				Server: config.ServerConfig{CORSAllowedOrigins: []string{"https://flux.test"}},
			}})
			e.Use(global.CORS())
			e.POST("/api/v1/test-only", func(c echo.Context) error { return c.NoContent(204) })
			req := httptest.NewRequest(http.MethodOptions, "/api/v1/test-only", nil)
			req.Header.Set("Origin", origin)
			req.Header.Set("Access-Control-Request-Method", "POST")
			req.Header.Set("Access-Control-Request-Headers", "authorization,content-type")
			response := httptest.NewRecorder()
			e.ServeHTTP(response, req)
			require.Equal(t, 403, response.Code)
			require.Empty(t, response.Header().Get("Access-Control-Allow-Origin"))
		})
	}
}

func TestSessionAllowedPreflight(t *testing.T) {
	e := echo.New()
	global := middleware.NewGlobalMiddlewares(&server.Server{Config: &config.Config{
		Server: config.ServerConfig{CORSAllowedOrigins: []string{"https://flux.test"}},
	}})
	e.Use(global.CORS())
	e.POST("/api/v1/test-only", func(c echo.Context) error { return c.NoContent(204) })
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/test-only", nil)
	req.Header.Set("Origin", "https://flux.test")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "authorization,content-type")
	response := httptest.NewRecorder()
	e.ServeHTTP(response, req)
	require.Equal(t, 204, response.Code)
	require.Equal(t, "https://flux.test", response.Header().Get("Access-Control-Allow-Origin"))
	require.Empty(t, response.Header().Get("Access-Control-Allow-Credentials"))
	require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
}
