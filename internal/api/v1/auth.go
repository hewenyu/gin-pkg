package v1

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hewenyu/gin-pkg/internal/model"
	"github.com/hewenyu/gin-pkg/internal/service"
	"github.com/hewenyu/gin-pkg/pkg/auth"
)

type AuthController struct {
	userService        service.UserService
	securityService    auth.SecurityService
	enableRegistration bool
}

func NewAuthController(userService service.UserService, securityService auth.SecurityService, enableRegistration bool) *AuthController {
	return &AuthController{
		userService:        userService,
		securityService:    securityService,
		enableRegistration: enableRegistration,
	}
}

// Register handles user registration
func (c *AuthController) Register(ctx *gin.Context) {
	if !c.enableRegistration {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "registration is disabled"})
		return
	}

	var input model.CreateUserInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default role if not provided
	if input.Role == "" {
		input.Role = "user"
	}

	user, err := c.userService.CreateUser(ctx, input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to response model
	userResponse := model.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		Role:      user.Role,
		Active:    user.Active,
		AvatarURL: user.AvatarURL,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}

	ctx.JSON(http.StatusCreated, userResponse)
}

// Login handles user authentication and returns JWT tokens
func (c *AuthController) Login(ctx *gin.Context) {
	var input model.LoginInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, user, err := c.userService.Login(ctx, input.Email, input.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Convert to response model
	userResponse := model.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		Role:      user.Role,
		Active:    user.Active,
		AvatarURL: user.AvatarURL,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}

	authResponse := model.AuthResponse{
		User:         userResponse,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	}

	ctx.JSON(http.StatusOK, authResponse)
}

// RefreshToken handles token refresh
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	var input model.RefreshTokenInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := c.userService.RefreshToken(ctx, input.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
	})
}

// GetNonce generates and returns a new nonce for request signing
func (c *AuthController) GetNonce(ctx *gin.Context) {
	nonce, err := c.securityService.GenerateNonce()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate nonce"})
		return
	}

	ctx.JSON(http.StatusOK, model.NonceResponse{Nonce: nonce})
}

// RegisterRoutes registers the auth routes
func (c *AuthController) RegisterRoutes(router *gin.RouterGroup) {
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/register", c.Register)
		authRoutes.POST("/login", c.Login)
		authRoutes.POST("/refresh", c.RefreshToken)
		authRoutes.GET("/nonce", c.GetNonce)
	}
}
