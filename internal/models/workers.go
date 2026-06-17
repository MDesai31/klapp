package models

import (
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Worker struct {
	ID         int
	WorkerName string
	Phone      string
	Active     bool
}

type WorkerModel struct {
	DB *sql.DB
}

// Authenticate identifies a worker by PIN alone (no separate username/ID
// entry), so it checks the entered PIN against every active worker's
// bcrypt hash until one matches. Fine at the expected scale (~15 workers).
func (m *WorkerModel) Authenticate(pin string) (Worker, error) {
	rows, err := m.DB.Query(`SELECT id, worker_name, pin, phone FROM workers WHERE active = TRUE`)
	if err != nil {
		return Worker{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var w Worker
		var hash string
		if err := rows.Scan(&w.ID, &w.WorkerName, &hash, &w.Phone); err != nil {
			return Worker{}, err
		}
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(pin)) == nil {
			w.Active = true
			return w, nil
		}
	}
	if err := rows.Err(); err != nil {
		return Worker{}, err
	}

	return Worker{}, ErrInvalidPIN
}

func (m *WorkerModel) Get(id int) (Worker, error) {
	stmt := `SELECT id, worker_name, phone, active FROM workers WHERE id = ?`

	var w Worker
	err := m.DB.QueryRow(stmt, id).Scan(&w.ID, &w.WorkerName, &w.Phone, &w.Active)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Worker{}, ErrNoRecord
		}
		return Worker{}, err
	}

	return w, nil
}

// List returns every worker, active or not, for the admin worker
// management page.
func (m *WorkerModel) List() ([]Worker, error) {
	rows, err := m.DB.Query(`SELECT id, worker_name, phone, active FROM workers ORDER BY worker_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Worker
	for rows.Next() {
		var w Worker
		if err := rows.Scan(&w.ID, &w.WorkerName, &w.Phone, &w.Active); err != nil {
			return nil, err
		}
		out = append(out, w)
	}

	return out, rows.Err()
}

func (m *WorkerModel) Create(name, pin, phone string) (int, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	stmt := `INSERT INTO workers (worker_name, pin, phone, active) VALUES (?, ?, ?, TRUE)`
	result, err := m.DB.Exec(stmt, name, hash, phone)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// SetActive activates or deactivates a worker. There's no hard delete -
// punch history references worker_id and should stay intact.
func (m *WorkerModel) SetActive(id int, active bool) error {
	result, err := m.DB.Exec(`UPDATE workers SET active = ? WHERE id = ?`, active, id)
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
