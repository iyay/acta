package usage

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/iyay/acta/internal/gitc"
)

// ErrNotRepository says the working directory is not inside a Git
// repository, so there is no scope to report on.
var ErrNotRepository = errors.New("not inside a Git repository")

// errOverflow refuses a total that would wrap, because a wrapped total looks
// like a real one.
var errOverflow = errors.New("usage totals overflow int64")

// fileLog is one scanned transcript and what the report decided about it.
type fileLog struct {
	rel     string
	log     Log
	matched bool // its own start is inside this repository
	include bool
}

// Read builds a usage report for the repository holding cwd, reading the
// transcripts below sourceDirectory. A missing folder is no data, not a
// failure; only a repository, source or arithmetic failure is an error.
func Read(cwd, sourceDirectory string) (Report, error) {
	repo, err := repoRoot(cwd)
	if err != nil {
		return Report{}, err
	}
	checkouts, err := checkoutsOf(repo)
	if err != nil {
		return Report{}, fmt.Errorf("list worktrees: %w", err)
	}
	report := Report{
		SchemaVersion: schemaVersion,
		Scope: Scope{
			Repository:      repo,
			Checkouts:       checkouts,
			SourceDirectory: sourceDirectory,
			HistoryCoverage: historyCoverage,
			Limitations:     []string{HistoryLimitation},
		},
		Sessions:    []SessionRow{},
		Models:      []ModelRow{},
		Diagnostics: []Diagnostic{},
	}

	paths, broken, err := scanLogs(sourceDirectory)
	if err != nil {
		return Report{}, err
	}
	var (
		partial      bool
		unclassified int
		pending      int
		invalid      int
		undated      int
		conflicts    int
	)
	// countUnclassified files one log that cannot be attributed, so the two
	// reasons keep the same bookkeeping.
	countUnclassified := func(fl fileLog) {
		report.Diagnostics = append(report.Diagnostics, fl.log.Diagnostics...)
		pending += fl.log.PendingLines
		unclassified++
		partial = true
	}
	for _, rel := range broken {
		report.Diagnostics = append(report.Diagnostics, Diagnostic{
			Code:    codeUnreadableLog,
			File:    new(rel),
			Message: "This entry could not be read and was skipped.",
		})
		unclassified++
		partial = true
	}
	var logs []fileLog
	for _, path := range paths {
		rel, err := filepath.Rel(sourceDirectory, path)
		if err != nil {
			return Report{}, err
		}
		rel = filepath.ToSlash(rel)
		log, err := readOne(path, rel)
		if err != nil {
			report.Diagnostics = append(report.Diagnostics, Diagnostic{
				Code:    codeUnreadableLog,
				File:    new(rel),
				Message: "This log could not be read and was skipped.",
			})
			unclassified++
			partial = true
			continue
		}
		if log.ThreadID == "" {
			// Without metadata this file cannot be attributed to a repo, so
			// it is not called unrelated either.
			countUnclassified(fileLog{log: log})
			continue
		}
		logs = append(logs, fileLog{rel: rel, log: log, matched: insideCheckouts(log.Cwd, checkouts)})
	}

	// A matched log establishes the root sessions it names, so a child that
	// started elsewhere joins its root. A root with no log of its own leaves
	// its group partial.
	established := map[string]bool{}
	rootKnown := map[string]bool{}
	for _, fl := range logs {
		if !fl.matched {
			continue
		}
		rootKnown[fl.log.ThreadID] = true
		named := false
		for _, rec := range fl.log.Records {
			if rec.RootID == "" {
				continue
			}
			established[rec.RootID] = true
			named = true
		}
		if !named {
			// A root log with no usage still names its own root session.
			established[fl.log.ThreadID] = true
		}
	}
	for i := range logs {
		if logs[i].matched {
			logs[i].include = true
			continue
		}
		for _, rec := range logs[i].log.Records {
			if rec.RootID != "" && established[rec.RootID] {
				logs[i].include = true
				break
			}
		}
	}

	var (
		copies        = map[string]*copyState{}
		validPerLog   = map[string]int{}
		threadsSeen   = map[string]bool{}
		matchedThread = map[string]bool{}
		missingRootAt = map[string]bool{}
	)
	var included []fileLog
	for _, fl := range logs {
		if fl.include {
			included = append(included, fl)
			continue
		}
		if fl.log.Cwd != "" {
			// A readable log whose start is inside another repository is
			// classified, so it stays out of this repository's coverage.
			continue
		}
		// A log that names no starting folder cannot be called unrelated.
		countUnclassified(fl)
		report.Diagnostics = append(report.Diagnostics, Diagnostic{
			Code:    codeMissingMetadata,
			File:    new(fl.rel),
			Message: "This log names no starting folder, so it cannot be attributed to a repository.",
		})
	}
	for _, fl := range included {
		report.Diagnostics = append(report.Diagnostics, fl.log.Diagnostics...)
		pending += fl.log.PendingLines
		if len(fl.log.Diagnostics) > 0 || fl.log.PendingLines > 0 {
			partial = true
		}
		threadsSeen[fl.log.ThreadID] = true
		if fl.matched {
			matchedThread[fl.log.ThreadID] = true
		}
		for _, rec := range fl.log.Records {
			if !rec.Valid {
				invalid++
				continue
			}
			validPerLog[fl.log.ThreadID]++
			keepBestCopy(copies, rec)
			if !rootKnown[rec.RootID] && !missingRootAt[fl.rel] {
				// The root log of this group is gone, so the group may be
				// missing usage that this run cannot see.
				missingRootAt[fl.rel] = true
				partial = true
				report.Diagnostics = append(report.Diagnostics, Diagnostic{
					Code:    codeMissingMetadata,
					File:    new(fl.rel),
					Message: "The root log of this group was not found, so its group may be incomplete.",
				})
			}
		}
	}

	var responses []Record
	for _, st := range copies {
		if st.conflict {
			conflicts++
			continue
		}
		responses = append(responses, st.rec)
	}
	for range conflicts {
		// A conflicting copy names no file that would be fair to blame.
		report.Diagnostics = append(report.Diagnostics, Diagnostic{
			Code:    codeConflictingResponse,
			Message: "Copies of one response disagree, so none of them was counted.",
		})
	}
	for _, rec := range responses {
		if rec.At.IsZero() {
			undated++
			partial = true
		}
	}
	if conflicts > 0 || invalid > 0 {
		partial = true
	}

	rows, models, span, err := aggregate(responses)
	if err != nil {
		return Report{}, err
	}
	withoutUsage := 0
	for thread := range matchedThread {
		if validPerLog[thread] == 0 {
			withoutUsage++
		}
	}
	report.Range = span.span()
	report.Sessions = rows
	report.Models = models
	report.Coverage = Coverage{
		Partial:                   partial,
		MatchedRootSessions:       len(established),
		IncludedThreadLogs:        len(threadsSeen),
		CountedResponses:          len(responses),
		ExcludedInvalidRecords:    invalid,
		ConflictingResponses:      conflicts,
		LogsWithoutSupportedUsage: withoutUsage,
		UnclassifiedLogs:          unclassified,
		PendingLines:              pending,
		UndatedResponses:          undated,
	}
	switch {
	case partial:
		report.Status = StatusPartial
	case len(responses) > 0:
		report.Status = StatusComplete
	default:
		report.Status = StatusNoData
	}
	sortDiagnostics(report.Diagnostics)
	return report, nil
}

// copyState holds one response identity while every copy of it is read.
// One disagreeing copy takes all of them out, whatever the scan order.
type copyState struct {
	rec      Record
	conflict bool
}

// keepBestCopy adds one copy of a response. Identical copies are harmless;
// copies that disagree are marked so no copy is counted at all.
func keepBestCopy(copies map[string]*copyState, rec Record) {
	st, ok := copies[rec.ResponseID]
	if !ok {
		copies[rec.ResponseID] = &copyState{rec: rec}
		return
	}
	if !sameUsage(st.rec, rec) {
		st.conflict = true
		return
	}
	// Identical copies can still be written at different times; the earliest
	// dated one keeps the range stable.
	if st.rec.At.IsZero() || (!rec.At.IsZero() && rec.At.Before(st.rec.At)) {
		st.rec.At = rec.At
	}
}

// sameUsage says whether two copies claim the same owner, model and usage.
// Only those decide a count; a different timestamp alone is not a conflict.
func sameUsage(a, b Record) bool {
	return a.RootID == b.RootID && a.Model == b.Model &&
		a.Input == b.Input && a.CachedInput == b.CachedInput && a.Output == b.Output
}

// rowKey is one root session and model pair.
type rowKey struct {
	session string
	model   string
}

// tokenTotal is a checked sum of the three counters and their responses.
type tokenTotal struct {
	fresh, cached, output int64
	count                 int
}

func (t *tokenTotal) add(fresh, cached, output int64, count int) error {
	var err error
	if t.fresh, err = addChecked(t.fresh, fresh); err != nil {
		return err
	}
	if t.cached, err = addChecked(t.cached, cached); err != nil {
		return err
	}
	if t.output, err = addChecked(t.output, output); err != nil {
		return err
	}
	if t.count > math.MaxInt-count {
		return errOverflow
	}
	t.count += count
	return nil
}

// addChecked adds two counts and refuses a result that would wrap.
func addChecked(a, b int64) (int64, error) {
	if b > 0 && a > math.MaxInt64-b {
		return 0, errOverflow
	}
	return a + b, nil
}

// aggregate turns the surviving responses into session rows, model rows and
// the report range. Fresh input is input minus cached, counted per response.
func aggregate(responses []Record) ([]SessionRow, []ModelRow, spanBuilder, error) {
	type rowState struct {
		key  rowKey
		tot  tokenTotal
		span spanBuilder
	}
	rows := map[rowKey]*rowState{}
	models := map[string]*tokenTotal{}
	var total spanBuilder
	for _, rec := range responses {
		model := rec.Model
		if model == "" {
			model = UnknownModel
		}
		fresh := rec.Input - rec.CachedInput
		key := rowKey{session: rec.RootID, model: model}
		st, ok := rows[key]
		if !ok {
			st = &rowState{key: key}
			rows[key] = st
		}
		if err := st.tot.add(fresh, rec.CachedInput, rec.Output, 1); err != nil {
			return nil, nil, spanBuilder{}, err
		}
		st.span.add(rec.At)
		if _, ok := models[model]; !ok {
			models[model] = &tokenTotal{}
		}
		if err := models[model].add(fresh, rec.CachedInput, rec.Output, 1); err != nil {
			return nil, nil, spanBuilder{}, err
		}
		total.add(rec.At)
	}

	ordered := make([]*rowState, 0, len(rows))
	for _, st := range rows {
		ordered = append(ordered, st)
	}
	sort.Slice(ordered, func(i, j int) bool {
		a, b := ordered[i], ordered[j]
		aDated, bDated := !a.span.first.IsZero(), !b.span.first.IsZero()
		if aDated != bDated {
			return aDated
		}
		if aDated && !a.span.first.Equal(b.span.first) {
			return a.span.first.Before(b.span.first)
		}
		if a.key.session != b.key.session {
			return a.key.session < b.key.session
		}
		return a.key.model < b.key.model
	})
	sessionRows := make([]SessionRow, 0, len(ordered))
	for _, st := range ordered {
		sessionRows = append(sessionRows, SessionRow{
			SessionID:         st.key.session,
			Model:             st.key.model,
			Range:             st.span.span(),
			FreshInputTokens:  st.tot.fresh,
			CachedInputTokens: st.tot.cached,
			OutputTokens:      st.tot.output,
			ResponseCount:     st.tot.count,
		})
	}

	modelRows := make([]ModelRow, 0, len(models))
	for name, tot := range models {
		modelRows = append(modelRows, ModelRow{
			Model:             name,
			FreshInputTokens:  tot.fresh,
			CachedInputTokens: tot.cached,
			OutputTokens:      tot.output,
			ResponseCount:     tot.count,
		})
	}
	sort.Slice(modelRows, func(i, j int) bool { return modelRows[i].Model < modelRows[j].Model })
	return sessionRows, modelRows, total, nil
}

// repoRoot finds the repository holding dir. Outside a repository it says
// so; a repository git cannot read is an operational failure.
func repoRoot(dir string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel")
	// A failure must read the same for every user, so git answers in the C
	// locale whatever the user's language is.
	cmd.Env = append(os.Environ(), "LC_ALL=C", "LANGUAGE=")
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		// Git uses one failure for "no repository" and for a repository it
		// cannot read, and its text is translated. A .git entry above dir is
		// the answer that does not depend on the message.
		if !hasGitEntry(dir) {
			return "", ErrNotRepository
		}
		msg := strings.TrimSpace(errOut.String())
		if msg == "" {
			return "", fmt.Errorf("find the repository: %w", err)
		}
		return "", fmt.Errorf("find the repository: %s", msg)
	}
	root := strings.TrimSpace(out.String())
	if root == "" {
		return "", ErrNotRepository
	}
	return canonical(root), nil
}

// hasGitEntry says whether dir or a parent holds a .git entry, so a
// repository exists even when git refuses to read it.
func hasGitEntry(dir string) bool {
	for p := canonical(dir); ; {
		if _, err := os.Stat(filepath.Join(p, ".git")); err == nil {
			return true
		}
		parent := filepath.Dir(p)
		if parent == p {
			return false
		}
		p = parent
	}
}

// checkoutsOf lists the repository's main checkout and its registered,
// existing worktrees, sorted, so the scope is stable across runs.
func checkoutsOf(repo string) ([]string, error) {
	wts, err := gitc.Worktrees(repo)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var out []string
	add := func(p string) {
		c := canonical(p)
		if !seen[c] {
			seen[c] = true
			out = append(out, c)
		}
	}
	add(repo)
	for _, w := range wts {
		add(w.Path)
	}
	sort.Strings(out)
	return out, nil
}

// insideCheckouts says whether dir is inside one of the checkouts. A nested
// repository belongs to itself, so a folder with its own .git inside a
// checkout does not count.
func insideCheckouts(dir string, checkouts []string) bool {
	if dir == "" {
		return false
	}
	dir = canonical(dir)
	for _, root := range checkouts {
		if !within(dir, root) {
			continue
		}
		if hasNestedRepo(dir, root) {
			continue
		}
		return true
	}
	return false
}

// within says whether dir is root or sits below it.
func within(dir, root string) bool {
	rel, err := filepath.Rel(root, dir)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// hasNestedRepo says whether a folder between dir and root holds a .git
// entry, which makes it another repository.
func hasNestedRepo(dir, root string) bool {
	for p := dir; ; p = filepath.Dir(p) {
		if p == root || p == "." || p == string(filepath.Separator) {
			return false
		}
		if _, err := os.Stat(filepath.Join(p, ".git")); err == nil {
			return true
		}
		if filepath.Dir(p) == p {
			return false
		}
	}
}

// canonical resolves symlinks so two paths to one folder compare equal. A
// recorded folder that no longer exists still gets its existing parents
// resolved, so it compares equal to the checkout that holds it.
func canonical(p string) string {
	abs, err := filepath.Abs(p)
	if err != nil {
		abs = filepath.Clean(p)
	}
	rest := ""
	for cur := abs; ; cur = filepath.Dir(cur) {
		if real, err := filepath.EvalSymlinks(cur); err == nil {
			if rest == "" {
				return real
			}
			return filepath.Join(real, rest)
		}
		base := filepath.Base(cur)
		if rest == "" {
			rest = base
		} else {
			rest = filepath.Join(base, rest)
		}
		if filepath.Dir(cur) == cur {
			return abs
		}
	}
}

// scanLogs lists the regular .jsonl files below dir, never following a
// symlink and skipping archived or compressed names. A missing folder is an
// empty list. broken names entries the walk could not read.
func scanLogs(dir string) (files, broken []string, err error) {
	if dir == "" {
		return nil, nil, nil
	}
	if _, err := os.Stat(dir); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
		rel, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			return relErr
		}
		rel = filepath.ToSlash(rel)
		if walkErr != nil {
			if path == dir {
				return walkErr
			}
			broken = append(broken, rel)
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.Type()&fs.ModeSymlink != 0 {
			// Never follow a link out of the tree, and never read a link
			// twice under two names.
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() || !strings.HasSuffix(path, ".jsonl") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return files, broken, nil
}

// readOne reads one file up to the size it had when opened, so a session
// that keeps writing cannot make one read unbounded.
func readOne(path, rel string) (Log, error) {
	f, err := os.Open(path)
	if err != nil {
		return Log{}, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return Log{}, err
	}
	return ReadLog(rel, io.LimitReader(f, st.Size()))
}
