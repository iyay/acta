package update

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewer(t *testing.T) {
	for _, tc := range []struct {
		cached, current string
		want            bool
	}{
		{"v0.1.47", "v0.1.46", true},
		{"v0.2.0", "v0.1.99", true},
		{"v1.0.0", "v0.9.9", true},
		{"v0.1.10", "v0.1.9", true}, // numbers, not text
		{"v0.1.46", "v0.1.46", false},
		{"v0.1.45", "v0.1.46", false},
		{"v0.1.9", "v0.1.10", false},
		{"0.1.47", "v0.1.46", false}, // missing v
		{"v0.1.47", "0.1.46", false},
		{"v0.1", "v0.1.46", false},
		{"v0.1.47-rc1", "v0.1.46", false},
		{"garbage", "v0.1.46", false},
		{"", "v0.1.46", false},
		{"v0.1.47", "", false},
		{"v0.1.47", "dev", false},
	} {
		if got := Newer(tc.cached, tc.current); got != tc.want {
			t.Errorf("Newer(%q, %q) = %v, want %v", tc.cached, tc.current, got, tc.want)
		}
	}
}

func TestCacheRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "update.json")
	if err := WriteCache(path, "v0.1.47"); err != nil {
		t.Fatal(err)
	}
	if got := ReadCache(path); got != "v0.1.47" {
		t.Fatalf("ReadCache = %q, want v0.1.47", got)
	}
	// No temp file is left beside the cache.
	entries, _ := os.ReadDir(filepath.Dir(path))
	if len(entries) != 1 {
		t.Fatalf("dir holds %d files, want only update.json", len(entries))
	}
}

func TestReadCacheBadFiles(t *testing.T) {
	dir := t.TempDir()
	for name, body := range map[string]string{
		"empty.json":   "",
		"corrupt.json": "{not json",
		"wrong.json":   `{"latest": 5}`,
	} {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if got := ReadCache(p); got != "" {
			t.Errorf("%s: ReadCache = %q, want empty", name, got)
		}
	}
	if got := ReadCache(filepath.Join(dir, "missing.json")); got != "" {
		t.Errorf("missing: ReadCache = %q, want empty", got)
	}
}

func TestStale(t *testing.T) {
	dir := t.TempDir()
	if !Stale(filepath.Join(dir, "missing.json"), time.Now()) {
		t.Error("missing file must be stale")
	}
	p := filepath.Join(dir, "update.json")
	if err := os.WriteFile(p, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	for _, tc := range []struct {
		age  time.Duration
		want bool
	}{
		{0, false},
		{23 * time.Hour, false},
		{25 * time.Hour, true},
	} {
		mt := now.Add(-tc.age)
		if err := os.Chtimes(p, mt, mt); err != nil {
			t.Fatal(err)
		}
		if got := Stale(p, now); got != tc.want {
			t.Errorf("age %v: Stale = %v, want %v", tc.age, got, tc.want)
		}
	}
}
