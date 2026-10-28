package usage

import (
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

// jsonl joins the lines of one transcript stream. Every complete line ends
// with a newline, the way a finished write looks.
func jsonl(lines ...string) string { return strings.Join(lines, "\n") + "\n" }

const (
	lineMetaRoot  = `{"timestamp":"2026-10-28T10:00:00Z","type":"session_meta","payload":{"id":"root-1","cwd":"/example/repo"}}`
	lineMetaChild = `{"timestamp":"2026-10-28T10:00:00Z","type":"session_meta","payload":{"id":"child-1","cwd":"/example/repo"}}`
	lineTurnModel = `{"timestamp":"2026-10-28T10:00:01Z","type":"turn_context","payload":{"model":"example-model"}}`
	lineUsage     = `{"timestamp":"2026-10-28T10:00:02.125Z","type":"token_usage_record","payload":{"thread_id":"root-1","session_id":"root-1","response_id":"response-1","usage":{"input_tokens":100,"cached_input_tokens":60,"output_tokens":10,"reasoning_output_tokens":4},"turn_token_usage":{"input_tokens":100,"cached_input_tokens":60,"output_tokens":10},"thread_token_usage":{"input_tokens":100,"cached_input_tokens":60,"output_tokens":10}}}`
)

// usageLine builds one usage record with the given payload fields, so a test
// can bend one field without repeating the whole line.
func usageLine(t *testing.T, thread, session, response, usage string) string {
	t.Helper()
	fields := []string{}
	if thread != "" {
		fields = append(fields, `"thread_id":`+thread)
	}
	if session != "" {
		fields = append(fields, `"session_id":`+session)
	}
	if response != "" {
		fields = append(fields, `"response_id":`+response)
	}
	if usage != "" {
		fields = append(fields, `"usage":`+usage)
	}
	return `{"timestamp":"2026-10-28T10:00:02.125Z","type":"token_usage_record","payload":{` +
		strings.Join(fields, ",") + `}}`
}

func mustReadLog(t *testing.T, stream string) Log {
	t.Helper()
	log, err := ReadLog("log.jsonl", strings.NewReader(stream))
	if err != nil {
		t.Fatalf("ReadLog: %v", err)
	}
	return log
}

func wantOneRecord(t *testing.T, log Log) Record {
	t.Helper()
	if len(log.Records) != 1 {
		t.Fatalf("records = %d, want 1: %+v", len(log.Records), log.Records)
	}
	return log.Records[0]
}

func diagCodes(log Log) []string {
	var codes []string
	for _, d := range log.Diagnostics {
		codes = append(codes, d.Code)
	}
	return codes
}

func wantDiag(t *testing.T, log Log, code string, line int) {
	t.Helper()
	for _, d := range log.Diagnostics {
		if d.Code != code {
			continue
		}
		if line == 0 {
			if d.Line != nil {
				t.Fatalf("%s: line = %d, want null", code, *d.Line)
			}
			return
		}
		if d.Line == nil || *d.Line != line {
			t.Fatalf("%s: line = %v, want %d", code, d.Line, line)
		}
		if d.File == nil || *d.File != "log.jsonl" {
			t.Fatalf("%s: file = %v, want log.jsonl", code, d.File)
		}
		return
	}
	t.Fatalf("no %s diagnostic; got %v", code, diagCodes(log))
}

// The supported shapes establish identity, model, timestamp and usage from
// their nested fields only.
func TestReadLogSupportedShapes(t *testing.T) {
	log := mustReadLog(t, jsonl(lineMetaRoot, lineTurnModel, lineUsage))
	if log.ThreadID != "root-1" || log.Cwd != "/example/repo" {
		t.Fatalf("identity = %q at %q", log.ThreadID, log.Cwd)
	}
	rec := wantOneRecord(t, log)
	wantAt := time.Date(2026, 10, 28, 10, 0, 2, 125000000, time.UTC)
	if !rec.At.Equal(wantAt) {
		t.Errorf("at = %v, want %v", rec.At, wantAt)
	}
	if rec.Model != "example-model" || rec.RootID != "root-1" || rec.ResponseID != "response-1" {
		t.Errorf("record = %+v", rec)
	}
	if rec.Input != 100 || rec.CachedInput != 60 || rec.Output != 10 || !rec.Valid {
		t.Errorf("usage = %+v", rec)
	}
	if len(log.Diagnostics) != 0 {
		t.Errorf("diagnostics = %v, want none", diagCodes(log))
	}
	if log.PendingLines != 0 {
		t.Errorf("pending lines = %d", log.PendingLines)
	}
}

// A later session_meta cannot take the file over, and a record of the other
// thread is inherited, so it never counts here.
func TestReadLogFirstIdentityWins(t *testing.T) {
	other := `{"timestamp":"2026-10-28T10:00:00Z","type":"session_meta","payload":{"id":"root-2","cwd":"/elsewhere/repo"}}`
	intruder := usageLine(t, `"root-2"`, `"root-2"`, `"response-2"`, `{"input_tokens":5,"cached_input_tokens":0,"output_tokens":1}`)
	log := mustReadLog(t, jsonl(lineMetaRoot, lineUsage, other, intruder))
	if log.ThreadID != "root-1" || log.Cwd != "/example/repo" {
		t.Fatalf("identity = %q at %q, want root-1", log.ThreadID, log.Cwd)
	}
	if len(log.Records) != 1 || log.Records[0].ResponseID != "response-1" {
		t.Fatalf("records = %+v, want only response-1", log.Records)
	}
	if len(log.Diagnostics) != 0 {
		t.Errorf("inherited records must be silent; got %v", diagCodes(log))
	}
}

// The per-response usage is the only count. Cumulative thread and turn
// totals, and the event_msg counter stream, never become usage again.
func TestReadLogCumulativeCountersIgnored(t *testing.T) {
	cumulative := `{"timestamp":"2026-10-28T10:00:03Z","type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":900,"cached_input_tokens":0,"output_tokens":90}}}}`
	childUsage := usageLine(t, `"child-1"`, `"root-1"`, `"response-1"`, `{"input_tokens":100,"cached_input_tokens":60,"output_tokens":10}`)
	log := mustReadLog(t, jsonl(lineMetaChild, childUsage, cumulative))
	rec := wantOneRecord(t, log)
	if rec.Input != 100 || rec.CachedInput != 60 || rec.Output != 10 {
		t.Fatalf("usage = %+v, want the per-response values", rec)
	}
	if len(log.Diagnostics) != 0 {
		t.Errorf("diagnostics = %v, want none", diagCodes(log))
	}
}

// A missing, empty or non-string model resets the current model to unknown.
func TestReadLogModelResetsToUnknown(t *testing.T) {
	noModel := `{"type":"turn_context","payload":{}}`
	numberModel := `{"type":"turn_context","payload":{"model":3}}`
	emptyModel := `{"type":"turn_context","payload":{"model":""}}`
	log := mustReadLog(t, jsonl(lineMetaRoot,
		lineTurnModel, usageLine(t, `"root-1"`, `"root-1"`, `"r-a"`, `{"input_tokens":1,"cached_input_tokens":0,"output_tokens":1}`),
		noModel, usageLine(t, `"root-1"`, `"root-1"`, `"r-b"`, `{"input_tokens":1,"cached_input_tokens":0,"output_tokens":1}`),
		lineTurnModel, numberModel, usageLine(t, `"root-1"`, `"root-1"`, `"r-c"`, `{"input_tokens":1,"cached_input_tokens":0,"output_tokens":1}`),
		lineTurnModel, emptyModel, usageLine(t, `"root-1"`, `"root-1"`, `"r-d"`, `{"input_tokens":1,"cached_input_tokens":0,"output_tokens":1}`),
	))
	if len(log.Records) != 4 {
		t.Fatalf("records = %d, want 4", len(log.Records))
	}
	want := []string{"example-model", "", "", ""}
	for i, rec := range log.Records {
		if rec.Model != want[i] {
			t.Errorf("record %d model = %q, want %q", i, rec.Model, want[i])
		}
	}
}

// Timestamps parse as RFC 3339 with a timezone and come out in UTC. A
// response with no usable timestamp keeps its tokens and is marked undated.
func TestReadLogTimestamps(t *testing.T) {
	offset := `{"timestamp":"2026-10-28T10:00:02.125+07:00","type":"token_usage_record","payload":{"thread_id":"root-1","session_id":"root-1","response_id":"r-offset","usage":{"input_tokens":10,"cached_input_tokens":0,"output_tokens":2}}}`
	bad := `{"timestamp":"yesterday","type":"token_usage_record","payload":{"thread_id":"root-1","session_id":"root-1","response_id":"r-bad","usage":{"input_tokens":20,"cached_input_tokens":0,"output_tokens":2}}}`
	missing := `{"type":"token_usage_record","payload":{"thread_id":"root-1","session_id":"root-1","response_id":"r-missing","usage":{"input_tokens":30,"cached_input_tokens":0,"output_tokens":2}}}`
	log := mustReadLog(t, jsonl(lineMetaRoot, offset, bad, missing))
	if len(log.Records) != 3 {
		t.Fatalf("records = %d, want 3", len(log.Records))
	}
	if got, want := log.Records[0].At.UTC(), time.Date(2026, 10, 28, 3, 0, 2, 125000000, time.UTC); !got.Equal(want) {
		t.Errorf("offset time = %v, want %v", got, want)
	}
	for _, rec := range []Record{log.Records[1], log.Records[2]} {
		if !rec.At.IsZero() || !rec.Valid {
			t.Errorf("undated record = %+v, want zero time and valid usage", rec)
		}
	}
	if got := diagCodes(log); len(got) != 2 || got[0] != codeInvalidTimestamp || got[1] != codeInvalidTimestamp {
		t.Errorf("diagnostics = %v, want two invalid timestamps", got)
	}
	wantDiag(t, log, codeInvalidTimestamp, 3)
}

// Counters must be present, nonnegative, whole and inside int64, and cached
// input cannot beat input. A payload with no usage object is unsupported.
func TestReadLogRejectsBadCounters(t *testing.T) {
	for _, tc := range []struct {
		name  string
		usage string
		code  string
	}{
		{"missing input", `{"cached_input_tokens":0,"output_tokens":1}`, codeInvalidUsage},
		{"negative cached", `{"input_tokens":10,"cached_input_tokens":-1,"output_tokens":1}`, codeInvalidUsage},
		{"fractional output", `{"input_tokens":10,"cached_input_tokens":0,"output_tokens":1.5}`, codeInvalidUsage},
		{"past int64", `{"input_tokens":100000000000000000000,"cached_input_tokens":0,"output_tokens":1}`, codeInvalidUsage},
		{"string input", `{"input_tokens":"10","cached_input_tokens":0,"output_tokens":1}`, codeInvalidUsage},
		{"cached above input", `{"input_tokens":10,"cached_input_tokens":11,"output_tokens":1}`, codeInvalidUsage},
		{"null output", `{"input_tokens":10,"cached_input_tokens":0,"output_tokens":null}`, codeInvalidUsage},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stream := jsonl(lineMetaRoot, usageLine(t, `"root-1"`, `"root-1"`, `"r-1"`, tc.usage))
			log := mustReadLog(t, stream)
			rec := wantOneRecord(t, log)
			if rec.Valid {
				t.Fatalf("record counted as valid: %+v", rec)
			}
			wantDiag(t, log, tc.code, 2)
		})
	}

	t.Run("no usage object", func(t *testing.T) {
		stream := jsonl(lineMetaRoot, usageLine(t, `"root-1"`, `"root-1"`, `"r-1"`, ""))
		log := mustReadLog(t, stream)
		if rec := wantOneRecord(t, log); rec.Valid {
			t.Fatalf("record counted as valid: %+v", rec)
		}
		wantDiag(t, log, codeUnsupportedUsage, 2)
	})
}

// Optional fields add nothing: reasoning tokens are already inside output,
// and cache writes and cumulative totals are not part of these counters.
func TestReadLogOptionalFieldsAddNoTokens(t *testing.T) {
	usage := `{"input_tokens":100,"cached_input_tokens":60,"cache_write_input_tokens":50,"output_tokens":10,"reasoning_output_tokens":4}`
	log := mustReadLog(t, jsonl(lineMetaRoot, usageLine(t, `"root-1"`, `"root-1"`, `"r-1"`, usage)))
	rec := wantOneRecord(t, log)
	if rec.Input != 100 || rec.CachedInput != 60 || rec.Output != 10 {
		t.Fatalf("usage = %+v, want 100/60/10", rec)
	}
}

// Event types this reader does not know are skipped without a diagnostic.
func TestReadLogUnknownEventsIgnored(t *testing.T) {
	stream := jsonl(lineMetaRoot,
		`{"type":"world_state","payload":{"full":"anything"}}`,
		`{"type":"response_item","payload":{"type":"message","content":[{"text":"hello"}]}}`,
		`{"type":"something_new","payload":{"x":1}}`,
	)
	log := mustReadLog(t, stream)
	if len(log.Records) != 0 || len(log.Diagnostics) != 0 {
		t.Fatalf("records = %d, diagnostics = %v; want none", len(log.Records), diagCodes(log))
	}
}

// One broken line does not lose the rest of the file.
func TestReadLogMalformedLineRecovers(t *testing.T) {
	broken := `{"type":"token_usage_record", oops`
	log := mustReadLog(t, jsonl(lineMetaRoot, lineUsage, broken, lineUsage))
	if len(log.Records) != 2 {
		t.Fatalf("records = %d, want 2", len(log.Records))
	}
	wantDiag(t, log, codeInvalidJSON, 3)
}

// A final line with no newline is a write in progress, so it is not read and
// the file reports it.
func TestReadLogPendingLine(t *testing.T) {
	stream := lineMetaRoot + "\n" + lineUsage
	log := mustReadLog(t, stream)
	if len(log.Records) != 0 {
		t.Fatalf("records = %+v, want none from the pending line", log.Records)
	}
	if log.PendingLines != 1 {
		t.Fatalf("pending lines = %d, want 1", log.PendingLines)
	}
	wantDiag(t, log, codePendingLine, 2)
}

// A file with no usable session_meta has no owner: every record in it is
// unattributed, and the report says the metadata is missing.
func TestReadLogMissingMetadata(t *testing.T) {
	log := mustReadLog(t, jsonl(lineUsage))
	if log.ThreadID != "" || len(log.Records) != 0 {
		t.Fatalf("identity = %q, records = %d; want no identity and no records", log.ThreadID, len(log.Records))
	}
	wantDiag(t, log, codeMissingMetadata, 0)
}

// A usage record needs its thread, root and response identities as nonempty
// strings, all read from nested payload fields.
func TestReadLogMissingIdentity(t *testing.T) {
	for _, tc := range []struct {
		name string
		line string
	}{
		{"no root", usageLine(t, `"root-1"`, "", `"r-1"`, `{"input_tokens":1,"cached_input_tokens":0,"output_tokens":1}`)},
		{"no response", usageLine(t, `"root-1"`, `"root-1"`, "", `{"input_tokens":1,"cached_input_tokens":0,"output_tokens":1}`)},
		{"no thread", usageLine(t, "", `"root-1"`, `"r-1"`, `{"input_tokens":1,"cached_input_tokens":0,"output_tokens":1}`)},
		{"number response", usageLine(t, `"root-1"`, `"root-1"`, `7`, `{"input_tokens":1,"cached_input_tokens":0,"output_tokens":1}`)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			log := mustReadLog(t, jsonl(lineMetaRoot, tc.line))
			rec := wantOneRecord(t, log)
			if rec.Valid {
				t.Fatalf("record counted as valid: %+v", rec)
			}
			wantDiag(t, log, codeMissingIdentity, 2)
		})
	}
}

// An event-level session_id is not a substitute for the nested root id, and
// a turn_context session_id says nothing about root membership.
func TestReadLogOnlyNestedFields(t *testing.T) {
	line := `{"timestamp":"2026-10-28T10:00:02Z","session_id":"root-1","type":"token_usage_record","payload":{"thread_id":"root-1","response_id":"r-1","usage":{"input_tokens":1,"cached_input_tokens":0,"output_tokens":1}}}`
	context := `{"type":"turn_context","session_id":"root-1","payload":{"model":"example-model","session_id":"root-1"}}`
	log := mustReadLog(t, jsonl(lineMetaRoot, context, line))
	rec := wantOneRecord(t, log)
	if rec.Valid || rec.RootID != "" {
		t.Fatalf("record = %+v, want an invalid record with no root id", rec)
	}
	wantDiag(t, log, codeMissingIdentity, 3)
}

// The reader keeps no conversation text: an unknown event holding a secret
// leaves nothing behind, and a failed read does not put the line in its error.
func TestReadLogKeepsNoConversationText(t *testing.T) {
	secret := "SECRET-PROMPT-TEXT"
	stream := jsonl(lineMetaRoot,
		`{"type":"response_item","payload":{"type":"message","content":[{"text":"`+secret+`"}]}}`,
		lineUsage,
	)
	log := mustReadLog(t, stream)
	all := strings.Join(diagCodes(log), " ")
	for _, rec := range log.Records {
		all += rec.ResponseID + rec.RootID + rec.Model
	}
	if strings.Contains(all, secret) {
		t.Fatalf("kept conversation text: %q", all)
	}

	failing := io.MultiReader(strings.NewReader(lineMetaRoot+"\n"+secret+"\n"), errReader{})
	_, err := ReadLog("log.jsonl", failing)
	if err == nil {
		t.Fatal("ReadLog: got no error from a failing reader")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("error quotes conversation text: %v", err)
	}
}

// errReader fails the way a disk does, after some bytes.
type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("disk gone") }
