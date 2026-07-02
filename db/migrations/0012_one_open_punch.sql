-- +goose Up
-- Guarantee at most one open punch per worker. Before the index existed, a
-- double-tap on punch-in could race past the check-then-insert in PunchIn
-- and leave two open punches (both then closed with the same end time,
-- double-counting the day's hours).
--
-- Defensively close all but the newest open punch per worker first: any
-- duplicates get end_time = start_time (zero duration) and non_compliant so
-- they surface at the top of the admin timesheet for correction, and the
-- index creation below can't fail on existing data.
UPDATE time_punches
SET end_time = start_time, non_compliant = TRUE
WHERE end_time IS NULL
  AND id NOT IN (
    SELECT MAX(id) FROM time_punches WHERE end_time IS NULL GROUP BY worker_id
  );

CREATE UNIQUE INDEX idx_time_punches_one_open ON time_punches(worker_id) WHERE end_time IS NULL;

-- +goose Down
DROP INDEX idx_time_punches_one_open;
