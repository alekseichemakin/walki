ALTER TABLE media
    ADD COLUMN IF NOT EXISTS mime_type  TEXT,
    ADD COLUMN IF NOT EXISTS bytes      BIGINT,
    ADD COLUMN IF NOT EXISTS s3_bucket  TEXT,
    ADD COLUMN IF NOT EXISTS s3_key     TEXT,
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMP NOT NULL DEFAULT NOW();

CREATE TABLE IF NOT EXISTS telegram_files
(
    media_id     BIGINT PRIMARY KEY REFERENCES media (id) ON DELETE CASCADE,
    file_id      TEXT      NOT NULL,
    content_type TEXT,
    chat_id      BIGINT,
    created_at   TIMESTAMP NOT NULL DEFAULT NOW()
);