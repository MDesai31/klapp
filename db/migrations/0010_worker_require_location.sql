-- +goose Up
ALTER TABLE workers ADD COLUMN require_location BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
SELECT 1; -- SQLite does not support DROP COLUMN portably; this migration is not reversible
