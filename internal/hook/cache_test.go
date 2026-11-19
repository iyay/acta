package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var cacheNow = time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)

// usageLine is one assistant line of a Claude Code transcript, with the
// given age and context split over the three input counters.
func usageLine(age time.Duration, context int, model string) string {
	ts := cacheNow.Add(-age).Format(time.RFC3339Nano)
	third := context / 3
	return fmt.Sprintf(`{"type":"assistant","timestamp":%q,"message":{"model":%q,"usage":{"input_tokens":%d,"cache_creation_input_tokens":%d,"cache_read_input_tokens":%d,"output_tokens":50}}}`,
		ts, model, third, third, context-2*third)
}

func writeTranscript(t *testing.T, lines ...string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "t.jsonl")
	if err := os.WriteFile(p, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

const userLine = `{"type":"user","timestamp":"2026-11-20T11:59:00Z","message":{"content":"hi"}}`

func TestCacheNudge(t *testing.T) {
	const model = "claude-opus-5-5"
	stale := usageLine(2*time.Hour, 200000, model)
	tests := []struct {
		name  string
		lines []string
		want  string // empty means no line
	}{
		{"stale and big gives the line", []string{userLine, stale, userLine},
			"acta cache: 2 h idle with ~200k tokens of context, so the prompt cache has likely expired and this turn pays for the whole context again. Tell the user in one line and suggest /compact to go on from a summary, or /clear for a new task."},
		{"30 min idle gives none", []string{usageLine(30*time.Minute, 200000, model)}, ""},
		{"exactly one hour gives none", []string{usageLine(time.Hour, 200000, model)}, ""},
		{"exactly 150000 gives the line", []string{usageLine(2*time.Hour, 150000, model)}, "~150k"},
		{"100k context gives none", []string{usageLine(2*time.Hour, 100000, model)}, ""},
		{"synthetic last line is skipped", []string{stale, usageLine(time.Minute, 300000, "<synthetic>")}, "~200k"},
		{"broken last line is skipped", []string{stale, `{"type":"assistant",`}, "~200k"},
		{"compact boundary after usage gives none", []string{stale, `{"type":"system","subtype":"compact_boundary","timestamp":"2026-11-20T11:00:00Z"}`, userLine}, ""},
		{"bad timestamp gives none", []string{stale, `{"type":"assistant","timestamp":"yesterday","message":{"model":"m","usage":{"input_tokens":200000}}}`}, ""},
		{"missing timestamp gives none", []string{stale, `{"type":"assistant","message":{"model":"m","usage":{"input_tokens":200000}}}`}, ""},
		{"gap 119 min", []string{usageLine(119*time.Minute, 200000, model)}, "acta cache: 119 min idle"},
		{"gap 47 h", []string{usageLine(47*time.Hour, 200000, model)}, "acta cache: 47 h idle"},
		{"gap 2 days", []string{usageLine(48*time.Hour, 200000, model)}, "acta cache: 2 days idle"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CacheNudge(writeTranscript(t, tc.lines...), cacheNow)
			if tc.want == "" {
				if got != "" {
					t.Fatalf("want no line, got %q", got)
				}
				return
			}
			if !strings.Contains(got, tc.want) {
				t.Fatalf("want %q in line, got %q", tc.want, got)
			}
		})
	}
}

func TestCacheNudgeMissingFile(t *testing.T) {
	if got := CacheNudge(filepath.Join(t.TempDir(), "none.jsonl"), cacheNow); got != "" {
		t.Fatalf("want no line, got %q", got)
	}
}

// A file over the tail size cuts its first tail line in the middle. That cut
// line must never count, even when what is left of it parses as usage.
func TestCacheNudgeCutFirstLine(t *testing.T) {
	stale := usageLine(2*time.Hour, 200000, "m")
	// JSON allows trailing spaces, so this line plus its newline fills the
	// tail exactly. Junk in front of it on the same line makes the whole line
	// broken, but the tail starts right where the usage part begins.
	tail := stale + strings.Repeat(" ", cacheTailBytes-len(stale)-1)
	p := writeTranscript(t, "junk"+tail)
	if got := CacheNudge(p, cacheNow); got != "" {
		t.Fatalf("cut line counted: %q", got)
	}
	// The same file with a whole stale line after it still gives the line.
	p = writeTranscript(t, "junk"+tail, stale)
	if got := CacheNudge(p, cacheNow); !strings.Contains(got, "~200k") {
		t.Fatalf("whole line after the cut not read: %q", got)
	}
}
