package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"cashone/domain/entity"
	"cashone/domain/errors"
	"cashone/domain/repository"
	"cashone/pkg/config"
)

// AuthService handles authentication-related business logic
type AuthService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	config           *config.Config
	log              *zap.SugaredLogger
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	config *config.Config,
	log *zap.SugaredLogger,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		config:           config,
		log:              log,
	}
}

// Register creates a new user account and generates authentication tokens
func (s *AuthService) Register(ctx context.Context, req *entity.RegisterRequest) (*entity.RegisterResponse, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, errors.ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := s.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &entity.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Name:         req.Name,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	authToken, err := s.GenerateTokens(ctx, user, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &entity.RegisterResponse{
		User:      user,
		AuthToken: authToken,
	}, nil
}

// Login authenticates a user and generates new authentication tokens
func (s *AuthService) Login(ctx context.Context, req *entity.LoginRequest) (*entity.LoginResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.ErrInvalidCredentials
	}

	// Verify password
	if err := s.VerifyPassword(req.Password, user.PasswordHash); err != nil {
		return nil, errors.ErrInvalidCredentials
	}

	// Generate tokens
	authToken, err := s.GenerateTokens(ctx, user, req.UserAgent, req.IP)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &entity.LoginResponse{
		User:      user,
		AuthToken: authToken,
	}, nil
}

// RefreshToken generates new authentication tokens using a valid refresh token
func (s *AuthService) RefreshToken(ctx context.Context, token string) (*entity.AuthToken, error) {
	// Get refresh token from database
	refreshToken, err := s.refreshTokenRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}
	if refreshToken == nil {
		return nil, errors.ErrInvalidToken
	}

	// Check if token is expired or revoked
	if refreshToken.ExpiresAt.Before(time.Now()) {
		return nil, errors.ErrTokenExpired
	}
	if refreshToken.RevokedAt != nil {
		return nil, errors.ErrInvalidToken
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, refreshToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.ErrInvalidToken
	}

	// Generate new tokens
	authToken, err := s.GenerateTokens(ctx, user, refreshToken.UserAgent, refreshToken.IP)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Revoke old refresh token
	if err := s.refreshTokenRepo.Revoke(ctx, token); err != nil {
		s.log.Errorw("Failed to revoke old refresh token", "error", err)
	}

	return authToken, nil
}

// Logout revokes the specified refresh token
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID, token string) error {
	err := s.refreshTokenRepo.Revoke(ctx, token)
	if err != nil {
		return errors.ErrDatabaseOperation
	}
	return nil
}

// ValidateToken validates and parses a JWT token, returning the claims if valid
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*entity.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &entity.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.Security.JWT.Secret), nil
	})

	if err != nil {
		return nil, errors.ErrInvalidToken
	}

	if claims, ok := token.Claims.(*entity.Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.ErrInvalidToken
}

// HashPassword generates a bcrypt hash of the provided password
func (s *AuthService) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// VerifyPassword checks if the provided password matches the hash
func (s *AuthService) VerifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// GenerateTokens generates new access and refresh tokens for a user
func (s *AuthService) GenerateTokens(ctx context.Context, user *entity.User, userAgent, ip string) (*entity.AuthToken, error) {
	// Generate access token
	now := time.Now()
	accessExp := now.Add(s.config.Security.JWT.AccessTokenExpiration)

	claims := &entity.Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExp),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.config.Security.JWT.Issuer,
			Subject:   user.ID.String(),
			Audience:  jwt.ClaimStrings{s.config.Security.JWT.Audience},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(s.config.Security.JWT.Secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token
	refreshToken := &entity.RefreshToken{
		UserID:    user.ID,
		Token:     uuid.New().String(),
		ExpiresAt: now.Add(s.config.Security.JWT.RefreshTokenExpiration),
		UserAgent: userAgent,
		IP:        ip,
	}

	if err := s.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return &entity.AuthToken{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.Token,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.config.Security.JWT.AccessTokenExpiration.Seconds()),
		ExpiresAt:    accessExp,
	}, nil
}

// RevokeAllUserTokens revokes all refresh tokens for a user
func (s *AuthService) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	return s.refreshTokenRepo.RevokeAllUserTokens(ctx, userID)
}

// GetActiveTokens returns all active refresh tokens for a user
func (s *AuthService) GetActiveTokens(ctx context.Context, userID uuid.UUID) ([]entity.RefreshToken, error) {
	return s.refreshTokenRepo.GetActiveByUserID(ctx, userID)
}
