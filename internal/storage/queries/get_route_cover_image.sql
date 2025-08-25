SELECT m.url
FROM route_version_media rvm
         JOIN media m ON rvm.media_id = m.id
WHERE rvm.route_version_id = $1
  AND rvm.is_cover = TRUE LIMIT 1