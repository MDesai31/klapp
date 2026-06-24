package models

import (
	"database/sql"
	"errors"
)

type Worker struct {
	ID         int
	WorkerName string
	PIN        string
	Phone      string
	Active     bool
}

type WorkerModel struct {
	DB *sql.DB
}

// Authenticate identifies a worker by PIN alone (no separate username/ID
// entry), so it checks the entered PIN against every active worker's PIN
// until one matches. PINs are stored in plaintext - they're short numeric
// codes for clocking in, not account passwords, and admins need to be able
// to look one up for a worker who forgot it. Fine at the expected scale
// (~15 workers).
func (m *WorkerModel) Authenticate(pin string) (Worker, error) {
	rows, err := m.DB.Query(`SELECT id, worker_name, pin, phone FROM workers WHERE active = TRUE`)
	if err != nil {
		return Worker{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var w Worker
		if err := rows.Scan(&w.ID, &w.WorkerName, &w.PIN, &w.Phone); err != nil {
			return Worker{}, err
		}
		if w.PIN == pin {
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
	stmt := `SELECT id, worker_name, pin, phone, active FROM workers WHERE id = ?`

	var w Worker
	err := m.DB.QueryRow(stmt, id).Scan(&w.ID, &w.WorkerName, &w.PIN, &w.Phone, &w.Active)
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
	rows, err := m.DB.Query(`SELECT id, worker_name, pin, phone, active FROM workers ORDER BY worker_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Worker
	for rows.Next() {
		var w Worker
		if err := rows.Scan(&w.ID, &w.WorkerName, &w.PIN, &w.Phone, &w.Active); err != nil {
			return nil, err
		}
		out = append(out, w)
	}

	return out, rows.Err()
}

// Update changes a worker's name, PIN, and phone. Active status is changed
// separately via SetActive. Returns ErrDuplicatePIN if another worker
// (active or not) already has that PIN.
func (m *WorkerModel) Update(id int, name, pin, phone string) error {
	inUse, err := m.pinInUse(pin, id)
	if err != nil {
		return err
	}
	if inUse {
		return ErrDuplicatePIN
	}

	result, err := m.DB.Exec(`UPDATE workers SET worker_name = ?, pin = ?, phone = ? WHERE id = ?`, name, pin, phone, id)
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

// Create adds a new worker. Returns ErrDuplicatePIN if another worker
// (active or not) already has that PIN.
func (m *WorkerModel) Create(name, pin, phone string) (int, error) {
	inUse, err := m.pinInUse(pin, 0)
	if err != nil {
		return 0, err
	}
	if inUse {
		return 0, ErrDuplicatePIN
	}

	stmt := `INSERT INTO workers (worker_name, pin, phone, active) VALUES (?, ?, ?, TRUE)`
	result, err := m.DB.Exec(stmt, name, pin, phone)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// pinInUse reports whether some worker other than excludeID already has
// pin. Pass 0 for excludeID when checking a new worker. PINs must be
// unique across all workers, active or not, so a reactivated worker can't
// collide with a PIN issued while it was inactive.
func (m *WorkerModel) pinInUse(pin string, excludeID int) (bool, error) {
	var exists bool
	err := m.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM workers WHERE pin = ? AND id != ?)`, pin, excludeID).Scan(&exists)
	return exists, err
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
