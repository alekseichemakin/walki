SELECT EXISTS (
    SELECT 1
    FROM orders
    WHERE user_id = $1 AND route_id = $2
      AND status = 'paid'
      AND (access_expiry IS NULL OR access_expiry > NOW())
)