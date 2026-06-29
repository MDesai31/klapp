-- +goose Up
ALTER TABLE time_punches ADD COLUMN non_compliant BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
SELECT 1; -- SQLite does not support DROP COLUMN portably; this migration is not reversible
