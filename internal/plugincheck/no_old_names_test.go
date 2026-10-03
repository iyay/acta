package plugincheck

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/hook"
)

// oldSkillRe matches the old skill names: the eleven pm: names, and the two
// acta names that became shape and slice. A bare "pm" with no colon is prose,
// and so is the word brainstorm or plan alone, so only the prefixed name fails.
var oldSkillRe = regexp.MustCompile(`pm:(brainstorm|plan|build|tdd|debug|review|land|bug|dispatch|setup|migrate)\b|acta:(brainstorm|plan)\b`)

// oldPmbRe matches the old command as its own word: "pmb tick", "`pmb`" or
// "./cmd/pmb". It skips "pm-board", "PMB-EXEC-OK" and "PM_VOICE_FILE",
// which are other names.
var oldPmbRe = regexp.MustCompile(`\bpmb\b`)

// oldRootRe matches the old root folder. ".pm.yaml" carries no slash and
// moved under its own rename, file by file.
var oldRootRe = regexp.MustCompile(`\.pm/`)

// oldVoiceRe matches the old config command. The Go package and the
// PM_VOICE_FILE variable keep the voice name, so only "acta voice" fails.
var oldVoiceRe = regexp.MustCompile(`\bacta voice\b`)

// aliasLines is the one old text the rename keeps: the README sentence that
// tells users pmb still runs as an alias. Paths read as walkPlugin reports
// them, values are 1-based line numbers, so a second old name anywhere else
// still fails.
var aliasLines = map[string][]int{
	"README.md": {20},
}

// oldNameProblems names every old name on lines the alias list does not skip.
func oldNameProblems(source, text string) []string {
	allowed := map[int]bool{}
	for _, n := range aliasLines[source] {
		allowed[n] = true
	}
	checks := []struct {
		name string
		re   *regexp.Regexp
	}{
		{"old skill name", oldSkillRe},
		{"old pmb command", oldPmbRe},
		{"old .pm/ root", oldRootRe},
		{"old acta voice command", oldVoiceRe},
	}
	var out []string
	for i, line := range strings.Split(text, "\n") {
		if allowed[i+1] {
			continue
		}
		for _, c := range checks {
			if hit := c.re.FindString(line); hit != "" {
				out = append(out, c.name+" "+hit)
			}
		}
	}
	return out
}

// TestNoOldNames fails when the old names come back: a pm: skill, a bare
// pmb command, or .pm/ as the root in any plugin file, in the text the hook
// really prints, or in the manifest and package names.
func TestNoOldNames(t *testing.T) {
	walkPlugin(t, func(rel, text string) {
		for _, p := range oldNameProblems(rel, text) {
			t.Errorf("%s: %s", rel, p)
		}
	})
	// The allowlist must still point at the alias sentence. Without this a
	// rewrite that deletes the sentence would pass on nothing.
	for rel, lines := range aliasLines {
		all := strings.Split(readFile(t, rel), "\n")
		for _, n := range lines {
			if n < 1 || n > len(all) || !strings.Contains(all[n-1], "pmb") {
				t.Errorf("alias allowlist %s:%d points at no pmb line", rel, n)
			}
		}
	}
	// The same check runs over what the session really shows, so the shown
	// text cannot drift back while the files stay clean.
	set := hook.Input{Voice: config.User{ChatLanguage: "Korean", Style: "adhd", RepoLanguage: "English"}, VoiceExists: true}
	shown := map[string]string{
		"first run":      hook.SessionStart(hook.Input{Voice: config.UserDefault()}),
		"voice set":      hook.SessionStart(set),
		"broken voice":   hook.SessionStart(hook.Input{Voice: config.UserDefault(), VoiceExists: true, VoiceErr: errors.New("x")}),
		"conflicts":      hook.SessionStart(hook.Input{Voice: config.UserDefault(), VoiceExists: true, Conflicts: []string{"superpowers@x"}}),
		"prompt set":     hook.Prompt(set),
		"prompt missing": hook.Prompt(hook.Input{Voice: config.UserDefault()}),
		"prompt broken":  hook.Prompt(hook.Input{Voice: config.UserDefault(), VoiceExists: true, VoiceErr: errors.New("x")}),
	}
	for name, text := range shown {
		for _, p := range oldNameProblems("hook "+name, text) {
			t.Errorf("hook %s: %s", name, p)
		}
	}
	// The plugin and the package carry the new name in their manifests.
	var pluginJSON struct{ Name string }
	if err := json.Unmarshal([]byte(readFile(t, ".claude-plugin", "plugin.json")), &pluginJSON); err != nil {
		t.Fatal(err)
	}
	if pluginJSON.Name != "acta" {
		t.Errorf("plugin.json name = %q, want acta", pluginJSON.Name)
	}
	var market struct {
		Name    string
		Plugins []struct{ Name string }
	}
	if err := json.Unmarshal([]byte(readFile(t, ".claude-plugin", "marketplace.json")), &market); err != nil {
		t.Fatal(err)
	}
	if market.Name != "acta-local" || len(market.Plugins) != 1 || market.Plugins[0].Name != "acta" {
		t.Errorf("marketplace.json = %+v, want acta-local holding acta", market)
	}
	var pkg struct{ Name string }
	if err := json.Unmarshal([]byte(readFile(t, "package.json")), &pkg); err != nil {
		t.Fatal(err)
	}
	if pkg.Name != "acta" {
		t.Errorf("package.json name = %q, want acta", pkg.Name)
	}
}
