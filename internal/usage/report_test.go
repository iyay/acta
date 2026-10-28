package usage

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// writeRawLog writes one transcript exactly as given, so a test can leave a
// final line unterminated.
func writeRawLog(t *testing.T, sessions, rel, body string) {
	t.Helper()
	path := filepath.Join(sessions, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func reportCodes(r Report) []string {
	var out []string
	for _, d := range r.Diagnostics {
		out = append(out, d.Code)
	}
	return out
}

func hasCode(r Report, code string) bool {
	for _, c := range reportCodes(r) {
		if c == code {
			return true
		}
	}
	return false
}

// A response that appears in a root log and in copies of a child log counts
// once, while copies that disagree are dropped as a whole.
func TestUsageReportCountsEachResponseOnce(t *testing.T) {
	repo := newRepo(t)
	sessions := t.TempDir()
	writeLog(t, sessions, "a/root.jsonl",
		metaLine("root-1", repo),
		modelLine("m1"),
		recordLine("2026-10-28T10:00:02Z", "root-1", "root-1", "r1", 100, 60, 10),
	)
	writeLog(t, sessions, "b/copy.jsonl",
		metaLine("child-1", repo),
		modelLine("m1"),
		recordLine("2026-10-28T10:00:02Z", "child-1", "root-1", "r1", 100, 60, 10),
	)
	writeLog(t, sessions, "c/conflict-a.jsonl",
		metaLine("child-2", repo),
		recordLine("2026-10-28T10:00:03Z", "child-2", "root-1", "r2", 10, 0, 1),
	)
	writeLog(t, sessions, "d/conflict-b.jsonl",
		metaLine("child-3", repo),
		recordLine("2026-10-28T10:00:03Z", "child-3", "root-1", "r2", 20, 0, 1),
	)

	r := mustRead(t, repo, sessions)
	if r.Coverage.CountedResponses != 1 || r.Coverage.ConflictingResponses != 1 {
		t.Fatalf("coverage = %+v, want one counted response and one conflict", r.Coverage)
	}
	if len(r.Sessions) != 1 {
		t.Fatalf("sessions = %v", sessionIDs(r))
	}
	row := r.Sessions[0]
	if row.FreshInputTokens != 40 || row.CachedInputTokens != 60 || row.OutputTokens != 10 || row.ResponseCount != 1 {
		t.Fatalf("row = %+v, want the single counted response", row)
	}
	if !hasCode(r, codeConflictingResponse) {
		t.Fatalf("diagnostics = %v, want a conflicting response", reportCodes(r))
	}
}

// Which copy is read first must not decide the outcome.
func TestUsageReportConflictIndependentOfScanOrder(t *testing.T) {
	repo := newRepo(t)
	build := func(names [2]string) Report {
		sessions := t.TempDir()
		writeLog(t, sessions, names[0],
			metaLine("child-a", repo),
			recordLine("2026-10-28T10:00:03Z", "child-a", "root-1", "r2", 10, 0, 1),
		)
		writeLog(t, sessions, names[1],
			metaLine("child-b", repo),
			recordLine("2026-10-28T10:00:03Z", "child-b", "root-1", "r2", 20, 0, 1),
		)
		return mustRead(t, repo, sessions)
	}
	forward := build([2]string{"a.jsonl", "b.jsonl"})
	backward := build([2]string{"b.jsonl", "a.jsonl"})
	if !reflect.DeepEqual(forward.Sessions, backward.Sessions) ||
		!reflect.DeepEqual(forward.Coverage, backward.Coverage) ||
		!reflect.DeepEqual(forward.Diagnostics, backward.Diagnostics) {
		t.Fatalf("scan order changed the report:\nforward %+v\nbackward %+v", forward, backward)
	}
	if forward.Coverage.CountedResponses != 0 || forward.Coverage.ConflictingResponses != 1 {
		t.Fatalf("coverage = %+v, want the conflict excluded", forward.Coverage)
	}
}

// Rows carry one root and model each, sort by first response, and never
// lose the unknown model group.
func TestUsageReportRowsAndModels(t *testing.T) {
	repo := newRepo(t)
	sessions := t.TempDir()
	writeLog(t, sessions, "a/root-1.jsonl",
		metaLine("root-1", repo),
		recordLine("2026-10-28T09:00:00Z", "root-1", "root-1", "r1", 10, 0, 1),
		modelLine("m-a"),
		recordLine("2026-10-28T10:00:00Z", "root-1", "root-1", "r2", 20, 5, 2),
		modelLine("m-b"),
		recordLine("2026-10-28T11:00:00Z", "root-1", "root-1", "r3", 30, 10, 3),
	)
	writeLog(t, sessions, "b/root-2.jsonl",
		metaLine("root-2", repo),
		modelLine("m-a"),
		recordLine("2026-10-28T08:00:00Z", "root-2", "root-2", "r4", 40, 0, 4),
		modelLine("m-b"),
		recordLine("no-a-time", "root-2", "root-2", "r5", 50, 0, 5),
	)

	r := mustRead(t, repo, sessions)
	want := []string{
		"root-2/m-a",
		"root-1/" + UnknownModel,
		"root-1/m-a",
		"root-1/m-b",
		"root-2/m-b",
	}
	if got := sessionIDs(r); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("sessions = %v, want %v", got, want)
	}
	if got, wantModels := modelNames(r), []string{"m-a", "m-b", UnknownModel}; strings.Join(got, ",") != strings.Join(wantModels, ",") {
		t.Fatalf("models = %v, want %v", got, wantModels)
	}
	if r.Range.FirstResponseAt == nil || *r.Range.FirstResponseAt != "2026-10-28T08:00:00Z" {
		t.Errorf("first response = %v", r.Range.FirstResponseAt)
	}
	if r.Range.LastResponseAt == nil || *r.Range.LastResponseAt != "2026-10-28T11:00:00Z" {
		t.Errorf("last response = %v", r.Range.LastResponseAt)
	}
	undated := r.Sessions[4]
	if undated.Range.FirstResponseAt != nil || undated.Range.LastResponseAt != nil {
		t.Errorf("undated row range = %+v, want null endpoints", undated.Range)
	}
	if r.Coverage.UndatedResponses != 1 || !r.Coverage.Partial {
		t.Errorf("coverage = %+v", r.Coverage)
	}
	// Fresh input is input minus cached, counted per response.
	if got := undated.FreshInputTokens; got != 50 {
		t.Errorf("undated fresh input = %d, want 50", got)
	}
	for _, m := range r.Models {
		var fresh, cached, output int64
		var count int
		for _, s := range r.Sessions {
			if s.Model != m.Model {
				continue
			}
			fresh += s.FreshInputTokens
			cached += s.CachedInputTokens
			output += s.OutputTokens
			count += s.ResponseCount
		}
		if m.FreshInputTokens != fresh || m.CachedInputTokens != cached || m.OutputTokens != output || m.ResponseCount != count {
			t.Errorf("model %s = %+v, want the sum of its rows", m.Model, m)
		}
	}
}

func modelNames(r Report) []string {
	var out []string
	for _, m := range r.Models {
		out = append(out, m.Model)
	}
	return out
}

// Coverage names every counter the spec defines, so a partial report says
// what it could not count.
func TestUsageReportCoverageCounters(t *testing.T) {
	repo := newRepo(t)
	sessions := t.TempDir()
	writeLog(t, sessions, "a/root-1.jsonl",
		metaLine("root-1", repo),
		modelLine("m1"),
		recordLine("2026-10-28T10:00:02Z", "root-1", "root-1", "r1", 10, 0, 1),
		recordLine("2026-10-28T10:00:03Z", "root-1", "root-1", "r2", 10, 11, 1),
		`{"timestamp":"2026-10-28T10:00:04Z","type":"token_usage_record","payload":{"thread_id":"root-1","session_id":"root-1","usage":{"input_tokens":1,"cached_input_tokens":0,"output_tokens":1}}}`,
	)
	writeRawLog(t, sessions, "b/root-2.jsonl",
		metaLine("root-2", repo)+"\n"+
			`{"timestamp":"2026-10-28T10:00:05Z","type":"token_usage_record","payload":{"thread_id":"root-2","session_id":"root-2","response_id":"r-pending","usage":{"input_tokens":1,"cached_input_tokens":0,"output_tokens":1}}}`)
	writeLog(t, sessions, "c/no-meta.jsonl",
		modelLine("m1"),
		recordLine("2026-10-28T10:00:06Z", "root-9", "root-9", "r9", 1, 0, 1),
	)

	r := mustRead(t, repo, sessions)
	want := Coverage{
		Partial:                   true,
		MatchedRootSessions:       2,
		IncludedThreadLogs:        2,
		CountedResponses:          1,
		ExcludedInvalidRecords:    2,
		ConflictingResponses:      0,
		LogsWithoutSupportedUsage: 1,
		UnclassifiedLogs:          1,
		PendingLines:              1,
		UndatedResponses:          0,
	}
	if r.Coverage != want {
		t.Fatalf("coverage = %+v\nwant      %+v", r.Coverage, want)
	}
}

// The JSON report follows the version 1 contract: stable names, integers,
// UTC times, null for an unknown range end, and no response identifiers.
func TestUsageReportJSONContract(t *testing.T) {
	repo := newRepo(t)
	sessions := t.TempDir()
	writeRawLog(t, sessions, "log.jsonl",
		metaLine("root-1", repo)+"\n"+
			`{"timestamp":"2026-10-28T10:00:01Z","type":"turn_context","payload":{"model":"example-model"}}`+"\n"+
			recordLine("2026-10-28T10:00:02.125Z", "root-1", "root-1", "response-1", 100, 60, 10)+"\n"+
			`{"timestamp":"2026-10-28T10:00:03Z","type":"token_usage_record","payload":`)

	r := mustRead(t, repo, sessions)
	if r.SchemaVersion != 1 || r.Status != statusPartial {
		t.Fatalf("schema = %d status = %q", r.SchemaVersion, r.Status)
	}
	if r.Scope.Repository != canonicalPath(t, repo) {
		t.Errorf("repository = %q", r.Scope.Repository)
	}
	if r.Scope.SourceDirectory != sessions {
		t.Errorf("source directory = %q, want %q", r.Scope.SourceDirectory, sessions)
	}
	if r.Scope.HistoryCoverage != "discovered_logs_only" || len(r.Scope.Limitations) != 1 {
		t.Errorf("scope = %+v", r.Scope)
	}
	if len(r.Sessions) != 1 {
		t.Fatalf("sessions = %v", sessionIDs(r))
	}
	row := r.Sessions[0]
	if row.SessionID != "root-1" || row.Model != "example-model" ||
		row.FreshInputTokens != 40 || row.CachedInputTokens != 60 || row.OutputTokens != 10 || row.ResponseCount != 1 {
		t.Errorf("row = %+v", row)
	}
	if row.Range.FirstResponseAt == nil || *row.Range.FirstResponseAt != "2026-10-28T10:00:02.125Z" {
		t.Errorf("row range = %+v", row.Range)
	}
	if r.Coverage.PendingLines != 1 || r.Coverage.CountedResponses != 1 {
		t.Errorf("coverage = %+v", r.Coverage)
	}
	if len(r.Diagnostics) != 1 || r.Diagnostics[0].Code != codePendingLine ||
		r.Diagnostics[0].File == nil || *r.Diagnostics[0].File != "log.jsonl" ||
		r.Diagnostics[0].Line == nil || *r.Diagnostics[0].Line != 4 {
		t.Fatalf("diagnostics = %+v", r.Diagnostics)
	}

	body, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(body, &keys); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"schema_version", "status", "scope", "range", "sessions", "models", "coverage", "diagnostics"} {
		if _, ok := keys[k]; !ok {
			t.Errorf("JSON has no %q", k)
		}
	}
	if strings.Contains(string(body), "response-1") {
		t.Errorf("JSON shows a response identifier: %s", body)
	}
}

// No matching session is no data, not a measured zero.
func TestUsageReportNoData(t *testing.T) {
	repo := newRepo(t)
	r := mustRead(t, repo, t.TempDir())
	if r.Status != statusNoData || r.Coverage.Partial {
		t.Fatalf("status = %q partial = %v", r.Status, r.Coverage.Partial)
	}
	if len(r.Sessions) != 0 || len(r.Models) != 0 || len(r.Diagnostics) != 0 {
		t.Fatalf("report = %+v, want empty rows", r)
	}
	if r.Range.FirstResponseAt != nil || r.Range.LastResponseAt != nil {
		t.Fatalf("range = %+v, want null endpoints", r.Range)
	}
	body, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"sessions":[]`) || !strings.Contains(string(body), `"models":[]`) {
		t.Fatalf("JSON empty rows must be arrays: %s", body)
	}
	if !strings.Contains(string(body), `"first_response_at":null`) {
		t.Fatalf("JSON range must be null: %s", body)
	}
}

// A total that would wrap int64 is refused instead of reported wrong.
func TestUsageReportOverflow(t *testing.T) {
	repo := newRepo(t)
	sessions := t.TempDir()
	writeLog(t, sessions, "a.jsonl",
		metaLine("root-1", repo),
		modelLine("m1"),
		recordLine("2026-10-28T10:00:02Z", "root-1", "root-1", "r1", math.MaxInt64, 0, 0),
		recordLine("2026-10-28T10:00:03Z", "root-1", "root-1", "r2", math.MaxInt64, 0, 0),
	)
	if _, err := Read(repo, sessions); err == nil {
		t.Fatal("Read: got no error from a wrapping total")
	}
}

// A response the turn gave no model for stays in its own group, and every
// counted response is one row of one root session.
func TestUsageReportUnknownModelKept(t *testing.T) {
	repo := newRepo(t)
	sessions := t.TempDir()
	writeLog(t, sessions, "a.jsonl",
		metaLine("root-1", repo),
		recordLine("2026-10-28T10:00:02Z", "root-1", "root-1", "r1", 10, 0, 1),
	)
	r := mustRead(t, repo, sessions)
	if got := sessionIDs(r); strings.Join(got, ",") != "root-1/"+UnknownModel {
		t.Fatalf("sessions = %v", got)
	}
	if r.Models[0].Model != UnknownModel || r.Models[0].ResponseCount != 1 {
		t.Fatalf("models = %+v", r.Models)
	}
}
