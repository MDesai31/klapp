// Package schedule builds and ships the printable pay-period schedule.
//
// The admin site never draws the PDF itself. It turns a pay period into the
// Payload below and posts it to the schedule listener on the home server
// (see schedule_listener/), which hands the rows to build_schedule.py. Both
// ends of that wire import this package, so the JSON has exactly one
// definition.
package schedule

import (
	"fmt"
	"sort"
	"time"

	"klapp/internal/models"
)

// Payload is one print job: every sheet to produce for a pay period.
type Payload struct {
	// PayPeriod is the period's first day, "2006-01-02".
	PayPeriod string `json:"pay_period"`
	// StartDate and EndDate bracket the 14 printed days, "2006-01-02".
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	// Sheets is one page per worker, in worker-name order.
	Sheets []Sheet `json:"sheets"`
}

// Sheet is one worker's page.
type Sheet struct {
	Name string `json:"name"`
	// Days holds exactly 14 rows, one per day of the period, in order.
	Days []Row `json:"days"`
	// Extra holds the overflow rows printed under TOTAL: a second (or
	// third) punch on a day that already filled its own row. The PDF
	// reserves four such rows and grows if there are more.
	Extra []Row `json:"extra"`
	// Total is the sheet's TOTAL cell, "H:MM", or "" for no closed hours.
	Total string `json:"total"`
}

// Row is one line of the sheet: the DATE, ENTRADA, SALIDA and HORAS cells.
// Every field may be empty, which prints an empty cell for someone to fill
// in by hand.
type Row struct {
	// Date is "2006-01-02"; the PDF formats the label itself.
	Date string `json:"date"`
	In   string `json:"in"`
	Out  string `json:"out"`
	// Hours is "H:MM" worked, or "" when the punch is still open.
	Hours string `json:"hours"`
}

// timeFormat is how ENTRADA and SALIDA are printed.
const timeFormat = "3:04 PM"

// Build turns a pay period's punches into one sheet per active worker who
// has punches in that period. Inactive workers are left out even if they
// have punches - a print job is for the crew being paid now, unlike the
// on-screen summary, which keeps them for history.
//
// A day's first punch fills that day's own row; further punches on the same
// day spill into the extra rows at the bottom of the sheet, so no recorded
// shift is ever silently merged away. A punch that is still open prints its
// ENTRADA with SALIDA and HORAS blank and does not count toward TOTAL - a
// printed sheet should not claim hours for a shift that has not ended.
func Build(payPeriod string, workers []models.Worker, punches []models.TimesheetRow) (Payload, error) {
	days, err := models.PayPeriodDays(payPeriod)
	if err != nil {
		return Payload{}, err
	}

	dayIndex := make(map[string]int, len(days))
	for i, d := range days {
		dayIndex[d] = i
	}

	// Punches arrive in the timesheet's display order, which is not the
	// order they happened in. Sort so "first punch of the day" means it.
	byWorker := make(map[int][]models.TimesheetRow)
	for _, p := range punches {
		byWorker[p.WorkerID] = append(byWorker[p.WorkerID], p)
	}
	for id := range byWorker {
		ps := byWorker[id]
		sort.Slice(ps, func(i, j int) bool {
			if ps[i].Day != ps[j].Day {
				return ps[i].Day < ps[j].Day
			}
			return ps[i].StartTime.Before(ps[j].StartTime)
		})
	}

	payload := Payload{
		PayPeriod: payPeriod,
		StartDate: days[0],
		EndDate:   days[len(days)-1],
	}

	for _, w := range workers {
		if !w.Active {
			continue
		}
		ps := byWorker[w.ID]
		if len(ps) == 0 {
			continue
		}

		sheet := Sheet{Name: w.WorkerName, Days: make([]Row, len(days))}
		for i, d := range days {
			sheet.Days[i].Date = d
		}

		var total time.Duration
		for _, p := range ps {
			idx, ok := dayIndex[p.Day]
			if !ok {
				// A punch whose pay_period and day disagree. Keep it
				// visible rather than dropping it on the floor.
				sheet.Extra = append(sheet.Extra, punchRow(p, &total))
				continue
			}
			row := punchRow(p, &total)
			if sheet.Days[idx].In == "" && sheet.Days[idx].Out == "" {
				sheet.Days[idx] = row
				continue
			}
			sheet.Extra = append(sheet.Extra, row)
		}

		sheet.Total = formatHours(total)
		payload.Sheets = append(payload.Sheets, sheet)
	}

	return payload, nil
}

// punchRow renders one punch and adds its worked time to total. An open
// punch contributes nothing.
func punchRow(p models.TimesheetRow, total *time.Duration) Row {
	row := Row{
		Date: p.Day,
		In:   p.StartTime.Local().Format(timeFormat),
	}

	if !p.EndTime.Valid {
		return row
	}
	end, err := time.Parse(time.RFC3339, p.EndTime.String)
	if err != nil {
		return row
	}

	worked := end.Sub(p.StartTime)
	row.Out = end.Local().Format(timeFormat)
	row.Hours = formatHours(worked)
	*total += worked
	return row
}

// formatHours renders a duration as the sheet's "H:MM", or "" for zero.
// Hours are not capped at 24 - a TOTAL of "86:30" is normal.
func formatHours(d time.Duration) string {
	if d <= 0 {
		return ""
	}
	m := int(d.Round(time.Minute).Minutes())
	return fmt.Sprintf("%d:%02d", m/60, m%60)
}
