package models

import (
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Admin struct {
	ID       int
	Username string
}

type AdminModel struct {
	DB *sql.DB
}

func (m *AdminModel) Authenticate(username, password string) (Admin, error) {
	var a Admin
	var hash string

	stmt := `SELECT id, username, password_hash FROM admins WHERE username = ?`
	err := m.DB.QueryRow(stmt, username).Scan(&a.ID, &a.Username, &hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Admin{}, ErrInvalidCredentials
		}
		return Admin{}, err
	}

	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return Admin{}, ErrInvalidCredentials
	}

	return a, nil
}

// Upsert creates the admin if the username is new, or replaces their
// password if it already exists. Used only by cmd/seedadmin - there's no
// signup page.
func (m *AdminModel) Upsert(username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO admins (username, password_hash) VALUES (?, ?)
		ON CONFLICT(username) DO UPDATE SET password_hash = excluded.password_hash`
	_, err = m.DB.Exec(stmt, username, string(hash))
	return err
}
