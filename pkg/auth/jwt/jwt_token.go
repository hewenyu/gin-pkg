package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTService implements TokenService
type JWTService struct {
	accessSecret           string
	refreshSecret          string
	accessTokenDuration    time.Duration
	refreshTokenDuration   time.Duration
	defaultAccessTokenExp  int64
	defaultRefreshTokenExp int64
	blacklistToken         func(tokenID string, expiration time.Duration) error
	isTokenBlacklisted     func(tokenID string) (bool, error)
}

// NewJWTService creates a new JWT service
func NewJWTService(
	accessSecret string,
	refreshSecret string,
	accessTokenDuration time.Duration,
	refreshTokenDuration time.Duration,
	defaultAccessTokenExp int64,
	defaultRefreshTokenExp int64,
	blacklistToken func(tokenID string, expiration time.Duration) error,
	isTokenBlacklisted func(tokenID string) (bool, error),
) TokenService {
	return &JWTService{
		accessSecret:           accessSecret,
		refreshSecret:          refreshSecret,
		accessTokenDuration:    accessTokenDuration,
		refreshTokenDuration:   refreshTokenDuration,
		defaultAccessTokenExp:  defaultAccessTokenExp,
		defaultRefreshTokenExp: defaultRefreshTokenExp,
		blacklistToken:         blacklistToken,
		isTokenBlacklisted:     isTokenBlacklisted,
	}
}

// GenerateTokenPair creates a new pair of access and refresh tokens
func (s *JWTService) GenerateTokenPair(userID string, email, role string) (*TokenPair, error) {
	// Generate access token
	accessTokenID := uuid.New().String()
	accessTokenExpiration := time.Now().Add(s.accessTokenDuration)
	accessClaims := Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		TokenType: string(AccessToken),
		TokenID:   accessTokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessTokenExpiration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "gin-pkg",
			Subject:   userID,
			ID:        accessTokenID,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.accessSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token
	refreshTokenID := uuid.New().String()
	refreshTokenExpiration := time.Now().Add(s.refreshTokenDuration)
	refreshClaims := Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		TokenType: string(RefreshToken),
		TokenID:   refreshTokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshTokenExpiration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "gin-pkg",
			Subject:   userID,
			ID:        refreshTokenID,
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.refreshSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    s.defaultAccessTokenExp,
	}, nil
}

// ValidateToken validates a JWT token
func (s *JWTService) ValidateToken(tokenString string, tokenType TokenType) (*Claims, error) {
	var secret string
	switch tokenType {
	case AccessToken:
		secret = s.accessSecret
	case RefreshToken:
		secret = s.refreshSecret
	default:
		return nil, errors.New("invalid token type")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Check if the token is of the correct type
	if TokenType(claims.TokenType) != tokenType {
		return nil, errors.New("token type mismatch")
	}

	// Check if the token is blacklisted
	isBlacklisted, err := s.isTokenBlacklisted(claims.TokenID)
	if err != nil {
		return nil, fmt.Errorf("failed to check token blacklist: %w", err)
	}
	if isBlacklisted {
		return nil, errors.New("token has been revoked")
	}

	return claims, nil
}

// RefreshTokens generates a new token pair using a valid refresh token
func (s *JWTService) RefreshTokens(refreshToken string) (*TokenPair, error) {
	claims, err := s.ValidateToken(refreshToken, RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Blacklist the used refresh token
	expiry := time.Until(claims.ExpiresAt.Time)
	if err := s.BlacklistToken(claims.TokenID, expiry); err != nil {
		return nil, fmt.Errorf("failed to blacklist refresh token: %w", err)
	}

	// Generate new token pair
	return s.GenerateTokenPair(claims.UserID, claims.Email, claims.Role)
}

// BlacklistToken adds a token to the blacklist
func (s *JWTService) BlacklistToken(tokenID string, expiration time.Duration) error {
	return s.blacklistToken(tokenID, expiration)
}

// IsTokenBlacklisted checks if a token is blacklisted
func (s *JWTService) IsTokenBlacklisted(tokenID string) (bool, error) {
	return s.isTokenBlacklisted(tokenID)
}
