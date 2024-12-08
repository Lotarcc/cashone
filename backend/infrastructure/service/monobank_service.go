package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"cashone/domain/entity"
	"cashone/domain/errors"
	"cashone/domain/repository"
	"cashone/domain/service"
)

// MonobankService implements the service.MonobankService interface
type MonobankService struct {
	monoRepo   repository.MonobankIntegrationRepository
	cardRepo   repository.CardRepository
	txRepo     repository.TransactionRepository
	userRepo   repository.UserRepository
	httpClient interface {
		Do(*http.Request) (*http.Response, error)
	}
	log *zap.SugaredLogger
}

type monobankClientInfo struct {
	ClientID    string            `json:"clientId"`
	Name        string            `json:"name"`
	WebHookURL  string            `json:"webHookUrl"`
	Permissions string            `json:"permissions"`
	Accounts    []monobankAccount `json:"accounts"`
}

type monobankAccount struct {
	ID           string   `json:"id"`
	SendID       string   `json:"sendId"`
	Balance      int64    `json:"balance"`
	CreditLimit  int64    `json:"creditLimit"`
	Type         string   `json:"type"`
	CurrencyCode int      `json:"currencyCode"`
	CashbackType string   `json:"cashbackType"`
	MaskedPan    []string `json:"maskedPan"`
	IBAN         string   `json:"iban"`
}

type monobankTransaction struct {
	ID              string `json:"id"`
	Time            int64  `json:"time"`
	Description     string `json:"description"`
	MCC             int    `json:"mcc"`
	OriginalMCC     int    `json:"originalMcc"`
	Hold            bool   `json:"hold"`
	Amount          int64  `json:"amount"`
	OperationAmount int64  `json:"operationAmount"`
	CurrencyCode    int    `json:"currencyCode"`
	CommissionRate  int64  `json:"commissionRate"`
	CashbackAmount  int64  `json:"cashbackAmount"`
	Balance         int64  `json:"balance"`
	Comment         string `json:"comment,omitempty"`
	ReceiptID       string `json:"receiptId,omitempty"`
	CounterEdrpou   string `json:"counterEdrpou,omitempty"`
	CounterIban     string `json:"counterIban,omitempty"`
	CounterName     string `json:"counterName,omitempty"`
}

// NewMonobankService creates a new Monobank service instance with the provided repositories and logger
func NewMonobankService(
	monoRepo repository.MonobankIntegrationRepository,
	cardRepo repository.CardRepository,
	txRepo repository.TransactionRepository,
	userRepo repository.UserRepository,
	log *zap.SugaredLogger,
) service.MonobankService {
	return &MonobankService{
		monoRepo:   monoRepo,
		cardRepo:   cardRepo,
		txRepo:     txRepo,
		userRepo:   userRepo,
		httpClient: &http.Client{Timeout: time.Duration(viper.GetInt("monobank.request_timeout")) * time.Second},
		log:        log,
	}
}

// SetHTTPClient sets a custom HTTP client for testing
func (s *MonobankService) SetHTTPClient(client interface {
	Do(*http.Request) (*http.Response, error)
}) {
	s.httpClient = client
}

// Connect implements service.MonobankService
func (s *MonobankService) Connect(ctx context.Context, userID uuid.UUID, token string) error {
	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}
	if user == nil {
		return errors.ErrUserNotFound
	}

	// Get client info from Monobank API
	clientInfo, err := s.getMonobankClientInfo(token)
	if err != nil {
		return err
	}

	// Create or update integration
	integration := &entity.MonobankIntegration{
		UserID:      userID,
		ClientID:    clientInfo.ClientID,
		Token:       token,
		WebhookURL:  clientInfo.WebHookURL,
		Permissions: clientInfo.Permissions,
	}

	// Check if integration already exists
	existing, err := s.monoRepo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}

	if existing != nil {
		integration.ID = existing.ID
		if err := s.monoRepo.Update(ctx, integration); err != nil {
			return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
		}
	} else {
		if err := s.monoRepo.Create(ctx, integration); err != nil {
			return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
		}
	}

	// Create or update cards
	for _, account := range clientInfo.Accounts {
		card := &entity.Card{
			UserID:            userID,
			CardName:          fmt.Sprintf("%s (%s)", account.Type, account.MaskedPan[0]),
			MaskedPan:         account.MaskedPan[0],
			Balance:           account.Balance,
			CreditLimit:       account.CreditLimit,
			CurrencyCode:      account.CurrencyCode,
			IsManual:          false,
			Type:              account.Type,
			MonobankAccountID: account.ID,
		}

		existingCard, err := s.cardRepo.GetByMonobankAccountID(ctx, account.ID)
		if err != nil {
			return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
		}

		if existingCard != nil {
			card.ID = existingCard.ID
			if err := s.cardRepo.Update(ctx, card); err != nil {
				return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
			}
		} else {
			if err := s.cardRepo.Create(ctx, card); err != nil {
				return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
			}
		}
	}

	return nil
}

// Disconnect implements service.MonobankService
func (s *MonobankService) Disconnect(ctx context.Context, userID uuid.UUID) error {
	// Check if integration exists
	integration, err := s.monoRepo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}
	if integration == nil {
		return errors.ErrMonobankIntegrationNotFound
	}

	return s.monoRepo.Delete(ctx, userID)
}

// SyncUserData implements service.MonobankService
func (s *MonobankService) SyncUserData(ctx context.Context, userID uuid.UUID) error {
	// Get integration
	integration, err := s.monoRepo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}
	if integration == nil {
		return errors.ErrMonobankIntegrationNotFound
	}

	// Get cards
	cards, err := s.cardRepo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}

	// Sync transactions for each card
	for i := range cards {
		if !cards[i].IsManual && cards[i].MonobankAccountID != "" {
			if err := s.syncCardTransactions(ctx, &cards[i], integration.Token); err != nil {
				s.log.Errorw("Failed to sync card transactions",
					"error", err,
					"card_id", cards[i].ID,
					"account_id", cards[i].MonobankAccountID,
				)
				continue // Continue with other cards even if one fails
			}
		}
	}

	return nil
}

// HandleWebhook implements service.MonobankService
func (s *MonobankService) HandleWebhook(ctx context.Context, data []byte) error {
	var webhook struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(data, &webhook); err != nil {
		return fmt.Errorf("%w: failed to parse webhook data", errors.ErrInvalidRequest)
	}

	switch webhook.Type {
	case "StatementItem":
		var statement struct {
			Account   string              `json:"account"`
			Statement monobankTransaction `json:"statementItem"`
		}
		if err := json.Unmarshal(webhook.Data, &statement); err != nil {
			return fmt.Errorf("%w: failed to parse statement data", errors.ErrInvalidRequest)
		}

		// Get card by account ID
		card, err := s.cardRepo.GetByMonobankAccountID(ctx, statement.Account)
		if err != nil {
			return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
		}
		if card == nil {
			return fmt.Errorf("%w: account %s", errors.ErrCardNotFound, statement.Account)
		}

		// Create transaction
		tx := s.convertMonobankTransaction(&statement.Statement, card)
		if err := s.txRepo.Create(ctx, tx); err != nil {
			return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
		}

	default:
		s.log.Warnw("Unknown webhook type", "type", webhook.Type)
	}

	return nil
}

// GetStatus implements service.MonobankService
func (s *MonobankService) GetStatus(ctx context.Context, userID uuid.UUID) (*entity.MonobankIntegration, error) {
	integration, err := s.monoRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}
	if integration == nil {
		return nil, errors.ErrMonobankIntegrationNotFound
	}
	return integration, nil
}

func (s *MonobankService) getMonobankClientInfo(token string) (*monobankClientInfo, error) {
	req, err := http.NewRequest("GET", viper.GetString("monobank.api_url")+"/personal/client-info", nil)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create request", errors.ErrInternal)
	}

	req.Header.Set("X-Token", token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to make request", errors.ErrMonobankAPIError)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, errors.ErrMonobankRateLimit
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.ErrMonobankTokenInvalid
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d", errors.ErrMonobankAPIError, resp.StatusCode)
	}

	var clientInfo monobankClientInfo
	if err := json.NewDecoder(resp.Body).Decode(&clientInfo); err != nil {
		return nil, fmt.Errorf("%w: failed to decode response", errors.ErrMonobankAPIError)
	}

	return &clientInfo, nil
}

func (s *MonobankService) syncCardTransactions(ctx context.Context, card *entity.Card, token string) error {
	// Get last transaction time
	lastTx, err := s.txRepo.GetByCardID(ctx, card.ID, 1, 0)
	if err != nil {
		return fmt.Errorf("%w: failed to get last transaction", errors.ErrDatabaseOperation)
	}

	var from int64
	if len(lastTx) > 0 {
		from = lastTx[0].TransactionDate.Unix()
	} else {
		// If no transactions, get last month
		from = time.Now().AddDate(0, -1, 0).Unix()
	}

	// Get transactions from Monobank API
	req, err := http.NewRequest("GET", fmt.Sprintf(
		"%s/personal/statement/%s/%d",
		viper.GetString("monobank.api_url"),
		card.MonobankAccountID,
		from,
	), nil)
	if err != nil {
		return fmt.Errorf("%w: failed to create request", errors.ErrInternal)
	}

	req.Header.Set("X-Token", token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: failed to make request", errors.ErrMonobankAPIError)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return errors.ErrMonobankRateLimit
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return errors.ErrMonobankTokenInvalid
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: status %d", errors.ErrMonobankAPIError, resp.StatusCode)
	}

	var transactions []monobankTransaction
	if err := json.NewDecoder(resp.Body).Decode(&transactions); err != nil {
		return fmt.Errorf("%w: failed to decode response", errors.ErrMonobankAPIError)
	}

	// Process transactions
	for _, monoTx := range transactions {
		// Check if transaction already exists
		existing, err := s.txRepo.GetByMonobankID(ctx, monoTx.ID)
		if err != nil {
			s.log.Errorw("Failed to check existing transaction",
				"error", err,
				"monobank_id", monoTx.ID,
			)
			continue
		}
		if existing != nil {
			continue
		}

		// Create new transaction
		tx := s.convertMonobankTransaction(&monoTx, card)
		if err := s.txRepo.Create(ctx, tx); err != nil {
			s.log.Errorw("Failed to create transaction",
				"error", err,
				"monobank_id", monoTx.ID,
			)
			continue
		}
	}

	return nil
}

func (s *MonobankService) convertMonobankTransaction(monoTx *monobankTransaction, card *entity.Card) *entity.Transaction {
	txType := "expense"
	if monoTx.Amount > 0 {
		txType = "income"
	}

	return &entity.Transaction{
		CardID:          card.ID,
		UserID:          card.UserID,
		Amount:          abs(monoTx.Amount),
		OperationAmount: abs(monoTx.OperationAmount),
		CurrencyCode:    monoTx.CurrencyCode,
		Type:            txType,
		Description:     monoTx.Description,
		MCC:             monoTx.MCC,
		CommissionRate:  monoTx.CommissionRate,
		CashbackAmount:  monoTx.CashbackAmount,
		BalanceAfter:    monoTx.Balance,
		Hold:            monoTx.Hold,
		TransactionDate: time.Unix(monoTx.Time, 0),
		MonobankID:      monoTx.ID,
		Comment:         monoTx.Comment,
	}
}

func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
