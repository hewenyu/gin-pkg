package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hewenyu/gin-pkg/internal/ent"
	"github.com/hewenyu/gin-pkg/internal/ent/user"
	"github.com/hewenyu/gin-pkg/internal/model"
	"github.com/hewenyu/gin-pkg/pkg/auth/jwt"
	"golang.org/x/crypto/bcrypt"
)

// DefaultUserService implements UserService
type DBUserService struct {
	client       *ent.Client
	tokenService jwt.TokenService
}

// NewUserService creates a new user service
func NewUserService(client *ent.Client, tokenService jwt.TokenService) UserService {
	return &DBUserService{
		client:       client,
		tokenService: tokenService,
	}
}

// CreateUser creates a new user
func (s *DBUserService) CreateUser(ctx context.Context, input model.CreateUserInput) (*ent.User, error) {
	// Check if user with the same email already exists
	exists, err := s.client.User.Query().Where(user.Email(input.Email)).Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing user: %w", err)
	}
	if exists {
		return nil, errors.New("user with this email already exists")
	}

	// Check if user with the same username already exists
	exists, err = s.client.User.Query().Where(user.Username(input.Username)).Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing user: %w", err)
	}
	if exists {
		return nil, errors.New("user with this username already exists")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create the user
	newUser, err := s.client.User.Create().
		SetEmail(input.Email).
		SetUsername(input.Username).
		SetPasswordHash(string(hashedPassword)).
		SetRole(input.Role).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return newUser, nil
}

// GetUserByID gets a user by ID
func (s *DBUserService) GetUserByID(ctx context.Context, id string) (*ent.User, error) {
	user, err := s.client.User.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// GetUserByEmail gets a user by email
func (s *DBUserService) GetUserByEmail(ctx context.Context, email string) (*ent.User, error) {
	user, err := s.client.User.Query().Where(user.Email(email)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// UpdateUser updates a user
func (s *DBUserService) UpdateUser(ctx context.Context, id string, input model.UpdateUserInput) (*ent.User, error) {
	// Get the user
	userToUpdate, err := s.client.User.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Build the update query
	updateQuery := s.client.User.UpdateOne(userToUpdate)

	if input.Username != "" {
		// Check if username is already taken
		if input.Username != userToUpdate.Username {
			exists, err := s.client.User.Query().
				Where(user.Username(input.Username)).
				Exist(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to check for existing username: %w", err)
			}
			if exists {
				return nil, errors.New("username is already taken")
			}
		}
		updateQuery = updateQuery.SetUsername(input.Username)
	}

	if input.AvatarURL != nil {
		updateQuery = updateQuery.SetAvatarURL(*input.AvatarURL)
	}

	if input.Active != nil {
		updateQuery = updateQuery.SetActive(*input.Active)
	}

	if input.Role != "" {
		updateQuery = updateQuery.SetRole(input.Role)
	}

	// Execute the update
	updatedUser, err := updateQuery.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return updatedUser, nil
}

// DeleteUser deletes a user
func (s *DBUserService) DeleteUser(ctx context.Context, id string) error {
	err := s.client.User.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// Login authenticates a user and returns JWT tokens
func (s *DBUserService) Login(ctx context.Context, email, password string) (*jwt.TokenPair, *ent.User, error) {
	// Get the user by email
	user, err := s.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, nil, err
	}

	// Check if the user is active
	if !user.Active {
		return nil, nil, errors.New("account is deactivated")
	}

	// Verify the password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, nil, errors.New("invalid credentials")
	}

	// Generate JWT tokens
	tokenPair, err := s.tokenService.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Update last login time
	_, err = s.client.User.UpdateOne(user).
		SetLastLogin(time.Now()).
		Save(ctx)
	if err != nil {
		// Non-critical error, log but don't fail the login
		// In a real implementation, you'd want to log this error
		fmt.Printf("Failed to update last login time: %v\n", err)
	}

	return tokenPair, user, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *DBUserService) RefreshToken(ctx context.Context, refreshToken string) (*jwt.TokenPair, error) {
	return s.tokenService.RefreshTokens(refreshToken)
}

// UpdatePassword updates a user's password
func (s *DBUserService) UpdatePassword(ctx context.Context, userID string, currentPassword, newPassword string) error {
	// Get the user
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify the current password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword))
	if err != nil {
		return errors.New("invalid current password")
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update the password
	_, err = s.client.User.UpdateOne(user).
		SetPasswordHash(string(hashedPassword)).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}
