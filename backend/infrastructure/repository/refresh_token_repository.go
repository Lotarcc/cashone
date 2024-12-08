package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"cashone/domain/entity"
	"cashone/domain/repository"
)

type refreshTokenRepository struct {
	db  *gorm.DB
	log *zap.SugaredLogger
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *gorm.DB, log *zap.SugaredLogger) repository.RefreshTokenRepository {
	return &refreshTokenRepository{
		db:  db,
		log: log,
	}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *entity.RefreshToken) error {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		r.log.Errorw("Failed to create refresh token", "error", err, "user_id", token.UserID)
		return err
	}
	return nil
}

func (r *refreshTokenRepository) GetByToken(ctx context.Context, token string) (*entity.RefreshToken, error) {
	var refreshToken entity.RefreshToken
	if err := r.db.WithContext(ctx).Where("token = ? AND revoked_at IS NULL", token).First(&refreshToken).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.log.Errorw("Failed to get refresh token", "error", err)
		return nil, err
	}
	return &refreshToken, nil
}

func (r *refreshTokenRepository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]entity.RefreshToken, error) {
	var tokens []entity.RefreshToken
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, time.Now()).
		Find(&tokens).Error; err != nil {
		r.log.Errorw("Failed to get active refresh tokens", "error", err, "user_id", userID)
		return nil, err
	}
	return tokens, nil
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, token string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&entity.RefreshToken{}).
		Where("token = ? AND revoked_at IS NULL", token).
		Update("revoked_at", now)

	if result.Error != nil {
		r.log.Errorw("Failed to revoke refresh token", "error", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *refreshTokenRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Model(&entity.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error; err != nil {
		r.log.Errorw("Failed to revoke all user tokens", "error", err, "user_id", userID)
		return err
	}
	return nil
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	if err := r.db.WithContext(ctx).
		Where("expires_at < ? OR revoked_at IS NOT NULL", time.Now()).
		Delete(&entity.RefreshToken{}).Error; err != nil {
		r.log.Errorw("Failed to delete expired refresh tokens", "error", err)
		return err
	}
	return nil
}

func (r *refreshTokenRepository) Update(ctx context.Context, token *entity.RefreshToken) error {
	if err := r.db.WithContext(ctx).Save(token).Error; err != nil {
		r.log.Errorw("Failed to update refresh token", "error", err, "id", token.ID)
		return err
	}
	return nil
}
