SELECT id,
       route_id,
       version_number,
       title,
       description,
       duration_minutes,
       length_km,
       theme,
       price,
       city,
       created_at
FROM route_versions
WHERE city = $1
ORDER BY created_at DESC