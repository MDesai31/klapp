package schedule

import (
	"database/sql"
	"testing"
	"time"

	"klapp/internal/models"
)

const testPeriod = "2026-06-08"

// testWorkers is the cast used by the tests below: two active workers, one
// deactivated. Build only ever prints active ones.
func testWorkers() []models.Worker {
	return []models.Worker{
		{ID: 1, WorkerName: "Juan Perez", Active: true},
		{ID: 2, WorkerName: "Maria Lopez", Active: true},
		{ID: 3, WorkerName: "Old Guy", Active: false},
	}
}

// punch builds one closed punch on day at the given local wall-clock hours.
// An endHour of -1 leaves the punch open.
func punch(workerID int, day string, startHour, endHour int) models.TimesheetRow {
	d, err := time.ParseInLocation("2006-01-02", day, time.Local)
	if err != nil {
		panic(err)
	}
	start := d.Add(time.Duration(startHour) * time.Hour)

	r := models.TimesheetRow{}
	r.ID = workerID*100 + startHour
	r.WorkerID = workerID
	r.PayPeriod = testPeriod
	r.Day = day
	r.StartTime = start
	if endHour >= 0 {
		end := d.Add(time.Duration(endHour) * time.Hour)
		r.EndTime = sql.NullString{String: end.Format(time.RFC3339), Valid: true}
	}
	return r
}

func TestBuildFillsDayRows(t *testing.T) {
	punches := []models.TimesheetRow{
		punch(1, "2026-06-08", 8, 17),
		punch(2, "2026-06-11", 9, 15),
	}

	p, err := Build(testPeriod, testWorkers(), punches)
	if err != nil {
		t.Fatal(err)
	}

	if p.StartDate != "2026-06-08" || p.EndDate != "2026-06-21" {
		t.Errorf("got period %s..%s, want 2026-06-08..2026-06-21", p.StartDate, p.EndDate)
	}
	if len(p.Sheets) != 2 {
		t.Fatalf("got %d sheets, want 2", len(p.Sheets))
	}

	juan := p.Sheets[0]
	if juan.Name != "Juan Perez" {
		t.Errorf("got first sheet %q, want Juan Perez", juan.Name)
	}
	if len(juan.Days) != 14 {
		t.Fatalf("got %d day rows, want 14", len(juan.Days))
	}

	// The punch lands on its own day, and every other day stays empty so
	// it prints as a blank line to fill in by hand.
	got := juan.Days[0]
	want := Row{Date: "2026-06-08", In: "8:00 AM", Out: "5:00 PM", Hours: "9:00"}
	if got != want {
		t.Errorf("day 0 = %+v, want %+v", got, want)
	}
	for i, d := range juan.Days[1:] {
		if d.In != "" || d.Out != "" || d.Hours != "" {
			t.Errorf("day %d = %+v, want empty", i+1, d)
		}
	}
	if juan.Total != "9:00" {
		t.Errorf("total = %q, want 9:00", juan.Total)
	}
	if len(juan.Extra) != 0 {
		t.Errorf("got %d extra rows, want none", len(juan.Extra))
	}
}

func TestBuildSpillsSecondPunchOfADay(t *testing.T) {
	// Out of chronological order on purpose: ForPayPeriod hands them back
	// in display order, so Build has to sort before deciding which punch
	// owns the day's row.
	punches := []models.TimesheetRow{
		punch(1, "2026-06-09", 13, 17),
		punch(1, "2026-06-09", 8, 12),
	}

	p, err := Build(testPeriod, testWorkers(), punches)
	if err != nil {
		t.Fatal(err)
	}

	sheet := p.Sheets[0]
	day := sheet.Days[1] // June 9
	if day.In != "8:00 AM" || day.Out != "12:00 PM" || day.Hours != "4:00" {
		t.Errorf("June 9 row = %+v, want the morning punch", day)
	}

	if len(sheet.Extra) != 1 {
		t.Fatalf("got %d extra rows, want 1", len(sheet.Extra))
	}
	want := Row{Date: "2026-06-09", In: "1:00 PM", Out: "5:00 PM", Hours: "4:00"}
	if sheet.Extra[0] != want {
		t.Errorf("extra row = %+v, want %+v", sheet.Extra[0], want)
	}

	// Both punches count, wherever they were printed.
	if sheet.Total != "8:00" {
		t.Errorf("total = %q, want 8:00", sheet.Total)
	}
}

func TestBuildLeavesOpenPunchUnclosedAndUncounted(t *testing.T) {
	punches := []models.TimesheetRow{
		punch(1, "2026-06-08", 8, 17),
		punch(1, "2026-06-09", 8, -1),
	}

	p, err := Build(testPeriod, testWorkers(), punches)
	if err != nil {
		t.Fatal(err)
	}

	open := p.Sheets[0].Days[1]
	if open.In != "8:00 AM" {
		t.Errorf("open punch In = %q, want 8:00 AM", open.In)
	}
	if open.Out != "" || open.Hours != "" {
		t.Errorf("open punch = %+v, want blank Out and Hours", open)
	}
	if p.Sheets[0].Total != "9:00" {
		t.Errorf("total = %q, want 9:00 (the open punch must not count)", p.Sheets[0].Total)
	}
}

func TestBuildSkipsInactiveAndHourlessWorkers(t *testing.T) {
	punches := []models.TimesheetRow{
		punch(3, "2026-06-08", 8, 16), // the deactivated worker
		punch(2, "2026-06-08", 8, 16),
	}

	p, err := Build(testPeriod, testWorkers(), punches)
	if err != nil {
		t.Fatal(err)
	}

	if len(p.Sheets) != 1 {
		t.Fatalf("got %d sheets, want only Maria's", len(p.Sheets))
	}
	if p.Sheets[0].Name != "Maria Lopez" {
		t.Errorf("got sheet for %q, want Maria Lopez", p.Sheets[0].Name)
	}
}

func TestBuildRejectsBadPeriod(t *testing.T) {
	if _, err := Build("not-a-date", testWorkers(), nil); err == nil {
		t.Error("expected an error for a malformed pay period")
	}
}

func TestFormatHours(t *testing.T) {
	tests := []struct {
		in   time.Duration
		want string
	}{
		{0, ""},
		{-time.Hour, ""},
		{30 * time.Minute, "0:30"},
		{9 * time.Hour, "9:00"},
		{6*time.Hour + 45*time.Minute, "6:45"},
		// Totals run past a day and must not wrap.
		{86*time.Hour + 30*time.Minute, "86:30"},
		// Seconds round to the nearest minute rather than truncating.
		{time.Hour + 29*time.Second, "1:00"},
		{time.Hour + 31*time.Second, "1:01"},
	}

	for _, tt := range tests {
		if got := formatHours(tt.in); got != tt.want {
			t.Errorf("formatHours(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
