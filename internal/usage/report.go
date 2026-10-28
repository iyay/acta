package usage

import (
	"sort"
	"time"
)

// The report statuses. Partial means some data could not be counted, never
// that the totals are zero.
const (
	StatusComplete = "complete"
	StatusPartial  = "partial"
	StatusNoData   = "no_data"
)

// schemaVersion is the contract of the JSON fields below.
const schemaVersion = 1

// historyCoverage says the report covers the logs that exist now.
const historyCoverage = "discovered_logs_only"

// HistoryLimitation is the sentence the report and the command help carry,
// so the two cannot drift apart.
const HistoryLimitation = "Sessions whose only checkout was deleted or moved may be absent."

// Report is one usage report. It is built to be printed as JSON as it is.
type Report struct {
	SchemaVersion int          `json:"schema_version"`
	Status        string       `json:"status"`
	Scope         Scope        `json:"scope"`
	Range         Span         `json:"range"`
	Sessions      []SessionRow `json:"sessions"`
	Models        []ModelRow   `json:"models"`
	Coverage      Coverage     `json:"coverage"`
	Diagnostics   []Diagnostic `json:"diagnostics"`
}

// Scope says which repository, which checkouts and which logs the report
// read, so nobody reads it as complete repository history.
type Scope struct {
	Repository      string   `json:"repository"`
	Checkouts       []string `json:"checkouts"`
	SourceDirectory string   `json:"source_directory"`
	HistoryCoverage string   `json:"history_coverage"`
	Limitations     []string `json:"limitations"`
}

// Span is the first and last response time of a row or of the whole report.
// A nil endpoint is a time no included response could give.
type Span struct {
	FirstResponseAt *string `json:"first_response_at"`
	LastResponseAt  *string `json:"last_response_at"`
}

// SessionRow is the usage of one root session under one model.
type SessionRow struct {
	SessionID         string `json:"session_id"`
	Model             string `json:"model"`
	Range             Span   `json:"range"`
	FreshInputTokens  int64  `json:"fresh_input_tokens"`
	CachedInputTokens int64  `json:"cached_input_tokens"`
	OutputTokens      int64  `json:"output_tokens"`
	ResponseCount     int    `json:"response_count"`
}

// ModelRow is the usage of one model across every root session.
type ModelRow struct {
	Model             string `json:"model"`
	FreshInputTokens  int64  `json:"fresh_input_tokens"`
	CachedInputTokens int64  `json:"cached_input_tokens"`
	OutputTokens      int64  `json:"output_tokens"`
	ResponseCount     int    `json:"response_count"`
}

// Coverage says how much of the discovered history reached the totals.
type Coverage struct {
	Partial                   bool `json:"partial"`
	MatchedRootSessions       int  `json:"matched_root_sessions"`
	IncludedThreadLogs        int  `json:"included_thread_logs"`
	CountedResponses          int  `json:"counted_responses"`
	ExcludedInvalidRecords    int  `json:"excluded_invalid_records"`
	ConflictingResponses      int  `json:"conflicting_responses"`
	LogsWithoutSupportedUsage int  `json:"logs_without_supported_usage"`
	UnclassifiedLogs          int  `json:"unclassified_logs"`
	PendingLines              int  `json:"pending_lines"`
	UndatedResponses          int  `json:"undated_responses"`
}

// formatTime writes a response time the way the JSON contract asks: UTC,
// RFC 3339, with fractional seconds only when they are not zero.
func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

// spanBuilder keeps the first and last dated response of a row or report.
// Undated responses are left out, because a made-up time would be worse.
type spanBuilder struct {
	first time.Time
	last  time.Time
}

func (b *spanBuilder) add(at time.Time) {
	if at.IsZero() {
		return
	}
	if b.first.IsZero() || at.Before(b.first) {
		b.first = at
	}
	if at.After(b.last) {
		b.last = at
	}
}

func (b *spanBuilder) span() Span {
	var s Span
	if !b.first.IsZero() {
		s.FirstResponseAt = new(formatTime(b.first))
	}
	if !b.last.IsZero() {
		s.LastResponseAt = new(formatTime(b.last))
	}
	return s
}

// sortDiagnostics puts the report's diagnostics in the order the contract
// asks for: file, then line, then code, with the null values last.
func sortDiagnostics(d []Diagnostic) {
	sort.SliceStable(d, func(i, j int) bool {
		a, b := d[i], d[j]
		if (a.File == nil) != (b.File == nil) {
			return a.File != nil
		}
		if a.File != nil && *a.File != *b.File {
			return *a.File < *b.File
		}
		if (a.Line == nil) != (b.Line == nil) {
			return a.Line != nil
		}
		if a.Line != nil && *a.Line != *b.Line {
			return *a.Line < *b.Line
		}
		return a.Code < b.Code
	})
}
