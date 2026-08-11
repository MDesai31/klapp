package main

import (
	"testing"
	"time"
)

func TestParseTimeToday(t *testing.T) {
	day := time.Date(2026, 6, 17, 15, 30, 0, 0, time.Local)

	tests := []struct {
		name    string
		value   string
		want    time.Time
		wantErr bool
	}{
		{"blank", "", time.Time{}, false},
		{"hh:mm", "07:05", time.Date(2026, 6, 17, 7, 5, 0, 0, time.Local), false},
		{"with seconds", "16:45:30", time.Date(2026, 6, 17, 16, 45, 0, 0, time.Local), false},
		{"garbage", "not a time", time.Time{}, true},
		{"out of range", "25:00", time.Time{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTimeToday(tt.value, day)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseTimeToday(%q) = %v, want an error", tt.value, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseTimeToday(%q): %v", tt.value, err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("parseTimeToday(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestBulkPunchFlash(t *testing.T) {
	at := time.Date(2026, 6, 17, 7, 0, 0, 0, time.Local)

	tests := []struct {
		name       string
		in, out    time.Time
		punchedIn  []string
		punchedOut []string
		skipped    []string
		want       string
	}{
		{
			name: "nothing selected",
			want: "No workers selected.",
		},
		{
			name:      "punched in now",
			punchedIn: []string{"Ana", "Beto"},
			want:      "Punched in Ana, Beto.",
		},
		{
			name:       "punched both ways at the current time",
			punchedIn:  []string{"Ana"},
			punchedOut: []string{"Beto"},
			want:       "Punched in Ana. Punched out Beto.",
		},
		{
			name:      "punched in at a chosen time",
			in:        at,
			punchedIn: []string{"Ana"},
			want:      "Punched in Ana at 7:00 AM.",
		},
		{
			name:       "punched out at a chosen time, with a skip",
			out:        at.Add(9 * time.Hour),
			punchedOut: []string{"Ana"},
			skipped:    []string{"Beto (not punched in)"},
			want:       "Punched out Ana at 4:00 PM. Skipped Beto (not punched in).",
		},
		{
			name:      "whole shift",
			in:        at.Add(time.Hour),
			out:       at.Add(10 * time.Hour),
			punchedIn: []string{"Ana", "Beto"},
			want:      "Punched Ana, Beto in at 8:00 AM and out at 5:00 PM.",
		},
		{
			name:      "whole shift, with a skip",
			in:        at.Add(time.Hour),
			out:       at.Add(10 * time.Hour),
			punchedIn: []string{"Ana"},
			skipped:   []string{"Beto (already punched in)"},
			want:      "Punched Ana in at 8:00 AM and out at 5:00 PM. Skipped Beto (already punched in).",
		},
		{
			name:    "everyone skipped",
			out:     at,
			skipped: []string{"Ana (that time is before their punch-in)"},
			want:    "Skipped Ana (that time is before their punch-in).",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bulkPunchFlash(tt.in, tt.out, tt.punchedIn, tt.punchedOut, tt.skipped)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
