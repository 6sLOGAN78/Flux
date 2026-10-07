package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/6sLOGAN78/flux/internal/handler"
	"github.com/labstack/echo/v4"
)

type isolatedRequest struct {
	Marker   string `json:"marker"`
	Optional string `json:"optional"`
}

func (*isolatedRequest) Validate() error { return nil }

// Serialize decoding to prove cross-request contamination without relying on
// scheduling or a race report. Both callbacks wait until both bodies are bound.
type synchronizedBinder struct{ mu sync.Mutex }

func (b *synchronizedBinder) Bind(v interface{}, c echo.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return (&echo.DefaultBinder{}).Bind(v, c)
}

func isolationWrapper(kind string, observe func(*isolatedRequest)) echo.HandlerFunc {
	newRequest := func() *isolatedRequest { return &isolatedRequest{} }
	switch kind {
	case "json":
		return handler.Handle(handler.NewHandler(nil),
			func(_ echo.Context,
				req *isolatedRequest) (isolatedRequest,
				error) {
				observe(req)
				return *req, nil
			}, http.StatusOK, newRequest)
	case "file":
		return handler.HandleFile(handler.NewHandler(nil),
			func(_ echo.Context,
				req *isolatedRequest) ([]byte,
				error) {
				observe(req)
				return []byte(req.Marker), nil
			}, http.StatusOK, newRequest, "marker.txt", "text/plain")
	default:
		return handler.HandleNoContent(handler.NewHandler(nil),
			func(c echo.Context,
				req *isolatedRequest) error {
				observe(req)
				c.Response().Header().Set("X-Marker", req.Marker)
				return nil
			}, http.StatusNoContent, newRequest)
	}
}

//nolint:gocognit // Keep this regression scenario and its ordered failure assertions together.
func TestWrappersIsolateConcurrentRequests(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"json", "file", "no_content"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			e := echo.New()
			e.Binder = &synchronizedBinder{}
			var bound sync.WaitGroup
			bound.Add(2)
			wrapped := isolationWrapper(kind, func(*isolatedRequest) { bound.Done(); bound.Wait() })
			var finished sync.WaitGroup
			for _, marker := range []string{"request-alpha", "request-beta"} {
				finished.Go(func() {
					body, err := json.Marshal(isolatedRequest{Marker: marker})
					if err != nil {
						t.Error(err)
						return
					}
					r := httptest.NewRequest(http.MethodPost,
						"/isolation",
						strings.NewReader(string(body)))
					r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
					response := httptest.NewRecorder()
					if err75 := wrapped(e.NewContext(r, response)); err75 != nil {
						t.Error(err75)
						return
					}
					got := response.Body.String()
					if kind == "json" {
						var decoded isolatedRequest
						if err82 := json.Unmarshal(response.Body.Bytes(),
							&decoded); err82 != nil {
							t.Error(err82)
							return
						}
						got = decoded.Marker
					}
					if kind == "no_content" {
						got = response.Header().Get("X-Marker")
					}
					if got != marker {
						t.Errorf("request %q received %q", marker, got)
					}
				})
			}
			finished.Wait()
		})
	}
}

func TestWrappersDoNotRetainOmittedFields(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"json", "file", "no_content"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			e := echo.New()
			var optional []string
			wrapped := isolationWrapper(kind,
				func(req *isolatedRequest) {
					optional = append(optional,
						req.Optional)
				})
			for _, body := range []string{`{"marker":"first","optional":"private-first-request"}`,
				`{"marker":"second"}`} {
				r := httptest.NewRequest(http.MethodPost, "/isolation", strings.NewReader(body))
				r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				if err := wrapped(e.NewContext(r, httptest.NewRecorder())); err != nil {
					t.Fatal(err)
				}
			}
			if optional[1] != "" {
				t.Errorf("second request retained %q", optional[1])
			}
		})
	}
}
