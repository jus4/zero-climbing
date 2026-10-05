-- +goose Up
ALTER TABLE users ADD CONSTRAINT users_alias_key UNIQUE (alias);

-- +goose Down
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_alias_key;
