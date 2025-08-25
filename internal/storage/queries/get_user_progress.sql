SELECT
    urp.current_point_id,
    rp.version_id,
    rp.title,
    rp.description,
    rp.latitude,
    rp.longitude,
    rp.order_index,
    rp.arrival_instructions
FROM route_progress urp
         JOIN route_points rp ON urp.current_point_id = rp.id
WHERE urp.user_id = $1 AND urp.order_id = $2