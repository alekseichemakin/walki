INSERT INTO users (telegram_id, username, full_name, created_at, updated_at)
VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) ON CONFLICT (telegram_id)
        DO
UPDATE SET
    username = EXCLUDED.username,
    full_name = EXCLUDED.full_name,
    updated_at = CURRENT_TIMESTAMP
    RETURNING id