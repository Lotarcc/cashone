package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Base struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type User struct {
	Base
	Email        string `gorm:"type:varchar(255);unique;not null" json:"email"`
	PasswordHash string `gorm:"type:varchar(255);not null" json:"-"`
	Name         string `gorm:"type:varchar(255)" json:"name"`
	Cards        []Card `gorm:"foreignKey:UserID" json:"cards,omitempty"`
}

type Card struct {
	Base
	UserID            uuid.UUID     `gorm:"type:uuid;not null" json:"user_id"`
	CardName          string        `gorm:"type:varchar(255);not null" json:"card_name"`
	MaskedPan         string        `gorm:"type:varchar(19)" json:"masked_pan,omitempty"`
	Balance           int64         `gorm:"not null;default:0" json:"balance"`
	CreditLimit       int64         `gorm:"default:0" json:"credit_limit"`
	CurrencyCode      int           `gorm:"not null" json:"currency_code"`
	IsManual          bool          `gorm:"not null;default:true" json:"is_manual"`
	Type              string        `gorm:"type:varchar(50)" json:"type,omitempty"`
	MonobankAccountID string        `gorm:"type:varchar(255)" json:"monobank_account_id,omitempty"`
	Transactions      []Transaction `gorm:"foreignKey:CardID" json:"transactions,omitempty"`
}

type Transaction struct {
	Base
	CardID                uuid.UUID  `gorm:"type:uuid;not null" json:"card_id"`
	UserID                uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	CategoryID            *uuid.UUID `gorm:"type:uuid" json:"category_id,omitempty"`
	Amount                int64      `gorm:"not null" json:"amount"`
	OperationAmount       int64      `json:"operation_amount,omitempty"`
	CurrencyCode          int        `gorm:"not null" json:"currency_code"`
	OperationCurrencyCode int        `json:"operation_currency_code,omitempty"`
	Type                  string     `gorm:"type:varchar(20);check:type IN ('expense', 'income', 'transfer')" json:"type"`
	Description           string     `gorm:"type:text" json:"description"`
	MCC                   int        `json:"mcc,omitempty"`
	CommissionRate        int64      `json:"commission_rate,omitempty"`
	CashbackAmount        int64      `json:"cashback_amount,omitempty"`
	BalanceAfter          int64      `json:"balance_after"`
	Hold                  bool       `gorm:"default:false" json:"hold"`
	TransactionDate       time.Time  `gorm:"not null" json:"transaction_date"`
	MonobankID            string     `gorm:"type:varchar(255)" json:"monobank_id,omitempty"`
	Comment               string     `gorm:"type:text" json:"comment,omitempty"`
}

type Category struct {
	Base
	Name     string     `gorm:"type:varchar(100);not null" json:"name"`
	ParentID *uuid.UUID `gorm:"type:uuid" json:"parent_id,omitempty"`
	UserID   uuid.UUID  `gorm:"type:uuid" json:"user_id"`
	Type     string     `gorm:"type:varchar(20);check:type IN ('expense', 'income', 'transfer')" json:"type"`
}

type MonobankIntegration struct {
	Base
	UserID      uuid.UUID `gorm:"type:uuid;unique;not null" json:"user_id"`
	ClientID    string    `gorm:"type:varchar(255);not null" json:"client_id"`
	Token       string    `gorm:"type:varchar(255);not null" json:"-"`
	WebhookURL  string    `gorm:"type:varchar(255)" json:"webhook_url,omitempty"`
	Permissions string    `gorm:"type:varchar(50)" json:"permissions,omitempty"`
}
