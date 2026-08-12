package schedule

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"
)

// Path is the listener endpoint a Payload is posted to.
const Path = "/print"

// Send posts p to the schedule listener at host:port and returns whatever
// the listener printed back (the PDF paths it wrote), trimmed of nothing.
//
// The listener sits on the WireGuard network and shells out to reportlab,
// which is not instant, so the timeout is generous rather than snappy.
func Send(ctx context.Context, host string, port int, p Payload) (string, error) {
	body, err := json.Marshal(p)
	if err != nil {
		return "", err
	}

	url := "http://" + net.JoinHostPort(host, strconv.Itoa(port)) + Path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		// Already reads `Post "http://host:port/print": ...`, so it needs
		// no wrapping to say where it was going.
		return "", err
	}
	defer resp.Body.Close()

	// The listener's replies are short; cap the read anyway so a wedged
	// server can't feed us an unbounded body.
	out, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s returned %s: %s", url, resp.Status, bytes.TrimSpace(out))
	}

	return string(bytes.TrimSpace(out)), nil
}
