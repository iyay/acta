package plugin

import "strings"

// OutputStyle is the body of the acta output style, without its frontmatter.
// Codex has no output styles, so the session note carries this text there.
func OutputStyle() string {
	raw, err := Files.ReadFile("output-styles/acta.md")
	if err != nil {
		// The file is embedded at build time, so this only happens when it
		// was removed from the plugin. The note then goes without a style.
		return ""
	}
	text := string(raw)
	if rest, ok := strings.CutPrefix(text, "---\n"); ok {
		if _, body, found := strings.Cut(rest, "\n---\n"); found {
			text = body
		}
	}
	return strings.TrimSpace(text) + "\n"
}
