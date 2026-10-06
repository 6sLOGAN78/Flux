package app

import (
	"context"
	"errors"
	"net"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/database"
	"github.com/6sLOGAN78/flux/internal/lib/email"
	"github.com/6sLOGAN78/flux/internal/lib/job"
	loggerPkg "github.com/6sLOGAN78/flux/internal/logger"
	"github.com/6sLOGAN78/flux/internal/server"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

func roleTestConfig() *config.Config {
	return &config.Config{Primary: config.Primary{Env: "test"}, Observability: config.DefaultObservabilityConfig(),
		API:        config.RoleConfig{ListenAddress: "127.0.0.1:0", DrainTimeout: time.Second, ReadinessTimeout: time.Second},
		Redirector: config.RoleConfig{ListenAddress: "127.0.0.1:0", DrainTimeout: time.Second, ReadinessTimeout: time.Second},
		Worker:     config.RoleConfig{ListenAddress: "127.0.0.1:0", DrainTimeout: time.Second, ReadinessTimeout: time.Second},
	}
}

func roleSpies(fail string, opened, closed *[]string, cause error) roleFactories {
	stage := func(name string) (func(context.Context) error, error) {
		if name == fail {
			return nil, cause
		}
		*opened = append(*opened, name)
		return func(context.Context) error { *closed = append(*closed, name); return cause }, nil
	}
	return roleFactories{
		logger: func(*config.Config) (*zerolog.Logger, *loggerPkg.LoggerService, func(context.Context) error, error) {
			close, err := stage("logger")
			log := zerolog.Nop()
			return &log, nil, close, err
		},
		database: func(context.Context, *server.Server) (*database.Database, func(context.Context) error, error) {
			close, err := stage("database")
			return &database.Database{}, close, err
		},
		redis: func(context.Context, *server.Server) (*redis.Client, func(context.Context) error, error) {
			close, err := stage("redis")
			return nil, close, err
		},
		producer: func(*server.Server) (*job.JobService, func(context.Context) error, error) {
			close, err := stage("producer")
			return nil, close, err
		},
		email: func(*server.Server) (*email.Client, error) {
			if fail == "email" {
				return nil, cause
			}
			*opened = append(*opened, "email")
			return nil, nil
		},
		consumer: func(*server.Server, *email.Client) (*job.JobService, func(context.Context) error, error) {
			close, err := stage("consumer")
			return nil, close, err
		},
		startConsumer: func(*job.JobService) error {
			if fail == "start consumer" {
				return cause
			}
			*opened = append(*opened, "start consumer")
			return nil
		},
		router: func(role config.Role, srv *server.Server) (*echo.Echo, error) {
			if fail == "router" {
				return nil, cause
			}
			*opened = append(*opened, "router")
			return defaultRoleRouter(role, srv)
		},
		listen: func(address string) (net.Listener, error) {
			if fail == "listener" {
				return nil, cause
			}
			*opened = append(*opened, "listener")
			return net.Listen("tcp", address)
		},
	}
}

func TestRoleResourceGraphs(t *testing.T) {
	for _, test := range []struct {
		role     config.Role
		producer bool
		want     []string
	}{
		{config.RoleAPI, false, []string{"logger", "database", "router", "listener"}},
		{config.RoleAPI, true, []string{"logger", "database", "redis", "producer", "router", "listener"}},
		{config.RoleRedirector, false, []string{"logger", "router", "listener"}},
		{config.RoleWorker, false, []string{"logger", "redis", "email", "consumer", "router", "listener", "start consumer"}},
	} {
		t.Run(string(test.role)+string(rune('0'+boolInt(test.producer))), func(t *testing.T) {
			cfg := roleTestConfig()
			cfg.API.ProducerEnabled = test.producer
			var opened, closed []string
			role, err := newRole(context.Background(), test.role, cfg, roleSpies("", &opened, &closed, nil))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(opened, test.want) {
				t.Fatalf("role graph %v, want %v", opened, test.want)
			}
			if role.Server.DB != nil && test.role != config.RoleAPI {
				t.Fatal("non-API allocated PostgreSQL")
			}
			if test.role != config.RoleAPI {
				for _, path := range []string{"/docs", "/static/openapi.json", "/api/v1/links"} {
					recorder := httptest.NewRecorder()
					role.HTTP.ServeHTTP(recorder, httptest.NewRequest("GET", path, nil))
					if recorder.Code != 404 {
						t.Fatalf("%s exposed %s: %d", test.role, path, recorder.Code)
					}
				}
			}
			if err := role.Close(context.Background()); err != nil {
				t.Fatal(err)
			}
			wantClosed := []string{"logger"}
			if test.role == config.RoleAPI {
				wantClosed = []string{"database", "logger"}
				if test.producer {
					wantClosed = []string{"producer", "redis", "database", "logger"}
				}
			}
			if test.role == config.RoleWorker {
				wantClosed = []string{"consumer", "redis", "logger"}
			}
			if !reflect.DeepEqual(closed, wantClosed) {
				t.Fatalf("close order %v, want %v", closed, wantClosed)
			}
			if err := role.Close(context.Background()); err != nil || !reflect.DeepEqual(closed, wantClosed) {
				t.Fatal("repeated close changed effects")
			}
		})
	}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func TestRoleConstructionFailureUnwinds(t *testing.T) {
	cause := errors.New("SECRET-MARKER")
	for _, role := range []config.Role{config.RoleAPI, config.RoleRedirector, config.RoleWorker} {
		stages := []string{"logger", "router", "listener"}
		if role == config.RoleAPI {
			stages = []string{"logger", "database", "redis", "producer", "router", "listener"}
		}
		if role == config.RoleWorker {
			stages = []string{"logger", "redis", "email", "consumer", "router", "listener", "start consumer"}
		}
		for _, fail := range stages {
			t.Run(string(role)+"/"+fail, func(t *testing.T) {
				cfg := roleTestConfig()
				cfg.API.ProducerEnabled = true
				var opened, closed []string
				runtime, err := newRole(context.Background(), role, cfg, roleSpies(fail, &opened, &closed, cause))
				if runtime != nil || !errors.Is(err, cause) || strings.Contains(err.Error(), "SECRET-MARKER") {
					t.Fatalf("unsafe or lost failure: %v", err)
				}
				var want []string
				for i := len(opened) - 1; i >= 0; i-- {
					switch opened[i] {
					case "logger", "database", "redis", "producer", "consumer":
						want = append(want, opened[i])
					}
				}
				if !reflect.DeepEqual(closed, want) {
					t.Fatalf("failed %s: closed %v, want %v", fail, closed, want)
				}
			})
		}
	}
}

func TestRoleServerReceivesOnlyExplicitResources(t *testing.T) {
	log := zerolog.Nop()
	srv, err := server.New(roleTestConfig(), &log, nil)
	if err != nil || srv.DB != nil || srv.Redis != nil || srv.Job != nil {
		t.Fatalf("server allocated undeclared dependencies: %v", err)
	}
}
