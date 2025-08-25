SELECT o.id,
       o.user_id,
       o.route_id,
       o.version_id,
       o.status,
       o.amount,
       o.created_at,
       o.paid_at,
       o.access_expiry,
       rv.title,
       rv.city,
       rv.length_km,
       rv.duration_minutes
FROM orders o
         JOIN route_versions rv ON o.version_id = rv.id
WHERE o.user_id = $1
ORDER BY o.created_at DESC