-- +goose Up
-- SQLite can't drop NOT NULL inline; recreate the table with nullable start_lat/start_lon
-- so workers who don't require location can punch in without coordinates.
CREATE TABLE time_punches_new (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    worker_id         INTEGER NOT NULL REFERENCES workers(id),
    pay_period        TEXT NOT NULL,
    day               TEXT NOT NULL,
    start_time        TEXT NOT NULL,
    end_time          TEXT,
    start_lat         REAL,
    start_lon         REAL,
    end_lat           REAL,
    end_lon           REAL,
    late              BOOLEAN NOT NULL DEFAULT FALSE,
    modified_by_admin BOOLEAN NOT NULL DEFAULT FALSE,
    non_compliant     BOOLEAN NOT NULL DEFAULT FALSE
);
INSERT INTO time_punches_new SELECT * FROM time_punches;
DROP TABLE time_punches;
ALTER TABLE time_punches_new RENAME TO time_punches;
CREATE INDEX idx_time_punches_worker_open ON time_punches(worker_id, end_time);

-- +goose Down
SELECT 1; -- not reversible without potential data loss
