-- name: CreateLocation :one
INSERT INTO locations (user_id, geojson, title, description)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetLocationsByUserID :many
SELECT * FROM locations WHERE user_id = $1 ORDER BY created_at DESC;

-- name: GetLocationByID :one
SELECT * FROM locations WHERE id = $1 LIMIT 1;

-- name: DeleteLocation :exec
DELETE FROM locations WHERE id = $1 AND user_id = $2;
