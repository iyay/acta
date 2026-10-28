// Package usage counts the model usage that Codex recorded in its local
// session transcripts. It reads those files only: no network, no writes.
package usage

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"
)

// The diagnostic codes the report may carry. They name a recoverable
// problem, never a fatal one.
const (
	codePendingLine         = "pending_line"
	codeInvalidJSON         = "invalid_json"
	codeMissingMetadata     = "missing_metadata"
	codeUnsupportedUsage    = "unsupported_usage"
	codeMissingIdentity     = "missing_identity"
	codeInvalidUsage        = "invalid_usage"
	codeInvalidTimestamp    = "invalid_timestamp"
	codeConflictingResponse = "conflicting_response"
	codeUnreadableLog       = "unreadable_log"
)

// UnknownModel is the group a response joins when no usable turn named its
// model. Dropping those tokens would hide real usage.
const UnknownModel = "unknown"

// Diagnostic is one recoverable problem found while reading. Messages are
// for people; the code is the stable part.
type Diagnostic struct {
	Code    string  `json:"code"`
	File    *string `json:"file"`
	Line    *int    `json:"line"`
	Message string  `json:"message"`
}

// Record is one usage record owned by the thread of a log file. A record
// with Valid false is kept so the report can count it as excluded.
type Record struct {
	ResponseID  string
	RootID      string
	Model       string    // "" when the turn named no usable model
	At          time.Time // zero when the timestamp was missing or invalid
	Input       int64
	CachedInput int64
	Output      int64
	Valid       bool
}

// Log is what one transcript file adds to a report: the thread that owns it,
// where that thread started, its usage records and what went wrong.
type Log struct {
	ThreadID     string
	Cwd          string
	Records      []Record
	Diagnostics  []Diagnostic
	PendingLines int
}

// maxLine bounds one line in memory. A supported record is a few hundred
// bytes; only conversation lines get near this, and they are skipped anyway.
const maxLine = 32 << 20

// ReadLog reads one transcript stream. It keeps the file's identity, its
// usable metadata and its usage records, and never retains conversation or
// tool text. The only error is a failing reader.
func ReadLog(name string, r io.Reader) (Log, error) {
	var log Log
	var candidates []candidate
	identitySet := false
	model := ""

	br := bufio.NewReaderSize(r, 64*1024)
	for line := 1; ; line++ {
		raw, complete, tooLong, err := readLine(br)
		if err == io.EOF {
			break
		}
		if err != nil {
			return Log{}, fmt.Errorf("read %s: %w", name, err)
		}
		if !complete {
			// Codex writes each record in one write, so a final line with no
			// newline is a write in progress. Read it next time.
			log.PendingLines++
			log.diag(name, codePendingLine, line, "The final line is not terminated and was not read.")
			break
		}
		if len(bytes.TrimSpace(raw)) == 0 {
			continue
		}
		if tooLong {
			log.diag(name, codeInvalidJSON, line, "This line is too large to read and was skipped.")
			continue
		}

		var ev event
		if err := json.Unmarshal(raw, &ev); err != nil {
			log.diag(name, codeInvalidJSON, line, "This line is not valid JSON and was skipped.")
			continue
		}
		switch ev.Type {
		case "session_meta":
			if identitySet {
				continue
			}
			var p metaFields
			if err := json.Unmarshal(ev.Payload, &p); err != nil {
				continue
			}
			id, ok := rawString(p.ID)
			if !ok || id == "" {
				continue
			}
			log.ThreadID = id
			if cwd, ok := rawString(p.Cwd); ok {
				log.Cwd = cwd
			}
			identitySet = true
		case "turn_context":
			// A model that is missing, empty or not a string means unknown,
			// never the machine's default.
			model = ""
			var p metaFields
			if err := json.Unmarshal(ev.Payload, &p); err == nil {
				if m, ok := rawString(p.Model); ok {
					model = m
				}
			}
		case "token_usage_record":
			candidates = append(candidates, readCandidate(ev, model, line))
		}
	}

	// Ownership needs the file identity, which may arrive late, so the
	// records are settled once the whole file is read.
	if log.ThreadID == "" {
		log.diag(name, codeMissingMetadata, 0, "This log has no usable session metadata, so its usage cannot be attributed.")
	} else {
		for _, c := range candidates {
			if c.threadOK && c.thread != log.ThreadID {
				continue // a record inherited from another thread; it counts there
			}
			rec := c.rec
			code, message := c.problem()
			if code == "" {
				rec.Valid = true
			} else {
				log.diag(name, code, c.line, message)
			}
			if !c.timeOK {
				log.diag(name, codeInvalidTimestamp, c.line, "This response has no usable timestamp, so it is counted without a date.")
			}
			log.Records = append(log.Records, rec)
		}
	}

	sort.SliceStable(log.Diagnostics, func(i, j int) bool {
		a, b := log.Diagnostics[i], log.Diagnostics[j]
		if a.Line == nil || b.Line == nil {
			// A file-level diagnostic comes last, because it names no line.
			return a.Line != nil
		}
		if *a.Line != *b.Line {
			return *a.Line < *b.Line
		}
		return a.Code < b.Code
	})
	return log, nil
}

// diag adds one diagnostic about this file. line 0 means the whole file.
func (l *Log) diag(name, code string, line int, message string) {
	d := Diagnostic{Code: code, Message: message}
	if name != "" {
		d.File = &name
	}
	if line > 0 {
		d.Line = &line
	}
	l.Diagnostics = append(l.Diagnostics, d)
}

// event is the shared envelope. Only the fields the report needs are read,
// so conversation text is never kept.
type event struct {
	Timestamp json.RawMessage `json:"timestamp"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
}

// metaFields holds the nested fields of session_meta and turn_context. They
// stay raw so a value of the wrong type is refused instead of guessed.
type metaFields struct {
	ID    json.RawMessage `json:"id"`
	Cwd   json.RawMessage `json:"cwd"`
	Model json.RawMessage `json:"model"`
}

// recordFields holds the nested fields of token_usage_record.
type recordFields struct {
	ThreadID   json.RawMessage `json:"thread_id"`
	SessionID  json.RawMessage `json:"session_id"`
	ResponseID json.RawMessage `json:"response_id"`
	Usage      json.RawMessage `json:"usage"`
}

// usageFields holds the per-response counters. The cumulative turn and
// thread totals are deliberately not read: they would count responses twice.
type usageFields struct {
	Input       json.RawMessage `json:"input_tokens"`
	CachedInput json.RawMessage `json:"cached_input_tokens"`
	Output      json.RawMessage `json:"output_tokens"`
}

// candidate is a parsed usage record waiting for the file identity. Its
// diagnostics wait too, because an inherited record must stay silent.
type candidate struct {
	rec        Record
	thread     string
	threadOK   bool
	rootOK     bool
	responseOK bool
	timeOK     bool
	usageCode  string // "" when the usage counters are usable
	usageMsg   string
	line       int
}

// problem says why this record cannot be counted, or "" when it can.
func (c candidate) problem() (code, message string) {
	switch {
	case !c.threadOK || !c.rootOK || !c.responseOK:
		return codeMissingIdentity, "This usage record has no usable thread, root session or response identity and is not counted."
	case c.usageCode != "":
		return c.usageCode, c.usageMsg
	}
	return "", ""
}

// readCandidate parses one usage record. It reads the timestamp and the
// counters here, but leaves the identity verdict to the caller.
func readCandidate(ev event, model string, line int) candidate {
	c := candidate{line: line}
	c.rec.Model = model
	c.rec.At, c.timeOK = parseTimestamp(ev.Timestamp)

	var p recordFields
	if err := json.Unmarshal(ev.Payload, &p); err != nil {
		c.usageCode = codeUnsupportedUsage
		c.usageMsg = "This usage record has no supported payload and is not counted."
		return c
	}
	c.thread, c.threadOK = rawString(p.ThreadID)
	c.threadOK = c.threadOK && c.thread != ""
	c.rec.RootID, c.rootOK = rawString(p.SessionID)
	c.rootOK = c.rootOK && c.rec.RootID != ""
	c.rec.ResponseID, c.responseOK = rawString(p.ResponseID)
	c.responseOK = c.responseOK && c.rec.ResponseID != ""

	input, cached, output, code, message := readUsage(p.Usage)
	c.rec.Input, c.rec.CachedInput, c.rec.Output = input, cached, output
	c.usageCode, c.usageMsg = code, message
	return c
}

// readUsage checks the three counters the report counts. A usage object that
// is absent or not an object is unsupported; bad values inside it are
// invalid. Cached input above input cannot happen and would skew fresh input.
func readUsage(m json.RawMessage) (input, cached, output int64, code, message string) {
	if !isObject(m) {
		return 0, 0, 0, codeUnsupportedUsage, "This usage record has no supported usage object and is not counted."
	}
	var u usageFields
	if err := json.Unmarshal(m, &u); err != nil {
		return 0, 0, 0, codeUnsupportedUsage, "This usage record has no supported usage object and is not counted."
	}
	in, okIn := counter(u.Input)
	ca, okCa := counter(u.CachedInput)
	out, okOut := counter(u.Output)
	switch {
	case !okIn || !okCa || !okOut:
		return 0, 0, 0, codeInvalidUsage, "This response has a missing, negative or unsafe token count and is not counted."
	case ca > in:
		return 0, 0, 0, codeInvalidUsage, "This response counts more cached input than input and is not counted."
	}
	return in, ca, out, "", ""
}

// counter accepts a whole, nonnegative number that fits int64.
func counter(m json.RawMessage) (int64, bool) {
	n, ok := rawInt64(m)
	if !ok || n < 0 {
		return 0, false
	}
	return n, true
}

// parseTimestamp reads the top-level RFC 3339 timestamp and normalizes it to
// UTC. An unusable value gives a zero time, which the report shows as undated.
func parseTimestamp(m json.RawMessage) (time.Time, bool) {
	s, ok := rawString(m)
	if !ok || s == "" {
		return time.Time{}, false
	}
	at, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return time.Time{}, false
	}
	return at.UTC(), true
}

// rawString reads one JSON value that must be a string.
func rawString(m json.RawMessage) (string, bool) {
	if len(m) == 0 {
		return "", false
	}
	var v any
	if err := json.Unmarshal(m, &v); err != nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// rawInt64 reads one JSON number that must be a whole number inside int64.
// json.Number keeps the digits, so a large count is refused, not rounded.
func rawInt64(m json.RawMessage) (int64, bool) {
	if len(m) == 0 {
		return 0, false
	}
	dec := json.NewDecoder(bytes.NewReader(m))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return 0, false
	}
	n, ok := v.(json.Number)
	if !ok {
		return 0, false
	}
	i, err := n.Int64()
	if err != nil {
		return 0, false
	}
	return i, true
}

// isObject says whether the raw value is a JSON object. null and every other
// value are not.
func isObject(m json.RawMessage) bool {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(m, &obj); err != nil {
		return false
	}
	return obj != nil
}

// readLine reads one line without its newline. complete is false for a last
// line with no newline, the write in progress case. tooLong marks a line
// that did not fit maxLine, which keeps one runaway line from eating memory.
func readLine(r *bufio.Reader) (line []byte, complete, tooLong bool, err error) {
	var buf []byte
	for {
		chunk, e := r.ReadSlice('\n')
		if len(buf) < maxLine {
			room := maxLine - len(buf)
			if len(chunk) > room {
				tooLong = true
				chunk = chunk[:room]
			}
			buf = append(buf, chunk...)
		} else {
			tooLong = true
		}
		if e == bufio.ErrBufferFull {
			continue
		}
		if e != nil {
			if e == io.EOF {
				if len(buf) == 0 && !tooLong {
					return nil, false, false, io.EOF
				}
				return buf, false, tooLong, nil
			}
			return nil, false, false, e
		}
		buf = buf[:len(buf)-1] // drop the newline
		if n := len(buf); n > 0 && buf[n-1] == '\r' {
			buf = buf[:n-1]
		}
		return buf, true, tooLong, nil
	}
}
