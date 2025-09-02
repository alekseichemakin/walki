package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"walki/internal/models"
)

type MediaRepo struct {
	db *pgxpool.Pool
}

func NewMediaRepo(db *pgxpool.Pool) *MediaRepo { return &MediaRepo{db: db} }

// --- Media ---

func (r *MediaRepo) GetByID(ctx context.Context, id int64) (*models.Media, error) {
	const q = `
SELECT id, type, url, filename, size_bytes, description, uploaded_by, uploaded_at, is_public,
       mime_type, s3_bucket, s3_key
FROM media
WHERE id = $1`
	var (
		m  models.Media
		ts time.Time
	)
	if err := r.db.QueryRow(ctx, q, id).Scan(
		&m.ID,
		&m.Type,
		&m.URL,
		&m.Filename,
		&m.SizeBytes,
		&m.Description,
		&m.UploadedBy,
		&ts,
		&m.IsPublic,
		&m.MimeType, // *string
		&m.S3Bucket, // *string
		&m.S3Key,    // *string
	); err != nil {
		return nil, err
	}
	m.UploadedAt = ts
	return &m, nil
}

// --- S3 storage (original) ---

func (r *MediaRepo) UpdateStorage(ctx context.Context, mediaID int64, bucket, key, mimeType string, sizeBytes int64) error {
	const q = `
UPDATE media
SET s3_bucket = $2,
    s3_key    = $3,
    mime_type = $4,
    size_bytes = $5
WHERE id = $1`
	ct, err := r.db.Exec(ctx, q, mediaID, bucket, key, mimeType, sizeBytes)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("media not found")
	}
	return nil
}

// --- Telegram cache ---

func (r *MediaRepo) GetTelegramFileID(ctx context.Context, mediaID int64) (string, bool, error) {
	const q = `SELECT file_id FROM telegram_files WHERE media_id = $1`
	var fid string
	if err := r.db.QueryRow(ctx, q, mediaID).Scan(&fid); err != nil {
		return "", false, nil
	}
	return fid, true, nil
}

func (r *MediaRepo) UpsertTelegramFileID(ctx context.Context, mediaID int64, fileID, contentType string, chatID *int64) error {
	const q = `
INSERT INTO telegram_files (media_id, file_id, content_type, chat_id)
VALUES ($1, $2, $3, $4)
ON CONFLICT (media_id) DO UPDATE
SET file_id = EXCLUDED.file_id,
    content_type = EXCLUDED.content_type,
    chat_id = EXCLUDED.chat_id,
    created_at = NOW()`
	_, err := r.db.Exec(ctx, q, mediaID, fileID, contentType, chatID)
	return err
}
