package update

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// CheckEvery is how long a cached answer stays good. A daily check is enough
// to hear about a release without asking GitHub on every run.
const CheckEvery = 24 * time.Hour

type cache struct {
	Latest  string    `json:"latest"`
	Checked time.Time `json:"checked"`
}

// ReadCache returns the newest version the last check saw. A missing or
// broken file gives "", which means no notice.
func ReadCache(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var c cache
	if json.Unmarshal(data, &c) != nil {
		return ""
	}
	return c.Latest
}

// WriteCache saves the latest version. It writes a temp file and renames it,
// so a reader never sees half a file.
func WriteCache(path, latest string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(cache{Latest: latest, Checked: time.Now().UTC()})
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".update-*.json")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return nil
}

// Stale says a new check is due: the file is missing, or it was last touched
// more than CheckEvery ago. The file time is used, not the time inside it,
// because a run touches the file before it starts a check.
func Stale(path string, now time.Time) bool {
	info, err := os.Stat(path)
	if err != nil {
		return true
	}
	return now.Sub(info.ModTime()) > CheckEvery
}

// Newer says cached is a higher release than current. Both must look like
// v1.2.3; anything else never shows a notice.
func Newer(cached, current string) bool {
	a, okA := parseVersion(cached)
	b, okB := parseVersion(current)
	if !okA || !okB {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return false
}

func parseVersion(v string) ([3]int, bool) {
	var out [3]int
	rest, ok := strings.CutPrefix(v, "v")
	if !ok {
		return out, false
	}
	parts := strings.Split(rest, ".")
	if len(parts) != 3 {
		return out, false
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return out, false
		}
		out[i] = n
	}
	return out, true
}
