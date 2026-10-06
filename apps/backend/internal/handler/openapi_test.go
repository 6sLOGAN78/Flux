package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/6sLOGAN78/flux/internal/handler"
	"github.com/labstack/echo/v4"
)

func TestOpenAPIAlternateDirectory(t *testing.T) {
	canonical, err := os.ReadFile("../../../../packages/openapi/openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(t.TempDir())
	h := handler.NewOpenAPIHandler(nil)
	r := echo.New()
	r.GET("/docs", h.ServeOpenAPIUI)
	r.GET("/static/openapi.json", h.ServeOpenAPISpec)

	for _, path := range []string{"/docs", "/static/openapi.json"} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			r.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
			if response.Header().Get("Cache-Control") != "no-cache" {
				t.Fatal("documentation must revalidate its version")
			}
			if path == "/static/openapi.json" {
				if !bytes.Equal(response.Body.Bytes(), canonical) || !json.Valid(response.Body.Bytes()) {
					t.Fatal("served contract differs from canonical authored artifact")
				}
				if !strings.HasPrefix(response.Header().Get("Content-Type"), "application/json") {
					t.Fatal("contract response is not JSON")
				}
				return
			}
			body := response.Body.String()
			for _, fragment := range []string{"<!DOCTYPE html>", "data-url=\"/static/openapi.json\"", "@scalar/api-reference@", "integrity=\"sha384-", "crossorigin=\"anonymous\""} {
				if !strings.Contains(body, fragment) {
					t.Fatalf("docs missing %q", fragment)
				}
			}
			policy := response.Header().Get("Content-Security-Policy")
			for _, fragment := range []string{"default-src 'none'", "script-src 'sha384-", "connect-src 'self'", "frame-ancestors 'none'", "base-uri 'none'"} {
				if !strings.Contains(policy, fragment) {
					t.Fatalf("docs CSP missing %q", fragment)
				}
			}
			if strings.Contains(policy, "unsafe-eval") {
				t.Fatal("docs must not permit eval")
			}
		})
	}
}
