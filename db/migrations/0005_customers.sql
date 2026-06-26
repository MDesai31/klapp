-- +goose Up
CREATE TABLE customers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    phone TEXT NOT NULL DEFAULT '',
    house_number TEXT NOT NULL,
    address TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_customers_house_number ON customers(house_number);

-- +goose Down
DROP INDEX idx_customers_house_number;
DROP TABLE customers;
