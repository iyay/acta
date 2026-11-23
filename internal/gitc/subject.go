package gitc

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// maxSubject is the longest subject a fold may write, so git log stays on one line.
const maxSubject = 72

// choreSubject reads "chore(<kind>): <verb> <names>".
var choreSubject = regexp.MustCompile(`^chore\(([^)]+)\): (\S+)(?: (.*))?$`)

// actaKind is the kind of a planning fold that holds more than one kind.
const actaKind = "acta"

// scopeSubject reads the scope of any "<type>(<scope>): ..." subject, so a
// "docs(spec)" line counts as spec too.
var scopeSubject = regexp.MustCompile(`^[a-z]+\(([^)]+)\)!?: `)

// actaSubject builds "chore(acta): <kinds> of <names>" for a planning fold
// that mixes kinds, or holds a subject with no kind. The kinds are the
// scopes of lines, in the order first seen, then the folders the files sit
// in with a plural "s" cut ("bugs" is bug). Names from had are cut first,
// so a name this write brings is the last to go. When the kinds leave no
// room for one name, the names win and the kinds are left out. It gives ""
// when not even one name fits.
func actaSubject(lines, folders, had, fresh []string) string {
	var kinds []string
	for _, ln := range lines {
		m := scopeSubject.FindStringSubmatch(strings.TrimSpace(ln))
		if m != nil && m[1] != actaKind && !slices.Contains(kinds, m[1]) {
			kinds = append(kinds, m[1])
		}
	}
	for _, f := range folders {
		if k := strings.TrimSuffix(f, "s"); !slices.Contains(kinds, k) {
			kinds = append(kinds, k)
		}
	}
	prefix := "chore(" + actaKind + "):"
	if len(kinds) > 0 {
		if s := render(prefix+" "+strings.Join(kinds, ", ")+" of", had, fresh); s != "" {
			return s
		}
	}
	return render(prefix, had, fresh)
}

// SubjectKind gives the kind of a planning chore subject, or "" when the
// subject is not one.
func SubjectKind(subject string) string {
	if m := choreSubject.FindStringSubmatch(subject); m != nil {
		return m[1]
	}
	return ""
}

// FoldSubject builds one subject that names what a run of folded chore
// commits holds: the names of each verb joined, verbs in the order first
// seen. Names that do not fit in maxSubject become "+N more", so none is
// lost without a count. It gives "" when the subjects are not all chores of
// one kind, when there are none, or when not even one name fits, so the
// caller keeps the subject it has.
func FoldSubject(subjects []string) string {
	if len(subjects) == 0 {
		return ""
	}
	if len(subjects) == 1 {
		return subjects[0]
	}
	type item struct{ verb, name string }
	kind := ""
	var items []item
	for _, s := range subjects {
		m := choreSubject.FindStringSubmatch(s)
		if m == nil || (kind != "" && m[1] != kind) {
			return ""
		}
		kind = m[1]
		names := []string{""}
		if m[3] != "" {
			names = strings.Split(m[3], ", ")
		}
		for _, n := range names {
			if it := (item{m[2], n}); !slices.Contains(items, it) {
				items = append(items, it)
			}
		}
	}
	// render shows the first k items, grouped by verb, and counts the rest.
	render := func(k int) string {
		var verbs []string
		names := map[string][]string{}
		for _, it := range items[:k] {
			if !slices.Contains(verbs, it.verb) {
				verbs = append(verbs, it.verb)
			}
			if it.name != "" {
				names[it.verb] = append(names[it.verb], it.name)
			}
		}
		groups := make([]string, 0, len(verbs))
		for _, v := range verbs {
			g := v
			if len(names[v]) > 0 {
				g += " " + strings.Join(names[v], ", ")
			}
			groups = append(groups, g)
		}
		out := "chore(" + kind + "): " + strings.Join(groups, "; ")
		if rest := len(items) - k; rest > 0 {
			out += ", +" + strconv.Itoa(rest) + " more"
		}
		return out
	}
	for k := len(items); k > 0; k-- {
		if out := render(k); len(out) <= maxSubject {
			return out
		}
	}
	// Not even one name fits, so give nothing and let the caller keep the
	// subject it has. A cut name would say less than the old subject.
	return ""
}
