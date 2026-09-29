package board

import "strings"

// Section is one "## " part of a body, in the order the kind wants it.
type Section struct {
	Name     string
	Required bool
}

// Schema lists the sections of each kind in order. It comes from SPEC-22.
var Schema = map[Kind][]Section{
	KindScratch: {{"Words", true}, {"Context", false}, {"Log", false}, {"Open questions", false}},
	KindBug:     {{"Symptom", true}, {"Root cause", false}, {"Repro", false}, {"Found in", false}, {"Context", false}},
	KindDebt:    {{"Notes", true}, {"Context", false}},
	KindStory:   {{"Why", true}, {"Design", true}, {"Testing", true}, {"Context", false}},
	KindPlan:    {{"Global Constraints", true}, {"File Map", true}, {"Waves", true}, {"Context", false}},
}

// SchemaOn says which kinds get "schema: 1" on new files. The other kinds
// wait for their own plan, so their files are never checked yet.
func SchemaOn(k Kind) bool { return k == KindScratch }

// HasSchema says whether a file opted in to the check. Files without it are
// old and are left alone.
func HasSchema(front map[string]any) bool {
	switch v := front["schema"].(type) {
	case int:
		return v == 1
	case string:
		return v == "1"
	}
	return false
}

// SectionNames gives the section names of a kind, in order.
func SectionNames(k Kind) []string {
	var out []string
	for _, s := range Schema[k] {
		out = append(out, s.Name)
	}
	return out
}

// CheckBody lists what is wrong with a body. Nil means it is fine.
func CheckBody(k Kind, body string) []string {
	want, ok := Schema[k]
	if !ok {
		return nil
	}
	pos := map[string]int{}
	for i, s := range want {
		pos[s.Name] = i
	}
	var probs []string
	seen := map[string]bool{}
	firstLine, last, extra := true, -1, ""
	for _, ln := range strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n") {
		ln = strings.TrimRight(ln, " \t")
		if firstLine && ln != "" {
			firstLine = false
			if !strings.HasPrefix(ln, "# ") {
				probs = append(probs, "missing # title")
			}
		}
		if !strings.HasPrefix(ln, "## ") {
			continue
		}
		name := strings.TrimPrefix(ln, "## ")
		seen[name] = true
		i, known := pos[name]
		if !known {
			// An extra section is fine only after every schema one, so
			// remember it until a schema section shows up below it.
			extra = name
			continue
		}
		if extra != "" {
			probs = append(probs, "## "+extra+" is out of order")
			extra = ""
		}
		if i < last {
			probs = append(probs, "## "+name+" is out of order")
			continue
		}
		last = i
	}
	for _, s := range want {
		if s.Required && !seen[s.Name] {
			probs = append(probs, "missing ## "+s.Name)
		}
	}
	return probs
}
