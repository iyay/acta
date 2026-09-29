package board

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/iyay/acta/internal/config"
)

// Kind is what an item is.
type Kind string

const (
	KindStory    Kind = "story"
	KindTask     Kind = "task"
	KindBug      Kind = "bug"
	KindPlan     Kind = "plan"
	KindDebt     Kind = "debt"
	KindDebtItem Kind = "debt-item"
	KindScratch  Kind = "scratch"
)

// Item is one story, task or bug.
type Item struct {
	ID           string
	Kind         Kind
	Title        string
	Date         string // YYYY-MM-DD from the file name, or ""
	Slug         string
	Status       string
	StatusSource string // "derived" or "frontmatter"
	Ref          string
	FixedIn      string
	Parent       string   // tasks only
	SpecID       string   // plans only: the spec or bug whose tasks the plan carries
	PlanID       string   // tasks only: the plan file the task belongs to
	Children     []string // stories, bugs and plans: their task ids, in file order
	Done         int      // task: ticked boxes; story or bug: done tasks
	Total        int      // task: all boxes; story or bug: all tasks
	Closes       []string // specs, plans and bugs: the items this one also finishes
	ClosedBy     []string // items whose closes names this one, in file order
	Path         string   // absolute file path
	Line         int      // 1-based line to open the file at
	Legacy       bool
	Worktree     string // branch of the worktree the file came from; "" for the main tree
	OnDisk       bool   // false for items read from a branch that is not checked out
	Author       string // the git author of the commit that first added the file this item lives in
	Problems     []string
	Body         string // file body without frontmatter; a task holds only its section
	TaskNum      string
	ShortID      string // number ID like PLAN-12 or PLAN-12.3; "" when the file has none
	Hash         string // permanent ID like PLAN-k3f2 or PLAN-k3f2.3; "" when none
	RawID        string // id as the frontmatter holds it, "" when the file has none
	RawHash      string // hash as the frontmatter holds it, "" when the file has none
	PlanPath     string // tasks only: the plan file, which holds the plan's IDs
	Agent        string // the agent that last ticked this open task, or the agents of its open tasks
	Started      bool   // tasks only: someone ran tick --start on it
	Created      string // YYYY-MM-DD from the frontmatter, or ""
	StartedOn    string // YYYY-MM-DD, the day work on it began
	Finished     string // YYYY-MM-DD, the day it was finished

	// plans counts the plans that hang on this item. A plan item is one plan
	// itself, so it counts itself.
	fmStatus string
	fmCloses []string
	fmParent string
	plans    int
	// countedOn names the items a plan already counts on, so a closes: list
	// naming one of them does not count the same plan a second time.
	countedOn []string
	specFile  bool
	seq       int
}

// Board holds every item of one repo, newest first.
type Board struct {
	Items []*Item
	byID  map[string]*Item
	alias map[string]*Item // short IDs (number and hash, any case) to items
}

var (
	storyStatuses   = []string{"draft", "approved", "in-progress", "done", "dropped"}
	bugStatuses     = []string{"open", "fixing", "fixed", "wontfix"}
	taskStatuses    = []string{"todo", "in-progress", "done"}
	debtStatuses    = []string{"open", "done", "wontfix"}
	scratchStatuses = []string{"raw", "brainstorming", "dropped"}
	datedName       = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})-(.+)$`)
)

// Allowed lists the status values the contract gives a kind.
func Allowed(k Kind) []string {
	switch k {
	case KindStory, KindPlan:
		return append([]string(nil), storyStatuses...)
	case KindBug:
		return append([]string(nil), bugStatuses...)
	case KindDebt, KindDebtItem:
		return append([]string(nil), debtStatuses...)
	case KindScratch:
		return append([]string(nil), scratchStatuses...)
	default:
		return append([]string(nil), taskStatuses...)
	}
}

// Closed says whether a status takes an item off the active lists.
func Closed(status string) bool {
	switch status {
	case "done", "fixed", "dropped", "wontfix", "specced":
		return true
	}
	return false
}

type planFile struct {
	id, path, date, slug string
	legacy               bool
	doc                  Doc
	tree                 string
	onDisk               bool
}

// Tree is another worktree of the same repo, read with its own config.
// When Files is not nil the tree is a branch read from git, not from disk.
type Tree struct {
	Cfg    config.Config
	Branch string
	Files  map[string][]byte
}

type srcFile struct {
	id, path, date, slug string
	kind                 Kind // "" for a plan
	legacy               bool
	doc                  Doc
	tree                 string // worktree or branch name; "" for the main tree
	onDisk               bool
	ticked               int
}

// Load reads one tree: the root folder and the legacy folders of cfg.
func Load(cfg config.Config) (*Board, error) { return LoadTrees(cfg, nil) }

// LoadTrees reads the main tree and the other worktrees into one board. A file
// found only in a worktree is shown from there; a file found in both is shown
// from the worktree only when its plan has more ticked boxes. A worktree that
// cannot be read is skipped, so it never hides the main board.
func LoadTrees(main config.Config, others []Tree) (*Board, error) {
	files, err := collect(Tree{Cfg: main})
	if err != nil {
		return nil, err
	}
	at := map[string]int{}
	for i, f := range files {
		at[fileKey(f)] = i
	}
	for _, t := range others {
		more, err := collect(t)
		if err != nil {
			continue
		}
		for _, f := range more {
			k := fileKey(f)
			i, ok := at[k]
			if !ok {
				at[k] = len(files)
				files = append(files, f)
				continue
			}
			if f.ticked > files[i].ticked {
				files[i] = f
			}
		}
	}

	b := &Board{byID: map[string]*Item{}, alias: map[string]*Item{}}
	var plans []planFile
	var debts []debtFile
	for _, f := range files {
		if f.kind == "" {
			plans = append(plans, planFile{id: f.id, path: f.path, date: f.date, slug: f.slug, legacy: f.legacy, doc: f.doc, tree: f.tree, onDisk: f.onDisk})
			continue
		}
		it := fileItem(f.kind, f.id, f.path, f.date, f.slug, f.legacy, f.doc)
		it.specFile = f.kind == KindStory
		it.Worktree = f.tree
		it.OnDisk = f.onDisk
		b.add(it)
		b.aliasItem(it)
		if f.kind == KindDebt {
			debts = append(debts, debtFile{it: it, doc: f.doc})
		}
	}
	// Plans link after every spec and bug is known, so order does not matter.
	for _, p := range plans {
		b.linkPlan(p)
	}
	// Debt files can name a plan as their parent, so they link once every
	// plan exists on the board.
	for _, d := range debts {
		b.linkDebt(d)
	}
	b.linkScratch(main.Dirs.Scratch)
	b.linkCloses()
	b.fillStarted(main, others)
	b.derive()
	b.fillAuthors(main.Root)
	b.fillAgents(main, others)
	b.sortItems()
	return b, nil
}

// collect reads every planning file of one tree. A missing folder has no
// files. A tree with Files set is a branch read from git: only its root folder
// counts, and its paths are shown as "<branch>:<path>".
func collect(t Tree) ([]srcFile, error) {
	if t.Files != nil {
		return collectFiles(t), nil
	}
	cfg, tree := t.Cfg, t.Branch
	type source struct {
		base, idRoot string
		dirs         config.Dirs
		legacy       bool
	}
	sources := []source{{base: cfg.Root, idRoot: cfg.Root, dirs: cfg.Dirs}}
	for _, l := range cfg.Legacy {
		sources = append(sources, source{base: l, idRoot: cfg.RepoRoot,
			dirs: config.Dirs{Specs: "specs", Plans: "plans", Bugs: "bugs"}, legacy: true})
	}
	var out []srcFile
	for _, s := range sources {
		parts := []struct {
			dir  string
			kind Kind
		}{{s.dirs.Specs, KindStory}, {s.dirs.Bugs, KindBug}, {s.dirs.Debt, KindDebt}, {s.dirs.Scratch, KindScratch}, {s.dirs.Plans, ""}}
		for _, part := range parts {
			// A legacy source has no debt folder, so an empty dir name is
			// skipped instead of globbing the whole legacy root.
			if part.dir == "" {
				continue
			}
			files, err := filepath.Glob(filepath.Join(s.base, part.dir, "*.md"))
			if err != nil {
				return nil, err
			}
			sort.Strings(files)
			for _, f := range files {
				src, err := os.ReadFile(f)
				if err != nil {
					return nil, err
				}
				doc := Parse(src)
				date, slug := splitName(filepath.Base(f))
				ticked := 0
				for _, t := range doc.Tasks {
					ticked += t.Done
				}
				out = append(out, srcFile{id: makeID(s.idRoot, f), path: f, date: date, slug: slug,
					kind: part.kind, legacy: s.legacy, doc: doc, tree: tree, onDisk: true, ticked: ticked})
			}
		}
	}
	return out, nil
}

func collectFiles(t Tree) []srcFile {
	cfg := t.Cfg
	rootRel, err := filepath.Rel(cfg.RepoRoot, cfg.Root)
	if err != nil {
		return nil
	}
	rootRel = filepath.ToSlash(rootRel)
	kinds := map[string]Kind{cfg.Dirs.Specs: KindStory, cfg.Dirs.Bugs: KindBug, cfg.Dirs.Debt: KindDebt, cfg.Dirs.Scratch: KindScratch, cfg.Dirs.Plans: ""}
	names := make([]string, 0, len(t.Files))
	for name := range t.Files {
		names = append(names, name)
	}
	sort.Strings(names)
	var out []srcFile
	for _, name := range names {
		rest, ok := strings.CutPrefix(name, rootRel+"/")
		if !ok || !strings.HasSuffix(rest, ".md") {
			continue
		}
		dir, file, ok := strings.Cut(rest, "/")
		kind, known := kinds[dir]
		if !ok || !known || strings.Contains(file, "/") {
			continue
		}
		doc := Parse(t.Files[name])
		date, slug := splitName(file)
		ticked := 0
		for _, ts := range doc.Tasks {
			ticked += ts.Done
		}
		out = append(out, srcFile{id: strings.TrimSuffix(rest, ".md"), path: t.Branch + ":" + name,
			date: date, slug: slug, kind: kind, doc: doc, tree: t.Branch, ticked: ticked})
	}
	return out
}

// fileKey keeps root and legacy files apart, since their ids use different roots.
func fileKey(f srcFile) string {
	if f.legacy {
		return "legacy:" + f.id
	}
	return "root:" + f.id
}

// Get finds an item by path ID, number ID or hash ID. Short IDs ignore case.
func (b *Board) Get(id string) *Item {
	if it := b.byID[id]; it != nil {
		return it
	}
	return b.alias[strings.ToLower(id)]
}

// List gives non-legacy items of kind k. Closed ones are left out unless all is set.
func (b *Board) List(k Kind, all bool) []*Item {
	var out []*Item
	for _, it := range b.Items {
		if it.Kind == k && !it.Legacy && (all || !Closed(it.Status)) {
			out = append(out, it)
		}
	}
	return out
}

// Untyped gives legacy stories and bugs. A legacy plan is a plan now, so it is
// left out. Closed ones are left out unless all is set.
func (b *Board) Untyped(all bool) []*Item {
	var out []*Item
	for _, it := range b.Items {
		if it.Legacy && it.Kind != KindTask && it.Kind != KindPlan && (all || !Closed(it.Status)) {
			out = append(out, it)
		}
	}
	return out
}

// Search matches title, ref and slug of every item, case-insensitive.
func (b *Board) Search(q string) []*Item {
	q = strings.ToLower(strings.TrimSpace(q))
	if q == "" {
		return nil
	}
	var out []*Item
	for _, it := range b.Items {
		if strings.Contains(strings.ToLower(it.Title), q) ||
			strings.Contains(strings.ToLower(it.Ref), q) ||
			strings.Contains(strings.ToLower(it.Slug), q) {
			out = append(out, it)
		}
	}
	return out
}

func (b *Board) add(it *Item) {
	if old := b.byID[it.ID]; old != nil {
		old.Problems = append(old.Problems, "duplicate id "+it.ID)
		return
	}
	it.seq = len(b.Items)
	b.Items = append(b.Items, it)
	b.byID[it.ID] = it
}

func fileItem(k Kind, id, path, date, slug string, legacy bool, doc Doc) *Item {
	it := &Item{ID: id, Kind: k, Title: doc.Title, Date: date, Slug: slug,
		Path: path, Line: 1, Legacy: legacy, Body: doc.Body}
	if it.Title == "" {
		it.Title = slug
	}
	if doc.FrontErr != nil {
		it.Problems = append(it.Problems, "frontmatter: "+doc.FrontErr.Error())
	}
	if t := field(doc.Front, "type"); t != "" {
		switch Kind(t) {
		case KindStory, KindBug:
			it.Kind = Kind(t)
		default:
			it.Problems = append(it.Problems, "unknown type "+t)
		}
	}
	it.Ref = field(doc.Front, "ref")
	it.FixedIn = field(doc.Front, "fixed_in")
	it.Created = dateField(doc.Front, "created")
	it.StartedOn = dateField(doc.Front, "started")
	it.Finished = dateField(doc.Front, "finished")
	it.fmStatus = field(doc.Front, "status")
	it.fmParent = field(doc.Front, "parent")
	it.fmCloses = listField(doc.Front, "closes")
	setIDs(it, Prefix(it.Kind, false), doc)
	return it
}

// linkPlan gives a plan its own item and hangs the plan's tasks under it. The
// spec or bug the plan names keeps the same tasks, so its own progress and
// status do not change. A Spec line counts its spec even when the plan hangs
// under a parent, or the spec never learns the plan exists (SPEC-16).
func (b *Board) linkPlan(p planFile) {
	var parent *Item
	var spec *Item
	var problems []string
	wroteParent := field(p.doc.Front, "parent") != ""
	if want := field(p.doc.Front, "parent"); want != "" {
		if it := b.byID[want]; it != nil && it.Kind != KindTask {
			parent = it
		} else {
			problems = append(problems, "parent "+want+" not found")
		}
	}
	if p.doc.SpecPath != "" {
		// The Spec line names the spec even when parent: is set, so a plan can
		// hang under debt and still count for the spec that built it. A written
		// parent that resolves to nothing stays broken; the spec does not take
		// the plan's place in the tree, the plan just loses it.
		spec = b.findSpec(p.doc.SpecPath)
		switch {
		case spec == nil:
			problems = append(problems, "spec "+p.doc.SpecPath+" not found")
		case parent == nil && !wroteParent:
			parent = spec
		case spec != parent:
			spec.plans++
		}
	}
	plan := fileItem(KindPlan, p.id, p.path, p.date, p.slug, p.legacy, p.doc)
	plan.Worktree = p.tree
	plan.OnDisk = p.onDisk
	plan.Problems = append(plan.Problems, problems...)
	b.add(plan)
	// A plan speaks PLAN, whatever kind its frontmatter or its folder claims.
	setIDs(plan, "PLAN", p.doc)
	b.aliasItem(plan)
	// A plan is one plan, so it counts itself when its status is derived.
	plan.plans = 1
	if parent != nil {
		plan.SpecID = parent.ID
		parent.plans++
	}
	// A plan can point at one spec from two sides: a parent and a Spec line.
	// Note every item that took the plan, so linkCloses knows which specs
	// already hold it and leaves them alone instead of counting it twice.
	if parent != nil {
		plan.countedOn = append(plan.countedOn, parent.ID)
	}
	if spec != nil && spec != parent {
		plan.countedOn = append(plan.countedOn, spec.ID)
	}
	for _, t := range p.doc.Tasks {
		id := p.id + "#task-" + t.Num
		if b.byID[id] != nil {
			plan.Problems = append(plan.Problems, "duplicate task "+t.Num+" in "+p.id)
			continue
		}
		title := t.Title
		if title == "" {
			title = "Task " + t.Num
		}
		holder := parent
		if holder == nil {
			holder = plan
		}
		task := &Item{ID: id, Kind: KindTask, Title: title, Date: p.date, Slug: p.slug,
			Ref: holder.Ref, Parent: holder.ID, PlanID: p.id, Done: t.Done, Total: t.Total,
			Path: p.path, Line: t.Line, Legacy: p.legacy, Body: t.Body, TaskNum: t.Num,
			PlanPath: p.path, Worktree: p.tree, OnDisk: p.onDisk}
		b.add(task)
		b.aliasTask(task, plan.ShortID, plan.Hash)
		plan.Children = append(plan.Children, id)
		if parent != nil {
			parent.Children = append(parent.Children, id)
		}
		if spec != nil && spec != parent {
			spec.Children = append(spec.Children, id)
		}
	}
}

// debtFile pairs an already-added debt Item with the doc it came from, so its
// checklist lines can be turned into debt-item children once every plan on
// the board is known.
type debtFile struct {
	it  *Item
	doc Doc
}

// itemStatus turns a checklist box state into a debt-item status.
func itemStatus(state byte) string {
	switch state {
	case 'x':
		return "done"
	case '-':
		return "wontfix"
	default:
		return "open"
	}
}

// linkDebt checks a debt file's parent plan and gives it one debt-item child
// per checklist line, the same way linkPlan gives a plan its task children.
func (b *Board) linkDebt(d debtFile) {
	if want := field(d.doc.Front, "parent"); want != "" {
		if p := b.byID[want]; p == nil || p.Kind != KindPlan {
			d.it.Problems = append(d.it.Problems, "parent "+want+" not found")
		}
	}
	for _, line := range d.doc.Items {
		id := d.it.ID + "#item-" + fmt.Sprintf("%d", line.Num)
		item := &Item{ID: id, Kind: KindDebtItem, Title: line.Text, Date: d.it.Date, Slug: d.it.Slug,
			Parent: d.it.ID, Path: d.it.Path, Line: line.Line, Legacy: d.it.Legacy,
			Status: itemStatus(line.State), StatusSource: "derived",
			// The checklist becomes the list in the detail pane, so the body
			// holds only the words the file has around it.
			Body: d.doc.Text, Worktree: d.it.Worktree, OnDisk: d.it.OnDisk}
		b.add(item)
		b.aliasDebtItem(item, d.it.ShortID, d.it.Hash, line.Num)
		d.it.Children = append(d.it.Children, id)
	}
}

// findSpec matches a **Spec:** path to a spec file by file name. A root spec
// wins over a legacy one with the same name.
func (b *Board) findSpec(path string) *Item {
	want := strings.TrimSuffix(filepath.Base(strings.TrimSpace(path)), ".md")
	var legacyHit *Item
	for _, it := range b.Items {
		if !it.specFile || strings.TrimSuffix(filepath.Base(it.Path), ".md") != want {
			continue
		}
		if !it.Legacy {
			return it
		}
		if legacyHit == nil {
			legacyHit = it
		}
	}
	return legacyHit
}

func (b *Board) derive() {
	for _, it := range b.Items {
		if it.Kind == KindTask {
			it.Status, it.StatusSource = taskStatus(it.Done, it.Total, it.Started), "derived"
		}
	}
	for _, it := range b.Items {
		if it.Kind == KindTask || it.Kind == KindDebtItem {
			continue
		}
		if it.Kind == KindDebt {
			done := 0
			for _, id := range it.Children {
				if Closed(b.byID[id].Status) {
					done++
				}
			}
			it.Done, it.Total = done, len(it.Children)
			it.StatusSource = "derived"
			if it.Total > 0 && done == it.Total {
				it.Status = "done"
			} else {
				it.Status = "open"
			}
			continue
		}
		done, started := 0, 0
		for _, id := range it.Children {
			switch b.byID[id].Status {
			case "done":
				done++
				started++
			case "in-progress":
				started++
			}
		}
		it.Done, it.Total = done, len(it.Children)
		switch {
		case it.Kind == KindScratch:
			// A scratch named in a closes list was turned into a spec just
			// like one that hangs under a spec as its parent.
			it.Status, it.StatusSource = scratchStatus(it.fmStatus, len(it.Children) > 0 || len(it.ClosedBy) > 0)
		case it.fmStatus == "":
			if st, ok := closedByStatus(b, it); ok {
				it.Status, it.StatusSource = st, "derived"
			} else {
				it.Status, it.StatusSource = parentStatus(it.Kind, it.plans, done, started, it.Total), "derived"
			}
		case it.Kind == KindStory && it.plans > 0:
			// A spec with a plan of its own decides that status from the plan's
			// boxes. A status written next to the spec can only be a leftover
			// from before the plan existed, so it loses and the board says so
			// rather than letting a stale word hide a finished spec (SPEC-6).
			it.Status, it.StatusSource = parentStatus(it.Kind, it.plans, done, started, it.Total), "derived"
			if it.fmStatus != "" && it.fmStatus != it.Status {
				it.Problems = append(it.Problems, "written status "+it.fmStatus+" ignored, derived "+it.Status)
			}
		default:
			it.Status, it.StatusSource = it.fmStatus, "frontmatter"
		}
		if it.fmStatus != "" && !contains(Allowed(it.Kind), it.fmStatus) {
			it.Problems = append(it.Problems, "unknown status "+it.fmStatus)
		}
	}
	// A spec with no plan of its own follows the specs and bugs that close
	// it. A spec may close another spec, so the board can hold a chain and
	// the order files were read in decides which end is settled first. The
	// loop repeats until a pass changes nothing, and no chain is longer than
	// the board, so it always ends.
	for pass := 0; pass <= len(b.Items); pass++ {
		settled := true
		for _, it := range b.Items {
			st, ok := closedByStatus(b, it)
			if ok && st != it.Status {
				it.Status, it.StatusSource = st, "derived"
				settled = false
			}
		}
		if settled {
			break
		}
	}
}

func taskStatus(done, total int, started bool) string {
	switch {
	case done == 0:
		// Started with no box ticked yet: work has begun, so it is in progress.
		if started {
			return "in-progress"
		}
		return "todo"
	case done == total:
		return "done"
	default:
		return "in-progress"
	}
}

func parentStatus(k Kind, plans, done, started, total int) string {
	notStarted, going, finished, none := "approved", "in-progress", "done", "draft"
	if k == KindBug {
		notStarted, going, finished, none = "open", "fixing", "fixed", "open"
	}
	switch {
	case plans == 0:
		return none
	case total > 0 && done == total:
		return finished
	case started > 0:
		return going
	default:
		return notStarted
	}
}

// sortItems puts the newest date first; tasks stay in file order under their plan.
func (b *Board) sortItems() {
	sort.SliceStable(b.Items, func(i, j int) bool {
		a, c := b.Items[i], b.Items[j]
		if a.Date != c.Date {
			return a.Date > c.Date
		}
		if pa, pc := planOf(a.ID), planOf(c.ID); pa != pc {
			return pa < pc
		}
		return a.seq < c.seq
	})
}

func planOf(id string) string {
	if i := strings.Index(id, "#"); i >= 0 {
		return id[:i]
	}
	return id
}

func makeID(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		rel = path
	}
	return strings.TrimSuffix(filepath.ToSlash(rel), ".md")
}

func splitName(base string) (date, slug string) {
	name := strings.TrimSuffix(base, ".md")
	if m := datedName.FindStringSubmatch(name); m != nil {
		return m[1], m[2]
	}
	return "", name
}

func field(front map[string]any, key string) string {
	v, ok := front[key]
	if !ok || v == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(v))
}

// dateField is a YYYY-MM-DD value from the frontmatter, or "". A bare date
// comes back from yaml as a time and a quoted one as a string, so both are
// read; anything that is not a real day is dropped.
func dateField(front map[string]any, key string) string {
	switch v := front[key].(type) {
	case string:
		if s := strings.TrimSpace(v); isDate(s) {
			return s
		}
	case time.Time:
		return v.Format("2006-01-02")
	}
	return ""
}

// isDate says whether the text is a real day, so a day that does not exist
// never reaches the reader.
func isDate(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
