-- +goose Up
CREATE TABLE workers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    worker_name TEXT NOT NULL,
    pin TEXT NOT NULL,
    phone TEXT,
    active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE time_punches (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    worker_id INTEGER NOT NULL REFERENCES workers(id),
    pay_period TEXT NOT NULL,
    day TEXT NOT NULL,
    start_time TEXT NOT NULL,
    end_time TEXT,
    start_lat REAL NOT NULL,
    start_lon REAL NOT NULL,
    end_lat REAL,
    end_lon REAL,
    late BOOLEAN NOT NULL DEFAULT FALSE,
    modified_by_admin BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_time_punches_worker_open ON time_punches(worker_id, end_time);

-- +goose Down
DROP TABLE time_punches;
DROP TABLE workers;
