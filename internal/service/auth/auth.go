package auth

import (
	"context"

	"github.com/hewenyu/gin-pkg/internal/ent"
	"github.com/hewenyu/gin-pkg/internal/service/user"
	"github.com/hewenyu/gin-pkg/pkg/auth/jwt"
	"github.com/hewenyu/gin-pkg/pkg/auth/security"
)

// AuthService defines the interface for authentication and authorization operations
type AuthService interface {
	Login(ctx context.Context, email, password string) (*jwt.TokenPair, *ent.User, error)
	RefreshToken(ctx context.Context, refreshToken string) (*jwt.TokenPair, error)
	GetNonce(ctx context.Context) (string, error)
	ValidateTimestamp(timestamp string) error
	ValidateSignature(params map[string]string, signature string) error
	ValidateNonce(nonce string) error
}

// DefaultAuthService implements AuthService
type DefaultAuthService struct {
	userService     user.UserService
	tokenService    jwt.TokenService
	securityService security.SecurityService
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userService user.UserService,
	tokenService jwt.TokenService,
	securityService security.SecurityService,
) AuthService {
	return &DefaultAuthService{
		userService:     userService,
		tokenService:    tokenService,
		securityService: securityService,
	}
}

// Login authenticates a user and returns JWT tokens
func (s *DefaultAuthService) Login(ctx context.Context, email, password string) (*jwt.TokenPair, *ent.User, error) {
	return s.userService.Login(ctx, email, password)
}

// RefreshToken creates a new token pair using a refresh token
func (s *DefaultAuthService) RefreshToken(ctx context.Context, refreshToken string) (*jwt.TokenPair, error) {
	return s.userService.RefreshToken(ctx, refreshToken)
}

// GetNonce generates a new nonce for request signing
func (s *DefaultAuthService) GetNonce(ctx context.Context) (string, error) {
	return s.securityService.GenerateNonce()
}

// ValidateTimestamp checks if the timestamp is within the valid window
func (s *DefaultAuthService) ValidateTimestamp(timestamp string) error {
	return s.securityService.ValidateTimestamp(timestamp, 0) // 0 will use the default validity window
}

// ValidateSignature verifies that the signature matches the request parameters
func (s *DefaultAuthService) ValidateSignature(params map[string]string, signature string) error {
	return s.securityService.ValidateSignature(params, signature)
}

// ValidateNonce checks if the nonce is valid and hasn't been used before
func (s *DefaultAuthService) ValidateNonce(nonce string) error {
	return s.securityService.ValidateNonce(nonce)
}
