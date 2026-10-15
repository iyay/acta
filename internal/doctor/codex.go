package doctor

import (
	"os"
	"path/filepath"
	"strings"
)

// codexHasPlugin says whether Codex has acta installed and turned on. Codex
// writes each installed plugin to config.toml as a [plugins."<name>@<market>"]
// table with an enabled line. Only that table's own lines count, so an enabled
// line under another plugin is not read as ours.
func codexHasPlugin(e Env) bool {
	if e.CodexDir == "" {
		return false
	}
	raw, err := os.ReadFile(filepath.Join(e.CodexDir, "config.toml"))
	if err != nil {
		return false
	}
	inActa := false
	for _, line := range strings.Split(string(raw), "\n") {
		// A user may edit the file by hand, so a trailing comment, Windows
		// line ends, tabs and either quote style must read the same as the
		// plain form Codex writes.
		line, _, _ = strings.Cut(line, "#")
		line = strings.Join(strings.Fields(line), "")
		line = strings.ReplaceAll(line, `'`, `"`)
		if strings.HasPrefix(line, "[") {
			inActa = strings.HasPrefix(line, `[plugins."acta@`)
			continue
		}
		if inActa && (line == "enabled=true" || line == `"enabled"=true`) {
			return true
		}
	}
	return false
}

// checkCodex reports acta in Codex. Codex is an extra harness, so a missing
// plugin only warns. Codex skips plugin hooks until the user trusts them, and
// nothing outside Codex shows whether they did, so the OK line says how.
func checkCodex(e Env) Result {
	r := Result{Name: "codex"}
	if codexHasPlugin(e) {
		r.Level, r.Msg = OK, "acta plugin enabled in Codex; trust its hooks once in /hooks"
		return r
	}
	r.Level, r.Msg = Warn, "Codex is installed but the acta plugin is not enabled in it"
	r.Fix = `install it: see the acta plugin README "## Codex"`
	return r
}
