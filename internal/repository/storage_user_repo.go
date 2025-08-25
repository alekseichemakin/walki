package repository

import (
	"context"
	"walki/internal/models"
	"walki/internal/storage"
)

type StorageUserRepo struct{ s *storage.Storage }

func NewStorageUserRepo(s *storage.Storage) UserRepository { return &StorageUserRepo{s: s} }

func (r *StorageUserRepo) Upsert(ctx context.Context, u *models.User) error {
	return r.s.SaveUser(u)
}
func (r *StorageUserRepo) ByTelegramID(ctx context.Context, tgID int64) (*models.User, error) {
	return r.s.GetUserByTelegramID(tgID)
}
