package v1

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hewenyu/gin-pkg/internal/model"
	"github.com/hewenyu/gin-pkg/internal/service/user"
)

type UserController struct {
	userService user.UserService
}

func NewUserController(userService user.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

// GetCurrentUser returns the currently authenticated user
func (c *UserController) GetCurrentUser(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	user, err := c.userService.GetUserByID(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Convert to response model
	userResponse := model.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		Role:      user.Role,
		Active:    user.Active,
		AvatarURL: &user.AvatarURL,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}

	ctx.JSON(http.StatusOK, userResponse)
}

// UpdateCurrentUser updates the current user's information
func (c *UserController) UpdateCurrentUser(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	var input model.UpdateUserInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := c.userService.UpdateUser(ctx, userID, input)
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
		AvatarURL: &user.AvatarURL,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}

	ctx.JSON(http.StatusOK, userResponse)
}

// ChangePassword changes the current user's password
func (c *UserController) ChangePassword(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	var input model.ChangePasswordInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.userService.UpdatePassword(ctx, userID, input.CurrentPassword, input.NewPassword); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "password updated successfully"})
}

// GetUser retrieves a user by ID (admin only)
func (c *UserController) GetUser(ctx *gin.Context) {
	userIDStr := ctx.Param("id")
	if userIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user ID is required"})
		return
	}

	user, err := c.userService.GetUserByID(ctx, userIDStr)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Convert to response model
	userResponse := model.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		Role:      user.Role,
		Active:    user.Active,
		AvatarURL: &user.AvatarURL,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}

	ctx.JSON(http.StatusOK, userResponse)
}

// UpdateUser updates a user (admin only)
func (c *UserController) UpdateUser(ctx *gin.Context) {
	userIDStr := ctx.Param("id")
	if userIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user ID is required"})
		return
	}

	var input model.UpdateUserInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := c.userService.UpdateUser(ctx, userIDStr, input)
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
		AvatarURL: &user.AvatarURL,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}

	ctx.JSON(http.StatusOK, userResponse)
}

// DeleteUser deletes a user (admin only)
func (c *UserController) DeleteUser(ctx *gin.Context) {
	userIDStr := ctx.Param("id")
	if userIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user ID is required"})
		return
	}

	if err := c.userService.DeleteUser(ctx, userIDStr); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

// RegisterRoutes registers the user routes
func (c *UserController) RegisterRoutes(router *gin.RouterGroup, authMiddleware, adminMiddleware gin.HandlerFunc) {
	// Routes for authenticated users
	userRoutes := router.Group("/users")
	userRoutes.Use(authMiddleware)
	{
		userRoutes.GET("/me", c.GetCurrentUser)
		userRoutes.PUT("/me", c.UpdateCurrentUser)
		userRoutes.POST("/change-password", c.ChangePassword)
	}

	// Routes for admin users
	adminRoutes := router.Group("/admin/users")
	adminRoutes.Use(authMiddleware, adminMiddleware)
	{
		adminRoutes.GET("/:id", c.GetUser)
		adminRoutes.PUT("/:id", c.UpdateUser)
		adminRoutes.DELETE("/:id", c.DeleteUser)
	}
}
