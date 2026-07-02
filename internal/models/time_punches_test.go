package models

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"
)

func fptr(f float64) *float64 { return &f }

func TestPunchInAndOut(t *testing.T) {
	db := newTestDB(t)
	tm := &TimePunchModel{DB: db}
	workerID := mustInsertWorker(t, db, "Manthan", "1234", true)

	now := time.Date(2026, 6, 17, 8, 0, 0, 0, time.Local)

	if _, err := tm.PunchIn(workerID, fptr(40.0), fptr(-73.0), now); err != nil {
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

	if _, err := tm.PunchIn(workerID, fptr(40.0), fptr(-73.0), now); !errors.Is(err, ErrAlreadyOpen) {
		t.Errorf("got error %v, want ErrAlreadyOpen", err)
	}

	if err := tm.PunchOut(workerID, fptr(40.0), fptr(-73.0), now.Add(8*time.Hour)); err != nil {
		t.Fatalf("PunchOut: %v", err)
	}

	if _, err := tm.Open(workerID); !errors.Is(err, ErrNoRecord) {
		t.Errorf("got error %v, want ErrNoRecord after punch out", err)
	}

	if err := tm.PunchOut(workerID, fptr(40.0), fptr(-73.0), now); !errors.Is(err, ErrNoRecord) {
		t.Errorf("got error %v, want ErrNoRecord for a second punch out", err)
	}
}

func TestPunchInDailyLimit(t *testing.T) {
	db := newTestDB(t)
	tm := &TimePunchModel{DB: db, DailyPunchLimit: 3}
	workerID := mustInsertWorker(t, db, "Manthan", "1234", true)

	now := time.Date(2026, 6, 17, 8, 0, 0, 0, time.Local)

	for i := 0; i < 3; i++ {
		start := now.Add(time.Duration(i) * time.Hour)
		if _, err := tm.PunchIn(workerID, fptr(40.0), fptr(-73.0), start); err != nil {
			t.Fatalf("PunchIn #%d: %v", i+1, err)
		}
		if err := tm.PunchOut(workerID, fptr(40.0), fptr(-73.0), start.Add(10*time.Minute)); err != nil {
			t.Fatalf("PunchOut #%d: %v", i+1, err)
		}
	}

	if _, err := tm.PunchIn(workerID, fptr(40.0), fptr(-73.0), now.Add(4*time.Hour)); !errors.Is(err, ErrDailyLimitExceeded) {
		t.Errorf("got error %v, want ErrDailyLimitExceeded", err)
	}

	// A punch the next day is unaffected - the limit is per calendar day.
	nextDay := now.AddDate(0, 0, 1)
	if _, err := tm.PunchIn(workerID, fptr(40.0), fptr(-73.0), nextDay); err != nil {
		t.Errorf("PunchIn on next day: %v", err)
	}
}

func TestPunchInNilCoordsStoresNull(t *testing.T) {
	db := newTestDB(t)
	tm := &TimePunchModel{DB: db}
	workerID := mustInsertWorker(t, db, "No-GPS Worker", "9876", true)

	now := time.Date(2026, 6, 17, 8, 0, 0, 0, time.Local)
	id, err := tm.PunchIn(workerID, nil, nil, now)
	if err != nil {
		t.Fatalf("PunchIn with nil coords: %v", err)
	}

	p, err := tm.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if p.StartLat.Valid || p.StartLon.Valid {
		t.Errorf("StartLat.Valid=%v StartLon.Valid=%v, want both false", p.StartLat.Valid, p.StartLon.Valid)
	}

	if err := tm.PunchOut(workerID, nil, nil, now.Add(8*time.Hour)); err != nil {
		t.Fatalf("PunchOut with nil coords: %v", err)
	}

	p, err = tm.Get(id)
	if err != nil {
		t.Fatalf("Get after PunchOut: %v", err)
	}
	if p.EndLat.Valid || p.EndLon.Valid {
		t.Errorf("EndLat.Valid=%v EndLon.Valid=%v, want both false", p.EndLat.Valid, p.EndLon.Valid)
	}
}

func TestPunchOutLate(t *testing.T) {
	db := newTestDB(t)
	tm := &TimePunchModel{DB: db}
	workerID := mustInsertWorker(t, db, "Thomas", "4321", true)

	now := time.Date(2026, 6, 17, 8, 0, 0, 0, time.Local)
	if _, err := tm.PunchIn(workerID, fptr(40.0), fptr(-73.0), now); err != nil {
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

	if _, err := tm.PunchIn(in, fptr(40.0), fptr(-73.0), now); err != nil {
		t.Fatalf("PunchIn: %v", err)
	}
	if _, err := tm.PunchIn(out, fptr(40.0), fptr(-73.0), now); err != nil {
		t.Fatalf("PunchIn: %v", err)
	}
	if err := tm.PunchOut(out, fptr(40.0), fptr(-73.0), now.Add(time.Hour)); err != nil {
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

	if _, err := tm.PunchIn(workerID, fptr(40.0), fptr(-73.0), week1); err != nil {
		t.Fatalf("PunchIn: %v", err)
	}
	if err := tm.PunchOut(workerID, fptr(40.0), fptr(-73.0), week1.Add(time.Hour)); err != nil {
		t.Fatalf("PunchOut: %v", err)
	}
	if _, err := tm.PunchIn(workerID, fptr(40.0), fptr(-73.0), week3); err != nil {
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
	id, err := tm.PunchIn(workerID, fptr(40.0), fptr(-73.0), now)
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

func TestAdminUpdateRemovesConflictingPunch(t *testing.T) {
	db := newTestDB(t)
	tm := &TimePunchModel{DB: db}
	workerID := mustInsertWorker(t, db, "Thomas", "9999", true)

	friday := time.Date(2026, 6, 19, 7, 0, 0, 0, time.Local)
	saturday := time.Date(2026, 6, 20, 8, 0, 0, 0, time.Local)

	// Existing Friday punch (correct day, already in DB).
	fridayID, err := tm.PunchIn(workerID, nil, nil, friday)
	if err != nil {
		t.Fatalf("PunchIn Friday: %v", err)
	}
	if err := tm.PunchOut(workerID, nil, nil, friday.Add(8*time.Hour)); err != nil {
		t.Fatalf("PunchOut Friday: %v", err)
	}

	// Saturday punch that the admin is moving to Friday (worker forgot to punch Friday).
	saturdayID, err := tm.PunchIn(workerID, nil, nil, saturday)
	if err != nil {
		t.Fatalf("PunchIn Saturday: %v", err)
	}
	if err := tm.PunchOut(workerID, nil, nil, saturday.Add(8*time.Hour)); err != nil {
		t.Fatalf("PunchOut Saturday: %v", err)
	}

	// Edit the Saturday punch → move it to Friday; it now overlaps the existing Friday entry.
	correctedStart := time.Date(2026, 6, 19, 8, 30, 0, 0, time.Local)
	correctedEnd := correctedStart.Add(8 * time.Hour)
	if err := tm.AdminUpdate(saturdayID, correctedStart, correctedEnd); err != nil {
		t.Fatalf("AdminUpdate: %v", err)
	}

	// The old Friday punch should be gone — replaced by the edited one.
	if _, err := tm.Get(fridayID); !errors.Is(err, ErrNoRecord) {
		t.Errorf("old Friday punch: got %v, want ErrNoRecord (should have been replaced)", err)
	}

	// The edited punch should still exist with the new times.
	p, err := tm.Get(saturdayID)
	if err != nil {
		t.Fatalf("Get edited punch: %v", err)
	}
	if p.Day != "2026-06-19" {
		t.Errorf("day = %q, want 2026-06-19", p.Day)
	}
}

func TestAdminUpdateDoesNotRemoveNonOverlappingPunch(t *testing.T) {
	db := newTestDB(t)
	tm := &TimePunchModel{DB: db}
	workerID := mustInsertWorker(t, db, "Thomas", "9999", true)

	// Two punches on different days with no time overlap.
	friday := time.Date(2026, 6, 19, 7, 0, 0, 0, time.Local)
	saturday := time.Date(2026, 6, 20, 8, 0, 0, 0, time.Local)

	fridayID, err := tm.PunchIn(workerID, nil, nil, friday)
	if err != nil {
		t.Fatalf("PunchIn Friday: %v", err)
	}
	if err := tm.PunchOut(workerID, nil, nil, friday.Add(8*time.Hour)); err != nil {
		t.Fatalf("PunchOut Friday: %v", err)
	}

	saturdayID, err := tm.PunchIn(workerID, nil, nil, saturday)
	if err != nil {
		t.Fatalf("PunchIn Saturday: %v", err)
	}
	if err := tm.PunchOut(workerID, nil, nil, saturday.Add(8*time.Hour)); err != nil {
		t.Fatalf("PunchOut Saturday: %v", err)
	}

	// Edit Saturday's start/end time but keep it on Saturday — no overlap with Friday.
	correctedStart := saturday.Add(30 * time.Minute)
	correctedEnd := correctedStart.Add(7 * time.Hour)
	if err := tm.AdminUpdate(saturdayID, correctedStart, correctedEnd); err != nil {
		t.Fatalf("AdminUpdate: %v", err)
	}

	// Friday punch must still be there.
	if _, err := tm.Get(fridayID); err != nil {
		t.Errorf("Friday punch should still exist after editing Saturday: %v", err)
	}
}

func TestAdminUpdateRecomputesDayAndPayPeriod(t *testing.T) {
	db := newTestDB(t)
	tm := &TimePunchModel{DB: db}
	workerID := mustInsertWorker(t, db, "Thomas", "9999", true)

	// Punch in on Saturday (2026-06-20, last day of period 1)
	saturday := time.Date(2026, 6, 20, 8, 0, 0, 0, time.Local)
	id, err := tm.PunchIn(workerID, nil, nil, saturday)
	if err != nil {
		t.Fatalf("PunchIn: %v", err)
	}
	p, err := tm.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if p.Day != "2026-06-20" {
		t.Fatalf("initial day = %q, want 2026-06-20", p.Day)
	}

	// Edit to Friday — should move the record to a different day
	friday := time.Date(2026, 6, 19, 8, 0, 0, 0, time.Local)
	if err := tm.AdminUpdate(id, friday, friday.Add(8*time.Hour)); err != nil {
		t.Fatalf("AdminUpdate: %v", err)
	}

	p, err = tm.Get(id)
	if err != nil {
		t.Fatalf("Get after update: %v", err)
	}
	if p.Day != "2026-06-19" {
		t.Errorf("day = %q, want 2026-06-19 (Friday)", p.Day)
	}
	if p.PayPeriod != payPeriodAnchor.Format("2006-01-02") {
		t.Errorf("pay_period = %q, want %q", p.PayPeriod, payPeriodAnchor.Format("2006-01-02"))
	}
}

func TestAdminUpdateCrossesPayPeriodBoundary(t *testing.T) {
	db := newTestDB(t)
	tm := &TimePunchModel{DB: db}
	workerID := mustInsertWorker(t, db, "Thomas", "9999", true)

	// Punch in on the first day of period 2 (2026-06-22)
	period2Start := time.Date(2026, 6, 22, 8, 0, 0, 0, time.Local)
	id, err := tm.PunchIn(workerID, nil, nil, period2Start)
	if err != nil {
		t.Fatalf("PunchIn: %v", err)
	}

	// Edit back to the last day of period 1 (2026-06-21)
	period1LastDay := time.Date(2026, 6, 21, 8, 0, 0, 0, time.Local)
	if err := tm.AdminUpdate(id, period1LastDay, period1LastDay.Add(8*time.Hour)); err != nil {
		t.Fatalf("AdminUpdate: %v", err)
	}

	p, err := tm.Get(id)
	if err != nil {
		t.Fatalf("Get after update: %v", err)
	}
	if p.Day != "2026-06-21" {
		t.Errorf("day = %q, want 2026-06-21", p.Day)
	}
	wantPeriod := payPeriodAnchor.Format("2006-01-02") // 2026-06-08
	if p.PayPeriod != wantPeriod {
		t.Errorf("pay_period = %q, want %q", p.PayPeriod, wantPeriod)
	}
}

func TestAdminCreate(t *testing.T) {
	db := newTestDB(t)
	tm := &TimePunchModel{DB: db}
	workerID := mustInsertWorker(t, db, "Manthan", "6666", true)

	// Closed punch
	start := time.Date(2026, 6, 19, 8, 0, 0, 0, time.Local)
	end := start.Add(8 * time.Hour)
	id, err := tm.AdminCreate(workerID, start, end)
	if err != nil {
		t.Fatalf("AdminCreate: %v", err)
	}
	p, err := tm.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if p.Day != "2026-06-19" {
		t.Errorf("day = %q, want 2026-06-19", p.Day)
	}
	if p.PayPeriod != payPeriodAnchor.Format("2006-01-02") {
		t.Errorf("pay_period = %q, want %q", p.PayPeriod, payPeriodAnchor.Format("2006-01-02"))
	}
	if !p.ModifiedByAdmin {
		t.Error("want modified_by_admin = true")
	}
	if !p.EndTime.Valid {
		t.Error("want end_time to be set")
	}

	// Open punch (no end time)
	idOpen, err := tm.AdminCreate(workerID, start.AddDate(0, 0, 1), time.Time{})
	if err != nil {
		t.Fatalf("AdminCreate open: %v", err)
	}
	pOpen, err := tm.Get(idOpen)
	if err != nil {
		t.Fatalf("Get open: %v", err)
	}
	if pOpen.EndTime.Valid {
		t.Error("want open punch to have no end_time")
	}
}

func TestDelete(t *testing.T) {
	db := newTestDB(t)
	tm := &TimePunchModel{DB: db}
	workerID := mustInsertWorker(t, db, "Manthan", "5555", true)

	id, err := tm.PunchIn(workerID, nil, nil, time.Now())
	if err != nil {
		t.Fatalf("PunchIn: %v", err)
	}

	if err := tm.Delete(id); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := tm.Get(id); !errors.Is(err, ErrNoRecord) {
		t.Errorf("Get after delete: got %v, want ErrNoRecord", err)
	}
	if err := tm.Delete(id); !errors.Is(err, ErrNoRecord) {
		t.Errorf("Delete non-existent: got %v, want ErrNoRecord", err)
	}
}

func TestAutoPunchOutNonCompliant(t *testing.T) {
	db := newTestDB(t)
	tm := &TimePunchModel{DB: db}

	alice := mustInsertWorker(t, db, "Alice", "1111", true)
	bob := mustInsertWorker(t, db, "Bob", "2222", true)
	carlos := mustInsertWorker(t, db, "Carlos", "3333", true)

	day := time.Date(2026, 6, 17, 0, 0, 0, 0, time.Local)
	cutoff := time.Date(2026, 6, 17, 21, 0, 0, 0, time.Local)

	if _, err := tm.PunchIn(alice, nil, nil, day.Add(7*time.Hour)); err != nil {
		t.Fatalf("PunchIn Alice: %v", err)
	}
	if _, err := tm.PunchIn(bob, nil, nil, day.Add(8*time.Hour)); err != nil {
		t.Fatalf("PunchIn Bob: %v", err)
	}
	if _, err := tm.PunchIn(carlos, nil, nil, day.Add(8*time.Hour)); err != nil {
		t.Fatalf("PunchIn Carlos: %v", err)
	}
	if err := tm.PunchOut(carlos, nil, nil, day.Add(17*time.Hour)); err != nil {
		t.Fatalf("PunchOut Carlos: %v", err)
	}

	n, err := tm.AutoPunchOutNonCompliant(cutoff)
	if err != nil {
		t.Fatalf("AutoPunchOutNonCompliant: %v", err)
	}
	if n != 2 {
		t.Errorf("closed %d punches, want 2", n)
	}

	for _, workerID := range []int{alice, bob} {
		var closed, nc bool
		err := db.QueryRow(`SELECT end_time IS NOT NULL, non_compliant FROM time_punches WHERE worker_id = ?`, workerID).Scan(&closed, &nc)
		if err != nil {
			t.Fatalf("query worker %d: %v", workerID, err)
		}
		if !closed {
			t.Errorf("worker %d: want punch closed", workerID)
		}
		if !nc {
			t.Errorf("worker %d: want non_compliant = true", workerID)
		}
	}

	var nc bool
	if err := db.QueryRow(`SELECT non_compliant FROM time_punches WHERE worker_id = ?`, carlos).Scan(&nc); err != nil {
		t.Fatalf("query Carlos: %v", err)
	}
	if nc {
		t.Error("Carlos: want non_compliant = false (closed normally)")
	}
}

func TestForPayPeriodNonCompliantOrdering(t *testing.T) {
	db := newTestDB(t)
	tm := &TimePunchModel{DB: db}
	workerID := mustInsertWorker(t, db, "Manthan", "7777", true)

	day1 := payPeriodAnchor.Add(8 * time.Hour)
	day2 := payPeriodAnchor.AddDate(0, 0, 1).Add(8 * time.Hour)

	// Normal punch on day1
	if _, err := tm.PunchIn(workerID, nil, nil, day1); err != nil {
		t.Fatalf("PunchIn day1: %v", err)
	}
	if err := tm.PunchOut(workerID, nil, nil, day1.Add(8*time.Hour)); err != nil {
		t.Fatalf("PunchOut day1: %v", err)
	}

	// Non-compliant punch on day2
	if _, err := tm.PunchIn(workerID, nil, nil, day2); err != nil {
		t.Fatalf("PunchIn day2: %v", err)
	}
	cutoff := time.Date(day2.Year(), day2.Month(), day2.Day(), 21, 0, 0, 0, time.Local)
	if _, err := tm.AutoPunchOutNonCompliant(cutoff); err != nil {
		t.Fatalf("AutoPunchOutNonCompliant: %v", err)
	}

	rows, err := tm.ForPayPeriod(payPeriodAnchor.Format("2006-01-02"))
	if err != nil {
		t.Fatalf("ForPayPeriod: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(rows))
	}
	// Non-compliant (day2) should sort first despite being a later date
	if !rows[0].NonCompliant {
		t.Errorf("rows[0].NonCompliant = false, want true (non-compliant should sort first)")
	}
	if rows[1].NonCompliant {
		t.Errorf("rows[1].NonCompliant = true, want false")
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
	id1, err := tm.PunchIn(alice, nil, nil, day1)
	if err != nil {
		t.Fatalf("PunchIn: %v", err)
	}
	if err := tm.PunchOut(alice, nil, nil, day1.Add(time.Hour)); err != nil {
		t.Fatalf("PunchOut: %v", err)
	}
	id2, err := tm.PunchIn(alice, nil, nil, day1.Add(2*time.Hour))
	if err != nil {
		t.Fatalf("PunchIn: %v", err)
	}
	if err := tm.PunchOut(alice, nil, nil, day1.Add(3*time.Hour)); err != nil {
		t.Fatalf("PunchOut: %v", err)
	}
	if _, err := tm.PunchIn(alice, nil, nil, day2); err != nil {
		t.Fatalf("PunchIn: %v", err)
	}
	if err := tm.PunchOut(alice, nil, nil, day2.Add(time.Hour)); err != nil {
		t.Fatalf("PunchOut: %v", err)
	}

	// Bob (inactive): one punch on day1 (5h) - still included since it's in this period.
	if _, err := tm.PunchIn(bob, nil, nil, day1); err != nil {
		t.Fatalf("PunchIn: %v", err)
	}
	if err := tm.PunchOut(bob, nil, nil, day1.Add(5*time.Hour)); err != nil {
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

func TestAdminUpdateClearedEndKeepsLaterPunches(t *testing.T) {
	db := newTestDB(t)
	tm := &TimePunchModel{DB: db}
	workerID := mustInsertWorker(t, db, "Manthan", "1234", true)

	mk := func(day int) int {
		start := time.Date(2026, 6, day, 8, 0, 0, 0, time.Local)
		id, err := tm.AdminCreate(workerID, start, start.Add(8*time.Hour))
		if err != nil {
			t.Fatalf("AdminCreate day %d: %v", day, err)
		}
		return id
	}
	first := mk(9)
	mk(10)
	mk(11)

	// Clearing the end time reopens the June 9 punch; it must NOT be treated
	// as overlapping (and deleting) the June 10 and 11 punches.
	if err := tm.AdminUpdate(first, time.Date(2026, 6, 9, 8, 0, 0, 0, time.Local), time.Time{}); err != nil {
		t.Fatalf("AdminUpdate: %v", err)
	}

	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM time_punches WHERE worker_id = ?`, workerID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Errorf("got %d punches after clearing an end time, want 3 (later punches must survive)", n)
	}

	open, err := tm.Open(workerID)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if open.ID != first {
		t.Errorf("open punch = %d, want %d", open.ID, first)
	}
}

func TestAdminUpdateClearedEndStillReplacesSameDayPunch(t *testing.T) {
	db := newTestDB(t)
	tm := &TimePunchModel{DB: db}
	workerID := mustInsertWorker(t, db, "Manthan", "1234", true)

	start := time.Date(2026, 6, 9, 8, 0, 0, 0, time.Local)
	edited, err := tm.AdminCreate(workerID, start, start.Add(2*time.Hour))
	if err != nil {
		t.Fatalf("AdminCreate: %v", err)
	}
	// A second punch later the same day overlaps the reopened range and
	// should be replaced, same as before.
	if _, err := tm.AdminCreate(workerID, start.Add(4*time.Hour), start.Add(6*time.Hour)); err != nil {
		t.Fatalf("AdminCreate same-day: %v", err)
	}

	if err := tm.AdminUpdate(edited, start, time.Time{}); err != nil {
		t.Fatalf("AdminUpdate: %v", err)
	}

	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM time_punches WHERE worker_id = ?`, workerID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("got %d punches, want 1 (same-day overlap should be replaced)", n)
	}
}

func TestAdminUpdateClearedEndConflictsWithOpenPunch(t *testing.T) {
	db := newTestDB(t)
	tm := &TimePunchModel{DB: db}
	workerID := mustInsertWorker(t, db, "Manthan", "1234", true)

	closedStart := time.Date(2026, 6, 9, 8, 0, 0, 0, time.Local)
	closed, err := tm.AdminCreate(workerID, closedStart, closedStart.Add(8*time.Hour))
	if err != nil {
		t.Fatalf("AdminCreate: %v", err)
	}
	if _, err := tm.PunchIn(workerID, nil, nil, time.Date(2026, 6, 11, 8, 0, 0, 0, time.Local)); err != nil {
		t.Fatalf("PunchIn: %v", err)
	}

	// Reopening the June 9 punch would give the worker two open punches.
	err = tm.AdminUpdate(closed, closedStart, time.Time{})
	if !errors.Is(err, ErrAlreadyOpen) {
		t.Errorf("got error %v, want ErrAlreadyOpen", err)
	}
}

func TestAdminUpdateEndBeforeStart(t *testing.T) {
	db := newTestDB(t)
	tm := &TimePunchModel{DB: db}
	workerID := mustInsertWorker(t, db, "Manthan", "1234", true)

	start := time.Date(2026, 6, 9, 8, 0, 0, 0, time.Local)
	id, err := tm.AdminCreate(workerID, start, start.Add(8*time.Hour))
	if err != nil {
		t.Fatalf("AdminCreate: %v", err)
	}

	err = tm.AdminUpdate(id, start, start.Add(-time.Hour))
	if !errors.Is(err, ErrEndBeforeStart) {
		t.Errorf("got error %v, want ErrEndBeforeStart", err)
	}
}

func TestAdminCreateEndBeforeStart(t *testing.T) {
	db := newTestDB(t)
	tm := &TimePunchModel{DB: db}
	workerID := mustInsertWorker(t, db, "Manthan", "1234", true)

	start := time.Date(2026, 6, 9, 8, 0, 0, 0, time.Local)
	if _, err := tm.AdminCreate(workerID, start, start.Add(-time.Hour)); !errors.Is(err, ErrEndBeforeStart) {
		t.Errorf("got error %v, want ErrEndBeforeStart", err)
	}
}

func TestAdminCreateSecondOpenPunchConflicts(t *testing.T) {
	db := newTestDB(t)
	tm := &TimePunchModel{DB: db}
	workerID := mustInsertWorker(t, db, "Manthan", "1234", true)

	if _, err := tm.PunchIn(workerID, nil, nil, time.Date(2026, 6, 9, 8, 0, 0, 0, time.Local)); err != nil {
		t.Fatalf("PunchIn: %v", err)
	}

	// The unique index (not just the handler-level Open() check) must stop a
	// second open punch - this is what closes the double-tap race.
	_, err := tm.AdminCreate(workerID, time.Date(2026, 6, 10, 8, 0, 0, 0, time.Local), time.Time{})
	if !errors.Is(err, ErrAlreadyOpen) {
		t.Errorf("got error %v, want ErrAlreadyOpen", err)
	}
}

func TestPunchOutLateAfterNightlySweep(t *testing.T) {
	db := newTestDB(t)
	tm := &TimePunchModel{DB: db}
	workerID := mustInsertWorker(t, db, "Thomas", "4321", true)

	start := time.Date(2026, 6, 10, 10, 0, 0, 0, time.Local)
	id, err := tm.PunchIn(workerID, nil, nil, start)
	if err != nil {
		t.Fatalf("PunchIn: %v", err)
	}

	// The 9 PM sweep closes the punch before the worker reports in.
	if _, err := tm.AutoPunchOutNonCompliant(time.Date(2026, 6, 10, 21, 0, 0, 0, time.Local)); err != nil {
		t.Fatalf("AutoPunchOutNonCompliant: %v", err)
	}

	// The late report must still land, correcting the swept end time.
	realEnd := time.Date(2026, 6, 10, 17, 30, 0, 0, time.Local)
	if err := tm.PunchOutLate(workerID, realEnd); err != nil {
		t.Fatalf("PunchOutLate after sweep: %v", err)
	}

	p, err := tm.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !p.Late {
		t.Error("expected late = true")
	}
	if !p.NonCompliant {
		t.Error("expected non_compliant to stay true so the admin still reviews it")
	}
	if got, want := p.EndTime.String, realEnd.UTC().Format(time.RFC3339); got != want {
		t.Errorf("end time = %q, want %q", got, want)
	}

	// A second late report must not find anything (late = TRUE now).
	if err := tm.PunchOutLate(workerID, realEnd.Add(time.Hour)); !errors.Is(err, ErrNoRecord) {
		t.Errorf("got error %v, want ErrNoRecord for a second late report", err)
	}
}

func TestPunchOutLateEndBeforeStart(t *testing.T) {
	db := newTestDB(t)
	tm := &TimePunchModel{DB: db}
	workerID := mustInsertWorker(t, db, "Thomas", "4321", true)

	start := time.Date(2026, 6, 10, 10, 0, 0, 0, time.Local)
	if _, err := tm.PunchIn(workerID, nil, nil, start); err != nil {
		t.Fatalf("PunchIn: %v", err)
	}

	err := tm.PunchOutLate(workerID, start.Add(-time.Hour))
	if !errors.Is(err, ErrEndBeforeStart) {
		t.Errorf("got error %v, want ErrEndBeforeStart", err)
	}
}

// TestPayPeriodStartEveryDayInOwnPeriod sweeps several years of dates -
// including DST transitions when the test runs in a zone that has them -
// and checks the invariants the hour-division bug used to break: every
// date falls inside its own 14-day period and every period starts on a
// Monday.
func TestPayPeriodStartEveryDayInOwnPeriod(t *testing.T) {
	for d := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local); d.Year() < 2031; d = d.AddDate(0, 0, 1) {
		start := payPeriodStart(d)
		end := start.AddDate(0, 0, 14)
		if d.Before(start) || !d.Before(end) {
			t.Errorf("date %s -> period start %s does not contain the date", d.Format("2006-01-02"), start.Format("2006-01-02"))
		}
		if start.Weekday() != time.Monday {
			t.Errorf("date %s -> period start %s is a %s, not Monday", d.Format("2006-01-02"), start.Format("2006-01-02"), start.Weekday())
		}
	}
}
