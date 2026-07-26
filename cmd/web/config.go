package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// Config holds tunable settings for the worker punch site that operators
// may want to adjust without recompiling - currently just the PIN-guessing
// throttle and the daily punch-in cap. Loaded from a JSON file (see
// config.example.json); any field left out of the file, or the file being
// absent entirely, falls back to the default below.
type Config struct {
	// PinLockoutThreshold is how many failed PIN attempts from one IP
	// within PinLockoutWindowMinutes trigger a lockout.
	PinLockoutThreshold int `json:"pin_lockout_threshold"`
	// PinLockoutWindowMinutes is the trailing window failed attempts are
	// counted over before the count resets.
	PinLockoutWindowMinutes int `json:"pin_lockout_window_minutes"`
	// PinLockoutCooldownMinutes is how long a locked-out IP is blocked.
	PinLockoutCooldownMinutes int `json:"pin_lockout_cooldown_minutes"`
	// PinCheckDelayMs is a fixed delay applied to every PIN check
	// (success or failure) to slow down scripted guessing.
	PinCheckDelayMs int `json:"pin_check_delay_ms"`
	// DailyPunchLimit caps how many times a worker can punch in per
	// calendar day. Zero means no limit.
	DailyPunchLimit int `json:"daily_punch_limit"`
	// PunchSiteURL is the worker-facing punch site's public URL, sent in
	// the admin dashboard's "Notify" text message links.
	PunchSiteURL string `json:"punch_site_url"`
}

func defaultConfig() Config {
	return Config{
		PinLockoutThreshold:       10,
		PinLockoutWindowMinutes:   15,
		PinLockoutCooldownMinutes: 15,
		PinCheckDelayMs:           250,
		DailyPunchLimit:           3,
		PunchSiteURL:              "https://work.klauslandscaping.com",
	}
}

// loadConfig reads path as JSON into a Config seeded with defaults. A
// missing file is not an error - the defaults are used as-is, so the app
// runs with no config file present. A malformed file is an error, since
// that indicates operator mistake rather than absence.
func loadConfig(path string) (Config, error) {
	cfg := defaultConfig()

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("reading config file %s: %w", path, err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config file %s: %w", path, err)
	}

	return cfg, nil
}
