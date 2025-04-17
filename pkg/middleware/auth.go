package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hewenyu/gin-pkg/pkg/auth"
)

// AuthMiddleware is middleware that validates JWT tokens
func AuthMiddleware(tokenService auth.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		// Check if the header starts with "Bearer "
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		// Extract the token
		tokenString := parts[1]

		// Validate the token
		claims, err := tokenService.ValidateToken(tokenString, auth.AccessToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid access token"})
			c.Abort()
			return
		}

		// Store the claims in the context for later use
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("tokenID", claims.TokenID)

		c.Next()
	}
}

// OptionalAuthMiddleware is middleware that validates JWT tokens if present
func OptionalAuthMiddleware(tokenService auth.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Check if the header starts with "Bearer "
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.Next()
			return
		}

		// Extract the token
		tokenString := parts[1]

		// Validate the token
		claims, err := tokenService.ValidateToken(tokenString, auth.AccessToken)
		if err != nil {
			c.Next()
			return
		}

		// Store the claims in the context for later use
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("tokenID", claims.TokenID)
		c.Set("authenticated", true)

		c.Next()
	}
}

// RoleMiddleware is middleware that checks if a user has the required role
func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			c.Abort()
			return
		}

		if role != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}
