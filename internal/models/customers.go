package models

import (
	"database/sql"
	"errors"
)

type Customer struct {
	ID          int
	Name        string
	Phone       string
	HouseNumber string
	Address     string
}

type CustomerModel struct {
	DB *sql.DB
}

func (m *CustomerModel) GetByHouseNumber(houseNumber string) ([]Customer, error) {
	rows, err := m.DB.Query(
		`SELECT id, name, phone, house_number, address FROM customers WHERE house_number = ? ORDER BY name`,
		houseNumber,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Customer
	for rows.Next() {
		var c Customer
		if err := rows.Scan(&c.ID, &c.Name, &c.Phone, &c.HouseNumber, &c.Address); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (m *CustomerModel) List() ([]Customer, error) {
	rows, err := m.DB.Query(`SELECT id, name, phone, house_number, address FROM customers ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Customer
	for rows.Next() {
		var c Customer
		if err := rows.Scan(&c.ID, &c.Name, &c.Phone, &c.HouseNumber, &c.Address); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (m *CustomerModel) Search(q string) ([]Customer, error) {
	rows, err := m.DB.Query(
		`SELECT id, name, phone, house_number, address FROM customers WHERE name LIKE ? OR house_number LIKE ? OR address LIKE ? ORDER BY name LIMIT 50`,
		"%"+q+"%", "%"+q+"%", "%"+q+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Customer
	for rows.Next() {
		var c Customer
		if err := rows.Scan(&c.ID, &c.Name, &c.Phone, &c.HouseNumber, &c.Address); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (m *CustomerModel) Get(id int) (Customer, error) {
	var c Customer
	err := m.DB.QueryRow(
		`SELECT id, name, phone, house_number, address FROM customers WHERE id = ?`, id,
	).Scan(&c.ID, &c.Name, &c.Phone, &c.HouseNumber, &c.Address)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Customer{}, ErrNoRecord
		}
		return Customer{}, err
	}
	return c, nil
}

func (m *CustomerModel) Create(name, phone, houseNumber, address string) (int, error) {
	result, err := m.DB.Exec(
		`INSERT INTO customers (name, phone, house_number, address) VALUES (?, ?, ?, ?)`,
		name, phone, houseNumber, address,
	)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}

func (m *CustomerModel) Update(id int, name, phone, houseNumber, address string) error {
	result, err := m.DB.Exec(
		`UPDATE customers SET name = ?, phone = ?, house_number = ?, address = ? WHERE id = ?`,
		name, phone, houseNumber, address, id,
	)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNoRecord
	}
	return nil
}
