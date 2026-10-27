-- +goose Up
-- Remove previously imported duplicates, keeping the oldest row (lowest id)
-- for each user/geometry pair.
DELETE FROM locations l
USING locations keep
WHERE l.user_id = keep.user_id
  AND l.geojson = keep.geojson
  AND l.id > keep.id;

CREATE UNIQUE INDEX locations_user_id_geojson_key
    ON locations (user_id, geojson);

-- +goose Down
DROP INDEX IF EXISTS locations_user_id_geojson_key;
