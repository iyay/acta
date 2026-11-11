package telemetry

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

// flushServer answers every POST with reply, given the day it got.
func flushServer(t *testing.T, reply func(d Day) (int, string)) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		raw, _ := io.ReadAll(r.Body)
		var d Day
		if err := json.Unmarshal(raw, &d); err != nil {
			t.Errorf("body is not JSON: %s", raw)
		}
		code, body := reply(d)
		w.WriteHeader(code)
		io.WriteString(w, body)
	}))
	t.Cleanup(ts.Close)
	return ts, &calls
}

// storeWithDays gives a store pointed at ts with one finished day and today.
func storeWithDays(t *testing.T, ts *httptest.Server) *Store {
	t.Helper()
	s := openTest(t)
	Endpoint = ts.URL
	setNow(t, time.Date(2026, 3, 1, 12, 0, 0, 0, time.Local))
	if err := s.Add("commands", "show"); err != nil {
		t.Fatal(err)
	}
	setNow(t, time.Date(2026, 3, 2, 12, 0, 0, 0, time.Local))
	if err := s.Add("commands", "list"); err != nil {
		t.Fatal(err)
	}
	return s
}

func clientFor(ts *httptest.Server) *http.Client {
	return newClient(ts.Client().Transport)
}

func dayKept(t *testing.T, s *Store, day string) bool {
	t.Helper()
	_, err := os.Stat(s.dayPath(day))
	return err == nil
}

func okReceipt(d Day) string {
	return `{"ok": true, "install_id": "` + d.InstallID + `", "day": "` + d.Day + `", "duplicate": false}`
}

func TestFlushSendsOnlyFinishedDaysWithTheContractBody(t *testing.T) {
	var got []map[string]any
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var m map[string]any
		json.Unmarshal(raw, &m)
		got = append(got, m)
		io.WriteString(w, `{"ok": true, "install_id": "`+m["install_id"].(string)+`", "day": "`+m["day"].(string)+`", "duplicate": false}`)
	}))
	defer ts.Close()
	s := storeWithDays(t, ts)
	if err := Flush(context.Background(), s, clientFor(ts)); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0]["day"] != "2026-03-01" {
		t.Fatalf("sent %v, want only 2026-03-01", got)
	}
	want := []string{"kind", "install_id", "day", "acta_version", "os", "arch", "sessions", "commands", "skills", "workflow", "blocks", "settings"}
	if len(got[0]) != len(want) {
		t.Fatalf("body has fields %v, want exactly %v", got[0], want)
	}
	for _, k := range want {
		if _, ok := got[0][k]; !ok {
			t.Fatalf("body misses %q: %v", k, got[0])
		}
	}
	if got[0]["kind"] != "acta_telemetry" {
		t.Fatalf("kind = %v", got[0]["kind"])
	}
	if dayKept(t, s, "2026-03-01") || !dayKept(t, s, "2026-03-02") {
		t.Fatal("want the sent day deleted and today kept")
	}
	if s.LastFlush().IsZero() {
		t.Fatal("flush did not record its attempt")
	}
}

func TestFlushReceipts(t *testing.T) {
	cases := []struct {
		name  string
		reply func(d Day) (int, string)
		kept  bool
	}{
		{"ok", func(d Day) (int, string) { return 200, okReceipt(d) }, false},
		{"duplicate", func(d Day) (int, string) {
			return 200, strings.Replace(okReceipt(d), `"duplicate": false`, `"duplicate": true`, 1)
		}, false},
		{"terminal refusal", func(d Day) (int, string) {
			return 200, `{"ok": false, "error": "invalid_day", "retryable": false}`
		}, false},
		{"retryable refusal", func(d Day) (int, string) {
			return 200, `{"ok": false, "error": "busy", "retryable": true}`
		}, true},
		{"wrong install id", func(d Day) (int, string) {
			return 200, strings.Replace(okReceipt(d), d.InstallID, strings.Repeat("0", 32), 1)
		}, true},
		{"wrong day", func(d Day) (int, string) {
			return 200, strings.Replace(okReceipt(d), d.Day, "2026-02-28", 1)
		}, true},
		{"empty body", func(d Day) (int, string) { return 200, "" }, true},
		{"malformed", func(d Day) (int, string) { return 200, `{"ok": tru` }, true},
		{"ok as a string", func(d Day) (int, string) {
			return 200, strings.Replace(okReceipt(d), `"ok": true`, `"ok": "true"`, 1)
		}, true},
		{"extra field", func(d Day) (int, string) {
			return 200, strings.Replace(okReceipt(d), `{`, `{"debug": 1, `, 1)
		}, true},
		{"refusal without retryable", func(d Day) (int, string) {
			return 200, `{"ok": false, "error": "invalid_day"}`
		}, true},
		{"server error", func(d Day) (int, string) { return 500, okReceipt(d) }, true},
		{"huge body", func(d Day) (int, string) {
			return 200, okReceipt(d) + strings.Repeat(" ", 17*1024)
		}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ts, _ := flushServer(t, c.reply)
			s := storeWithDays(t, ts)
			if err := Flush(context.Background(), s, clientFor(ts)); err != nil {
				t.Fatal(err)
			}
			if got := dayKept(t, s, "2026-03-01"); got != c.kept {
				t.Fatalf("day kept = %v, want %v", got, c.kept)
			}
		})
	}
}

func TestFlushRefusesRedirectsOffGoogle(t *testing.T) {
	var hit atomic.Int32
	plain := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit.Add(1)
	}))
	defer plain.Close()
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, plain.URL, http.StatusFound)
	}))
	defer ts.Close()
	s := storeWithDays(t, ts)
	if err := Flush(context.Background(), s, clientFor(ts)); err != nil {
		t.Fatal(err)
	}
	if hit.Load() != 0 {
		t.Fatal("flush followed a redirect to plain http")
	}
	if !dayKept(t, s, "2026-03-01") {
		t.Fatal("a refused redirect deleted the day")
	}
}

func TestFlushTimesOut(t *testing.T) {
	ts, _ := flushServer(t, func(d Day) (int, string) {
		time.Sleep(300 * time.Millisecond)
		return 200, okReceipt(d)
	})
	s := storeWithDays(t, ts)
	c := clientFor(ts)
	c.Timeout = 50 * time.Millisecond
	if err := Flush(context.Background(), s, c); err != nil {
		t.Fatal(err)
	}
	if !dayKept(t, s, "2026-03-01") {
		t.Fatal("a timed out send deleted the day")
	}
}

func TestFlushStopsWhenAnotherFlushRuns(t *testing.T) {
	ts, calls := flushServer(t, func(d Day) (int, string) { return 200, okReceipt(d) })
	s := storeWithDays(t, ts)
	f, err := os.OpenFile(filepath.Join(s.dir, "flush.lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatal(err)
	}
	if err := Flush(context.Background(), s, clientFor(ts)); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 0 {
		t.Fatal("a second flush sent while the first held the lock")
	}
}

func TestFlushSendsNothingWhenSwitchedOff(t *testing.T) {
	ts, calls := flushServer(t, func(d Day) (int, string) { return 200, okReceipt(d) })
	s := storeWithDays(t, ts)
	t.Setenv("DO_NOT_TRACK", "1")
	if err := Flush(context.Background(), s, clientFor(ts)); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 0 {
		t.Fatal("flush sent under DO_NOT_TRACK")
	}
}
