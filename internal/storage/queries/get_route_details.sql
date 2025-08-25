SELECT r.id,
       r.status,
       r.is_visible,
       r.created_by,
       r.created_at,
       r.updated_at,
       v.id,
       v.version_number,
       v.title,
       v.description,
       v.duration_minutes,
       v.length_km,
       v.theme,
       v.price,
       v.city,
       v.created_at
FROM routes r
         JOIN route_versions v ON r.id = v.route_id
WHERE r.id = $1
ORDER BY v.version_number DESC LIMIT 1