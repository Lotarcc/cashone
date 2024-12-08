package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"cashone/domain/entity"
	"cashone/domain/errors"
	"cashone/domain/repository"
	"cashone/domain/service"
)

type userService struct {
	userRepo repository.UserRepository
	log      *zap.SugaredLogger
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository, log *zap.SugaredLogger) service.UserService {
	return &userService{
		userRepo: userRepo,
		log:      log,
	}
}

func (s *userService) Create(ctx context.Context, user *entity.User) error {
	// Validate user data
	if err := s.validateUser(user); err != nil {
		return fmt.Errorf("%w: %v", errors.ErrInvalidUserData, err)
	}

	// Check if user with this email already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, user.Email)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}
	if existingUser != nil {
		return errors.ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("%w: failed to hash password", errors.ErrInternal)
	}
	user.PasswordHash = string(hashedPassword)

	// Generate UUID if not provided
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	// Create user
	if err := s.userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}

	s.log.Infow("User created successfully", "id", user.ID, "email", user.Email)
	return nil
}

func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}
	if user == nil {
		return nil, errors.ErrUserNotFound
	}
	return user, nil
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	if email == "" {
		return nil, fmt.Errorf("%w: email is required", errors.ErrValidation)
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}
	if user == nil {
		return nil, errors.ErrUserNotFound
	}
	return user, nil
}

func (s *userService) Update(ctx context.Context, user *entity.User) error {
	// Validate user data
	if err := s.validateUser(user); err != nil {
		return fmt.Errorf("%w: %v", errors.ErrInvalidUserData, err)
	}

	// Check if user exists
	existingUser, err := s.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}
	if existingUser == nil {
		return errors.ErrUserNotFound
	}

	// If password is being updated, hash it
	if user.PasswordHash != existingUser.PasswordHash {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("%w: failed to hash password", errors.ErrInternal)
		}
		user.PasswordHash = string(hashedPassword)
	}

	// Update user
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}

	s.log.Infow("User updated successfully", "id", user.ID, "email", user.Email)
	return nil
}

func (s *userService) Delete(ctx context.Context, id uuid.UUID) error {
	// Check if user exists
	existingUser, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}
	if existingUser == nil {
		return errors.ErrUserNotFound
	}

	// Delete user
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}

	s.log.Infow("User deleted successfully", "id", id)
	return nil
}

func (s *userService) validateUser(user *entity.User) error {
	if user == nil {
		return errors.ErrInvalidUserData
	}

	var validationErrors []string

	if user.Email == "" {
		validationErrors = append(validationErrors, "email is required")
	}
	if user.PasswordHash == "" {
		validationErrors = append(validationErrors, "password is required")
	}
	if user.Name == "" {
		validationErrors = append(validationErrors, "name is required")
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("%w: %v", errors.ErrValidation, validationErrors)
	}

	return nil
}
