package entity

import (
	"time"

	"github.com/google/uuid"
)

// Base contains common fields for all entities
type Base struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
}

// User represents a user in the system
type User struct {
	Base
	Email         string     `gorm:"type:varchar(255);not null;unique" json:"email"`
	Name          string     `gorm:"type:varchar(255);not null" json:"name"`
	PasswordHash  string     `gorm:"type:varchar(255);not null" json:"-"`
	EmailVerified bool       `gorm:"not null;default:false" json:"email_verified"`
	LastLoginAt   *time.Time `json:"last_login_at"`
}

// Card represents a bank card
type Card struct {
	Base
	UserID            uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Name              string    `gorm:"type:varchar(255);not null" json:"name"`
	CardName          string    `gorm:"type:varchar(255)" json:"card_name"`
	MaskedPan         string    `gorm:"type:varchar(255)" json:"masked_pan"`
	MonobankID        string    `gorm:"type:varchar(255);unique" json:"monobank_id"`
	MonobankAccountID string    `gorm:"type:varchar(255)" json:"monobank_account_id"`
	Balance           int64     `gorm:"not null" json:"balance"`
	CreditLimit       int64     `gorm:"not null;default:0" json:"credit_limit"`
	CurrencyCode      int       `gorm:"not null" json:"currency_code"`
	Type              string    `gorm:"type:varchar(50)" json:"type"`
	IsManual          bool      `gorm:"not null;default:false" json:"is_manual"`
}

// Category represents a transaction category
type Category struct {
	Base
	UserID   uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	ParentID *uuid.UUID `gorm:"type:uuid" json:"parent_id"`
	Name     string     `gorm:"type:varchar(255);not null" json:"name"`
	Type     string     `gorm:"type:varchar(50);not null" json:"type"`
}

// CategoryTree represents a category with its children
type CategoryTree struct {
	Category
	Children []CategoryTree `json:"children"`
}

// Transaction represents a financial transaction
type Transaction struct {
	Base
	UserID          uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	CardID          uuid.UUID  `gorm:"type:uuid;not null" json:"card_id"`
	CategoryID      *uuid.UUID `gorm:"type:uuid" json:"category_id"`
	Amount          int64      `gorm:"not null" json:"amount"`
	OperationAmount int64      `gorm:"not null" json:"operation_amount"`
	CurrencyCode    int        `gorm:"not null" json:"currency_code"`
	Type            string     `gorm:"type:varchar(50);not null" json:"type"`
	Description     string     `gorm:"type:varchar(255)" json:"description"`
	Comment         string     `gorm:"type:varchar(255)" json:"comment"`
	TransactionDate time.Time  `gorm:"not null" json:"transaction_date"`
	MonobankID      string     `gorm:"type:varchar(255);unique" json:"monobank_id"`
	MCC             int        `gorm:"not null;default:0" json:"mcc"`
	CommissionRate  int64      `gorm:"not null;default:0" json:"commission_rate"`
	CashbackAmount  int64      `gorm:"not null;default:0" json:"cashback_amount"`
	BalanceAfter    int64      `gorm:"not null" json:"balance_after"`
	Hold            bool       `gorm:"not null;default:false" json:"hold"`
}

// TransactionSearchParams represents search parameters for transactions
type TransactionSearchParams struct {
	Query      string     `json:"query"`
	Type       string     `json:"type"`
	CategoryID *uuid.UUID `json:"category_id"`
	CardID     *uuid.UUID `json:"card_id"`
	FromDate   *time.Time `json:"from_date"`
	ToDate     *time.Time `json:"to_date"`
	MinAmount  *int64     `json:"min_amount"`
	MaxAmount  *int64     `json:"max_amount"`
}

// MonobankIntegration represents a user's Monobank integration
type MonobankIntegration struct {
	Base
	UserID      uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Token       string    `gorm:"type:varchar(255);not null" json:"token"`
	ClientID    string    `gorm:"type:varchar(255)" json:"client_id"`
	WebhookURL  string    `gorm:"type:varchar(255)" json:"webhook_url"`
	Permissions string    `gorm:"type:text" json:"permissions"`
	Active      bool      `gorm:"not null;default:true" json:"active"`
	LastSync    time.Time `gorm:"not null" json:"last_sync"`
	SyncError   *string   `gorm:"type:text" json:"sync_error"`
}
