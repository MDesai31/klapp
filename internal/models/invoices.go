package models

import (
	"database/sql"
	"errors"
)

type Invoice struct {
	ID           int
	SubmittedBy  int
	WorkerName   string
	Date         string
	HouseNumber  string
	CustomerName string
	CustomerID   *int
	TimeArrived  string
	TimeLeft     string
	NoOfWorkers  int
	Comments     string
	Reviewed     bool
	CreatedAt    string
	Jobs         []string
	Materials    []string
}

type InvoiceModel struct {
	DB *sql.DB
}

func (m *InvoiceModel) Create(
	submittedBy int,
	date, houseNumber, customerName string,
	customerID *int,
	timeArrived, timeLeft string,
	noOfWorkers int,
	comments string,
	jobs, materials []string,
) (int, error) {
	tx, err := m.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		`INSERT INTO invoices (submitted_by, date, house_number, customer_name, customer_id, time_arrived, time_left, no_of_workers, comments)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		submittedBy, date, houseNumber, customerName, customerID, timeArrived, timeLeft, noOfWorkers, comments,
	)
	if err != nil {
		return 0, err
	}

	invoiceID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, j := range jobs {
		if _, err := tx.Exec(`INSERT INTO invoice_jobs (invoice_id, description) VALUES (?, ?)`, invoiceID, j); err != nil {
			return 0, err
		}
	}

	for _, mat := range materials {
		if _, err := tx.Exec(`INSERT INTO invoice_materials_used (invoice_id, material) VALUES (?, ?)`, invoiceID, mat); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return int(invoiceID), nil
}

const invoicePageSize = 25

// List returns one page of invoices (newest first) and the total count.
func (m *InvoiceModel) List(page int) ([]Invoice, int, error) {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * invoicePageSize

	var total int
	if err := m.DB.QueryRow(`SELECT COUNT(*) FROM invoices`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := m.DB.Query(
		`SELECT i.id, i.submitted_by, w.worker_name, i.date, i.house_number, i.customer_name,
		        i.time_arrived, i.time_left, i.no_of_workers, i.reviewed, i.created_at
		 FROM invoices i
		 JOIN workers w ON w.id = i.submitted_by
		 ORDER BY i.date DESC, i.id DESC
		 LIMIT ? OFFSET ?`,
		invoicePageSize, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []Invoice
	for rows.Next() {
		var inv Invoice
		if err := rows.Scan(
			&inv.ID, &inv.SubmittedBy, &inv.WorkerName, &inv.Date,
			&inv.HouseNumber, &inv.CustomerName,
			&inv.TimeArrived, &inv.TimeLeft, &inv.NoOfWorkers,
			&inv.Reviewed, &inv.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, inv)
	}
	return out, total, rows.Err()
}

func (m *InvoiceModel) Get(id int) (Invoice, error) {
	var inv Invoice
	var customerID sql.NullInt64
	err := m.DB.QueryRow(
		`SELECT i.id, i.submitted_by, w.worker_name, i.date, i.house_number, i.customer_name,
		        i.customer_id, i.time_arrived, i.time_left, i.no_of_workers, i.comments, i.reviewed, i.created_at
		 FROM invoices i
		 JOIN workers w ON w.id = i.submitted_by
		 WHERE i.id = ?`,
		id,
	).Scan(
		&inv.ID, &inv.SubmittedBy, &inv.WorkerName, &inv.Date,
		&inv.HouseNumber, &inv.CustomerName, &customerID,
		&inv.TimeArrived, &inv.TimeLeft, &inv.NoOfWorkers,
		&inv.Comments, &inv.Reviewed, &inv.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Invoice{}, ErrNoRecord
		}
		return Invoice{}, err
	}
	if customerID.Valid {
		v := int(customerID.Int64)
		inv.CustomerID = &v
	}

	jobs, err := m.loadJobs(id)
	if err != nil {
		return Invoice{}, err
	}
	inv.Jobs = jobs

	mats, err := m.loadMaterials(id)
	if err != nil {
		return Invoice{}, err
	}
	inv.Materials = mats

	return inv, nil
}

func (m *InvoiceModel) SetReviewed(id int) error {
	result, err := m.DB.Exec(`UPDATE invoices SET reviewed = TRUE WHERE id = ?`, id)
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

func (m *InvoiceModel) ListByCustomer(customerID int) ([]Invoice, error) {
	rows, err := m.DB.Query(
		`SELECT i.id, i.submitted_by, w.worker_name, i.date, i.house_number, i.customer_name,
		        i.time_arrived, i.time_left, i.no_of_workers, i.reviewed, i.created_at
		 FROM invoices i
		 JOIN workers w ON w.id = i.submitted_by
		 WHERE i.customer_id = ?
		 ORDER BY i.date DESC, i.id DESC`,
		customerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Invoice
	for rows.Next() {
		var inv Invoice
		if err := rows.Scan(
			&inv.ID, &inv.SubmittedBy, &inv.WorkerName, &inv.Date,
			&inv.HouseNumber, &inv.CustomerName,
			&inv.TimeArrived, &inv.TimeLeft, &inv.NoOfWorkers,
			&inv.Reviewed, &inv.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, inv)
	}
	return out, rows.Err()
}

func (m *InvoiceModel) loadJobs(invoiceID int) ([]string, error) {
	rows, err := m.DB.Query(`SELECT description FROM invoice_jobs WHERE invoice_id = ? ORDER BY id`, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (m *InvoiceModel) loadMaterials(invoiceID int) ([]string, error) {
	rows, err := m.DB.Query(`SELECT material FROM invoice_materials_used WHERE invoice_id = ? ORDER BY id`, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
