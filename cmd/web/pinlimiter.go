package main

import (
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// errPinLockedOut is returned by authenticateWorker when the caller's IP
// has failed too many PIN checks recently. It's a web-layer concern (tied
// to an HTTP client, not a worker record), so it lives here rather than in
// internal/models alongside the business-data sentinel errors.
var errPinLockedOut = errors.New("too many failed PIN attempts from this address")

// pinLimiter throttles PIN guessing on the public punch site by source IP.
// The punch site sits on cellular-only traffic (see docs/reference/security.md), so
// carrier-grade NAT means many unrelated workers can share one IP - the
// threshold is set high enough in config that a handful of mistyped PINs
// across a shared connection won't trip it, while a scripted attacker
// hammering the endpoint from one IP will.
type pinLimiter struct {
	threshold int
	window    time.Duration
	cooldown  time.Duration

	now func() time.Time

	mu      sync.Mutex
	entries map[string]*pinLimiterEntry
}

type pinLimiterEntry struct {
	failures    int
	windowStart time.Time
	lockedUntil time.Time
	lastSeen    time.Time
}

func newPinLimiter(threshold int, window, cooldown time.Duration) *pinLimiter {
	return &pinLimiter{
		threshold: threshold,
		window:    window,
		cooldown:  cooldown,
		now:       time.Now,
		entries:   make(map[string]*pinLimiterEntry),
	}
}

// allow reports whether ip is currently permitted to attempt a PIN check.
func (l *pinLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	e, ok := l.entries[ip]
	if !ok {
		return true
	}
	return l.now().After(e.lockedUntil)
}

// recordFailure counts a failed PIN check against ip, locking it out for
// the configured cooldown once the threshold is reached within the window.
func (l *pinLimiter) recordFailure(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	e, ok := l.entries[ip]
	if !ok || now.Sub(e.windowStart) > l.window {
		e = &pinLimiterEntry{windowStart: now}
		l.entries[ip] = e
	}
	e.lastSeen = now
	e.failures++
	if e.failures >= l.threshold {
		e.lockedUntil = now.Add(l.cooldown)
	}
}

// recordSuccess clears any failure history for ip.
func (l *pinLimiter) recordSuccess(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, ip)
}

// cleanupLoop periodically evicts entries that are no longer relevant, so
// the map doesn't grow unbounded over the life of the process.
func (l *pinLimiter) cleanupLoop() {
	for {
		time.Sleep(l.window + l.cooldown)
		l.cleanup()
	}
}

func (l *pinLimiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := l.now().Add(-(l.window + l.cooldown))
	for ip, e := range l.entries {
		if e.lastSeen.Before(cutoff) && l.now().After(e.lockedUntil) {
			delete(l.entries, ip)
		}
	}
}

// clientIP returns the IP address a request should be attributed to for
// rate-limiting purposes. In production the worker site is only reachable
// via Caddy on 127.0.0.1 (see deploy/klapp.service, deploy/Caddyfile), so
// the client IP Caddy reports in X-Forwarded-For is trusted - but only
// when the direct peer actually is loopback. A directly-reachable client
// (e.g. a dev run bound to all interfaces) could otherwise spoof the
// header and dodge the PIN lockout.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}

	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			if first := strings.TrimSpace(strings.Split(fwd, ",")[0]); first != "" {
				return first
			}
		}
	}

	return host
}
