package plugincheck

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// evalCases ties each eval case to the one skill phrase it guards. When a skill
// loses that phrase the case stops testing anything, so the case files and the
// skill have to keep the same words.
var evalCases = []struct {
	folder string
	skill  string
	phrase string
}{
	{"side-idea-to-scratch", "scratch", "never"},
	{"note-to-scratch", "scratch", "note this"},
	{"second-brainstorm-choices", "shape", "One Architectural brainstorm per session"},
	{"one-file-fix-no-brainstorm", "shape", "Spike and Bounded"},
	{"brainstorm-files-scratch-first", "shape", "status brainstorming"},
	{"answers-appended", "shape", "acta scratch add"},
}

// TestEvalCases checks every case folder: it exists, it names the skill phrase
// it guards, and it caps its own quota.
func TestEvalCases(t *testing.T) {
	for _, c := range evalCases {
		t.Run(c.folder, func(t *testing.T) {
			dir := filepath.Join(pluginRoot(t), "evals", c.folder)
			if !exists(dir) {
				t.Fatalf("plugin/evals/%s is missing", c.folder)
			}
			text := caseText(t, dir)
			if !strings.Contains(text, "# guards: "+c.phrase) {
				t.Errorf("case does not say it guards %q", c.phrase)
			}
			if !strings.Contains(text, "max_turns:") || !strings.Contains(text, "timeout_seconds:") {
				t.Error("case must set max_turns and timeout_seconds, or one run eats the quota")
			}
			skill := readFile(t, "skills", c.skill, "SKILL.md")
			if !strings.Contains(skill, c.phrase) {
				t.Errorf("skills/%s/SKILL.md no longer says %q", c.skill, c.phrase)
			}
		})
	}
}

// TestEvalLLMGradersAreOnlyTheTwoJudgementCalls keeps the judge, which costs a
// model call, on the two cases that cannot be checked with files or commands.
func TestEvalLLMGradersAreOnlyTheTwoJudgementCalls(t *testing.T) {
	want := map[string]bool{"second-brainstorm-choices": true, "one-file-fix-no-brainstorm": true}
	for _, c := range evalCases {
		files, err := filepath.Glob(filepath.Join(pluginRoot(t), "evals", c.folder, "graders", "*.md"))
		if err != nil {
			t.Fatal(err)
		}
		if len(files) == 0 {
			t.Errorf("case %s has no graders", c.folder)
		}
		for _, f := range files {
			if strings.Contains(readFile(t, "evals", c.folder, "graders", filepath.Base(f)), "type: llm") != want[c.folder] {
				t.Errorf("case %s grader %s: an llm grader belongs only to the two judgement cases", c.folder, filepath.Base(f))
			}
		}
	}
}

// TestEvalScriptFlags keeps the run cheap and local. The user pays quota, so a
// cost ceiling would abort the suite and publishing would leak the report.
// --scaffold gives the four bash cases the scratch repo they work in, and
// --output-dir keeps the report out of the plugin tree, where a test here fails
// on the absolute user paths inside it.
func TestEvalScriptFlags(t *testing.T) {
	p := filepath.Join("..", "..", "scripts", "eval")
	st, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode()&0o111 == 0 {
		t.Error("scripts/eval must be executable")
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	txt := string(b)
	for _, want := range []string{"--model sonnet", "--ablation none", "--no-publish", "--allow-tools Bash", "--scaffold", "--output-dir", "--judge-model haiku"} {
		if !strings.Contains(txt, want) {
			t.Errorf("scripts/eval must pass %s", want)
		}
	}
	if strings.Contains(txt, "--max-cost-usd") {
		t.Error("scripts/eval must not set --max-cost-usd, the user is on a subscription")
	}
}

// caseText joins every file of one case folder, so a row can be checked without
// the test caring which file holds the phrase.
func caseText(t *testing.T, dir string) string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		t.Fatal(err)
	}
	files2, err := filepath.Glob(filepath.Join(dir, "graders", "*.md"))
	if err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for _, f := range append(files, files2...) {
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		b.Write(raw)
	}
	return b.String()
}

// TestScaffoldsRefuseNonEmptyDir runs every scaffold in a folder that already
// holds a file. A scaffold runs git init, git config and a commit of the whole
// folder, so a hand run inside a real repo would change that repo. It must
// stop first and leave the folder as it was.
func TestScaffoldsRefuseNonEmptyDir(t *testing.T) {
	scaffolds, err := filepath.Glob(filepath.Join(pluginRoot(t), "evals", "*", "scaffold.sh"))
	if err != nil || len(scaffolds) == 0 {
		t.Fatalf("no scaffolds found: %v", err)
	}
	for _, s := range scaffolds {
		t.Run(filepath.Base(filepath.Dir(s)), func(t *testing.T) {
			dir := t.TempDir()
			keep := filepath.Join(dir, "keep.txt")
			if err := os.WriteFile(keep, []byte("mine\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("bash", s)
			cmd.Dir = dir
			cmd.Env = []string{"PATH=/usr/bin:/bin", "HOME=" + t.TempDir()}
			out, err := cmd.CombinedOutput()
			if err == nil {
				t.Fatalf("scaffold ran in a non-empty folder:\n%s", out)
			}
			entries, _ := os.ReadDir(dir)
			if len(entries) != 1 || entries[0].Name() != "keep.txt" {
				t.Fatalf("scaffold changed the folder: %v", entries)
			}
		})
	}
}
