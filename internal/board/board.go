package board

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"pm-board/internal/config"
)

// Kind is what an item is.
type Kind string

const (
	KindStory Kind = "story"
	KindTask  Kind = "task"
	KindBug   Kind = "bug"
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
	Children     []string // stories and bugs: their task ids, in file order
	Done         int      // task: ticked boxes; story or bug: done tasks
	Total        int      // task: all boxes; story or bug: all tasks
	Path         string   // absolute file path
	Line         int      // 1-based line to open the file at
	Legacy       bool
	Problems     []string
	Body         string // file body without frontmatter; a task holds only its section
	TaskNum      string

	fmStatus string
	plans    int
	specFile bool
	seq      int
}

// Board holds every item of one repo, newest first.
type Board struct {
	Items []*Item
	byID  map[string]*Item
}

var (
	storyStatuses = []string{"draft", "approved", "in-progress", "done", "dropped"}
	bugStatuses   = []string{"open", "fixing", "fixed", "wontfix"}
	taskStatuses  = []string{"todo", "doing", "done"}
	datedName     = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})-(.+)$`)
)

// Allowed lists the status values the contract gives a kind.
func Allowed(k Kind) []string {
	switch k {
	case KindStory:
		return append([]string(nil), storyStatuses...)
	case KindBug:
		return append([]string(nil), bugStatuses...)
	default:
		return append([]string(nil), taskStatuses...)
	}
}

// Closed says whether a status takes an item off the active lists.
func Closed(status string) bool {
	switch status {
	case "done", "fixed", "dropped", "wontfix":
		return true
	}
	return false
}

type planFile struct {
	id, path, date, slug string
	legacy               bool
	doc                  Doc
}

// Load reads the root folder and the legacy folders of cfg. A missing folder
// is not an error; it just has no items.
func Load(cfg config.Config) (*Board, error) {
	b := &Board{byID: map[string]*Item{}}
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

	var plans []planFile
	for _, s := range sources {
		parts := []struct {
			dir  string
			kind Kind
		}{{s.dirs.Specs, KindStory}, {s.dirs.Bugs, KindBug}, {s.dirs.Plans, ""}}
		for _, part := range parts {
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
				id := makeID(s.idRoot, f)
				date, slug := splitName(filepath.Base(f))
				if part.kind == "" {
					plans = append(plans, planFile{id: id, path: f, date: date, slug: slug, legacy: s.legacy, doc: doc})
					continue
				}
				it := fileItem(part.kind, id, f, date, slug, s.legacy, doc)
				it.specFile = part.kind == KindStory
				b.add(it)
			}
		}
	}
	// Plans link after every spec and bug is known, so order does not matter.
	for _, p := range plans {
		b.linkPlan(p)
	}
	b.derive()
	b.sortItems()
	return b, nil
}

// Get returns the item with this id, or nil.
func (b *Board) Get(id string) *Item { return b.byID[id] }

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

// Untyped gives legacy stories and bugs. Closed ones are left out unless all is set.
func (b *Board) Untyped(all bool) []*Item {
	var out []*Item
	for _, it := range b.Items {
		if it.Legacy && it.Kind != KindTask && (all || !Closed(it.Status)) {
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
	it.fmStatus = field(doc.Front, "status")
	return it
}

// linkPlan finds the parent of a plan's tasks: frontmatter parent, then the
// **Spec:** line, then the plan stands in for a story of its own.
func (b *Board) linkPlan(p planFile) {
	var parent *Item
	problem := ""
	if want := field(p.doc.Front, "parent"); want != "" {
		if it := b.byID[want]; it != nil && it.Kind != KindTask {
			parent = it
		} else {
			problem = "parent " + want + " not found"
		}
	} else if p.doc.SpecPath != "" {
		if parent = b.findSpec(p.doc.SpecPath); parent == nil {
			problem = "spec " + p.doc.SpecPath + " not found"
		}
	}
	if parent == nil {
		parent = fileItem(KindStory, p.id, p.path, p.date, p.slug, p.legacy, p.doc)
		if problem != "" {
			parent.Problems = append(parent.Problems, problem)
		}
		b.add(parent)
	}
	parent.plans++
	for _, t := range p.doc.Tasks {
		id := p.id + "#task-" + t.Num
		if b.byID[id] != nil {
			parent.Problems = append(parent.Problems, "duplicate task "+t.Num+" in "+p.id)
			continue
		}
		title := t.Title
		if title == "" {
			title = "Task " + t.Num
		}
		b.add(&Item{ID: id, Kind: KindTask, Title: title, Date: p.date, Slug: p.slug,
			Ref: parent.Ref, Parent: parent.ID, Done: t.Done, Total: t.Total,
			Path: p.path, Line: t.Line, Legacy: p.legacy, Body: t.Body, TaskNum: t.Num})
		parent.Children = append(parent.Children, id)
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
			it.Status, it.StatusSource = taskStatus(it.Done, it.Total), "derived"
		}
	}
	for _, it := range b.Items {
		if it.Kind == KindTask {
			continue
		}
		done, started := 0, 0
		for _, id := range it.Children {
			switch b.byID[id].Status {
			case "done":
				done++
				started++
			case "doing":
				started++
			}
		}
		it.Done, it.Total = done, len(it.Children)
		if it.fmStatus == "" {
			it.Status, it.StatusSource = parentStatus(it.Kind, it.plans, done, started, it.Total), "derived"
			continue
		}
		it.Status, it.StatusSource = it.fmStatus, "frontmatter"
		if !contains(Allowed(it.Kind), it.fmStatus) {
			it.Problems = append(it.Problems, "unknown status "+it.fmStatus)
		}
	}
}

func taskStatus(done, total int) string {
	switch {
	case done == 0:
		return "todo"
	case done == total:
		return "done"
	default:
		return "doing"
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

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
