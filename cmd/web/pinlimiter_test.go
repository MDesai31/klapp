package main

import (
	"net/http"
	"testing"
	"time"
)

func newTestPinLimiter(threshold int, window, cooldown time.Duration) (*pinLimiter, *fakeClock) {
	clock := &fakeClock{t: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	l := newPinLimiter(threshold, window, cooldown)
	l.now = clock.Now
	return l, clock
}

type fakeClock struct{ t time.Time }

func (c *fakeClock) Now() time.Time          { return c.t }
func (c *fakeClock) Advance(d time.Duration) { c.t = c.t.Add(d) }

func TestPinLimiterLocksOutAfterThreshold(t *testing.T) {
	l, clock := newTestPinLimiter(3, 15*time.Minute, 15*time.Minute)

	for i := 0; i < 2; i++ {
		l.recordFailure("1.2.3.4")
		if !l.allow("1.2.3.4") {
			t.Fatalf("locked out after only %d failures, want allowed", i+1)
		}
	}

	l.recordFailure("1.2.3.4")
	if l.allow("1.2.3.4") {
		t.Fatalf("expected lockout after reaching threshold")
	}

	// Other IPs are unaffected.
	if !l.allow("5.6.7.8") {
		t.Errorf("unrelated IP should not be locked out")
	}

	clock.Advance(15*time.Minute + time.Second)
	if !l.allow("1.2.3.4") {
		t.Errorf("expected lockout to expire after cooldown")
	}
}

func TestPinLimiterSuccessClearsFailures(t *testing.T) {
	l, _ := newTestPinLimiter(3, 15*time.Minute, 15*time.Minute)

	l.recordFailure("1.2.3.4")
	l.recordFailure("1.2.3.4")
	l.recordSuccess("1.2.3.4")
	l.recordFailure("1.2.3.4")
	l.recordFailure("1.2.3.4")

	if !l.allow("1.2.3.4") {
		t.Errorf("expected allowed - success should have reset the failure count")
	}
}

func TestPinLimiterWindowResetsOldFailures(t *testing.T) {
	l, clock := newTestPinLimiter(3, 15*time.Minute, 15*time.Minute)

	l.recordFailure("1.2.3.4")
	l.recordFailure("1.2.3.4")

	clock.Advance(16 * time.Minute)

	l.recordFailure("1.2.3.4")
	if !l.allow("1.2.3.4") {
		t.Errorf("expected allowed - earlier failures fell outside the window and shouldn't count")
	}
}

func TestClientIPPrefersForwardedFor(t *testing.T) {
	r, err := http.NewRequest("POST", "/punch", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.RemoteAddr = "127.0.0.1:54321"
	r.Header.Set("X-Forwarded-For", "203.0.113.5, 127.0.0.1")

	if got := clientIP(r); got != "203.0.113.5" {
		t.Errorf("clientIP() = %q, want %q", got, "203.0.113.5")
	}
}

func TestClientIPIgnoresForwardedForFromDirectClients(t *testing.T) {
	r, err := http.NewRequest("POST", "/punch", nil)
	if err != nil {
		t.Fatal(err)
	}
	// Not loopback, so this didn't come through Caddy - a spoofed
	// X-Forwarded-For must not let the client pick its own limiter key.
	r.RemoteAddr = "203.0.113.9:54321"
	r.Header.Set("X-Forwarded-For", "8.8.8.8")

	if got := clientIP(r); got != "203.0.113.9" {
		t.Errorf("clientIP() = %q, want %q", got, "203.0.113.9")
	}
}

func TestClientIPFallsBackToRemoteAddr(t *testing.T) {
	r, err := http.NewRequest("POST", "/punch", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.RemoteAddr = "203.0.113.9:54321"

	if got := clientIP(r); got != "203.0.113.9" {
		t.Errorf("clientIP() = %q, want %q", got, "203.0.113.9")
	}
}
