package hook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"time"
)

// Claude Code keeps the prompt cache for about an hour. After a longer break
// the next turn pays full price for the whole context again, which hurts only
// when the context is big. These limits are fixed on purpose.
const (
	cacheIdleLimit  = time.Hour
	cacheMinContext = 150000
	cacheTailBytes  = 512 * 1024
)

// transcriptLine holds the few fields of a transcript line the nudge needs.
type transcriptLine struct {
	Type      string `json:"type"`
	Subtype   string `json:"subtype"`
	Timestamp string `json:"timestamp"`
	Message   struct {
		Model string `json:"model"`
		Usage *struct {
			Input         int `json:"input_tokens"`
			CacheCreation int `json:"cache_creation_input_tokens"`
			CacheRead     int `json:"cache_read_input_tokens"`
		} `json:"usage"`
	} `json:"message"`
}

// CacheNudge returns one line for the agent when the session sat idle long
// enough for the prompt cache to expire and the context is big. It returns ""
// in every other case, a file it cannot read included, so the hook never
// fails because of it.
func CacheNudge(transcriptPath string, now time.Time) string {
	tail, cut, err := readTail(transcriptPath, cacheTailBytes)
	if err != nil {
		return ""
	}
	lines := bytes.Split(tail, []byte("\n"))
	// When the tail starts in the middle of a line, that first piece is not a
	// whole line, so it must never count.
	first := 0
	if cut {
		first = 1
	}
	for i := len(lines) - 1; i >= first; i-- {
		var ln transcriptLine
		if json.Unmarshal(lines[i], &ln) != nil {
			continue
		}
		// The user just compacted, so the old context is gone.
		if ln.Type == "system" && ln.Subtype == "compact_boundary" {
			return ""
		}
		u := ln.Message.Usage
		if ln.Type != "assistant" || u == nil || ln.Message.Model == "<synthetic>" {
			continue
		}
		at, err := time.Parse(time.RFC3339Nano, ln.Timestamp)
		if err != nil {
			return ""
		}
		idle := now.Sub(at)
		context := u.Input + u.CacheCreation + u.CacheRead
		if idle <= cacheIdleLimit || context < cacheMinContext {
			return ""
		}
		return fmt.Sprintf("acta cache: %s idle with ~%dk tokens of context, so the prompt cache has likely expired and this turn pays for the whole context again. Tell the user in one line and suggest /compact to go on from a summary, or /clear for a new task.",
			idleGap(idle), int(math.Round(float64(context)/1000)))
	}
	return ""
}

// idleGap prints a break as minutes, hours or days, whichever reads best.
func idleGap(idle time.Duration) string {
	m := math.Round(idle.Minutes())
	switch {
	case m < 120:
		return fmt.Sprintf("%d min", int(m))
	case m < 2880:
		return fmt.Sprintf("%d h", int(math.Round(m/60)))
	default:
		return fmt.Sprintf("%d days", int(math.Round(m/1440)))
	}
}

// readTail returns the last n bytes of a file, and whether it left the start
// of the file out.
func readTail(path string, n int64) ([]byte, bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, false, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, false, err
	}
	start := info.Size() - n
	if start < 0 {
		start = 0
	}
	buf := make([]byte, info.Size()-start)
	if _, err := f.ReadAt(buf, start); err != nil && err != io.EOF {
		return nil, false, err
	}
	return buf, start > 0, nil
}
