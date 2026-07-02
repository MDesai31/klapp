package main

import (
	"testing"
	"time"
)

func TestNightlySweepAfter(t *testing.T) {
	loc := time.Local

	tests := []struct {
		name string
		now  time.Time
		want time.Time
	}{
		{
			name: "afternoon rolls to same evening",
			now:  time.Date(2026, 7, 2, 15, 0, 0, 0, loc),
			want: time.Date(2026, 7, 2, 21, 0, 0, 0, loc),
		},
		{
			name: "exactly 9 PM rolls to next day",
			now:  time.Date(2026, 7, 2, 21, 0, 0, 0, loc),
			want: time.Date(2026, 7, 3, 21, 0, 0, 0, loc),
		},
		{
			name: "after 9 PM rolls to next day",
			now:  time.Date(2026, 7, 2, 22, 30, 0, 0, loc),
			want: time.Date(2026, 7, 3, 21, 0, 0, 0, loc),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nightlySweepAfter(tt.now); !got.Equal(tt.want) {
				t.Errorf("nightlySweepAfter(%v) = %v, want %v", tt.now, got, tt.want)
			}
			// The startup catch-up cutoff is derived as next minus one day;
			// it must be the most recent 9 PM at or before now.
			catchUp := nightlySweepAfter(tt.now).AddDate(0, 0, -1)
			if catchUp.After(tt.now) {
				t.Errorf("catch-up cutoff %v is in the future of %v", catchUp, tt.now)
			}
			if tt.now.Sub(catchUp) >= 24*time.Hour {
				t.Errorf("catch-up cutoff %v is more than a day before %v", catchUp, tt.now)
			}
		})
	}
}
