SELECT id FROM route_versions
WHERE route_id = $1
ORDER BY version_number DESC
    LIMIT 1