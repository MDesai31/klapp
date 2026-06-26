-- +goose Up
CREATE TABLE invoices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    submitted_by INTEGER NOT NULL REFERENCES workers(id),
    date TEXT NOT NULL,
    house_number TEXT NOT NULL,
    customer_name TEXT NOT NULL DEFAULT '',
    customer_id INTEGER REFERENCES customers(id),
    time_arrived TEXT NOT NULL,
    time_left TEXT NOT NULL,
    no_of_workers INTEGER NOT NULL DEFAULT 1,
    comments TEXT NOT NULL DEFAULT '',
    reviewed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE invoice_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    invoice_id INTEGER NOT NULL REFERENCES invoices(id),
    description TEXT NOT NULL
);

CREATE TABLE invoice_materials_used (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    invoice_id INTEGER NOT NULL REFERENCES invoices(id),
    material TEXT NOT NULL
);

CREATE INDEX idx_invoices_date ON invoices(date);
CREATE INDEX idx_invoices_reviewed ON invoices(reviewed);
CREATE INDEX idx_invoices_customer ON invoices(customer_id);

-- +goose Down
DROP INDEX idx_invoices_customer;
DROP INDEX idx_invoices_reviewed;
DROP INDEX idx_invoices_date;
DROP TABLE invoice_materials_used;
DROP TABLE invoice_jobs;
DROP TABLE invoices;
