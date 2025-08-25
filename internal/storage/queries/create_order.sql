INSERT INTO orders (user_id,
                    route_id,
                    version_id,
                    status,
                    amount,
                    created_at,
                    paid_at,
                    access_expiry)
VALUES ($1, $2, $3, 'paid', $4, $5, $6,$7)
    RETURNING id, user_id, route_id, version_id, status, amount, created_at, paid_at, access_expiry