package write

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"pm-board/internal/board"
	"pm-board/internal/config"
	"pm-board/internal/gitc"
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
// other item has. Existing values never change. One commit holds the run.
func AssignIDs(cfg config.Config, b *board.Board, only []string) ([]string, Outcome, error) {
	next, taken := scanIDs(b)
	cands := candidates(b, only)
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
				return nil, Outcome{}, err
			}
			src = out
		}
		if hash == "" {
			hash = freeHash(taken)
			out, err := SetField(src, "hash", hash)
			if err != nil {
				return nil, Outcome{}, err
			}
			src = out
		}
		if string(src) == string(orig) {
			continue
		}
		if err := os.WriteFile(c.file, src, 0o644); err != nil {
			return nil, Outcome{}, err
		}
		taken[hash] = true
		rel, _ := filepath.Rel(cfg.Root, c.file)
		changes = append(changes, id+" ("+filepath.ToSlash(rel)+")")
		paths = append(paths, c.file)
	}
	if len(paths) == 0 {
		return nil, Outcome{}, nil
	}
	sort.Strings(paths)
	r := gitc.CommitPaths(cfg.RepoRoot, paths, "pm: assign short ids")
	return changes, Outcome{Committed: r.Committed, Skipped: !r.Committed, Reason: r.Reason}, nil
}

// FixDuplicates gives each later holder of a number the next free number.
// The first item to reach the branch keeps its number, hashes never move.
func FixDuplicates(cfg config.Config, b *board.Board) ([]string, Outcome, error) {
	next, _ := scanIDs(b)
	groups := map[string][]*board.Item{}
	for _, it := range b.Items {
		if it.Kind == board.KindTask || it.ShortID == "" || it.Legacy || it.Worktree != "" || !it.OnDisk {
			continue
		}
		groups[it.ShortID] = append(groups[it.ShortID], it)
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
			ai, _ := gitc.AddedAt(cfg.RepoRoot, g[i].Path)
			aj, _ := gitc.AddedAt(cfg.RepoRoot, g[j].Path)
			if ai == 0 || aj == 0 {
				return aj == 0 && ai != 0
			}
			return ai < aj
		})
		prefix := strings.TrimSuffix(k, "-"+numPart(k))
		for _, it := range g[1:] {
			id := fmt.Sprintf("%s-%d", prefix, next[prefix])
			next[prefix]++
			src, err := os.ReadFile(it.Path)
			if err != nil {
				return nil, Outcome{}, err
			}
			out, err := SetField(src, "id", id)
			if err != nil {
				return nil, Outcome{}, err
			}
			if err := os.WriteFile(it.Path, out, 0o644); err != nil {
				return nil, Outcome{}, err
			}
			rel, _ := filepath.Rel(cfg.Root, it.Path)
			changes = append(changes, k+" -> "+id+" ("+filepath.ToSlash(rel)+")")
			paths = append(paths, it.Path)
		}
	}
	sort.Strings(paths)
	r := gitc.CommitPaths(cfg.RepoRoot, paths, "pm: fix duplicate short ids")
	return changes, Outcome{Committed: r.Committed, Skipped: !r.Committed, Reason: r.Reason}, nil
}

type cand struct {
	it     *board.Item
	file   string
	prefix string
	id     string
	hash   string
}

func candidates(b *board.Board, only []string) []cand {
	seen := map[string]bool{}
	var out []cand
	add := func(it *board.Item, file, prefix, id, hash string) {
		if seen[file] || (id != "" && hash != "") {
			return
		}
		if len(only) > 0 && !named(only, it) {
			return
		}
		seen[file] = true
		out = append(out, cand{it: it, file: file, prefix: prefix, id: id, hash: hash})
	}
	for _, it := range b.Items {
		if it.Kind == board.KindTask || it.Legacy || it.Worktree != "" || !it.OnDisk {
			continue
		}
		file := it.Path
		if it.PlanPath != "" {
			file = it.PlanPath
		}
		add(it, file, prefixOf(it), it.ShortID, bareHash(it.Hash))
	}
	for _, it := range b.Items {
		if it.Kind != board.KindTask || it.Legacy || it.Worktree != "" || !it.OnDisk || it.PlanPath == "" {
			continue
		}
		id, hash := it.ShortID, it.Hash
		if i := strings.Index(id, "."); i >= 0 {
			id = id[:i]
		}
		if i := strings.Index(hash, "."); i >= 0 {
			hash = hash[:i]
		}
		add(it, it.PlanPath, "PLAN", id, bareHash(hash))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].file < out[j].file })
	return out
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
	next := map[string]int{"SPEC": 1, "PLAN": 1, "BUG": 1}
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

func pathID(cfg config.Config, it *board.Item) string {
	rel, err := filepath.Rel(cfg.Root, it.Path)
	if err != nil {
		return it.ID
	}
	return strings.TrimSuffix(filepath.ToSlash(rel), ".md")
}
