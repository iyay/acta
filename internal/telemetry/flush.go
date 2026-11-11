package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// maxReceipt is the most of a reply a flush reads. A bigger reply is not a
// receipt, so the day stays for the next try.
const maxReceipt = 16 * 1024

var errRedirect = errors.New("telemetry redirect refused")

// NewClient is the client a real flush uses: 20 seconds per request and only
// the Google content redirect that Apps Script answers with.
func NewClient() *http.Client { return newClient(nil) }

func newClient(transport http.RoundTripper) *http.Client {
	return &http.Client{Transport: transport, Timeout: 20 * time.Second, CheckRedirect: func(r *http.Request, via []*http.Request) error {
		// Apps Script answers a POST with a GET redirect to its content host.
		// Anything else could send the day, or the user, somewhere unknown.
		if len(via) > 3 || r.URL.Scheme != "https" || r.URL.Host != "script.googleusercontent.com" || r.URL.User != nil || r.Method != http.MethodGet {
			return errRedirect
		}
		return nil
	}}
}

// Payload is the exact body one day is sent with.
func Payload(d Day) ([]byte, error) { return json.Marshal(d) }

// Flush sends every finished day once. A day goes away only when the collector
// took it, or refused it for good. Every other answer keeps it for next time.
// It never returns a send failure: telemetry must stay out of the user's way.
func Flush(ctx context.Context, s *Store, c *http.Client) error {
	if !s.Enabled() {
		return nil
	}
	unlock, ok := s.flushLock()
	if !ok {
		return nil
	}
	defer unlock()
	if err := s.SetLastFlush(now()); err != nil {
		return err
	}
	days, err := s.Days()
	if err != nil {
		return err
	}
	for _, d := range days {
		if d.Day == Today() {
			continue
		}
		if send(ctx, c, d) {
			if err := s.Delete(d.Day); err != nil {
				return err
			}
		}
	}
	return nil
}

// send posts one day and says whether the day is done with: taken, or
// refused with retryable false.
func send(ctx context.Context, c *http.Client, d Day) bool {
	body, err := Payload(d)
	if err != nil {
		return false
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, Endpoint, bytes.NewReader(body))
	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxReceipt+1))
	if err != nil || len(data) > maxReceipt {
		return false
	}
	var r struct {
		OK        *bool  `json:"ok"`
		InstallID string `json:"install_id"`
		Day       string `json:"day"`
		Duplicate *bool  `json:"duplicate"`
		Error     string `json:"error"`
		Retryable *bool  `json:"retryable"`
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if dec.Decode(&r) != nil || dec.Decode(new(any)) != io.EOF || r.OK == nil {
		return false
	}
	if *r.OK {
		return r.InstallID == d.InstallID && r.Day == d.Day && r.Duplicate != nil
	}
	return r.Retryable != nil && !*r.Retryable
}

// flushLock lets one flush run at a time. A second one gives up at once,
// since the first is already sending the same days.
func (s *Store) flushLock() (func(), bool) {
	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		return nil, false
	}
	f, err := os.OpenFile(filepath.Join(s.dir, "flush.lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, false
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, false
	}
	return func() {
		syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
	}, true
}
