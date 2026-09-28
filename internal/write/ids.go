package write

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/gitc"
)

// randHash is a var so a test can force a clash.
var randHash = func() string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	const chars = letters + "0123456789"
	var b [4]byte
	rand.Read(b[:])
	out := []byte{letters[int(b[0])%len(letters)]}
	for _, c := range b[1:] {
		out = append(out, chars[int(c)%len(chars)])
	}
	return string(out)
}

// AssignIDs gives every non-legacy spec, plan and bug in the root folder that
// lacks an id or hash a number past the highest of its prefix and a hash no
// other item has. A value already written in a file is never changed, so a
// run writes a field only when the file has none, and a file it cannot read
// is skipped with its reason instead of ending the run. One commit holds the
// run.
func AssignIDs(cfg config.Config, b *board.Board, only []string) ([]string, Outcome, error) {
	next, taken := scanIDs(b)
	cands, skips := candidates(cfg, b, only)
	var changes, paths []string
	for _, c := range cands {
		orig, err := os.ReadFile(c.file)
		if err != nil {
			return nil, Outcome{}, err
		}
		src := orig
		id, hash := c.id, c.hash
		if id == "" {
			id = fmt.Sprintf("%s-%d", c.prefix, next[c.prefix])
			next[c.prefix]++
			out, err := SetField(src, "id", id)
			if err != nil {
				return nil, Outcome{Skips: skips}, err
			}
			src = out
		}
		if hash == "" {
			hash = freeHash(taken)
			out, err := SetField(src, "hash", hash)
			if err != nil {
				return nil, Outcome{Skips: skips}, err
			}
			src = out
		}
		if string(src) == string(orig) {
			continue
		}
		if err := os.WriteFile(c.file, src, 0o644); err != nil {
			return nil, Outcome{Skips: skips}, err
		}
		taken[hash] = true
		rel, _ := filepath.Rel(cfg.Root, c.file)
		changes = append(changes, id+" ("+filepath.ToSlash(rel)+")")
		paths = append(paths, c.file)
	}
	if len(paths) == 0 {
		return nil, Outcome{Skips: skips}, nil
	}
	sort.Strings(paths)
	if !cfg.AutoCommit {
		return changes, Outcome{Reason: "auto_commit is off", Skips: skips}, nil
	}
	r := gitc.CommitPaths(cfg.RepoRoot, paths, "acta: assign short ids")
	return changes, Outcome{Committed: r.Committed, Skipped: !r.Committed, Reason: r.Reason, Skips: skips}, nil
}

// FixDuplicates gives each later holder of a number the next free number.
// The file that reached the branch first keeps its number, so the order
// follows the first-parent history and not the clock. Hashes never move.
func FixDuplicates(cfg config.Config, b *board.Board) ([]string, Outcome, error) {
	next, _ := scanIDs(b)
	groups := map[string][]idFile{}
	for _, f := range idFiles(b) {
		if f.id == "" || badReason(f) != "" {
			continue
		}
		groups[f.id] = append(groups[f.id], f)
	}
	var keys []string
	for k, g := range groups {
		if len(g) > 1 {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		return nil, Outcome{}, nil
	}
	sort.Strings(keys)
	var changes, paths []string
	for _, k := range keys {
		g := groups[k]
		sort.SliceStable(g, func(i, j int) bool {
			ai, _ := gitc.FirstSeen(cfg.RepoRoot, g[i].file)
			aj, _ := gitc.FirstSeen(cfg.RepoRoot, g[j].file)
			return ai < aj
		})
		prefix := strings.TrimSuffix(k, "-"+numPart(k))
		for _, f := range g[1:] {
			id := fmt.Sprintf("%s-%d", prefix, next[prefix])
			next[prefix]++
			src, err := os.ReadFile(f.file)
			if err != nil {
				return nil, Outcome{}, err
			}
			out, err := SetField(src, "id", id)
			if err != nil {
				return nil, Outcome{}, err
			}
			if err := os.WriteFile(f.file, out, 0o644); err != nil {
				return nil, Outcome{}, err
			}
			rel, _ := filepath.Rel(cfg.Root, f.file)
			changes = append(changes, k+" -> "+id+" ("+filepath.ToSlash(rel)+")")
			paths = append(paths, f.file)
		}
	}
	sort.Strings(paths)
	if !cfg.AutoCommit {
		return changes, Outcome{Reason: "auto_commit is off"}, nil
	}
	r := gitc.CommitPaths(cfg.RepoRoot, paths, "acta: fix duplicate short ids")
	return changes, Outcome{Committed: r.Committed, Skipped: !r.Committed, Reason: r.Reason}, nil
}

// idFile is one plan, spec or bug file with the values its own frontmatter
// holds, so a tool works on the files and not on the board's reading of them.
type idFile struct {
	file, id, hash, prefix, frontErr string
}

// idFiles lists every plan, spec and bug file on disk. A plan file is its own
// item on the board now, so it is reached like any other.
func idFiles(b *board.Board) []idFile {
	var out []idFile
	for _, it := range b.Items {
		// A task lives inside its plan file and a debt-item lives inside
		// its debt file, so neither is a file of its own.
		if it.Kind == board.KindTask || it.Kind == board.KindDebtItem || it.Legacy || it.Worktree != "" || !it.OnDisk {
			continue
		}
		file := it.Path
		if it.PlanPath != "" {
			file = it.PlanPath
		}
		why := ""
		for _, p := range it.Problems {
			if strings.HasPrefix(p, "frontmatter: ") {
				why = p
				break
			}
		}
		out = append(out, idFile{file, it.RawID, it.RawHash, filePrefix(it, file), why})
	}
	return out
}

// filePrefix is the prefix a file's own id must carry: PLAN for a plan file,
// and for anything else the kind the file ends up as.
func filePrefix(it *board.Item, file string) string {
	if strings.Contains(filepath.ToSlash(file), "/plans/") {
		return "PLAN"
	}
	return string(board.Prefix(it.Kind, false))
}

// badReason says why a file must be left alone: frontmatter that will not
// parse, or a value in the wrong shape. A value someone wrote down is never
// replaced with a generated one.
func badReason(f idFile) string {
	if f.frontErr != "" {
		return f.frontErr
	}
	if f.id != "" && !board.IsID(f.id, f.prefix) {
		return "bad id " + f.id
	}
	if f.hash != "" && !board.IsHash(f.hash) {
		return "bad hash " + f.hash
	}
	return ""
}

type cand struct {
	it     *board.Item
	file   string
	prefix string
	id     string
	hash   string
}

// candidates lists the files to write, and the files to leave alone with the
// reason. What a file already holds comes from the file, so a value is
// written only where the file has none.
func candidates(cfg config.Config, b *board.Board, only []string) ([]cand, []string) {
	present := map[string]idFile{}
	for _, f := range idFiles(b) {
		present[f.file] = f
	}
	seen := map[string]bool{}
	var out []cand
	var skips []string
	add := func(it *board.Item, file, prefix string) {
		if seen[file] {
			return
		}
		if len(only) > 0 && !named(only, it) {
			return
		}
		seen[file] = true
		f := present[file]
		if why := badReason(f); why != "" {
			skips = append(skips, "skip "+fileID(cfg, file)+": "+why)
			if f.frontErr != "" {
				// Frontmatter that will not parse takes no field at all.
				return
			}
		}
		if f.id != "" && f.hash != "" {
			return
		}
		out = append(out, cand{it: it, file: file, prefix: prefix, id: f.id, hash: f.hash})
	}
	for _, it := range b.Items {
		if it.Kind == board.KindTask || it.Kind == board.KindDebtItem || it.Legacy || it.Worktree != "" || !it.OnDisk {
			continue
		}
		file := it.Path
		if it.PlanPath != "" {
			file = it.PlanPath
		}
		add(it, file, prefixOf(it))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].file < out[j].file })
	sort.Strings(skips)
	return out, skips
}

func named(only []string, it *board.Item) bool {
	for _, o := range only {
		if o == it.ID || strings.HasSuffix(it.Path, "/"+o+".md") || strings.HasSuffix(it.PlanPath, "/"+o+".md") {
			return true
		}
	}
	return false
}

func prefixOf(it *board.Item) string {
	if isPlanFile(it) {
		return "PLAN"
	}
	if it.ShortID != "" {
		return strings.TrimSuffix(it.ShortID, "-"+numPart(it.ShortID))
	}
	if it.Hash != "" {
		if p, _, ok := strings.Cut(it.Hash, "-"); ok {
			return p
		}
	}
	if it.PlanPath != "" {
		return "PLAN"
	}
	return string(board.Prefix(it.Kind, false))
}

func isPlanFile(it *board.Item) bool {
	for _, p := range []string{it.Path, it.PlanPath} {
		if strings.Contains(filepath.ToSlash(p), "/plans/") {
			return true
		}
	}
	return false
}

func scanIDs(b *board.Board) (map[string]int, map[string]bool) {
	next := map[string]int{"SPEC": 1, "PLAN": 1, "BUG": 1, "DEBT": 1, "SCRATCH": 1}
	taken := map[string]bool{}
	note := func(prefix, id, hash string) {
		if n, err := strconv.Atoi(strings.TrimPrefix(id, prefix+"-")); err == nil && !strings.Contains(id, ".") {
			if n+1 > next[prefix] {
				next[prefix] = n + 1
			}
		}
		if hash != "" {
			taken[bareHash(hash)] = true
		}
	}
	for _, it := range b.Items {
		if it.Kind == board.KindTask {
			id, hash := it.ShortID, it.Hash
			if i := strings.Index(id, "."); i >= 0 {
				id = id[:i]
			}
			if i := strings.Index(hash, "."); i >= 0 {
				hash = hash[:i]
			}
			if id != "" {
				note(strings.TrimSuffix(id, "-"+numPart(id)), id, hash)
			} else if hash != "" {
				taken[bareHash(hash)] = true
			}
			continue
		}
		if it.ShortID == "" && it.Hash == "" {
			continue
		}
		prefix := prefixOf(it)
		if it.ShortID != "" {
			note(prefix, it.ShortID, it.Hash)
		} else {
			taken[bareHash(it.Hash)] = true
		}
	}
	return next, taken
}

func numPart(id string) string {
	if i := strings.LastIndex(id, "-"); i >= 0 {
		return id[i+1:]
	}
	return id
}

func bareHash(hash string) string {
	if _, h, ok := strings.Cut(hash, "-"); ok {
		return h
	}
	return hash
}

func freeHash(taken map[string]bool) string {
	for {
		if h := randHash(); !taken[h] {
			return h
		}
	}
}

// fileID names a file the way a person types it: its path under the root,
// without the .md.
func fileID(cfg config.Config, file string) string {
	rel, err := filepath.Rel(cfg.Root, file)
	if err != nil {
		return strings.TrimSuffix(filepath.Base(file), ".md")
	}
	return strings.TrimSuffix(filepath.ToSlash(rel), ".md")
}
