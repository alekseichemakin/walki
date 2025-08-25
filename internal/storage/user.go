package storage

import (
	"context"
	"fmt"
	"walki/internal/models"
)

func (s *Storage) SaveUser(user *models.User) error {
	query := s.queries["create_user.sql"]
	err := s.db.QueryRow(context.Background(), query,
		user.TelegramID,
		user.Username,
		user.FullName,
	).Scan(&user.ID)

	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	return nil
}

func (s *Storage) GetUserByTelegramID(telegramID int64) (*models.User, error) {
	var user models.User
	query := s.queries["get_user_by_telegram_id.sql"]

	err := s.db.QueryRow(context.Background(), query, telegramID).Scan(
		&user.ID, &user.TelegramID, &user.Username, &user.FullName,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}
