-- +goose Up
CREATE TABLE job_descriptions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    description TEXT NOT NULL UNIQUE
);

-- +goose Down
DROP TABLE job_descriptions;
