-- name: CreateLocationIfNew :execrows
INSERT INTO locations (user_id, geojson, title, description, is_public)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id, geojson) DO NOTHING;

-- name: GetLocationsByUserID :many
SELECT * FROM locations WHERE user_id = $1 ORDER BY created_at DESC;

-- name: GetPublicPointLocations :many
SELECT l.*, u.alias
FROM locations l
JOIN users u ON u.id = l.user_id
WHERE l.geojson->>'type' = 'Point' AND l.is_public = true
ORDER BY l.id;

-- name: GetPrivatePointLocationsByUser :many
SELECT l.*, u.alias
FROM locations l
JOIN users u ON u.id = l.user_id
WHERE l.geojson->>'type' = 'Point' AND l.is_public = false AND l.user_id = $1
ORDER BY l.id;

-- name: GetLocationByID :one
SELECT * FROM locations WHERE id = $1 LIMIT 1;

-- name: DeleteLocation :execrows
DELETE FROM locations WHERE id = $1 AND user_id = $2;

-- name: UpdateLocation :one
UPDATE locations SET title = $3, description = $4, is_public = $5
WHERE id = $1 AND user_id = $2
RETURNING *;
