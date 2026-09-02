package server

import (
	"context"
	"net/http"
	"time"

	"flux/apps/backend/internal/config"
	"flux/apps/backend/internal/database"
	"flux/apps/backend/internal/db"
	"flux/apps/backend/internal/errs"
	"flux/apps/backend/internal/handler"
	"flux/apps/backend/internal/logger"
	customMiddleware "flux/apps/backend/internal/middleware"
	"flux/apps/backend/internal/model/analytics"
	"flux/apps/backend/internal/repository"
	"flux/apps/backend/internal/router"
	"flux/apps/backend/internal/service"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// Server represents the API server instance.
type Server struct {
	Echo               *echo.Echo
	Config             *config.Config
	DBPool             *pgxpool.Pool
	AnalyticsPublisher *service.RedisAnalyticsPublisher
	ClickHouseConsumer *service.RedisAnalyticsConsumer
	ConversionConsumer *service.RedisConversionConsumer
	DomainWorker       *service.DomainVerificationWorker
	WebhookWorker      *service.WebhookWorker
	WebhookRetryWorker *service.WebhookRetryWorker
}

// NewServer initializes and wires all dependencies for the server.
func NewServer(cfg *config.Config) (*Server, error) {
	logger.InitLogger("debug", "console")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	dbPool, dbErr := database.InitDBPool(ctx, cfg.GetDatabaseURL())
	cancel()

	if dbErr != nil {
		log.Warn().Err(dbErr).Msg("postgresql connection ping failed via pgx/v5")
	} else {
		log.Info().Msg("successfully connected and pinged postgresql database via pgx/v5")
		migCtx, migCancel := context.WithTimeout(context.Background(), 10*time.Second)
		if migErr := database.MigrateDSN(migCtx, &log.Logger, cfg.GetDatabaseURL()); migErr != nil {
			log.Warn().Err(migErr).Msg("database migration warning")
		}
		migCancel()
	}

	if cfg.Clerk.SecretKey != "" {
		clerk.SetKey(cfg.Clerk.SecretKey)
	} else {
		log.Warn().Msg("CLERK_SECRET_KEY is empty, authentication may fail")
	}

	var pub *service.RedisAnalyticsPublisher
	var pubInterface analytics.AnalyticsPublisher
	var convPubInterface analytics.ConversionPublisher
	var chConsumer *service.RedisAnalyticsConsumer
	var convConsumer *service.RedisConversionConsumer
	var analyticsProvider repository.AnalyticsProvider
	var redirectCache repository.RedirectCache
	var webhookWorker *service.WebhookWorker
	var webhookRetryWorker *service.WebhookRetryWorker

	if cfg.GetRedisURL() != "" {
		redisClient := redis.NewClient(&redis.Options{Addr: cfg.GetRedisURL()})
		
		// Initialize the Redirect Cache
		redirectCache = repository.NewRedisRedirectCache(redisClient)

		pub = service.NewRedisAnalyticsPublisher(redisClient, cfg.Redis.AnalyticsStream, 5000)
		pub.Start()
		pubInterface = pub
		
		// Conversion publisher
		convPubInterface = service.NewRedisConversionPublisher(redisClient, "analytics:conversions")

		if cfg.GetClickHouseURL() != "" {
			chConn, err := db.InitClickHouse(cfg.GetClickHouseURL())
			if err == nil {
				_ = db.MigrateClickHouseSchema(context.Background(), chConn)
				analyticsProvider = repository.NewClickHouseAnalyticsRepository(chConn)
				
				chConsumer = service.NewRedisAnalyticsConsumer(redisClient, chConn, cfg.Redis.AnalyticsStream)
				chConsumer.Start()

				convConsumer = service.NewRedisConversionConsumer(redisClient, chConn, "analytics:conversions")
				convConsumer.Start()
			} else {
				log.Warn().Err(err).Msg("failed to connect to ClickHouse, consumer will not start")
			}
			
			// Webhook Delivery Worker
			deliveryTimeout, _ := time.ParseDuration(cfg.Webhook.DeliveryTimeout)
			if deliveryTimeout == 0 {
				deliveryTimeout = 10 * time.Second
			}
			webhookRepo := repository.NewWebhookRepository(dbPool)
			webhookWorker = service.NewWebhookWorker(redisClient, webhookRepo, cfg.Redis.AnalyticsStream, &cfg.Webhook)
			webhookWorker.Start()
			webhookRetryWorker = service.NewWebhookRetryWorker(webhookRepo, &cfg.Webhook)
			webhookRetryWorker.Start()
		}
	} else {
		pubInterface = service.NewLogAnalyticsPublisher()
	}

	redirectRepo := repository.NewPostgresRedirectRepository(dbPool)
	redirectSvc := service.NewRedirectService(redirectRepo, redirectCache)
	billingRepo := repository.NewBillingRepository(dbPool)

	redirectHandler := handler.NewRedirectHandler(redirectSvc, pubInterface)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsProvider, billingRepo)

	linkRepo := repository.NewLinkRepository(dbPool)
	campaignRepo := repository.NewCampaignRepository(dbPool)
	domainRepo := repository.NewDomainRepository(dbPool)
	
	linksHandler := handler.NewLinksHandler(service.NewLinkService(linkRepo, redirectCache, campaignRepo, billingRepo))
	
	campaignSvc := service.NewCampaignService(campaignRepo, linkRepo, redirectCache)
	campaignHandler := handler.NewCampaignHandler(campaignSvc)

	domainSvc := service.NewDomainService(domainRepo, redirectCache, cfg.Server.PlatformDomain)
	domainHandler := handler.NewDomainHandler(domainSvc)

	domainWorker := service.NewDomainVerificationWorker(domainRepo, redirectCache, nil, 0)
	if dbPool != nil {
		domainWorker.Start()
	}

	tlsAuthHandler := handler.NewTLSAuthHandler(domainRepo, cfg.Auth.InternalAPIKey)

	userRepo := repository.NewUserRepository(dbPool)
	
	var trackingHandler *handler.TrackingHandler
	var limiterStore customMiddleware.LimiterStore
	if convPubInterface != nil {
		trackingHandler = handler.NewTrackingHandler(userRepo, convPubInterface)
	}

	if cfg.GetRedisURL() != "" {
		redisClient := redis.NewClient(&redis.Options{Addr: cfg.GetRedisURL()})
		limiterStore = customMiddleware.NewRedisSlidingWindowLimiter(redisClient)
	}

	e := echo.New()
	e.IPExtractor = echo.ExtractIPFromRealIPHeader()
	e.HTTPErrorHandler = errs.CustomHTTPErrorHandler
	customMiddleware.RegisterGlobalMiddlewares(e)
	e.Use(customMiddleware.TracingMiddleware(cfg.Observability.NewRelic.LicenseKey, "flux-backend"))

	stripeWebhookHandler := handler.NewStripeWebhookHandler(dbPool, billingRepo, cfg)

	billingHandler := handler.NewBillingHandler(billingRepo, cfg)
	router.InitRouter(e, dbPool, userRepo, redirectHandler, analyticsHandler, linksHandler, campaignHandler, domainHandler, tlsAuthHandler, trackingHandler, limiterStore, billingHandler)
	router.RegisterWebhookRoutes(e, stripeWebhookHandler)

	return &Server{
		Echo:               e,
		Config:             cfg,
		DBPool:             dbPool,
		AnalyticsPublisher: pub,
		ClickHouseConsumer: chConsumer,
		ConversionConsumer: convConsumer,
		DomainWorker:       domainWorker,
		WebhookWorker:      webhookWorker,
		WebhookRetryWorker: webhookRetryWorker,
	}, nil
}

// Start launches the Echo HTTP listener.
func (s *Server) Start() error {
	log.Info().Msgf("starting flux Echo v4 api server on port %s...", s.Config.Server.Port)
	if err := s.Echo.Start(":" + s.Config.Server.Port); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Stop shuts down the server gracefully, draining any pending background tasks.
func (s *Server) Stop(ctx context.Context) error {
	log.Info().Msg("shutting down HTTP server to stop accepting new requests...")
	err := s.Echo.Shutdown(ctx)

	if s.DomainWorker != nil {
		log.Info().Msg("stopping domain verification worker gracefully...")
		workerCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		s.DomainWorker.Stop(workerCtx)
		cancel()
	}
	
	if s.WebhookRetryWorker != nil {
		log.Info().Msg("stopping webhook retry worker gracefully...")
		s.WebhookRetryWorker.Stop(5 * time.Second)
	}

	if s.WebhookWorker != nil {
		log.Info().Msg("stopping webhook worker gracefully...")
		s.WebhookWorker.Stop(5 * time.Second)
	}

	if s.AnalyticsPublisher != nil {
		log.Info().Msg("draining redis analytics publisher queue...")
		s.AnalyticsPublisher.Stop(5 * time.Second)
	}

	if s.ClickHouseConsumer != nil {
		log.Info().Msg("stopping clickhouse consumer gracefully...")
		s.ClickHouseConsumer.Stop(5 * time.Second)
	}
	
	if s.ConversionConsumer != nil {
		log.Info().Msg("stopping conversion consumer gracefully...")
		s.ConversionConsumer.Stop(5 * time.Second)
	}

	return err
}
