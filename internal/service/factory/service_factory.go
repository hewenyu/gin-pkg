package factory

import (
	"time"

	"github.com/hewenyu/gin-pkg/internal/ent"
	"github.com/hewenyu/gin-pkg/internal/service/auth"
	"github.com/hewenyu/gin-pkg/internal/service/user"
	"github.com/hewenyu/gin-pkg/pkg/auth/jwt"
	"github.com/hewenyu/gin-pkg/pkg/auth/security"
	"github.com/hewenyu/gin-pkg/pkg/util"
)

// ServiceFactory provides methods to create service instances
type ServiceFactory struct {
	dbClient    *ent.Client
	redisClient *util.RedisClient
}

// NewServiceFactory creates a new service factory
func NewServiceFactory(dbClient *ent.Client, redisClient *util.RedisClient) *ServiceFactory {
	return &ServiceFactory{
		dbClient:    dbClient,
		redisClient: redisClient,
	}
}

// CreateTokenService creates a new JWT token service
func (f *ServiceFactory) CreateTokenService(
	accessSecret string,
	refreshSecret string,
	accessTokenDuration time.Duration,
	refreshTokenDuration time.Duration,
	defaultAccessTokenExp int64,
	defaultRefreshTokenExp int64,
) jwt.TokenService {
	return jwt.NewJWTService(
		accessSecret,
		refreshSecret,
		accessTokenDuration,
		refreshTokenDuration,
		defaultAccessTokenExp,
		defaultRefreshTokenExp,
		f.redisClient.BlacklistToken,
		f.redisClient.IsTokenBlacklisted,
	)
}

// CreateSecurityService creates a new security service
func (f *ServiceFactory) CreateSecurityService(
	signatureSecret string,
	nonceValidityDuration time.Duration,
) security.SecurityService {
	return security.NewSecurityService(
		signatureSecret,
		nonceValidityDuration,
		f.redisClient.StoreNonce,
		f.redisClient.GetNonce,
		f.redisClient.InvalidateNonce,
	)
}

// CreateUserService creates a new user service
func (f *ServiceFactory) CreateUserService(tokenService jwt.TokenService) user.UserService {
	return user.NewUserService(f.dbClient, tokenService)
}

// CreateAuthService creates a new authentication service
func (f *ServiceFactory) CreateAuthService(
	userService user.UserService,
	tokenService jwt.TokenService,
	securityService security.SecurityService,
) auth.AuthService {
	return auth.NewAuthService(userService, tokenService, securityService)
}
