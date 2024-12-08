package entity

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// RegisterRequest represents the registration request data
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required"`
}

// RegisterResponse represents the registration response data
type RegisterResponse struct {
	User      *User      `json:"user"`
	AuthToken *AuthToken `json:"auth_token"`
}

// LoginRequest represents the login request data
type LoginRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required"`
	UserAgent string `json:"user_agent"`
	IP        string `json:"ip"`
}

// LoginResponse represents the login response data
type LoginResponse struct {
	User      *User      `json:"user"`
	AuthToken *AuthToken `json:"auth_token"`
}

// AuthToken represents an authentication token pair
type AuthToken struct {
	TokenType    string    `json:"token_type"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// RefreshToken represents a refresh token in the database
type RefreshToken struct {
	Base
	UserID    uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	Token     string     `gorm:"type:varchar(255);not null;unique" json:"token"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	RevokedAt *time.Time `gorm:"" json:"revoked_at"`
	UserAgent string     `gorm:"type:varchar(255)" json:"user_agent"`
	IP        string     `gorm:"type:varchar(45)" json:"ip"`
	Revoked   bool       `gorm:"not null;default:false" json:"revoked"`
}

// Claims represents the JWT claims
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}
