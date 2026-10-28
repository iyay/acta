package usage

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runGit runs one git command in dir and fails the test when it does.
func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// newRepo makes a real repository with one commit, so worktrees and
// rev-parse answer the way they do for a user.
func newRepo(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "config", "user.email", "usage@example.com")
	runGit(t, dir, "config", "user.name", "Usage Test")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "README.md")
	runGit(t, dir, "commit", "-q", "-m", "init")
	return dir
}

// canonicalPath resolves the temp folder symlink macOS puts in front of it.
func canonicalPath(t *testing.T, p string) string {
	t.Helper()
	real, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatalf("resolve %s: %v", p, err)
	}
	return real
}

// writeLog writes one transcript below the sessions folder.
func writeLog(t *testing.T, sessions, rel string, lines ...string) {
	t.Helper()
	path := filepath.Join(sessions, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	body := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func metaLine(id, cwd string) string {
	return fmt.Sprintf(`{"timestamp":"2026-10-28T10:00:00Z","type":"session_meta","payload":{"id":%q,"cwd":%q}}`, id, cwd)
}

func modelLine(model string) string {
	return fmt.Sprintf(`{"type":"turn_context","payload":{"model":%q}}`, model)
}

func recordLine(at, thread, root, response string, input, cached, output int64) string {
	return fmt.Sprintf(`{"timestamp":%q,"type":"token_usage_record","payload":{"thread_id":%q,"session_id":%q,"response_id":%q,"usage":{"input_tokens":%d,"cached_input_tokens":%d,"output_tokens":%d}}}`, at, thread, root, response, input, cached, output)
}

func mustRead(t *testing.T, cwd, sessions string) Report {
	t.Helper()
	r, err := Read(cwd, sessions)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	return r
}

// sessionIDs lists the report's session rows as "root/model", so an order
// or membership mistake reads plainly.
func sessionIDs(r Report) []string {
	var out []string
	for _, s := range r.Sessions {
		out = append(out, s.SessionID+"/"+s.Model)
	}
	return out
}

// The current checkout and its registered worktrees share one scope, while
// an unrelated repository and a nested one stay out of it.
func TestUsageReadScope(t *testing.T) {
	repo := newRepo(t)
	wt := filepath.Join(t.TempDir(), "wt")
	runGit(t, repo, "worktree", "add", "-q", wt, "-b", "feat")
	nested := filepath.Join(repo, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	runGit(t, nested, "init", "-q", "-b", "main")
	other := newRepo(t)

	sessions := t.TempDir()
	writeLog(t, sessions, "a/root.jsonl",
		metaLine("root-1", repo),
		modelLine("m1"),
		recordLine("2026-10-28T10:00:02Z", "root-1", "root-1", "r1", 100, 60, 10),
	)
	writeLog(t, sessions, "b/wt.jsonl",
		metaLine("root-2", filepath.Join(wt, "sub")),
		modelLine("m2"),
		recordLine("2026-10-28T10:00:03Z", "root-2", "root-2", "r2", 10, 0, 1),
	)
	writeLog(t, sessions, "c/other.jsonl",
		metaLine("root-3", other),
		modelLine("m3"),
		recordLine("2026-10-28T10:00:04Z", "root-3", "root-3", "r3", 10, 0, 1),
	)
	writeLog(t, sessions, "d/nested.jsonl",
		metaLine("root-4", nested),
		modelLine("m4"),
		recordLine("2026-10-28T10:00:05Z", "root-4", "root-4", "r4", 10, 0, 1),
	)

	r := mustRead(t, repo, sessions)
	if got, want := sessionIDs(r), []string{"root-1/m1", "root-2/m2"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("sessions = %v, want %v", got, want)
	}
	if got, want := r.Coverage.CountedResponses, 2; got != want {
		t.Errorf("counted responses = %d, want %d", got, want)
	}
	wantCheckouts := []string{canonicalPath(t, repo), canonicalPath(t, wt)}
	if strings.Join(r.Scope.Checkouts, ",") != strings.Join(wantCheckouts, ",") {
		t.Errorf("checkouts = %v, want %v", r.Scope.Checkouts, wantCheckouts)
	}
	if r.Scope.Repository != canonicalPath(t, repo) {
		t.Errorf("repository = %q", r.Scope.Repository)
	}
	if r.Status != StatusComplete {
		t.Errorf("status = %q, want complete", r.Status)
	}
}

// A child counts toward the root session that matched the repository, a
// child of an unknown root only joins when its own start does, and that
// group says it is incomplete.
func TestUsageReadChildRoots(t *testing.T) {
	repo := newRepo(t)
	elsewhere := filepath.Join(t.TempDir(), "elsewhere")

	sessions := t.TempDir()
	writeLog(t, sessions, "a/root.jsonl",
		metaLine("root-1", repo),
		modelLine("m1"),
		recordLine("2026-10-28T10:00:02Z", "root-1", "root-1", "r1", 100, 0, 10),
	)
	writeLog(t, sessions, "b/child.jsonl",
		metaLine("child-1", elsewhere),
		recordLine("2026-10-28T10:00:03Z", "child-1", "root-1", "r2", 50, 0, 5),
	)
	writeLog(t, sessions, "c/orphan-child.jsonl",
		metaLine("child-2", elsewhere),
		recordLine("2026-10-28T10:00:04Z", "child-2", "root-9", "r3", 7, 0, 1),
	)
	writeLog(t, sessions, "d/rootless.jsonl",
		metaLine("child-3", repo),
		recordLine("2026-10-28T10:00:05Z", "child-3", "root-8", "r4", 9, 0, 1),
	)

	r := mustRead(t, repo, sessions)
	want := []string{"root-1/m1", "root-1/" + UnknownModel, "root-8/" + UnknownModel}
	if got := sessionIDs(r); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("sessions = %v, want %v", got, want)
	}
	var root1Fresh, root1Count int64
	var root8 *SessionRow
	for i := range r.Sessions {
		row := &r.Sessions[i]
		switch {
		case row.SessionID == "root-1":
			root1Fresh += row.FreshInputTokens
			root1Count += int64(row.ResponseCount)
		case row.SessionID == "root-8":
			root8 = row
		}
	}
	if root1Count != 2 || root1Fresh != 150 {
		t.Errorf("root-1 totals = %d responses, %d fresh input; want the child too", root1Count, root1Fresh)
	}
	if root8.ResponseCount != 1 {
		t.Errorf("root-8 row = %+v", root8)
	}
	if !r.Coverage.Partial {
		t.Error("a group without its root log must be partial")
	}
	if r.Coverage.MatchedRootSessions != 2 {
		t.Errorf("matched root sessions = %d, want 2", r.Coverage.MatchedRootSessions)
	}
	if !hasCode(r, codeMissingMetadata) {
		t.Errorf("diagnostics = %v, want missing metadata for the rootless group", reportCodes(r))
	}
}

// Symlinks, compressed logs and other files are never opened, and the walk
// gives the same answer on every run.
func TestUsageReadScan(t *testing.T) {
	repo := newRepo(t)
	sessions := t.TempDir()
	writeLog(t, sessions, "real.jsonl",
		metaLine("root-1", repo),
		modelLine("m1"),
		recordLine("2026-10-28T10:00:02Z", "root-1", "root-1", "r1", 10, 0, 1),
	)
	writeLog(t, sessions, "hidden/copy.jsonl.zst",
		metaLine("root-2", repo),
		recordLine("2026-10-28T10:00:03Z", "root-2", "root-2", "r2", 10, 0, 1),
	)
	writeLog(t, sessions, "notes.txt",
		metaLine("root-3", repo),
		recordLine("2026-10-28T10:00:04Z", "root-3", "root-3", "r3", 10, 0, 1),
	)
	// A file symlink and a folder symlink both point at a real log.
	outside := t.TempDir()
	writeLog(t, outside, "linked.jsonl",
		metaLine("root-4", repo),
		recordLine("2026-10-28T10:00:05Z", "root-4", "root-4", "r4", 10, 0, 1),
	)
	if err := os.Symlink(filepath.Join(outside, "linked.jsonl"), filepath.Join(sessions, "linked.jsonl")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(sessions, "linkdir")); err != nil {
		t.Fatal(err)
	}

	r := mustRead(t, repo, sessions)
	if got, want := sessionIDs(r), []string{"root-1/m1"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("sessions = %v, want %v", got, want)
	}
	if r.Coverage.CountedResponses != 1 || r.Coverage.UnclassifiedLogs != 0 {
		t.Errorf("coverage = %+v", r.Coverage)
	}
}

// A missing sessions folder is no data, not an error. Neither the folder nor
// the transcripts change when a report is read.
func TestUsageReadMissingSourceAndNoWrites(t *testing.T) {
	repo := newRepo(t)
	sessions := filepath.Join(t.TempDir(), "absent")
	r, err := Read(repo, sessions)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if r.Status != StatusNoData || r.Coverage.Partial {
		t.Fatalf("status = %q partial = %v, want no data and not partial", r.Status, r.Coverage.Partial)
	}
	if r.Sessions == nil || r.Models == nil || r.Diagnostics == nil {
		t.Fatal("empty arrays must not be nil, so JSON shows []")
	}

	sessions = t.TempDir()
	writeLog(t, sessions, "a.jsonl",
		metaLine("root-1", repo),
		modelLine("m1"),
		recordLine("2026-10-28T10:00:02Z", "root-1", "root-1", "r1", 10, 0, 1),
	)
	before := snapshotTree(t, sessions)
	mustRead(t, repo, sessions)
	after := snapshotTree(t, sessions)
	if fmt.Sprint(before) != fmt.Sprint(after) {
		t.Fatalf("reading changed the source tree:\nbefore %v\nafter  %v", before, after)
	}
}

// Outside a repository there is nothing to scope a report to.
func TestUsageReadNotRepository(t *testing.T) {
	cwd := t.TempDir()
	_, err := Read(cwd, t.TempDir())
	if !errors.Is(err, ErrNotRepository) {
		t.Fatalf("err = %v, want ErrNotRepository", err)
	}

	// A repository git refuses to read is an operational failure, not a
	// missing repository, so the real cause reaches the user.
	repo := newRepo(t)
	cfg, err := os.OpenFile(filepath.Join(repo, ".git", "config"), os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fmt.Fprint(cfg, "this line has no equals sign\n"); err != nil {
		t.Fatal(err)
	}
	if err := cfg.Close(); err != nil {
		t.Fatal(err)
	}
	_, err = Read(repo, t.TempDir())
	if err == nil {
		t.Fatal("Read: got no error from a repository git cannot read")
	}
	if errors.Is(err, ErrNotRepository) {
		t.Fatalf("broken git reported as a missing repository: %v", err)
	}
	if !strings.Contains(err.Error(), "bad config") {
		t.Fatalf("err = %v, want git's own reason", err)
	}

	// A checkout that lost its HEAD reads exactly like "not a repository";
	// the .git entry is what tells the two apart.
	lost := newRepo(t)
	if err := os.Remove(filepath.Join(lost, ".git", "HEAD")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LC_ALL", "de_DE.UTF-8")
	_, err = Read(lost, t.TempDir())
	if err == nil {
		t.Fatal("Read: got no error from a checkout without HEAD")
	}
	if errors.Is(err, ErrNotRepository) {
		t.Fatalf("a broken checkout reported as a missing repository: %v", err)
	}
}

// A source folder that cannot be listed is a fatal failure, while a single
// unreadable log only makes the report partial.
func TestUsageReadUnreadableSourceAndFile(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root reads anything, so the permission check cannot fail")
	}
	repo := newRepo(t)

	t.Run("source folder", func(t *testing.T) {
		sessions := t.TempDir()
		if err := os.Chmod(sessions, 0o000); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(sessions, 0o755) })
		if _, err := Read(repo, sessions); err == nil {
			t.Fatal("Read: got no error from an unreadable source folder")
		}
	})

	t.Run("one log", func(t *testing.T) {
		sessions := t.TempDir()
		writeLog(t, sessions, "locked.jsonl",
			metaLine("root-1", repo),
			recordLine("2026-10-28T10:00:02Z", "root-1", "root-1", "r1", 10, 0, 1),
		)
		locked := filepath.Join(sessions, "locked.jsonl")
		if err := os.Chmod(locked, 0o000); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(locked, 0o644) })
		r := mustRead(t, repo, sessions)
		if !r.Coverage.Partial || r.Coverage.UnclassifiedLogs != 1 {
			t.Fatalf("coverage = %+v, want a partial report naming one unclassified log", r.Coverage)
		}
		if !hasCode(r, codeUnreadableLog) {
			t.Fatalf("diagnostics = %v, want an unreadable log", reportCodes(r))
		}
	})
}

// A log with usage but no metadata cannot be attributed, so its tokens stay
// out of the report and the coverage says why.
func TestUsageReadUnclassifiedMetadata(t *testing.T) {
	repo := newRepo(t)

	t.Run("complete record", func(t *testing.T) {
		sessions := t.TempDir()
		writeLog(t, sessions, "no-meta.jsonl",
			modelLine("m1"),
			recordLine("2026-10-28T10:00:02Z", "root-1", "root-1", "r1", 10, 0, 1),
		)
		r := mustRead(t, repo, sessions)
		if len(r.Sessions) != 0 || r.Coverage.CountedResponses != 0 {
			t.Fatalf("sessions = %v, want none", sessionIDs(r))
		}
		if r.Coverage.UnclassifiedLogs != 1 || !r.Coverage.Partial || r.Status != StatusPartial {
			t.Fatalf("coverage = %+v status = %q", r.Coverage, r.Status)
		}
		if r.Coverage.PendingLines != 0 {
			t.Fatalf("pending lines = %d, want none", r.Coverage.PendingLines)
		}
	})

	t.Run("pending line", func(t *testing.T) {
		sessions := t.TempDir()
		writeRawLog(t, sessions, "no-meta.jsonl",
			modelLine("m1")+"\n"+
				recordLine("2026-10-28T10:00:02Z", "root-1", "root-1", "r1", 10, 0, 1))
		r := mustRead(t, repo, sessions)
		if r.Coverage.UnclassifiedLogs != 1 || r.Coverage.PendingLines != 1 {
			t.Fatalf("coverage = %+v, want one unclassified log with one pending line", r.Coverage)
		}
	})
}

// A log that names a thread but no starting folder cannot be attributed, so
// it is unclassified rather than silently unrelated.
func TestUsageReadMissingStartingFolder(t *testing.T) {
	repo := newRepo(t)
	sessions := t.TempDir()
	writeLog(t, sessions, "no-cwd.jsonl",
		`{"type":"session_meta","payload":{"id":"root-1"}}`,
		recordLine("2026-10-28T10:00:02Z", "root-1", "root-1", "r1", 10, 0, 1),
	)
	r := mustRead(t, repo, sessions)
	if len(r.Sessions) != 0 || r.Coverage.CountedResponses != 0 {
		t.Fatalf("sessions = %v, want none", sessionIDs(r))
	}
	if r.Coverage.UnclassifiedLogs != 1 || !r.Coverage.Partial || r.Status != StatusPartial {
		t.Fatalf("coverage = %+v status = %q", r.Coverage, r.Status)
	}
	if !hasCode(r, codeMissingMetadata) {
		t.Fatalf("diagnostics = %v, want the missing start named", reportCodes(r))
	}

	pending := t.TempDir()
	writeRawLog(t, pending, "no-cwd.jsonl",
		`{"type":"session_meta","payload":{"id":"root-1"}}`+"\n"+
			recordLine("2026-10-28T10:00:02Z", "root-1", "root-1", "r1", 10, 0, 1))
	r = mustRead(t, repo, pending)
	if r.Coverage.PendingLines != 1 {
		t.Fatalf("pending lines = %d, want the skipped final line counted", r.Coverage.PendingLines)
	}
}

// Paths from a transcript are arguments, never shell text.
func TestUsageReadPathsAreLiteral(t *testing.T) {
	base := t.TempDir()
	repo := filepath.Join(base, "repo with space $(touch pwned)")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	runGit(t, repo, "init", "-q", "-b", "main")
	runGit(t, repo, "config", "user.email", "usage@example.com")
	runGit(t, repo, "config", "user.name", "Usage Test")

	sessions := t.TempDir()
	writeLog(t, sessions, "a.jsonl",
		metaLine("root-1", filepath.Join(repo, "sub`touch pwned`")),
		recordLine("2026-10-28T10:00:02Z", "root-1", "root-1", "r1", 10, 0, 1),
	)
	if err := os.MkdirAll(filepath.Join(repo, "sub`touch pwned`"), 0o755); err != nil {
		t.Fatal(err)
	}
	r := mustRead(t, repo, sessions)
	if got := sessionIDs(r); len(got) != 1 {
		t.Fatalf("sessions = %v, want the space and shell path to match", got)
	}
	for _, dir := range []string{repo, base} {
		if _, err := os.Stat(filepath.Join(dir, "pwned")); err == nil {
			t.Fatalf("a transcript path was run as shell text in %s", dir)
		}
	}
}

// snapshotTree records every file below dir with its bytes, so a test can
// prove a read changed nothing.
func snapshotTree(t *testing.T, dir string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		out[rel] = string(body)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}
