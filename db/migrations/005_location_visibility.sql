-- +goose Up
ALTER TABLE locations ADD COLUMN is_public BOOLEAN NOT NULL DEFAULT false;

CREATE INDEX idx_locations_is_public ON locations(is_public);

-- +goose Down
DROP INDEX IF EXISTS idx_locations_is_public;

ALTER TABLE locations DROP COLUMN is_public;
