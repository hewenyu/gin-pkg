package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType represents the type of JWT token
type TokenType string

const (
	// AccessToken is used for API access
	AccessToken TokenType = "access"
	// RefreshToken is used to get a new access token
	RefreshToken TokenType = "refresh"
)

// Claims represents the JWT claims
type Claims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
	TokenID   string `json:"token_id"`
	jwt.RegisteredClaims
}

// TokenPair contains both access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// TokenService defines the interface for JWT token operations
type TokenService interface {
	GenerateTokenPair(userID string, email, role string) (*TokenPair, error)
	ValidateToken(tokenString string, tokenType TokenType) (*Claims, error)
	RefreshTokens(refreshToken string) (*TokenPair, error)
	BlacklistToken(tokenID string, expiration time.Duration) error
	IsTokenBlacklisted(tokenID string) (bool, error)
}
