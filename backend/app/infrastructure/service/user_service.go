package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"cashone/domain/entity"
	"cashone/domain/repository"
	"cashone/domain/service"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserExists         = errors.New("user with this email already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidPassword    = errors.New("invalid password")
)

type userService struct {
	userRepo repository.UserRepository
	log      *zap.SugaredLogger
}

func NewUserService(userRepo repository.UserRepository, log *zap.SugaredLogger) service.UserService {
	return &userService{
		userRepo: userRepo,
		log:      log,
	}
}

func (s *userService) Register(ctx context.Context, email, password, name string) (*entity.User, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		s.log.Errorw("Failed to check existing user", "error", err, "email", email)
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, ErrUserExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Errorw("Failed to hash password", "error", err)
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &entity.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		Name:         name,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.log.Errorw("Failed to create user", "error", err, "email", email)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Clear password hash before returning
	user.PasswordHash = ""
	return user, nil
}

func (s *userService) Login(ctx context.Context, email, password string) (*entity.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		s.log.Errorw("Failed to get user by email", "error", err, "email", email)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, ErrInvalidCredentials
		}
		s.log.Errorw("Failed to compare password hash", "error", err)
		return nil, fmt.Errorf("failed to verify password: %w", err)
	}

	// Clear password hash before returning
	user.PasswordHash = ""
	return user, nil
}

func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Errorw("Failed to get user by ID", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Clear password hash before returning
	user.PasswordHash = ""
	return user, nil
}

func (s *userService) UpdateProfile(ctx context.Context, id uuid.UUID, name string) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Errorw("Failed to get user for update", "error", err, "id", id)
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	user.Name = name
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.log.Errorw("Failed to update user profile", "error", err, "id", id)
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (s *userService) ChangePassword(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Errorw("Failed to get user for password change", "error", err, "id", id)
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidPassword
		}
		s.log.Errorw("Failed to compare password hash", "error", err)
		return fmt.Errorf("failed to verify password: %w", err)
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		s.log.Errorw("Failed to hash new password", "error", err)
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.PasswordHash = string(hashedPassword)
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.log.Errorw("Failed to update user password", "error", err, "id", id)
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}
