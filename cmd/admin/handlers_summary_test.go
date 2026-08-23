package main

import (
	"errors"
	"testing"
)

func TestPrintFailureReason(t *testing.T) {
	execErr := errors.New("exit status 1")

	tests := []struct {
		name string
		out  string
		want string
	}{
		{
			"printsched's own message, prefix stripped",
			"printsched: no active workers have hours in pay period 2026-06-08\n",
			"no active workers have hours in pay period 2026-06-08",
		},
		{
			"last line wins",
			"sent 2 sheet(s) for 2026-06-08\nprintsched: posting to http://10.9.0.7:5555/print: connection refused\n",
			"posting to http://10.9.0.7:5555/print: connection refused",
		},
		{
			"no output at all falls back to the exec error",
			"",
			"exit status 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := printFailureReason([]byte(tt.out), execErr); got != tt.want {
				t.Errorf("printFailureReason() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintFailureReasonTruncates(t *testing.T) {
	long := make([]byte, 500)
	for i := range long {
		long[i] = 'x'
	}

	got := printFailureReason(long, errors.New("boom"))
	if len(got) != 203 {
		t.Errorf("got a %d-char reason, want 200 chars plus an ellipsis", len(got))
	}
}
