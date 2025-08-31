package service

import (
	"context"
	"walki/internal/models"
	"walki/internal/repository"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Register сохраняет нового пользователя или обновляет существующего
func (s *UserService) Register(ctx context.Context, user *models.User) error {
	return s.repo.Upsert(ctx, user)
}

// GetByTelegramID получает пользователя по его Telegram ID
func (s *UserService) GetByTelegramID(ctx context.Context, tgID int64) (*models.User, error) {
	return s.repo.ByTelegramID(ctx, tgID)
}
