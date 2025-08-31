package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"walki/internal/models"
)

type UserRepo struct{ db *pgxpool.Pool }

func NewUserRepo(db *pgxpool.Pool) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) Upsert(ctx context.Context, u *models.User) error {
	const q = `
	  INSERT INTO users (telegram_id, username, full_name, created_at, updated_at)
	  VALUES ($1,$2,$3,NOW(),NOW())
	  ON CONFLICT (telegram_id) DO UPDATE
	    SET username=EXCLUDED.username,
	        full_name=EXCLUDED.full_name,
	        updated_at=NOW();
	`
	_, err := r.db.Exec(ctx, q, u.TelegramID, u.Username, u.FullName)
	return err
}

func (r *UserRepo) ByTelegramID(ctx context.Context, tgID int64) (*models.User, error) {
	const q = `SELECT id, telegram_id, username, full_name, created_at, updated_at FROM users WHERE telegram_id=$1`
	var u models.User
	if err := r.db.QueryRow(ctx, q, tgID).
		Scan(&u.ID, &u.TelegramID, &u.Username, &u.FullName, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}
