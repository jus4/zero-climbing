-- name: CreateLocation :one
INSERT INTO locations (user_id, geojson, title, description)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetLocationsByUserID :many
SELECT * FROM locations WHERE user_id = $1 ORDER BY created_at DESC;

-- name: GetPointLocations :many
SELECT l.*, u.alias
FROM locations l
JOIN users u ON u.id = l.user_id
WHERE l.geojson->>'type' = 'Point';

-- name: GetLocationByID :one
SELECT * FROM locations WHERE id = $1 LIMIT 1;

-- name: DeleteLocation :execrows
DELETE FROM locations WHERE id = $1 AND user_id = $2;

-- name: UpdateLocation :one
UPDATE locations SET title = $3, description = $4
WHERE id = $1 AND user_id = $2
RETURNING *;
