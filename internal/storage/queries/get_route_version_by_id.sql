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
WHERE id = $1