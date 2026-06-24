package models

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestPunchInAndOut(t *testing.T) {
	db := newTestDB(t)
	tm := &TimePunchModel{DB: db}
	workerID := mustInsertWorker(t, db, "Manthan", "1234", true)

	now := time.Date(2026, 6, 17, 8, 0, 0, 0, time.Local)

	if _, err := tm.PunchIn(workerID, 40.0, -73.0, now); err != nil {
		t.Fatalf("PunchIn: %v", err)
	}

	open, err := tm.Open(workerID)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if open.WorkerID != workerID {
		t.Errorf("got worker %d, want %d", open.WorkerID, workerID)
	}
	if open.EndTime.Valid {
		t.Errorf("expected no end time yet, got %v", open.EndTime)
	}

	if _, err := tm.PunchIn(workerID, 40.0, -73.0, now); !errors.Is(err, ErrAlreadyOpen) {
		t.Errorf("got error %v, want ErrAlreadyOpen", err)
	}

	if err := tm.PunchOut(workerID, 40.0, -73.0, now.Add(8*time.Hour)); err != nil {
		t.Fatalf("PunchOut: %v", err)
	}

	if _, err := tm.Open(workerID); !errors.Is(err, ErrNoRecord) {
		t.Errorf("got error %v, want ErrNoRecord after punch out", err)
	}

	if err := tm.PunchOut(workerID, 40.0, -73.0, now); !errors.Is(err, ErrNoRecord) {
		t.Errorf("got error %v, want ErrNoRecord for a second punch out", err)
	}
}

func TestPunchOutLate(t *testing.T) {
	db := newTestDB(t)
	tm := &TimePunchModel{DB: db}
	workerID := mustInsertWorker(t, db, "Thomas", "4321", true)

	now := time.Date(2026, 6, 17, 8, 0, 0, 0, time.Local)
	if _, err := tm.PunchIn(workerID, 40.0, -73.0, now); err != nil {
		t.Fatalf("PunchIn: %v", err)
	}

	endTime := time.Date(2026, 6, 17, 21, 30, 0, 0, time.Local)
	if err := tm.PunchOutLate(workerID, endTime); err != nil {
		t.Fatalf("PunchOutLate: %v", err)
	}

	if _, err := tm.Open(workerID); !errors.Is(err, ErrNoRecord) {
		t.Errorf("got error %v, want ErrNoRecord after late punch out", err)
	}

	var late bool
	err := db.QueryRow(`SELECT late FROM time_punches WHERE worker_id = ?`, workerID).Scan(&late)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if !late {
		t.Error("expected late = true")
	}
}

func TestPunchOutLateWithNoOpenPunch(t *testing.T) {
	db := newTestDB(t)
	tm := &TimePunchModel{DB: db}
	workerID := mustInsertWorker(t, db, "Idle Worker", "0000", true)

	err := tm.PunchOutLate(workerID, time.Now())
	if !errors.Is(err, ErrNoRecord) {
		t.Errorf("got error %v, want ErrNoRecord", err)
	}
}

func TestPayPeriodStart(t *testing.T) {
	anchor := payPeriodAnchor

	tests := []struct {
		name string
		day  time.Time
		want time.Time
	}{
		{"anchor day itself", anchor, anchor},
		{"one day into the period", anchor.AddDate(0, 0, 1), anchor},
		{"last day of the period", anchor.AddDate(0, 0, 13), anchor},
		{"first day of next period", anchor.AddDate(0, 0, 14), anchor.AddDate(0, 0, 14)},
		{"one period before anchor", anchor.AddDate(0, 0, -1), anchor.AddDate(0, 0, -14)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := payPeriodStart(tt.day)
			if !got.Equal(tt.want) {
				t.Errorf("payPeriodStart(%v) = %v, want %v", tt.day, got, tt.want)
			}
		})
	}
}

func TestDashboardStatus(t *testing.T) {
	db := newTestDB(t)
	tm := &TimePunchModel{DB: db}

	in := mustInsertWorker(t, db, "Punched In", "1111", true)
	out := mustInsertWorker(t, db, "Punched Out", "2222", true)
	notIn := mustInsertWorker(t, db, "Not In", "3333", true)
	mustInsertWorker(t, db, "Inactive", "4444", false)

	now := time.Date(2026, 6, 17, 8, 0, 0, 0, time.Local)
	day := now.Format("2006-01-02")

	if _, err := tm.PunchIn(in, 40.0, -73.0, now); err != nil {
		t.Fatalf("PunchIn: %v", err)
	}
	if _, err := tm.PunchIn(out, 40.0, -73.0, now); err != nil {
		t.Fatalf("PunchIn: %v", err)
	}
	if err := tm.PunchOut(out, 40.0, -73.0, now.Add(time.Hour)); err != nil {
		t.Fatalf("PunchOut: %v", err)
	}

	rows, err := tm.DashboardStatus(day)
	if err != nil {
		t.Fatalf("DashboardStatus: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("got %d rows, want 3 (inactive worker excluded)", len(rows))
	}

	status := map[int]DashboardRow{}
	for _, r := range rows {
		status[r.WorkerID] = r
	}

	if !status[in].StartTime.Valid || status[in].EndTime.Valid {
		t.Errorf("worker %d: want punched in with no end time, got %+v", in, status[in])
	}
	if !status[out].StartTime.Valid || !status[out].EndTime.Valid {
		t.Errorf("worker %d: want punched out, got %+v", out, status[out])
	}
	if status[notIn].StartTime.Valid {
		t.Errorf("worker %d: want no punch at all, got %+v", notIn, status[notIn])
	}
}

func TestDashboardRowStatusLabel(t *testing.T) {
	t.Run("not in", func(t *testing.T) {
		r := DashboardRow{}
		if got, want := r.StatusLabel(), "Not in"; got != want {
			t.Errorf("StatusLabel() = %q, want %q", got, want)
		}
	})

	t.Run("punched out", func(t *testing.T) {
		start := time.Date(2026, 6, 17, 8, 0, 0, 0, time.Local)
		end := start.Add(7*time.Hour + 45*time.Minute)
		r := DashboardRow{
			StartTime: sql.NullString{String: start.UTC().Format(time.RFC3339), Valid: true},
			EndTime:   sql.NullString{String: end.UTC().Format(time.RFC3339), Valid: true},
		}
		want := fmt.Sprintf("Out at %s (7h 45m worked)", end.Format("3:04 PM"))
		if got := r.StatusLabel(); got != want {
			t.Errorf("StatusLabel() = %q, want %q", got, want)
		}
	})

	t.Run("punched in", func(t *testing.T) {
		start := time.Now().Add(-90 * time.Minute)
		r := DashboardRow{
			StartTime: sql.NullString{String: start.UTC().Format(time.RFC3339), Valid: true},
		}
		want := fmt.Sprintf("In since %s (1h 30m)", start.Format("3:04 PM"))
		if got := r.StatusLabel(); got != want {
			t.Errorf("StatusLabel() = %q, want %q", got, want)
		}
	})
}

func TestForPayPeriodAndPayPeriods(t *testing.T) {
	db := newTestDB(t)
	tm := &TimePunchModel{DB: db}
	workerID := mustInsertWorker(t, db, "Manthan", "1234", true)

	week1 := payPeriodAnchor.Add(8 * time.Hour)
	week3 := payPeriodAnchor.AddDate(0, 0, 14).Add(8 * time.Hour)

	if _, err := tm.PunchIn(workerID, 40.0, -73.0, week1); err != nil {
		t.Fatalf("PunchIn: %v", err)
	}
	if err := tm.PunchOut(workerID, 40.0, -73.0, week1.Add(time.Hour)); err != nil {
		t.Fatalf("PunchOut: %v", err)
	}
	if _, err := tm.PunchIn(workerID, 40.0, -73.0, week3); err != nil {
		t.Fatalf("PunchIn: %v", err)
	}

	periods, err := tm.PayPeriods()
	if err != nil {
		t.Fatalf("PayPeriods: %v", err)
	}
	wantPeriods := []string{
		payPeriodStart(week3).Format("2006-01-02"),
		payPeriodStart(week1).Format("2006-01-02"),
	}
	if len(periods) != 2 || periods[0] != wantPeriods[0] || periods[1] != wantPeriods[1] {
		t.Errorf("got periods %v, want %v (most recent first)", periods, wantPeriods)
	}

	rows, err := tm.ForPayPeriod(payPeriodStart(week1).Format("2006-01-02"))
	if err != nil {
		t.Fatalf("ForPayPeriod: %v", err)
	}
	if len(rows) != 1 || rows[0].WorkerName != "Manthan" {
		t.Errorf("got %+v, want a single row for Manthan", rows)
	}
}

func TestAdminUpdate(t *testing.T) {
	db := newTestDB(t)
	tm := &TimePunchModel{DB: db}
	workerID := mustInsertWorker(t, db, "Manthan", "1234", true)

	now := time.Date(2026, 6, 17, 8, 0, 0, 0, time.Local)
	id, err := tm.PunchIn(workerID, 40.0, -73.0, now)
	if err != nil {
		t.Fatalf("PunchIn: %v", err)
	}

	correctedStart := now.Add(-15 * time.Minute)
	correctedEnd := now.Add(8 * time.Hour)
	if err := tm.AdminUpdate(id, correctedStart, correctedEnd); err != nil {
		t.Fatalf("AdminUpdate: %v", err)
	}

	p, err := tm.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !p.ModifiedByAdmin {
		t.Error("want modified_by_admin = true after AdminUpdate")
	}
	if !p.StartTime.Equal(correctedStart.UTC().Truncate(time.Second)) {
		t.Errorf("got start time %v, want %v", p.StartTime, correctedStart)
	}
	if !p.EndTime.Valid {
		t.Error("want end time to be set")
	}

	if err := tm.AdminUpdate(id+999, correctedStart, correctedEnd); !errors.Is(err, ErrNoRecord) {
		t.Errorf("got error %v, want ErrNoRecord for unknown punch", err)
	}
}

func TestPayPeriodDays(t *testing.T) {
	days, err := PayPeriodDays("2026-06-08")
	if err != nil {
		t.Fatalf("PayPeriodDays: %v", err)
	}
	if len(days) != 14 {
		t.Fatalf("got %d days, want 14", len(days))
	}
	if days[0] != "2026-06-08" || days[13] != "2026-06-21" {
		t.Errorf("got first/last day %q/%q, want 2026-06-08/2026-06-21", days[0], days[13])
	}
}

func TestDayLabel(t *testing.T) {
	got := DayLabel("2026-06-08")
	want := time.Date(2026, 6, 8, 0, 0, 0, 0, time.Local).Format("Mon 1/2")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPayPeriodSummary(t *testing.T) {
	db := newTestDB(t)
	tm := &TimePunchModel{DB: db}

	period := payPeriodAnchor.Format("2006-01-02")
	day1 := payPeriodAnchor.Add(8 * time.Hour)
	day2 := payPeriodAnchor.AddDate(0, 0, 1).Add(8 * time.Hour)

	alice := mustInsertWorker(t, db, "Alice", "1111", true)
	mustInsertWorker(t, db, "Zane", "2222", true)
	bob := mustInsertWorker(t, db, "Bob", "3333", false)
	mustInsertWorker(t, db, "Yolanda", "4444", false) // inactive, no punches - excluded

	// Alice: two punches on day1 (in/out twice, 1h + 1h), one punch on day2 (1h).
	id1, err := tm.PunchIn(alice, 0, 0, day1)
	if err != nil {
		t.Fatalf("PunchIn: %v", err)
	}
	if err := tm.PunchOut(alice, 0, 0, day1.Add(time.Hour)); err != nil {
		t.Fatalf("PunchOut: %v", err)
	}
	id2, err := tm.PunchIn(alice, 0, 0, day1.Add(2*time.Hour))
	if err != nil {
		t.Fatalf("PunchIn: %v", err)
	}
	if err := tm.PunchOut(alice, 0, 0, day1.Add(3*time.Hour)); err != nil {
		t.Fatalf("PunchOut: %v", err)
	}
	if _, err := tm.PunchIn(alice, 0, 0, day2); err != nil {
		t.Fatalf("PunchIn: %v", err)
	}
	if err := tm.PunchOut(alice, 0, 0, day2.Add(time.Hour)); err != nil {
		t.Fatalf("PunchOut: %v", err)
	}

	// Bob (inactive): one punch on day1 (5h) - still included since it's in this period.
	if _, err := tm.PunchIn(bob, 0, 0, day1); err != nil {
		t.Fatalf("PunchIn: %v", err)
	}
	if err := tm.PunchOut(bob, 0, 0, day1.Add(5*time.Hour)); err != nil {
		t.Fatalf("PunchOut: %v", err)
	}

	rows, err := tm.PayPeriodSummary(period)
	if err != nil {
		t.Fatalf("PayPeriodSummary: %v", err)
	}

	wantNames := []string{"Alice", "Bob", "Zane"}
	if len(rows) != len(wantNames) {
		t.Fatalf("got %d worker rows, want %d (%v)", len(rows), len(wantNames), rows)
	}
	byName := make(map[string]PayPeriodSummaryRow, len(rows))
	for i, r := range rows {
		byName[r.WorkerName] = r
		if r.WorkerName != wantNames[i] {
			t.Errorf("got worker order %d = %q, want %q (alphabetical)", i, r.WorkerName, wantNames[i])
		}
	}

	aliceRow := byName["Alice"]
	if got, want := aliceRow.Days[0].Worked, 2*time.Hour; got != want {
		t.Errorf("Alice day0 worked = %v, want %v", got, want)
	}
	if got, want := aliceRow.Days[0].PunchIDs, []int{id1, id2}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("Alice day0 punch IDs = %v, want %v", got, want)
	}
	if got, want := aliceRow.Days[1].Worked, time.Hour; got != want {
		t.Errorf("Alice day1 worked = %v, want %v", got, want)
	}
	if got, want := aliceRow.Total, 3*time.Hour; got != want {
		t.Errorf("Alice total = %v, want %v", got, want)
	}
	if got := aliceRow.Days[2].Display(); got != "-" {
		t.Errorf("Alice day2 (no punch) = %q, want %q", got, "-")
	}

	bobRow := byName["Bob"]
	if got, want := bobRow.Total, 5*time.Hour; got != want {
		t.Errorf("Bob total = %v, want %v", got, want)
	}

	zaneRow := byName["Zane"]
	if got := zaneRow.TotalDisplay(); got != "-" {
		t.Errorf("Zane (no punches) total = %q, want %q", got, "-")
	}
	for i, d := range zaneRow.Days {
		if got := d.Display(); got != "-" {
			t.Errorf("Zane day %d = %q, want %q", i, got, "-")
		}
	}
}
