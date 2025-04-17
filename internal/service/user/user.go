package user

import (
	"context"

	"github.com/hewenyu/gin-pkg/internal/ent"
	"github.com/hewenyu/gin-pkg/internal/model"
	"github.com/hewenyu/gin-pkg/pkg/auth/jwt"
)

// UserService defines the interface for user operations
type UserService interface {
	CreateUser(ctx context.Context, input model.CreateUserInput) (*ent.User, error)
	GetUserByID(ctx context.Context, id string) (*ent.User, error)
	GetUserByEmail(ctx context.Context, email string) (*ent.User, error)
	UpdateUser(ctx context.Context, id string, input model.UpdateUserInput) (*ent.User, error)
	DeleteUser(ctx context.Context, id string) error
	Login(ctx context.Context, email, password string) (*jwt.TokenPair, *ent.User, error)
	RefreshToken(ctx context.Context, refreshToken string) (*jwt.TokenPair, error)
	UpdatePassword(ctx context.Context, userID string, currentPassword, newPassword string) error
}
