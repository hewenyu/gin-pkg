package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hewenyu/gin-pkg/config"
	"github.com/hewenyu/gin-pkg/internal/ent"
	"github.com/hewenyu/gin-pkg/internal/router"
	"github.com/hewenyu/gin-pkg/internal/service/auth"
	"github.com/hewenyu/gin-pkg/internal/service/factory"
	"github.com/hewenyu/gin-pkg/internal/service/user"
	"github.com/hewenyu/gin-pkg/pkg/auth/jwt"
	"github.com/hewenyu/gin-pkg/pkg/auth/security"
	"github.com/hewenyu/gin-pkg/pkg/logger"
	"github.com/hewenyu/gin-pkg/pkg/util"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// App represents the application
type App struct {
	config          *config.Config
	router          *gin.Engine
	dbClient        *ent.Client
	redisClient     *util.RedisClient
	serviceFactory  *factory.ServiceFactory
	tokenService    jwt.TokenService
	securityService security.SecurityService
	userService     user.UserService
	authService     auth.AuthService
	server          *http.Server
}

// NewApp creates a new application instance
func NewApp(configPath string) (*App, error) {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Set up Gin
	router := gin.Default()

	return &App{
		config: cfg,
		router: router,
	}, nil
}

// Initialize sets up the application components
func (a *App) Initialize() error {
	var err error

	// Initialize database connection
	a.dbClient, err = a.setupDatabase()
	if err != nil {
		return err
	}
	logger.Info("Database connection established")

	// Initialize Redis connection
	a.redisClient, err = a.setupRedis()
	if err != nil {
		return err
	}
	logger.Info("Redis connection established")

	// Create service factory
	a.serviceFactory = factory.NewServiceFactory(a.dbClient, a.redisClient)
	logger.Info("Service factory created")

	// Initialize services
	a.tokenService = a.serviceFactory.CreateTokenService(
		a.config.Auth.AccessTokenSecret,
		a.config.Auth.RefreshTokenSecret,
		a.config.Auth.AccessTokenDuration,
		a.config.Auth.RefreshTokenDuration,
		a.config.Auth.DefaultAccessTokenExp,
		a.config.Auth.DefaultRefreshTokenExp,
	)
	logger.Debug("Token service initialized")

	a.securityService = a.serviceFactory.CreateSecurityService(
		a.config.Security.SignatureSecret,
		a.config.Security.NonceValidityDuration,
	)
	logger.Debug("Security service initialized")

	a.userService = a.serviceFactory.CreateUserService(a.tokenService)
	a.authService = a.serviceFactory.CreateAuthService(a.userService, a.tokenService, a.securityService)
	logger.Debug("User and auth services initialized")

	// Set up routes
	router.Setup(
		a.router,
		a.userService,
		a.tokenService,
		a.securityService,
		a.config.Auth.EnableRegistration,
		a.config.Security.TimestampValidityWindow,
	)
	logger.Info("API routes configured")

	// Initialize HTTP server
	a.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", a.config.Server.Port),
		Handler:      a.router,
		ReadTimeout:  a.config.Server.ReadTimeout,
		WriteTimeout: a.config.Server.WriteTimeout,
	}
	logger.Info("HTTP server initialized")

	return nil
}

// Run starts the application
func (a *App) Run() error {
	// Start HTTP server in a goroutine
	go func() {
		logger.Infof("Server listening on port %d", a.config.Server.Port)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shut down server
	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	logger.Info("Server exiting")
	return nil
}

// Cleanup performs cleanup operations
func (a *App) Cleanup() {
	if a.dbClient != nil {
		a.dbClient.Close()
		logger.Debug("Database connection closed")
	}
	if a.redisClient != nil {
		a.redisClient.Close()
		logger.Debug("Redis connection closed")
	}
}

// setupDatabase initializes the database connection
func (a *App) setupDatabase() (*ent.Client, error) {
	dbURL := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		a.config.Database.Username,
		a.config.Database.Password,
		a.config.Database.Host,
		a.config.Database.Port,
		a.config.Database.Database,
		a.config.Database.SSLMode,
	)
	client, err := ent.Open(a.config.Database.Driver, dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run schema migrations
	if err := client.Schema.Create(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	return client, nil
}

// setupRedis initializes the Redis connection
func (a *App) setupRedis() (*util.RedisClient, error) {
	redis, err := util.NewRedisClient(
		a.config.Redis.Host,
		a.config.Redis.Port,
		a.config.Redis.Password,
		a.config.Redis.DB,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	return redis, nil
}
