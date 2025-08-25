SELECT m.url, m.type, m.description
FROM route_version_media rvm
         JOIN media m ON rvm.media_id = m.id
WHERE rvm.route_version_id = $1
ORDER BY rvm.display_order