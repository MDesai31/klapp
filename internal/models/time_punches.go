package models

import (
	"database/sql"
	"errors"
	"fmt"
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
	NonCompliant    bool
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

// StatusLabel renders this row's status for the admin dashboard, including
// elapsed time for an open punch or total worked time for a closed one.
func (r DashboardRow) StatusLabel() string {
	if !r.StartTime.Valid {
		return "Not in"
	}
	start, err := time.Parse(time.RFC3339, r.StartTime.String)
	if err != nil {
		return "?"
	}
	if !r.EndTime.Valid {
		return fmt.Sprintf("In since %s (%s)", start.Local().Format("3:04 PM"), formatDuration(time.Since(start)))
	}
	end, err := time.Parse(time.RFC3339, r.EndTime.String)
	if err != nil {
		return "?"
	}
	return fmt.Sprintf("Out at %s (%s worked)", end.Local().Format("3:04 PM"), formatDuration(end.Sub(start)))
}

// formatDuration renders a duration as "Xh Ym", or just "Ym" under an hour.
func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h == 0 {
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%dh %dm", h, m)
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

// PayPeriodDays returns the 14 day strings ("2006-01-02") that make up the
// pay period starting at payPeriod, for the admin summary tab's columns.
func PayPeriodDays(payPeriod string) ([]string, error) {
	start, err := time.ParseInLocation("2006-01-02", payPeriod, time.Local)
	if err != nil {
		return nil, err
	}

	days := make([]string, 14)
	for i := range days {
		days[i] = start.AddDate(0, 0, i).Format("2006-01-02")
	}
	return days, nil
}

// DayLabel formats a "2006-01-02" date as a short column header, e.g.
// "Mon 6/8".
func DayLabel(date string) string {
	t, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return date
	}
	return t.Format("Mon 1/2")
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTimePunch(row scanner) (TimePunch, error) {
	var p TimePunch
	var start string

	err := row.Scan(&p.ID, &p.WorkerID, &p.PayPeriod, &p.Day, &start, &p.EndTime,
		&p.StartLat, &p.StartLon, &p.EndLat, &p.EndLon, &p.Late, &p.ModifiedByAdmin, &p.NonCompliant)
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
	start_lat, start_lon, end_lat, end_lon, late, modified_by_admin, non_compliant`

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
// day and pay_period are recalculated from startTime so moving a punch
// across a day boundary is reflected correctly.
func (m *TimePunchModel) AdminUpdate(id int, startTime, endTime time.Time) error {
	var end any
	if !endTime.IsZero() {
		end = endTime.UTC().Format(time.RFC3339)
	}

	day := startTime.Local().Format("2006-01-02")
	payPeriod := payPeriodStart(startTime.Local()).Format("2006-01-02")

	stmt := `UPDATE time_punches SET start_time = ?, end_time = ?, day = ?, pay_period = ?, modified_by_admin = TRUE WHERE id = ?`
	result, err := m.DB.Exec(stmt, startTime.UTC().Format(time.RFC3339), end, day, payPeriod, id)
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
// Non-compliant punches sort first so they appear at the top.
func (m *TimePunchModel) ForPayPeriod(payPeriod string) ([]TimesheetRow, error) {
	stmt := `SELECT tp.id, tp.worker_id, w.worker_name, tp.pay_period, tp.day, tp.start_time, tp.end_time,
		tp.start_lat, tp.start_lon, tp.end_lat, tp.end_lon, tp.late, tp.modified_by_admin, tp.non_compliant
		FROM time_punches tp
		JOIN workers w ON w.id = tp.worker_id
		WHERE tp.pay_period = ?
		ORDER BY tp.non_compliant DESC, tp.day, w.worker_name`

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
			&r.StartLat, &r.StartLon, &r.EndLat, &r.EndLon, &r.Late, &r.ModifiedByAdmin, &r.NonCompliant)
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

// PayPeriodSummaryRow is one worker's total worked time per day of a pay
// period, for the admin summary tab.
type PayPeriodSummaryRow struct {
	WorkerID   int
	WorkerName string
	HourlyRate float64
	Days       []DaySummary // one per day of the pay period, in order
	Total      time.Duration
}

// TotalDisplay renders the row's pay-period total, or "-" if the worker
// has no punches in the period at all.
func (r PayPeriodSummaryRow) TotalDisplay() string {
	if r.Total == 0 {
		return "-"
	}
	return formatDuration(r.Total)
}

// SalaryDisplay renders the pre-tax salary owed for the period, or "-" if
// the worker has no hourly rate configured or no hours worked.
func (r PayPeriodSummaryRow) SalaryDisplay() string {
	if r.HourlyRate == 0 || r.Total == 0 {
		return "-"
	}
	return fmt.Sprintf("$%.2f", r.Total.Hours()*r.HourlyRate)
}

// DaySummary is a worker's worked time on a single day within a pay
// period summary. PunchIDs is usually one punch, but can hold more than
// one if the worker punched in and out twice in a day.
type DaySummary struct {
	Date     string
	Worked   time.Duration
	PunchIDs []int
}

// Display renders the day's worked time for the summary grid, or "-" if
// the worker didn't punch in that day.
func (d DaySummary) Display() string {
	if len(d.PunchIDs) == 0 {
		return "-"
	}
	return formatDuration(d.Worked)
}

// PayPeriodSummary builds the admin summary tab's grid: every active
// worker, plus any inactive worker with punches in this period (so past
// summaries stay complete after someone is deactivated), each with their
// worked time broken down by day.
func (m *TimePunchModel) PayPeriodSummary(payPeriod string) ([]PayPeriodSummaryRow, error) {
	days, err := PayPeriodDays(payPeriod)
	if err != nil {
		return nil, err
	}
	dayIndex := make(map[string]int, len(days))
	for i, d := range days {
		dayIndex[d] = i
	}

	stmt := `SELECT w.id, w.worker_name, w.hourly_rate, tp.id, tp.day, tp.start_time, tp.end_time
		FROM workers w
		LEFT JOIN time_punches tp ON tp.worker_id = w.id AND tp.pay_period = ?
		WHERE w.active = TRUE OR tp.id IS NOT NULL
		ORDER BY w.worker_name, tp.day, tp.start_time`

	rows, err := m.DB.Query(stmt, payPeriod)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var order []int
	byWorker := make(map[int]*PayPeriodSummaryRow)

	for rows.Next() {
		var workerID int
		var workerName string
		var hourlyRate float64
		var punchID sql.NullInt64
		var day, start, end sql.NullString
		if err := rows.Scan(&workerID, &workerName, &hourlyRate, &punchID, &day, &start, &end); err != nil {
			return nil, err
		}

		r, ok := byWorker[workerID]
		if !ok {
			r = &PayPeriodSummaryRow{WorkerID: workerID, WorkerName: workerName, HourlyRate: hourlyRate, Days: make([]DaySummary, len(days))}
			for i, d := range days {
				r.Days[i].Date = d
			}
			byWorker[workerID] = r
			order = append(order, workerID)
		}

		if !punchID.Valid {
			continue
		}
		idx, ok := dayIndex[day.String]
		if !ok {
			continue
		}

		startTime, err := time.Parse(time.RFC3339, start.String)
		if err != nil {
			return nil, err
		}
		worked := time.Since(startTime)
		if end.Valid {
			endTime, err := time.Parse(time.RFC3339, end.String)
			if err != nil {
				return nil, err
			}
			worked = endTime.Sub(startTime)
		}

		r.Days[idx].Worked += worked
		r.Days[idx].PunchIDs = append(r.Days[idx].PunchIDs, int(punchID.Int64))
		r.Total += worked
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]PayPeriodSummaryRow, len(order))
	for i, id := range order {
		out[i] = *byWorker[id]
	}
	return out, nil
}

// AutoPunchOutNonCompliant closes every open punch whose start_time is
// before cutoff, sets non_compliant = TRUE, and returns the number of rows
// updated. Called nightly at 9 PM.
func (m *TimePunchModel) AutoPunchOutNonCompliant(cutoff time.Time) (int, error) {
	cutoffStr := cutoff.UTC().Format(time.RFC3339)
	result, err := m.DB.Exec(`
		UPDATE time_punches SET end_time = ?, non_compliant = TRUE
		WHERE end_time IS NULL AND start_time < ?`,
		cutoffStr, cutoffStr)
	if err != nil {
		return 0, err
	}
	n, err := result.RowsAffected()
	return int(n), err
}

// Delete removes a punch by ID. ErrNoRecord if it doesn't exist.
func (m *TimePunchModel) Delete(id int) error {
	result, err := m.DB.Exec(`DELETE FROM time_punches WHERE id = ?`, id)
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

// AdminCreate inserts a new punch on behalf of an admin (no GPS coords).
// A zero endTime leaves the punch open.
func (m *TimePunchModel) AdminCreate(workerID int, startTime, endTime time.Time) (int, error) {
	day := startTime.Local().Format("2006-01-02")
	payPeriod := payPeriodStart(startTime.Local()).Format("2006-01-02")

	var end any
	if !endTime.IsZero() {
		end = endTime.UTC().Format(time.RFC3339)
	}

	stmt := `INSERT INTO time_punches
		(worker_id, pay_period, day, start_time, end_time, start_lat, start_lon, modified_by_admin)
		VALUES (?, ?, ?, ?, ?, 0, 0, TRUE)`
	result, err := m.DB.Exec(stmt, workerID, payPeriod, day, startTime.UTC().Format(time.RFC3339), end)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}
