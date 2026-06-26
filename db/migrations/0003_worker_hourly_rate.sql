-- +goose Up
ALTER TABLE workers ADD COLUMN hourly_rate REAL NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE workers DROP COLUMN hourly_rate;
