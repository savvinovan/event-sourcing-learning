-- +goose Up
ALTER TABLE events ADD COLUMN schema_version INT NOT NULL DEFAULT 1;

-- +goose Down
ALTER TABLE events DROP COLUMN schema_version;
