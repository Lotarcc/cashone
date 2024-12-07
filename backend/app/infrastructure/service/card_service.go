package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"cashone/domain/entity"
	"cashone/domain/repository"
	"cashone/domain/service"
)

var (
	ErrCardNotFound    = errors.New("card not found")
	ErrUnauthorized    = errors.New("unauthorized access to card")
	ErrInvalidCardData = errors.New("invalid card data")
	ErrMonobankCard    = errors.New("cannot modify monobank card manually")
)

type cardService struct {
	cardRepo repository.CardRepository
	userRepo repository.UserRepository
	log      *zap.SugaredLogger
}

func NewCardService(
	cardRepo repository.CardRepository,
	userRepo repository.UserRepository,
	log *zap.SugaredLogger,
) service.CardService {
	return &cardService{
		cardRepo: cardRepo,
		userRepo: userRepo,
		log:      log,
	}
}

func (s *cardService) CreateManualCard(ctx context.Context, userID uuid.UUID, card *entity.Card) error {
	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.log.Errorw("Failed to get user for card creation", "error", err, "user_id", userID)
		return fmt.Errorf("failed to verify user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Validate card data
	if err := s.validateCardData(card); err != nil {
		return err
	}

	// Set card properties
	card.UserID = userID
	card.IsManual = true
	card.MonobankAccountID = "" // Ensure no Monobank ID for manual cards

	// Create the card
	if err := s.cardRepo.Create(ctx, card); err != nil {
		s.log.Errorw("Failed to create card",
			"error", err,
			"user_id", userID,
			"card_name", card.CardName,
		)
		return fmt.Errorf("failed to create card: %w", err)
	}

	return nil
}

func (s *cardService) GetUserCards(ctx context.Context, userID uuid.UUID) ([]entity.Card, error) {
	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.log.Errorw("Failed to get user for cards retrieval", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to verify user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Get all cards for the user
	cards, err := s.cardRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.log.Errorw("Failed to get user cards", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to get cards: %w", err)
	}

	return cards, nil
}

func (s *cardService) UpdateCard(ctx context.Context, userID uuid.UUID, card *entity.Card) error {
	// Get existing card
	existingCard, err := s.cardRepo.GetByID(ctx, card.ID)
	if err != nil {
		s.log.Errorw("Failed to get card for update", "error", err, "card_id", card.ID)
		return fmt.Errorf("failed to get card: %w", err)
	}
	if existingCard == nil {
		return ErrCardNotFound
	}

	// Verify card ownership
	if existingCard.UserID != userID {
		s.log.Warnw("Unauthorized card update attempt",
			"user_id", userID,
			"card_id", card.ID,
			"card_owner_id", existingCard.UserID,
		)
		return ErrUnauthorized
	}

	// Prevent modification of Monobank cards
	if !existingCard.IsManual {
		return ErrMonobankCard
	}

	// Validate updated card data
	if err := s.validateCardData(card); err != nil {
		return err
	}

	// Preserve unchangeable fields
	card.UserID = existingCard.UserID
	card.IsManual = true
	card.MonobankAccountID = ""

	// Update the card
	if err := s.cardRepo.Update(ctx, card); err != nil {
		s.log.Errorw("Failed to update card",
			"error", err,
			"card_id", card.ID,
			"user_id", userID,
		)
		return fmt.Errorf("failed to update card: %w", err)
	}

	return nil
}

func (s *cardService) DeleteCard(ctx context.Context, userID, cardID uuid.UUID) error {
	// Get existing card
	card, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		s.log.Errorw("Failed to get card for deletion", "error", err, "card_id", cardID)
		return fmt.Errorf("failed to get card: %w", err)
	}
	if card == nil {
		return ErrCardNotFound
	}

	// Verify card ownership
	if card.UserID != userID {
		s.log.Warnw("Unauthorized card deletion attempt",
			"user_id", userID,
			"card_id", cardID,
			"card_owner_id", card.UserID,
		)
		return ErrUnauthorized
	}

	// Prevent deletion of Monobank cards
	if !card.IsManual {
		return ErrMonobankCard
	}

	// Delete the card
	if err := s.cardRepo.Delete(ctx, cardID); err != nil {
		s.log.Errorw("Failed to delete card",
			"error", err,
			"card_id", cardID,
			"user_id", userID,
		)
		return fmt.Errorf("failed to delete card: %w", err)
	}

	return nil
}

func (s *cardService) validateCardData(card *entity.Card) error {
	if card.CardName == "" {
		return fmt.Errorf("%w: card name is required", ErrInvalidCardData)
	}

	if card.CurrencyCode <= 0 {
		return fmt.Errorf("%w: invalid currency code", ErrInvalidCardData)
	}

	if card.Balance < 0 && card.CreditLimit == 0 {
		return fmt.Errorf("%w: negative balance not allowed without credit limit", ErrInvalidCardData)
	}

	if card.CreditLimit < 0 {
		return fmt.Errorf("%w: credit limit cannot be negative", ErrInvalidCardData)
	}

	return nil
}
