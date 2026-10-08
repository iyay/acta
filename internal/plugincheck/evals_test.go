package plugincheck

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/evalomp"
)

// evalCases ties each eval case to the one phrase it guards. The file is a path
// inside the plugin folder: a skill, the output style, or the hooks file. When
// that file loses the phrase the case stops testing anything, so the case files
// and that file have to keep the same words.
var evalCases = []struct {
	folder string
	file   string
	phrase string
}{
	{"side-idea-to-scratch", "skills/scratch/SKILL.md", "never"},
	{"note-to-scratch", "skills/scratch/SKILL.md", "note this"},
	{"second-brainstorm-choices", "skills/shape/SKILL.md", "One Architectural brainstorm per session"},
	{"one-file-fix-no-brainstorm", "skills/shape/SKILL.md", "Spike and Bounded"},
	{"brainstorm-files-scratch-first", "skills/shape/SKILL.md", "status brainstorming"},
	{"answers-appended", "skills/shape/SKILL.md", "acta scratch add"},
	{"frame-no-brainstorming-status", "skills/frame/SKILL.md", "the `brainstorming` status"},
	{"probe-round", "skills/shape/probe.md", "five questions at most"},
	{"style-short-answer", "output-styles/acta.md", "Open with the answer"},
	{"wiki-hint", "hooks/hooks.json", "Bash|PowerShell|Read|Edit|Write|MultiEdit"},
	{"wiki-close", "skills/build/SKILL.md", "acta wiki check <parent>..HEAD"},
	{"state-resume", "skills/build/SKILL.md", "acta state set plans/<stem> <part>"},
	{"repo-override-ask", "skills/setup/SKILL.md", "is committed and shared"},
	{"polish-light-review", "skills/review/SKILL.md", "Light review: the orchestrator reads the full diff"},
	{"planning-commit-folds", "skills/shape/SKILL.md", "acta commit <path> -m \"<message>\""},
	{"polish-full-review", "skills/review/SKILL.md", "## Review tiers"},
}

// TestEvalCases checks every case folder: it exists, it names the phrase it
// guards, and it caps its own quota.
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
			guarded := readFile(t, c.file)
			if !strings.Contains(guarded, c.phrase) {
				t.Errorf("%s no longer says %q", c.file, c.phrase)
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

// TestEvalsHoldNoRoutingCases keeps the default run cheap. The routing set is
// a baseline measure that lives in plugin/evals-routing/, so a default
// scripts/eval, which runs plugin/evals/ with no filter, never sees it.
func TestEvalsHoldNoRoutingCases(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join(pluginRoot(t), "evals", "routing-*"))
	if err != nil {
		t.Fatal(err)
	}
	var dirs []string
	for _, m := range matches {
		st, err := os.Stat(m)
		if err != nil {
			t.Fatal(err)
		}
		if st.IsDir() {
			dirs = append(dirs, filepath.Base(m))
		}
	}
	if len(dirs) != 0 {
		t.Errorf("plugin/evals/ holds routing cases %v; they live in plugin/evals-routing/", dirs)
	}
}

// TestEvalScriptPassesArgsThrough runs scripts/eval under /bin/bash with stub
// claude and go binaries that only log arguments. The script adds no case
// list of its own: with no arguments the runner covers plugin/evals/ as is,
// and the routing baseline runs with --eval-dir evals-routing.
func TestEvalScriptPassesArgsThrough(t *testing.T) {
	script, err := filepath.Abs(filepath.Join("..", "..", "scripts", "eval"))
	if err != nil {
		t.Fatal(err)
	}
	origPath := os.Getenv("PATH")
	// runEval runs scripts/eval with /bin/bash and stub claude and go
	// binaries that only log their arguments, and returns those logged
	// arguments, one per line.
	runEval := func(args ...string) []string {
		bin := t.TempDir()
		log := filepath.Join(t.TempDir(), "argv.log")
		stub := "#!/bin/sh\nprintf '%s\\n' \"$@\" >> \"$EVAL_STUB_LOG\"\n"
		for _, name := range []string{"claude", "go"} {
			if err := os.WriteFile(filepath.Join(bin, name), []byte(stub), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		t.Setenv("EVAL_STUB_LOG", log)
		t.Setenv("PATH", bin+string(os.PathListSeparator)+origPath)
		t.Setenv("HOME", t.TempDir())
		t.Setenv("TMPDIR", t.TempDir())
		cmd := exec.Command("/bin/bash", append([]string{script}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("/bin/bash scripts/eval %v: %v\n%s", args, err, out)
		}
		raw, err := os.ReadFile(log)
		if err != nil {
			t.Fatalf("stub logged nothing for args %v: %v", args, err)
		}
		return strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
	}
	// evalDir pulls the --eval-dir value out of logged arguments.
	evalDir := func(argv []string) (string, bool) {
		for i, a := range argv {
			if v, ok := strings.CutPrefix(a, "--eval-dir="); ok {
				return v, true
			}
			if a == "--eval-dir" && i+1 < len(argv) {
				return argv[i+1], true
			}
		}
		return "", false
	}
	// caseArgs pulls every --case value out of logged arguments.
	caseArgs := func(argv []string) []string {
		var out []string
		for i, a := range argv {
			if v, ok := strings.CutPrefix(a, "--case="); ok {
				out = append(out, v)
			} else if a == "--case" && i+1 < len(argv) {
				out = append(out, argv[i+1])
			}
		}
		return out
	}
	// No args on the Claude path: no case list added.
	if globs := caseArgs(runEval()); len(globs) != 0 {
		t.Errorf("default run gained a case list: %v", globs)
	}
	// --eval-dir evals-routing passes through untouched, no --case added.
	argv := runEval("--eval-dir", "evals-routing")
	if dir, ok := evalDir(argv); !ok || dir != "evals-routing" {
		t.Errorf("--eval-dir evals-routing became %v", argv)
	}
	if globs := caseArgs(argv); len(globs) != 0 {
		t.Errorf("--eval-dir run gained a case list: %v", globs)
	}
	// --eval-dir=evals-routing passes through untouched, no --case added.
	argv = runEval("--eval-dir=evals-routing")
	if dir, ok := evalDir(argv); !ok || dir != "evals-routing" {
		t.Errorf("--eval-dir=evals-routing became %v", argv)
	}
	if globs := caseArgs(argv); len(globs) != 0 {
		t.Errorf("--eval-dir= run gained a case list: %v", globs)
	}
	// --case x passes through with no added list.
	argv = runEval("--case", "x")
	if globs := caseArgs(argv); len(globs) != 1 || globs[0] != "x" {
		t.Errorf("--case x became %v, want only [x]", globs)
	}
	// --omp with no args routes to acta eval-omp with no case list.
	argv = runEval("--omp")
	found := false
	for _, a := range argv {
		if a == "eval-omp" {
			found = true
		}
	}
	if !found {
		t.Errorf("--omp run never reached acta eval-omp: %v", argv)
	}
	if globs := caseArgs(argv); len(globs) != 0 {
		t.Errorf("--omp run gained a case list: %v", globs)
	}
	// --omp --eval-dir evals-routing reaches eval-omp untouched, no --case.
	argv = runEval("--omp", "--eval-dir", "evals-routing")
	if dir, ok := evalDir(argv); !ok || dir != "evals-routing" {
		t.Errorf("--omp --eval-dir evals-routing became %v", argv)
	}
	if globs := caseArgs(argv); len(globs) != 0 {
		t.Errorf("--omp --eval-dir run gained a case list: %v", globs)
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
// stop first and leave the folder as it was. Both eval dirs count: the
// routing baseline lives in plugin/evals-routing/.
func TestScaffoldsRefuseNonEmptyDir(t *testing.T) {
	var scaffolds []string
	for _, dir := range []string{"evals", "evals-routing"} {
		found, err := filepath.Glob(filepath.Join(pluginRoot(t), dir, "*", "scaffold.sh"))
		if err != nil {
			t.Fatal(err)
		}
		scaffolds = append(scaffolds, found...)
	}
	if len(scaffolds) == 0 {
		t.Fatal("no scaffolds found")
	}
	for _, s := range scaffolds {
		t.Run(filepath.Join(filepath.Base(filepath.Dir(filepath.Dir(s))), filepath.Base(filepath.Dir(s))), func(t *testing.T) {
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

// The omp runner reads these patterns as JavaScript regex, so each must compile there.
func TestEvalGraderPatternsCompileInOmp(t *testing.T) {
	cases, err := evalomp.LoadCases(filepath.Join(pluginRoot(t), "evals"))
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range evalomp.PatternProblems(cases) {
		t.Error(p)
	}
}
