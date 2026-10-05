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
	var b [7]byte
	rand.Read(b[:])
	out := []byte{letters[int(b[0])%len(letters)]}
	for _, c := range b[1:] {
		out = append(out, chars[int(c)%len(chars)])
	}
	return string(out)
}

// AssignIDs gives every non-legacy spec, plan and bug in the root folder the
// id and hash it lacks, and moves the ones written the old way to the new
// form. A value already in the new format is never replaced, and a value in
// the wrong shape is a person's typo, so that is left alone with its reason
// instead of ending the run. A first id also writes the day the file was
// made, and a spec that grew out of a scratch idea closes that idea. A file
// that asked for the schema check and fails it takes no id at all. One commit
// holds the run.
func AssignIDs(cfg config.Config, b *board.Board, only []string) ([]string, Outcome, error) {
	next, taken := scanIDs(b)
	cands, skips := candidates(cfg, b, only)
	var changes, paths []string
	migrated := false
	for _, c := range cands {
		orig, err := os.ReadFile(c.file)
		if err != nil {
			return nil, Outcome{}, err
		}
		doc := board.Parse(orig)
		if board.HasSchema(doc.Front) {
			if probs := board.CheckBody(c.it.Kind, doc.Body); len(probs) > 0 {
				// A file that asked for the check and fails it takes nothing,
				// and the other files still get their ids.
				word := string(c.it.Kind)
				if c.it.Kind == board.KindStory {
					word = "spec"
				}
				skips = append(skips, fmt.Sprintf("skip %s: %s %s: %s", fileID(cfg, c.file), word, filepath.Base(c.file), probs[0]))
				continue
			}
		}
		src := orig
		id, hash := c.id, c.hash
		switch {
		case id == "":
			id = board.FormatID(c.prefix, next[c.prefix])
			next[c.prefix]++
			writes := [][2]string{{"id", id}, {"created", Now().Format(stampLayout)}}
			if board.SchemaOn(c.it.Kind) && !hasField(orig, "schema") {
				writes = append(writes, [2]string{"schema", "1"})
			}
			for _, w := range writes {
				out, err := SetField(src, w[0], w[1])
				if err != nil {
					return nil, Outcome{Skips: skips}, err
				}
				src = out
			}
			if parent := closeScratchParent(b, c.it.Kind, doc.Front); parent != "" {
				// The idea became a spec, so it closed today. Its date rides
				// the same commit as the id.
				psrc, err := os.ReadFile(parent)
				if err != nil {
					return nil, Outcome{Skips: skips}, err
				}
				pout, err := MarkFinishedOnce(psrc)
				if err != nil {
					return nil, Outcome{Skips: skips}, err
				}
				if err := os.WriteFile(parent, pout, 0o644); err != nil {
					return nil, Outcome{Skips: skips}, err
				}
				paths = append(paths, parent)
			}
		case c.oldID():
			id = c.it.ShortID
			migrated = true
			out, err := SetField(src, "id", id)
			if err != nil {
				return nil, Outcome{Skips: skips}, err
			}
			src = out
		}
		switch {
		case hash == "":
			hash = freeHash(taken)
		case c.oldHash():
			hash = longHash(hash, taken)
			migrated = true
		}
		if hash != c.hash {
			out, err := SetField(src, "hash", hash)
			if err != nil {
				return nil, Outcome{Skips: skips}, err
			}
			src = out
		}
		if list, changed := migratedCloses(b, c.closes); changed {
			// The line is copied straight, because SetField would write the
			// list as one quoted string and the reader would see one entry.
			src = setCloses(src, c.closesList, list)
			migrated = true
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
	what := "assign short ids"
	if migrated {
		what = "migrate ids to 3-letter prefix"
	}
	msg := Subject(cfg, paths, what)
	r := gitc.CommitPaths(cfg.RepoRoot, paths, msg)
	return changes, Outcome{Committed: r.Committed, Skipped: !r.Committed, Reason: r.Reason, Skips: skips}, nil
}

// closeScratchParent gives the file of the scratch idea a spec grew out of.
// Any other kind of parent is closed by its own command, so it gets nothing.
func closeScratchParent(b *board.Board, k board.Kind, front map[string]any) string {
	if k != board.KindStory {
		return ""
	}
	want, _ := front["parent"].(string)
	if !strings.HasPrefix(want, "scratch/") {
		return ""
	}
	it := b.Get(want)
	if it == nil || !it.OnDisk || it.Kind != board.KindScratch {
		return ""
	}
	return it.Path
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
		k := strings.ToUpper(board.Canon(f.id))
		groups[k] = append(groups[k], f)
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
			id := board.FormatID(prefix, next[prefix])
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
	r := gitc.CommitPaths(cfg.RepoRoot, paths, Subject(cfg, paths, "fix duplicate short ids"))
	return changes, Outcome{Committed: r.Committed, Skipped: !r.Committed, Reason: r.Reason}, nil
}

// idFile is one plan, spec or bug file with the values its own frontmatter
// holds, so a tool works on the files and not on the board's reading of them.
type idFile struct {
	file, id, hash, frontErr, bad string
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
		// The board has already read the frontmatter, so its own complaint is
		// the reason to leave a file alone. An old id or a 4 character hash
		// draws no complaint: `acta id` rewrites those, not this tool.
		why, bad := "", ""
		for _, p := range it.Problems {
			switch {
			case why == "" && strings.HasPrefix(p, "frontmatter: "):
				why = p
			case bad == "" && (strings.HasPrefix(p, "bad id ") || strings.HasPrefix(p, "bad hash ")):
				bad = p
			}
		}
		out = append(out, idFile{file, it.RawID, it.RawHash, why, bad})
	}
	return out
}

// badReason says why a file must be left alone: frontmatter that will not
// parse, or a value in the wrong shape. A value someone wrote down is never
// replaced with a generated one.
func badReason(f idFile) string {
	if f.frontErr != "" {
		return f.frontErr
	}
	return f.bad
}

type cand struct {
	it         *board.Item
	file       string
	prefix     string
	id, hash   string
	closes     []string
	closesList bool
}

// oldID says the id in the file is not the new form the board already reads
// it as, so it is written the way the rest of the tree spells it.
func (c cand) oldID() bool { return c.id != "" && c.it.ShortID != "" && c.id != c.it.ShortID }

// oldHash says the hash is the old 4 characters. A hash of any other wrong
// shape is a problem on the file, so the file never gets here with one.
func (c cand) oldHash() bool { return len(c.hash) == 4 }

// migratedCloses gives the closes: list the way the board reads every entry,
// and says whether any of them moved. An entry that names nothing keeps the
// words it was written with, so a typo is never lost.
func migratedCloses(b *board.Board, entries []string) ([]string, bool) {
	out := make([]string, len(entries))
	moved := false
	for i, e := range entries {
		out[i] = e
		if it := b.Get(e); it != nil && it.ShortID != "" && it.ShortID != e {
			out[i] = it.ShortID
			moved = true
		}
	}
	return out, moved
}

// readCloses gives the closes: entries as the file writes them, and whether
// the file holds a list. A file that will not read is left to the writer,
// which stops the run and says why.
func readCloses(path string) ([]string, bool) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	for i, line := range strings.Split(string(raw), "\n") {
		if i > 0 && strings.TrimSpace(line) == "---" {
			break
		}
		v, ok := strings.CutPrefix(line, "closes:")
		if !ok {
			continue
		}
		v = strings.TrimSpace(v)
		// A comment after the list rides along inside the last entry, so a
		// rewritten line ends with that comment. Rare here, and no entry is
		// ever dropped.
		if list, ok := strings.CutPrefix(v, "["); ok {
			return splitList(strings.TrimSuffix(strings.TrimSpace(list), "]")), true
		}
		return splitList(v), v != ""
	}
	return nil, false
}

func splitList(v string) []string {
	var out []string
	for _, e := range strings.Split(v, ",") {
		if e = strings.TrimSpace(e); e != "" {
			out = append(out, e)
		}
	}
	return out
}

// setCloses rewrites the closes: line and copies every other byte, because a
// list has to stay a list: written as one quoted string the reader would see
// a single entry naming nothing.
func setCloses(src []byte, list bool, entries []string) []byte {
	value := strings.Join(entries, ", ")
	if list {
		value = "[" + value + "]"
	}
	lines := strings.SplitAfter(string(src), "\n")
	for i, line := range lines {
		if i > 0 && strings.TrimSpace(line) == "---" {
			break
		}
		if !strings.HasPrefix(line, "closes:") {
			continue
		}
		// Keep the line's own ending, so a CRLF file does not grow a second
		// style in one block.
		head := strings.TrimRight(line, "\r\n")
		lines[i] = "closes: " + value + strings.TrimPrefix(line, head)
		return []byte(strings.Join(lines, ""))
	}
	return src
}

// longHash keeps the 4 characters a file already has and adds 3 more, so an
// old hash still names the same item, and draws again until it clashes with
// nothing.
func longHash(old string, taken map[string]bool) string {
	for {
		if h := old + randHash()[4:]; !taken[h] {
			return h
		}
	}
}

// candidates lists the files to write, and the files to leave alone with the
// reason. What a file already holds comes from the file, so a value is
// written only where the file has none, and a file still holding the old form
// is a candidate even though it is not empty.
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
		c := cand{it: it, file: file, prefix: prefix, id: f.id, hash: f.hash}
		c.closes, c.closesList = readCloses(file)
		if c.id != "" && c.hash != "" && !c.oldID() && !c.oldHash() {
			if _, moved := migratedCloses(b, c.closes); !moved {
				return
			}
		}
		out = append(out, c)
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
		return "PLN"
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
		return "PLN"
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
	next := map[string]int{"SPC": 1, "PLN": 1, "BUG": 1, "DBT": 1, "SCR": 1}
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
