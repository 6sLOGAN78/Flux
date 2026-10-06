package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/database"
	"github.com/6sLOGAN78/flux/internal/lib/job"
	loggerPkg "github.com/6sLOGAN78/flux/internal/logger"
	"github.com/6sLOGAN78/flux/internal/observability"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// Server receives role-owned resources and adapts them for existing packages.
// It never allocates infrastructure or starts background consumers.
type Server struct {
	Role       config.Role
	Config     *config.Config
	Logger     *zerolog.Logger
	Telemetry  *observability.Telemetry
	DB         *database.Database
	Redis      *redis.Client
	Job        *job.JobService
	httpServer *http.Server
}

// Dependencies are allocated and closed by the composing app role.
type Dependencies struct {
	DB    *database.Database
	Redis *redis.Client
	Job   *job.JobService
}

// New retains the legacy call signature while accepting explicit dependencies.
func New(cfg *config.Config, log *zerolog.Logger, _ *loggerPkg.LoggerService, owned ...Dependencies) (*Server, error) {
	if cfg == nil || log == nil {
		return nil, errors.New("server configuration and logger required")
	}
	srv := &Server{Config: cfg, Logger: log}
	if len(owned) > 1 {
		return nil, errors.New("one explicit dependency graph required")
	}
	if len(owned) == 1 {
		srv.DB = owned[0].DB
		srv.Redis = owned[0].Redis
		srv.Job = owned[0].Job
	}
	return srv, nil
}

func (s *Server) SetupHTTPServer(handler http.Handler) {
	s.SetupHTTPServerAt(":"+s.Config.Server.Port, handler)
}

// SetupHTTPServerAt uses the selected role address with legacy HTTP timeouts.
func (s *Server) SetupHTTPServerAt(address string, handler http.Handler) {
	s.httpServer = &http.Server{
		Addr: address, Handler: handler,
		ReadTimeout:       time.Duration(s.Config.Server.ReadTimeout) * time.Second,
		ReadHeaderTimeout: time.Duration(s.Config.Server.ReadTimeout) * time.Second,
		WriteTimeout:      time.Duration(s.Config.Server.WriteTimeout) * time.Second,
		IdleTimeout:       time.Duration(s.Config.Server.IdleTimeout) * time.Second,
	}
}

func (s *Server) Start() error {
	if s.httpServer == nil {
		return errors.New("HTTP server not initialized")
	}
	return s.httpServer.ListenAndServe()
}

// Serve starts HTTP on a listener already registered by the owning role.
func (s *Server) Serve(listener net.Listener) error {
	if s.httpServer == nil {
		return errors.New("HTTP server not initialized")
	}
	return s.httpServer.Serve(listener)
}

// Shutdown drains HTTP only; the role closes shared resources afterwards.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	err := s.httpServer.Shutdown(ctx)
	if err != nil {
		err = errors.Join(err, s.httpServer.Close())
	}
	return err
}
