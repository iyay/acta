package write

import "strings"

// BugTemplate is what a new bug file starts as before the user fills it in.
func BugTemplate(title, ref string) []byte {
	return withRef("# "+title+"\n\n## Symptom\n\n## Root cause\n\n## Repro\n\n## Found in\n", ref)
}

// BugFile builds a bug file from a body an agent sent. A leading "# " line in
// the body is used as the title when no title was given, and is dropped.
func BugFile(title, slug, ref string, body []byte) []byte {
	text := strings.TrimLeft(string(body), "\n")
	if strings.HasPrefix(text, "# ") {
		first, rest, _ := strings.Cut(text, "\n")
		if title == "" {
			title = strings.TrimSpace(first[2:])
		}
		text = strings.TrimLeft(rest, "\n")
	}
	if title == "" {
		title = strings.ReplaceAll(slug, "-", " ")
	}
	return withRef("# "+title+"\n\n"+text, ref)
}

// HasSymptom says whether content has a "## Symptom" heading, the one part a
// bug file must have.
func HasSymptom(content []byte) bool {
	for _, ln := range strings.Split(string(content), "\n") {
		if strings.TrimRight(ln, " \t\r") == "## Symptom" {
			return true
		}
	}
	return false
}

func withRef(content, ref string) []byte {
	if ref == "" {
		return []byte(content)
	}
	out, err := SetField([]byte(content), "ref", ref)
	if err != nil {
		// SetField cannot fail on a file with no frontmatter.
		panic(err)
	}
	return out
}
