package plugincheck

import (
	"strings"
	"testing"
)

// TestSkillDebug pins the evidence contract of the debug skill. Text checks
// cannot prove behaviour; the debug-* eval cases do that. These only keep the
// four phases, the safety rules and the recipe triggers from being cut.
func TestSkillDebug(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name: "debug",
		// race.md holds the copyable race probe, which made the folder longer.
		// The known-flake rerun rule in Phase 1 adds two more lines.
		// The required-proof steps, the recent-changes step and the hang recipe
		// added lines to SKILL.md; private probe checks extend the recipes.
		MaxLines: 360,
		Must: []string{
			// The four phases and their exit evidence.
			"Symptom and feedback loop", "Narrow and distinguish", "Prove the mechanism", "Fix, verify and clean up",
			"exact symptom", "not proven", "observed", "unverified",
			// Safety and the acta gates.
			"outside the repo", "Redact secrets", "never push unless the user asks", "live systems",
			"acta:bug", "acta:shape", "acta:tdd", "acta:land", "Three failed fixes",
			// Recipes, each with the evidence that matches it.
			"force the interleaving", "smallest command that keeps the failing order",
			"count the work", "unmeasured", "failure count",
			"root-cause-tracing.md", "condition-based-waiting.md", "defense-in-depth.md",
			"trust boundar",
		},
		MustNot: []string{
			"superpowers:", "CREATION-LOG", "test-pressure", "test-academic", "git-bug",
			"add diagnostic instrumentation", "SMALLEST possible change", "When you can't trace manually, add instrumentation:",
			// Old advice the design replaced: blanket checks, made-up numbers,
			// and printing secrets to find a bug.
			"Validate at EVERY layer", "Add validation at each layer", "Bug impossible",
			"95%", "1847 tests", "60% → 100%", "40% faster",
			"security find-identity", "env | grep",
		},
	})
}

// TestDebugRerunsKnownFlake reads SKILL.md on its own. A test that an open
// bug already marks as flaky gets one lone rerun before any root-cause work,
// so a known flake is logged instead of debugged again. The rule has to sit
// in Phase 1, before the narrowing starts. It covers only a test met while
// doing other work: a test the current bug is about still gets debugged, or
// a tracked flake could never be fixed. Debug can run with no plan, so the
// plan id is only passed when there is one.
func TestDebugRerunsKnownFlake(t *testing.T) {
	txt := readSkill(t, "skills/debug/SKILL.md")
	for _, want := range []string{
		"While you work on something else, a failing test that an open bug lists under `flaky_tests`",
		"acta list --type bug --json",
		"runs once more alone first",
		"a pass runs `acta flake record <name>`, plus `--plan <plan id>` when there is a plan, and needs no debugging",
		"A test the current bug or task is fixing is not covered",
	} {
		if !strings.Contains(txt, want) {
			t.Errorf("debug/SKILL.md missing %q", want)
		}
	}
	if strings.Contains(txt, "a pass runs `acta flake record <name> --plan <plan id>`") {
		t.Error("debug/SKILL.md still asks for a plan id that a debug with no plan does not have")
	}
	rule, phase2 := strings.Index(txt, "runs once more alone first"), strings.Index(txt, "## Phase 2")
	if rule < 0 || phase2 < 0 || rule > phase2 {
		t.Errorf("debug/SKILL.md must give the flake rerun before Phase 2 (rule %d, Phase 2 %d)", rule, phase2)
	}
}

// TestRaceProbeStartsFresh reads race.md on its own. A probe that reuses one
// state file adds to the last run's count, so a rerun or the lock control can
// print the right number on code that is still racy.
func TestRaceProbeStartsFresh(t *testing.T) {
	txt := readSkill(t, "skills/debug/race.md")
	if !strings.Contains(txt, "tempfile.mkdtemp()") {
		t.Error("debug/race.md must give each run a fresh folder with tempfile.mkdtemp()")
	}
	if strings.Contains(txt, "/tmp/probe") {
		t.Error("debug/race.md still uses a fixed /tmp/probe path")
	}
}

// TestDebugBisectRunsInATempClone keeps two debug rules from clashing. Phases
// 1 to 3 change no file in the repo, and bisect checks out other commits in
// the working tree, so every line that names bisect must send it to a clone
// and show the clone command to copy.
func TestDebugBisectRunsInATempClone(t *testing.T) {
	txt := readSkill(t, "skills/debug/SKILL.md")
	for _, want := range []string{
		"`git bisect` and a checkout of an older commit run only in a temp clone",
		`git clone "$source_root" history && cd history && pwd -P && git bisect start`,
		"never in the repo's own working tree",
		`cd "$probe_root/history" && pwd -P && git bisect good <sha>`,
	} {
		if !strings.Contains(txt, want) {
			t.Errorf("debug/SKILL.md missing %q", want)
		}
	}
	for _, line := range strings.Split(txt, "\n") {
		if strings.Contains(line, "git bisect") && !strings.Contains(line, "temp clone") {
			t.Errorf("debug/SKILL.md names git bisect with no temp clone on the same line: %q", line)
		}
	}
}

// TestDebugProbeFolderIsChecked keeps a probe from writing to the repo. An
// agent once chained mktemp, cd and sed, the first step failed, and sed
// changed a tracked file. The ground rules show one pattern to copy: make the
// folder, change into it, and read pwd before any write. No line may offer
// less work when a step fails; that reads as leave to skip the check.
func TestDebugProbeFolderIsChecked(t *testing.T) {
	txt := readSkill(t, "skills/debug/SKILL.md")
	for _, want := range []string{
		// The write rides on the same line as the check: a harness may start
		// every shell call in the repo again, so a cd in an earlier call
		// proves nothing about the next one.
		"Every shell call may start in the repo again",
		"on the same line",
		"Join the write with `&&`, never with `;` or on a new line: only then does a step that fails stop the line before the write.",
		// The old words, kept on purpose: without "temp copy" here agents ran
		// the user's own repro command in the repo (debug-race, 37 of 40).
		"A command that writes files runs only in a temp copy.",
		// Words an agent acts on; cutting them cost controls before.
		"(in the temp copy)",
	} {
		if !strings.Contains(txt, want) {
			t.Errorf("debug/SKILL.md missing %q", want)
		}
	}
	trace := readSkill(t, "skills/debug/root-cause-tracing.md")
	if !strings.Contains(trace, "in the checked temp folder the ground rules of `SKILL.md` show") {
		t.Error("root-cause-tracing.md must point its temp copy at the checked folder of the ground rules")
	}
	if race := readSkill(t, "skills/debug/race.md"); !strings.Contains(race, "on a copy of the code in the checked temp folder the ground rules of `SKILL.md` show") {
		t.Error("race.md must run its probe on a copy in the checked folder of the ground rules")
	}
	// A bare cd that a later call relies on is the trap, so the old two-call
	// pattern must not come back.
	if strings.Contains(txt, "Go on only when `pwd` prints that folder") {
		t.Error("debug/SKILL.md still checks pwd in one call and writes in the next")
	}
	for _, bad := range []string{"No copy?", "if the copy fails", "when the folder cannot be made"} {
		if strings.Contains(txt, bad) || strings.Contains(trace, bad) {
			t.Errorf("the debug skill holds a fallback line %q", bad)
		}
	}
}

// TestDebugProofIsRequired keeps the strongest checks of Phase 3 from being
// optional. A cause is stated only after a probe that could have shown it
// wrong, the control runs or the report says why it could not, and every
// report has the same four lines, so the command that reproduces the bug is
// never left out of the reply.
func TestDebugProofIsRequired(t *testing.T) {
	txt := readSkill(t, "skills/debug/SKILL.md")
	for _, want := range []string{
		"Before you state the cause, write down the result that would show it wrong, then run that probe.",
		"Run a control: the same probe with the suspected cause removed (in the temp copy) makes the symptom go away.",
		"A control that cannot run: say why under Unverified.",
		"Report in four lines, whether the cause is proven or not",
		"**Cause:**", "**Repro:**", "**Evidence:**", "**Unverified:**",
		// The exit that stops in Phase 1 uses the same four lines.
		"Report it in the four lines of Phase 3, with Cause \"not proven\" and Repro \"none yet\" plus what you tried.",
		"a command the user can paste",
		"or \"not proven\" plus what you ruled out",
	} {
		if !strings.Contains(txt, want) {
			t.Errorf("debug/SKILL.md missing %q", want)
		}
	}
	if strings.Contains(txt, "Use a control when you can") {
		t.Error("debug/SKILL.md still makes the control optional")
	}
}

// TestDebugRecentChangesAndOtherCallers keeps two cheap steps in the skill.
// What changed lately is the first thing to look at, since most new failures
// follow a change. And a proven cause is searched for in the other callers,
// so the report says where else the same defect sits. Both steps only read.
func TestDebugRecentChangesAndOtherCallers(t *testing.T) {
	txt := readSkill(t, "skills/debug/SKILL.md")
	for _, want := range []string{
		"Start with what changed lately: `git log` on the files involved, dependency versions, and the environment",
		"A change that lines up with the first failure is the first thing to probe.",
		"Once the cause is proven, search the other callers for the same pattern",
		"or \"none found\" with the search you ran",
		"after the four lines of the report",
		"Only read them: a fix for another caller is its own plan.",
	} {
		if !strings.Contains(txt, want) {
			t.Errorf("debug/SKILL.md missing %q", want)
		}
	}
	phase2, phase3 := strings.Index(txt, "## Phase 2"), strings.Index(txt, "## Phase 3")
	if i := strings.Index(txt, "Start with what changed lately"); i < phase2 || i > phase3 {
		t.Error("the changed-lately step must sit in Phase 2")
	}
	if i := strings.Index(txt, "Once the cause is proven, search the other callers"); i < phase3 {
		t.Error("the other-callers step must sit in Phase 3")
	}
}

// TestDebugHangRecipe keeps a recipe for every symptom the description
// names. A hang is proven by a stack dump that shows who waits on whom, never
// by a longer timeout, which only hides it. "Three failed fixes" is a rule
// about fixes, so it sits in Phase 4, the only phase that writes one.
func TestDebugHangRecipe(t *testing.T) {
	txt := readSkill(t, "skills/debug/SKILL.md")
	for _, want := range []string{
		"- **Hang.**",
		"a stack dump of the stuck process",
		"who waits on whom",
		"A longer timeout is not evidence",
		"Two dumps a few seconds apart tell a hang from slow work.",
		"set up before the hang",
		"macOS `sample <pid>`",
		"hang, slow",
	} {
		if !strings.Contains(txt, want) {
			t.Errorf("debug/SKILL.md missing %q", want)
		}
	}
	phase3, phase4 := strings.Index(txt, "## Phase 3"), strings.Index(txt, "## Phase 4")
	recipes := strings.Index(txt, "## Recipes")
	i := strings.Index(txt, "Three failed fixes")
	if i < phase4 || i > recipes {
		t.Errorf("Three failed fixes must sit in Phase 4 (at %d, phase 3 %d, phase 4 %d, recipes %d)", i, phase3, phase4, recipes)
	}
	if strings.Count(txt, "Three failed fixes") != 1 {
		t.Error("Three failed fixes must be said once")
	}
}

// Controls copied from another checkout cannot prove the reported defect.
func TestDebugControlSource(t *testing.T) {
	txt := readSkill(t, "skills/debug/SKILL.md")
	for _, want := range []string{"source_root", "absolute root of the checkout under investigation", "before changing directories", `cp -R "$source_root/<what the probe needs>" .`, "Record that absolute path", "Each later shell call"} {
		if !strings.Contains(txt, want) {
			t.Errorf("debug control source missing %q", want)
		}
	}
	if strings.Contains(txt, "$OLDPWD") {
		t.Error("control source still depends on shell history")
	}
}

// Every probe writer uses fresh private storage, including incidental outputs.
func TestDebugProbeIsolation(t *testing.T) {
	txt := readSkill(t, "skills/debug/SKILL.md")
	for _, want := range []string{"umask 077", "exclusive", "never `mkdir -p`", "test ! -L", "unset TMPDIR", "cache", "polluter", "profile", "File-tool writes", "verified absolute", "metadata-only exceptions", "Before the first repro", "Treat tests, imports and profilers as writers", "Never append a repro", "Any probe of another caller"} {
		if !strings.Contains(txt, want) {
			t.Errorf("debug isolation missing %q", want)
		}
	}
	for _, name := range []string{"root-cause-tracing.md", "race.md"} {
		if !strings.Contains(readSkill(t, "skills/debug/"+name), "private probe") {
			t.Errorf("%s omits private probe boundary", name)
		}
	}
}
