package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"cashone/domain/entity"
	"cashone/domain/errors"
	"cashone/domain/repository"
	"cashone/domain/service"
)

type cardService struct {
	cardRepo repository.CardRepository
	userRepo repository.UserRepository
	log      *zap.SugaredLogger
}

// NewCardService creates a new card service
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

func (s *cardService) Create(ctx context.Context, card *entity.Card) error {
	// Validate card data
	if err := s.validateCard(card); err != nil {
		return fmt.Errorf("%w: %v", errors.ErrInvalidCardData, err)
	}

	// Check if user exists
	user, err := s.userRepo.GetByID(ctx, card.UserID)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}
	if user == nil {
		return errors.ErrUserNotFound
	}

	// Check if card with this masked PAN already exists for the user
	existingCards, err := s.cardRepo.GetByUserID(ctx, card.UserID)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}
	for _, existingCard := range existingCards {
		if existingCard.MaskedPan == card.MaskedPan {
			return errors.ErrCardAlreadyExists
		}
	}

	// Generate UUID if not provided
	if card.ID == uuid.Nil {
		card.ID = uuid.New()
	}

	// Create card
	if err := s.cardRepo.Create(ctx, card); err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}

	s.log.Infow("Card created successfully",
		"id", card.ID,
		"user_id", card.UserID,
		"masked_pan", card.MaskedPan,
	)
	return nil
}

func (s *cardService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Card, error) {
	card, err := s.cardRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}
	if card == nil {
		return nil, errors.ErrCardNotFound
	}
	return card, nil
}

func (s *cardService) GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Card, error) {
	// Check if user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}
	if user == nil {
		return nil, errors.ErrUserNotFound
	}

	// Get user's cards
	cards, err := s.cardRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}
	return cards, nil
}

func (s *cardService) Update(ctx context.Context, card *entity.Card) error {
	// Validate card data
	if err := s.validateCard(card); err != nil {
		return fmt.Errorf("%w: %v", errors.ErrInvalidCardData, err)
	}

	// Check if card exists
	existingCard, err := s.cardRepo.GetByID(ctx, card.ID)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}
	if existingCard == nil {
		return errors.ErrCardNotFound
	}

	// Check if user exists
	user, err := s.userRepo.GetByID(ctx, card.UserID)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}
	if user == nil {
		return errors.ErrUserNotFound
	}

	// Update card
	if err := s.cardRepo.Update(ctx, card); err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}

	s.log.Infow("Card updated successfully",
		"id", card.ID,
		"user_id", card.UserID,
		"masked_pan", card.MaskedPan,
	)
	return nil
}

func (s *cardService) Delete(ctx context.Context, id uuid.UUID) error {
	// Check if card exists
	existingCard, err := s.cardRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}
	if existingCard == nil {
		return errors.ErrCardNotFound
	}

	// Delete card
	if err := s.cardRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabaseOperation, err)
	}

	s.log.Infow("Card deleted successfully", "id", id)
	return nil
}

func (s *cardService) validateCard(card *entity.Card) error {
	if card == nil {
		return errors.ErrInvalidCardData
	}

	var validationErrors []string

	if card.UserID == uuid.Nil {
		validationErrors = append(validationErrors, "user ID is required")
	}
	if card.CardName == "" {
		validationErrors = append(validationErrors, "card name is required")
	}
	if card.MaskedPan == "" {
		validationErrors = append(validationErrors, "masked PAN is required")
	}
	if card.CurrencyCode == 0 {
		validationErrors = append(validationErrors, "currency code is required")
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("%w: %v", errors.ErrValidation, validationErrors)
	}

	return nil
}
