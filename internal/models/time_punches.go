package models

import (
	"database/sql"
	"errors"
	"time"
)

type TimePunch struct {
	ID              int
	WorkerID        int
	PayPeriod       string
	Day             string
	StartTime       time.Time
	EndTime         sql.NullString // RFC3339 string; empty/invalid until punched out
	StartLat        float64
	StartLon        float64
	EndLat          sql.NullFloat64
	EndLon          sql.NullFloat64
	Late            bool
	ModifiedByAdmin bool
}

// TimesheetRow is a punch joined with the worker's name, for admin views.
type TimesheetRow struct {
	TimePunch
	WorkerName string
}

// EndTimeDisplay renders the end time for admin templates, or "-" if the
// punch is still open.
func (p TimePunch) EndTimeDisplay() string {
	if !p.EndTime.Valid {
		return "-"
	}
	t, err := time.Parse(time.RFC3339, p.EndTime.String)
	if err != nil {
		return "-"
	}
	return t.Local().Format("3:04 PM")
}

// StartTimeInputValue and EndTimeInputValue format the punch for an
// <input type="datetime-local"> on the admin edit form.
func (p TimePunch) StartTimeInputValue() string {
	return p.StartTime.Local().Format("2006-01-02T15:04")
}

func (p TimePunch) EndTimeInputValue() string {
	if !p.EndTime.Valid {
		return ""
	}
	t, err := time.Parse(time.RFC3339, p.EndTime.String)
	if err != nil {
		return ""
	}
	return t.Local().Format("2006-01-02T15:04")
}

// DashboardRow is one active worker's punch status for a given day. Punch
// fields are unset (NULL) if the worker hasn't punched in that day.
type DashboardRow struct {
	WorkerID   int
	WorkerName string
	PunchID    sql.NullInt64
	StartTime  sql.NullString
	EndTime    sql.NullString
}

// StatusLabel renders this row's status for the admin dashboard.
func (r DashboardRow) StatusLabel() string {
	if !r.StartTime.Valid {
		return "Not in"
	}
	start, err := time.Parse(time.RFC3339, r.StartTime.String)
	if err != nil {
		return "?"
	}
	if !r.EndTime.Valid {
		return "In since " + start.Local().Format("3:04 PM")
	}
	end, err := time.Parse(time.RFC3339, r.EndTime.String)
	if err != nil {
		return "?"
	}
	return "Out at " + end.Local().Format("3:04 PM")
}

type TimePunchModel struct {
	DB *sql.DB
}

// payPeriodAnchor is the start of a known pay period (a Monday). Pay
// periods are fixed 14-day blocks counted forward from this date.
var payPeriodAnchor = time.Date(2026, 6, 8, 0, 0, 0, 0, time.Local)

func payPeriodStart(t time.Time) time.Time {
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	days := int(t.Sub(payPeriodAnchor).Hours() / 24)
	periodIndex := days / 14
	if days < 0 && days%14 != 0 {
		periodIndex--
	}
	return payPeriodAnchor.AddDate(0, 0, periodIndex*14)
}

// CurrentPayPeriod returns the start date ("2006-01-02") of the pay
// period containing t, for defaulting the admin timesheet view.
func CurrentPayPeriod(t time.Time) string {
	return payPeriodStart(t).Format("2006-01-02")
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTimePunch(row scanner) (TimePunch, error) {
	var p TimePunch
	var start string

	err := row.Scan(&p.ID, &p.WorkerID, &p.PayPeriod, &p.Day, &start, &p.EndTime,
		&p.StartLat, &p.StartLon, &p.EndLat, &p.EndLon, &p.Late, &p.ModifiedByAdmin)
	if err != nil {
		return TimePunch{}, err
	}

	p.StartTime, err = time.Parse(time.RFC3339, start)
	if err != nil {
		return TimePunch{}, err
	}

	return p, nil
}

const timePunchColumns = `id, worker_id, pay_period, day, start_time, end_time,
	start_lat, start_lon, end_lat, end_lon, late, modified_by_admin`

// Open returns the worker's punch that has no end_time yet, if any.
func (m *TimePunchModel) Open(workerID int) (TimePunch, error) {
	stmt := `SELECT ` + timePunchColumns + ` FROM time_punches WHERE worker_id = ? AND end_time IS NULL`

	p, err := scanTimePunch(m.DB.QueryRow(stmt, workerID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return TimePunch{}, ErrNoRecord
		}
		return TimePunch{}, err
	}

	return p, nil
}

// Get fetches a single punch by ID, for the admin edit form.
func (m *TimePunchModel) Get(id int) (TimePunch, error) {
	stmt := `SELECT ` + timePunchColumns + ` FROM time_punches WHERE id = ?`

	p, err := scanTimePunch(m.DB.QueryRow(stmt, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return TimePunch{}, ErrNoRecord
		}
		return TimePunch{}, err
	}

	return p, nil
}

// PunchIn records a new punch-in. It refuses to create one if the worker
// already has an open punch.
func (m *TimePunchModel) PunchIn(workerID int, lat, lon float64, now time.Time) (int, error) {
	if _, err := m.Open(workerID); err == nil {
		return 0, ErrAlreadyOpen
	} else if !errors.Is(err, ErrNoRecord) {
		return 0, err
	}

	day := now.Format("2006-01-02")
	payPeriod := payPeriodStart(now).Format("2006-01-02")

	stmt := `INSERT INTO time_punches (worker_id, pay_period, day, start_time, start_lat, start_lon)
		VALUES (?, ?, ?, ?, ?, ?)`
	result, err := m.DB.Exec(stmt, workerID, payPeriod, day, now.UTC().Format(time.RFC3339), lat, lon)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// PunchOut closes the worker's open punch. ErrNoRecord means there was
// nothing open to close.
func (m *TimePunchModel) PunchOut(workerID int, lat, lon float64, now time.Time) error {
	stmt := `UPDATE time_punches SET end_time = ?, end_lat = ?, end_lon = ?
		WHERE worker_id = ? AND end_time IS NULL`
	result, err := m.DB.Exec(stmt, now.UTC().Format(time.RFC3339), lat, lon, workerID)
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

// PunchOutLate closes the worker's open punch via the late-notice link.
// No location is captured for a late punch-out - the worker is just
// entering a finish time after the fact.
func (m *TimePunchModel) PunchOutLate(workerID int, endTime time.Time) error {
	stmt := `UPDATE time_punches SET end_time = ?, late = TRUE
		WHERE worker_id = ? AND end_time IS NULL`
	result, err := m.DB.Exec(stmt, endTime.UTC().Format(time.RFC3339), workerID)
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

// AdminUpdate overwrites a punch's start/end time as corrected by an
// admin. A zero endTime clears it back to open (still punched in).
func (m *TimePunchModel) AdminUpdate(id int, startTime, endTime time.Time) error {
	var end any
	if !endTime.IsZero() {
		end = endTime.UTC().Format(time.RFC3339)
	}

	stmt := `UPDATE time_punches SET start_time = ?, end_time = ?, modified_by_admin = TRUE WHERE id = ?`
	result, err := m.DB.Exec(stmt, startTime.UTC().Format(time.RFC3339), end, id)
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

// DashboardStatus reports every active worker's punch status for the
// given day ("2006-01-02") - in, out, or not punched in at all.
func (m *TimePunchModel) DashboardStatus(day string) ([]DashboardRow, error) {
	stmt := `SELECT w.id, w.worker_name, tp.id, tp.start_time, tp.end_time
		FROM workers w
		LEFT JOIN time_punches tp ON tp.worker_id = w.id AND tp.day = ?
		WHERE w.active = TRUE
		ORDER BY w.worker_name`

	rows, err := m.DB.Query(stmt, day)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []DashboardRow
	for rows.Next() {
		var r DashboardRow
		if err := rows.Scan(&r.WorkerID, &r.WorkerName, &r.PunchID, &r.StartTime, &r.EndTime); err != nil {
			return nil, err
		}
		out = append(out, r)
	}

	return out, rows.Err()
}

// ForPayPeriod lists every punch in the given pay period (its start date,
// "2006-01-02"), across all workers, for the admin timesheet.
func (m *TimePunchModel) ForPayPeriod(payPeriod string) ([]TimesheetRow, error) {
	stmt := `SELECT tp.id, tp.worker_id, w.worker_name, tp.pay_period, tp.day, tp.start_time, tp.end_time,
		tp.start_lat, tp.start_lon, tp.end_lat, tp.end_lon, tp.late, tp.modified_by_admin
		FROM time_punches tp
		JOIN workers w ON w.id = tp.worker_id
		WHERE tp.pay_period = ?
		ORDER BY tp.day, w.worker_name`

	rows, err := m.DB.Query(stmt, payPeriod)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []TimesheetRow
	for rows.Next() {
		var r TimesheetRow
		var start string
		err := rows.Scan(&r.ID, &r.WorkerID, &r.WorkerName, &r.PayPeriod, &r.Day, &start, &r.EndTime,
			&r.StartLat, &r.StartLon, &r.EndLat, &r.EndLon, &r.Late, &r.ModifiedByAdmin)
		if err != nil {
			return nil, err
		}
		r.StartTime, err = time.Parse(time.RFC3339, start)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}

	return out, rows.Err()
}

// PayPeriods lists every pay period that has at least one punch, most
// recent first - used to let an admin browse pay period history.
func (m *TimePunchModel) PayPeriods() ([]string, error) {
	rows, err := m.DB.Query(`SELECT DISTINCT pay_period FROM time_punches ORDER BY pay_period DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		out = append(out, p)
	}

	return out, rows.Err()
}
