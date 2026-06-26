-- +goose Up
ALTER TABLE workers ADD COLUMN language TEXT NOT NULL DEFAULT 'spanish';

-- +goose Down
ALTER TABLE workers DROP COLUMN language;
