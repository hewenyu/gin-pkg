package router

import (
	"time"

	"github.com/gin-gonic/gin"
	v1 "github.com/hewenyu/gin-pkg/internal/api/v1"
	"github.com/hewenyu/gin-pkg/internal/service/user"
	"github.com/hewenyu/gin-pkg/pkg/auth/jwt"
	"github.com/hewenyu/gin-pkg/pkg/auth/security"
	"github.com/hewenyu/gin-pkg/pkg/middleware"
)

// Setup configures the API routes
func Setup(
	router *gin.Engine,
	userService user.UserService,
	tokenService jwt.TokenService,
	securityService security.SecurityService,
	enableRegistration bool,
	timestampValidityWindow time.Duration,
) {
	// Set up middleware
	authMiddleware := middleware.AuthMiddleware(tokenService)
	securityMiddleware := middleware.SecurityMiddleware(securityService, timestampValidityWindow)
	adminMiddleware := middleware.RoleMiddleware("admin")

	// Set up API v1 routes
	apiV1 := router.Group("/api/v1")
	apiV1.Use(securityMiddleware)

	// Initialize controllers
	authController := v1.NewAuthController(userService, securityService, enableRegistration)
	userController := v1.NewUserController(userService)

	// Register routes
	authController.RegisterRoutes(apiV1)
	userController.RegisterRoutes(apiV1, authMiddleware, adminMiddleware)
}
