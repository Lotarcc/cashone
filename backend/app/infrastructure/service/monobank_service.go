package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"cashone/domain/entity"
	"cashone/domain/repository"
	"cashone/domain/service"
)

var (
	ErrMonobankAPIError     = errors.New("monobank API error")
	ErrInvalidMonobankToken = errors.New("invalid monobank token")
	ErrMonobankRateLimit    = errors.New("monobank rate limit exceeded")
)

type monobankService struct {
	monoRepo   repository.MonobankIntegrationRepository
	cardRepo   repository.CardRepository
	txRepo     repository.TransactionRepository
	userRepo   repository.UserRepository
	httpClient *http.Client
	log        *zap.SugaredLogger
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

func NewMonobankService(
	monoRepo repository.MonobankIntegrationRepository,
	cardRepo repository.CardRepository,
	txRepo repository.TransactionRepository,
	userRepo repository.UserRepository,
	log *zap.SugaredLogger,
) service.MonobankService {
	return &monobankService{
		monoRepo:   monoRepo,
		cardRepo:   cardRepo,
		txRepo:     txRepo,
		userRepo:   userRepo,
		httpClient: &http.Client{Timeout: time.Duration(viper.GetInt("monobank.request_timeout")) * time.Second},
		log:        log,
	}
}

func (s *monobankService) IntegrateMonobank(ctx context.Context, userID uuid.UUID, token string) error {
	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.log.Errorw("Failed to get user for monobank integration", "error", err, "user_id", userID)
		return fmt.Errorf("failed to verify user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
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
		s.log.Errorw("Failed to check existing integration", "error", err, "user_id", userID)
		return fmt.Errorf("failed to check existing integration: %w", err)
	}

	if existing != nil {
		integration.ID = existing.ID
		if err := s.monoRepo.Update(ctx, integration); err != nil {
			s.log.Errorw("Failed to update monobank integration", "error", err, "user_id", userID)
			return fmt.Errorf("failed to update integration: %w", err)
		}
	} else {
		if err := s.monoRepo.Create(ctx, integration); err != nil {
			s.log.Errorw("Failed to create monobank integration", "error", err, "user_id", userID)
			return fmt.Errorf("failed to create integration: %w", err)
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
			s.log.Errorw("Failed to check existing card", "error", err, "account_id", account.ID)
			return fmt.Errorf("failed to check existing card: %w", err)
		}

		if existingCard != nil {
			card.ID = existingCard.ID
			if err := s.cardRepo.Update(ctx, card); err != nil {
				s.log.Errorw("Failed to update monobank card", "error", err, "account_id", account.ID)
				return fmt.Errorf("failed to update card: %w", err)
			}
		} else {
			if err := s.cardRepo.Create(ctx, card); err != nil {
				s.log.Errorw("Failed to create monobank card", "error", err, "account_id", account.ID)
				return fmt.Errorf("failed to create card: %w", err)
			}
		}
	}

	return nil
}

func (s *monobankService) SyncMonobankData(ctx context.Context, userID uuid.UUID) error {
	// Get integration
	integration, err := s.monoRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.log.Errorw("Failed to get monobank integration", "error", err, "user_id", userID)
		return fmt.Errorf("failed to get integration: %w", err)
	}
	if integration == nil {
		return errors.New("monobank integration not found")
	}

	// Get cards
	cards, err := s.cardRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.log.Errorw("Failed to get user cards", "error", err, "user_id", userID)
		return fmt.Errorf("failed to get cards: %w", err)
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

func (s *monobankService) GetMonobankIntegration(ctx context.Context, userID uuid.UUID) (*entity.MonobankIntegration, error) {
	return s.monoRepo.GetByUserID(ctx, userID)
}

func (s *monobankService) DisconnectMonobank(ctx context.Context, userID uuid.UUID) error {
	return s.monoRepo.Delete(ctx, userID)
}

func (s *monobankService) HandleWebhook(ctx context.Context, data []byte) error {
	var webhook struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(data, &webhook); err != nil {
		s.log.Errorw("Failed to unmarshal webhook data", "error", err)
		return fmt.Errorf("failed to parse webhook data: %w", err)
	}

	switch webhook.Type {
	case "StatementItem":
		var statement struct {
			Account   string              `json:"account"`
			Statement monobankTransaction `json:"statementItem"`
		}
		if err := json.Unmarshal(webhook.Data, &statement); err != nil {
			s.log.Errorw("Failed to unmarshal statement data", "error", err)
			return fmt.Errorf("failed to parse statement data: %w", err)
		}

		// Get card by account ID
		card, err := s.cardRepo.GetByMonobankAccountID(ctx, statement.Account)
		if err != nil {
			s.log.Errorw("Failed to get card by account ID",
				"error", err,
				"account_id", statement.Account,
			)
			return fmt.Errorf("failed to get card: %w", err)
		}
		if card == nil {
			return fmt.Errorf("card not found for account: %s", statement.Account)
		}

		// Create transaction
		tx := s.convertMonobankTransaction(&statement.Statement, card)
		if err := s.txRepo.Create(ctx, tx); err != nil {
			s.log.Errorw("Failed to create transaction from webhook",
				"error", err,
				"account_id", statement.Account,
			)
			return fmt.Errorf("failed to create transaction: %w", err)
		}

	default:
		s.log.Warnw("Unknown webhook type", "type", webhook.Type)
		return fmt.Errorf("unknown webhook type: %s", webhook.Type)
	}

	return nil
}

func (s *monobankService) getMonobankClientInfo(token string) (*monobankClientInfo, error) {
	req, err := http.NewRequest("GET", viper.GetString("monobank.api_url")+"/personal/client-info", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Token", token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, ErrMonobankRateLimit
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d", ErrMonobankAPIError, resp.StatusCode)
	}

	var clientInfo monobankClientInfo
	if err := json.NewDecoder(resp.Body).Decode(&clientInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &clientInfo, nil
}

func (s *monobankService) syncCardTransactions(ctx context.Context, card *entity.Card, token string) error {
	// Get last transaction time
	lastTx, err := s.txRepo.GetByCardID(ctx, card.ID, 1, 0)
	if err != nil {
		return fmt.Errorf("failed to get last transaction: %w", err)
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
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Token", token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return ErrMonobankRateLimit
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: status %d", ErrMonobankAPIError, resp.StatusCode)
	}

	var transactions []monobankTransaction
	if err := json.NewDecoder(resp.Body).Decode(&transactions); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
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

func (s *monobankService) convertMonobankTransaction(monoTx *monobankTransaction, card *entity.Card) *entity.Transaction {
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
