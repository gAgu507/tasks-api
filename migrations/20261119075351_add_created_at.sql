-- +goose Up
ALTER TABLE tasks ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- +goose Down
ALTER TABLE tasks DROP COLUMN created_at;