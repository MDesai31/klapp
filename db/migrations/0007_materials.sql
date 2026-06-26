-- +goose Up
CREATE TABLE materials (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    unit TEXT NOT NULL DEFAULT '',
    price REAL NOT NULL DEFAULT 0
);

-- +goose Down
DROP TABLE materials;
