SELECT id,
       telegram_id,
       username,
       full_name,
       created_at,
       updated_at
FROM users
WHERE telegram_id = $1