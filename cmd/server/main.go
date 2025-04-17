package main

// import (
// 	"context"
// 	"flag"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"os"
// 	"os/signal"
// 	"syscall"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"github.com/hewenyu/gin-pkg/config"
// 	v1 "github.com/hewenyu/gin-pkg/internal/api/v1"
// 	"github.com/hewenyu/gin-pkg/internal/ent"
// 	"github.com/hewenyu/gin-pkg/internal/service"
// 	"github.com/hewenyu/gin-pkg/pkg/auth"
// 	"github.com/hewenyu/gin-pkg/pkg/middleware"
// 	"github.com/hewenyu/gin-pkg/pkg/util"

// 	_ "github.com/lib/pq" // PostgreSQL driver
// )

// func main() {
// 	// Parse command line flags
// 	configPath := flag.String("config", "config/default.yaml", "path to configuration file")
// 	flag.Parse()

// 	// Load configuration
// 	cfg, err := config.Load(*configPath)
// 	if err != nil {
// 		log.Fatalf("Failed to load configuration: %v", err)
// 	}

// 	// Set up Gin
// 	r := gin.Default()

// 	// Connect to database
// 	dbURL := fmt.Sprintf(
// 		"postgresql://%s:%s@%s:%d/%s?sslmode=%s",
// 		cfg.Database.Username,
// 		cfg.Database.Password,
// 		cfg.Database.Host,
// 		cfg.Database.Port,
// 		cfg.Database.Database,
// 		cfg.Database.SSLMode,
// 	)
// 	client, err := ent.Open(cfg.Database.Driver, dbURL)
// 	if err != nil {
// 		log.Fatalf("Failed to connect to database: %v", err)
// 	}
// 	defer client.Close()

// 	// Run schema migrations
// 	if err := client.Schema.Create(context.Background()); err != nil {
// 		log.Fatalf("Failed to run database migrations: %v", err)
// 	}

// 	// Set up Redis client
// 	redis, err := util.NewRedisClient(
// 		cfg.Redis.Host,
// 		cfg.Redis.Port,
// 		cfg.Redis.Password,
// 		cfg.Redis.DB,
// 	)
// 	if err != nil {
// 		log.Fatalf("Failed to connect to Redis: %v", err)
// 	}
// 	defer redis.Close()

// 	// Initialize services
// 	tokenService := auth.NewJWTService(
// 		cfg.Auth.AccessTokenSecret,
// 		cfg.Auth.RefreshTokenSecret,
// 		cfg.Auth.AccessTokenDuration,
// 		cfg.Auth.RefreshTokenDuration,
// 		cfg.Auth.DefaultAccessTokenExp,
// 		cfg.Auth.DefaultRefreshTokenExp,
// 		redis.BlacklistToken,
// 		redis.IsTokenBlacklisted,
// 	)

// 	securityService := auth.NewSecurityService(
// 		cfg.Security.SignatureSecret,
// 		cfg.Security.NonceValidityDuration,
// 		redis.StoreNonce,
// 		redis.GetNonce,
// 		redis.InvalidateNonce,
// 	)

// 	userService := service.NewUserService(client, tokenService)

// 	// Initialize middleware
// 	authMiddleware := middleware.AuthMiddleware(tokenService)
// 	securityMiddleware := middleware.SecurityMiddleware(securityService, cfg.Security.TimestampValidityWindow)
// 	adminMiddleware := middleware.RoleMiddleware("admin")

// 	// Set up API v1 routes
// 	apiV1 := r.Group("/api/v1")
// 	apiV1.Use(securityMiddleware)

// 	// Initialize controllers
// 	authController := v1.NewAuthController(userService, securityService, cfg.Auth.EnableRegistration)
// 	userController := v1.NewUserController(userService)

// 	// Register routes
// 	authController.RegisterRoutes(apiV1)
// 	userController.RegisterRoutes(apiV1, authMiddleware, adminMiddleware)

// 	// Start server
// 	srv := &http.Server{
// 		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
// 		Handler:      r,
// 		ReadTimeout:  cfg.Server.ReadTimeout,
// 		WriteTimeout: cfg.Server.WriteTimeout,
// 	}

// 	// Run server in a goroutine so that it doesn't block
// 	go func() {
// 		log.Printf("Server listening on port %d", cfg.Server.Port)
// 		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
// 			log.Fatalf("Failed to start server: %v", err)
// 		}
// 	}()

// 	// Wait for interrupt signal to gracefully shut down the server
// 	quit := make(chan os.Signal, 1)
// 	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
// 	<-quit
// 	log.Println("Shutting down server...")

// 	// Create a deadline to wait for
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	// Shut down server
// 	if err := srv.Shutdown(ctx); err != nil {
// 		log.Fatalf("Server forced to shutdown: %v", err)
// 	}

// 	log.Println("Server exiting")
// }
