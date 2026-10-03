package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// sendPlan is the plan of the demo repo: three first-round tasks, one fix round.
var sendPlan = strings.Replace(briefPlanHead+briefFixOne, "depth: minimal", "id: PLAN-1\nhash: aaaa\ndepth: minimal", 1)

const sendPlanPath = ".acta/plans/2026-10-03-demo.md"

// sendEnv is one test setup: a repo on branch dispatch-send, the scripted
// herdr, a rules file, and the herdr variables send needs.
type sendEnv struct {
	dir   string
	rules string
	h     *scriptedHerdr
}

func newSendEnv(t *testing.T) sendEnv { return newSendEnvOn(t, "main") }

// newSendEnvOn makes a main checkout on branch mainBranch and a worktree on
// branch dispatch-send next to it. The test runs in the worktree, like the
// real command does.
func newSendEnvOn(t *testing.T, mainBranch string) sendEnv {
	t.Helper()
	fastWaits(t, time.Second, time.Second, 0, time.Millisecond)
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(root, "main")
	dir := filepath.Join(root, "worktree")
	git := func(args ...string) {
		t.Helper()
		if out, err := exec.Command("git", append([]string{"-C", main}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	writeSendFile(t, filepath.Join(main, sendPlanPath), sendPlan)
	git("init", "-q", "-b", mainBranch)
	git("config", "user.name", "test")
	git("config", "user.email", "test@example.com")
	git("add", ".")
	git("commit", "-q", "-m", "init")
	git("worktree", "add", "-q", "-b", "dispatch-send", dir)
	sendGit(t, dir, "config", "user.name", "test")
	sendGit(t, dir, "config", "user.email", "test@example.com")
	rules := filepath.Join(t.TempDir(), "house-rules.md")
	writeSendFile(t, rules, "rules")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("HERDR_ENV", "1")
	h := newScriptedHerdr(t)
	return sendEnv{dir: dir, rules: rules, h: h}
}

func writeSendFile(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

// send runs `acta dispatch send` in the repo with the usual flags first.
func (e sendEnv) send(t *testing.T, stdin string, extra ...string) (int, string, string) {
	t.Helper()
	args := append([]string{"--plan", sendPlanPath, "--rules", e.rules}, extra...)
	return e.run(t, stdin, append([]string{"dispatch", "send"}, args...)...)
}

func (e sendEnv) run(t *testing.T, stdin string, args ...string) (int, string, string) {
	t.Helper()
	t.Chdir(e.dir)
	var stdout, stderr strings.Builder
	code := Run(args, strings.NewReader(stdin), false, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func (e sendEnv) head(t *testing.T) string {
	t.Helper()
	return dispatchGitOut(t, e.dir, "rev-parse", "HEAD")
}

func (e sendEnv) briefPath() string {
	return filepath.Join(e.dir, ".claude", "dispatch", "dispatch-send-brief.md")
}

func (e sendEnv) recordFile() string { return filepath.Join(e.dir, ".acta", ".dispatch.json") }

func (e sendEnv) readBrief(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(e.briefPath())
	if err != nil {
		t.Fatalf("no brief: %v", err)
	}
	return string(raw)
}

// herdrReady scripts a herdr that already knows the agent: found, idle, the
// goal mark shows on the first read, and the recent pane text is as given.
func (e sendEnv) herdrReady(recent string) {
	e.h.out("agent_get", agentJSON("wM:p5", "idle"))
	e.h.out("agent_read_visible", "> ")
	e.h.out("agent_read_visible", "🎯 Goal")
	e.h.out("agent_read_recent-unwrapped", recent)
}

const slugGet = "agent get dispatch-send"

func wantNothingSent(t *testing.T, e sendEnv) {
	t.Helper()
	if c := e.h.calls(); len(c) != 0 {
		t.Fatalf("herdr was called: %q", c)
	}
	for _, p := range []string{e.recordFile(), e.briefPath()} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Fatalf("%s was written: %v", p, err)
		}
	}
}

func TestDispatchSendFirstRoundNewTab(t *testing.T) {
	e := newSendEnv(t)
	e.h.fail("agent_get", notFoundJSON, "") // the slug is unknown
	e.h.out("agent_get", agentJSON("wM:p5", ""))
	e.h.out("agent_get", agentJSON("wM:p5", "idle"))
	e.h.out("tab_create", tabJSON("wM:p5"))
	e.h.out("agent_read_visible", "> ")
	e.h.out("agent_read_visible", "🎯 Goal")
	e.h.out("agent_read_recent-unwrapped", todoCard+"Task 1 first, Task 2 second, Task 3 third")

	code, stdout, stderr := e.send(t, "")
	if code != exitOK {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
	repo, _ := filepath.EvalSymlinks(e.dir)
	brief := filepath.Join(repo, ".claude", "dispatch", "dispatch-send-brief.md")
	want := strings.Join([]string{
		"slug: dispatch-send",
		"pane: wM:p5",
		"base: " + e.head(t),
		"brief: " + brief,
		"checkpoint: ok",
		"watcher: herdr agent wait dispatch-send --until idle --until done",
	}, "\n") + "\n"
	if stdout != want {
		t.Fatalf("stdout\n got: %q\nwant: %q", stdout, want)
	}
	goal := "/goal Demo Plan ultrathink orchestrate. Read the hand-off at " + brief +
		" first and follow every line. REPLY-BACK: after the last commit the build skill runs `acta reply-back`."
	e.h.wantCalls([]string{
		slugGet,
		"tab create --cwd " + repo + " --workspace wM --label dispatch-send --no-focus",
		"pane run wM:p5 omp",
		"agent get wM:p5",
		"agent rename wM:p5 dispatch-send",
		slugGet,
		"agent read dispatch-send --source visible --lines 10",
		"agent prompt dispatch-send " + goal,
		"agent read dispatch-send --source visible --lines 10",
		"agent read dispatch-send --source recent-unwrapped --lines 60",
	}, e.h.collapsed())

	rec := readRecord(t, e.dir)
	if rec.Pane != "wM:p9" || rec.Base != e.head(t) || rec.Plan != sendPlanPath || rec.Round != "dispatch-send" {
		t.Fatalf("record %+v", rec)
	}
	b := e.readBrief(t)
	for _, w := range []string{"plans/2026-10-03-demo#task-1", "plans/2026-10-03-demo#task-3", "parent main", "base " + e.head(t), e.rules} {
		if !strings.Contains(b, w) {
			t.Errorf("brief misses %q\n%s", w, b)
		}
	}
	if strings.Contains(b, "task-4") {
		t.Errorf("first round took the fix task\n%s", b)
	}
}

// The main checkout is not always on main. The brief must name the branch it
// is really on, because that is where the worktree gets merged.
func TestDispatchSendParentIsTheMainCheckoutBranch(t *testing.T) {
	e := newSendEnvOn(t, "master")
	e.herdrReady(todoCard + "Task 1, Task 2, Task 3")
	if code, _, stderr := e.send(t, ""); code != exitOK {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
	b := e.readBrief(t)
	if !strings.Contains(b, "parent master,") || strings.Contains(b, "parent main,") {
		t.Fatalf("brief does not name master as the parent\n%s", b)
	}
}

// A main checkout on a detached head has no branch to merge into: refuse
// before anything is sent.
func TestDispatchSendDetachedMainCheckoutExits3(t *testing.T) {
	e := newSendEnv(t)
	sendGit(t, filepath.Join(filepath.Dir(e.dir), "main"), "checkout", "-q", "--detach")
	code, _, stderr := e.send(t, "")
	if code != exitOther || !strings.Contains(stderr, "detached") {
		t.Fatalf("exit %d stderr %q", code, stderr)
	}
	wantNothingSent(t, e)
}

func TestDispatchSendFixRoundReusesTab(t *testing.T) {
	e := newSendEnv(t)
	e.herdrReady(todoCard + "working on Task 4 now")
	code, stdout, stderr := e.send(t, "", "--round", "r1")
	if code != exitOK {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
	if !strings.Contains(stdout, "checkpoint: ok\n") {
		t.Fatalf("stdout %q", stdout)
	}
	if e.h.count("tab create") != 0 || e.h.count("pane run") != 0 {
		t.Fatalf("a fix round opened a tab: %q", e.h.calls())
	}
	if rec := readRecord(t, e.dir); rec.Round != "r1" {
		t.Fatalf("record round %q, want r1", rec.Round)
	}
	b := e.readBrief(t)
	if !strings.Contains(b, "task-4") || strings.Contains(b, "task-1") {
		t.Fatalf("fix round brief wrong\n%s", b)
	}
}

func TestDispatchSendPolishNoteFromStdin(t *testing.T) {
	e := newSendEnv(t)
	// The demo plan gains a polish task, the way the review skill appends
	// one after a CLEAN round.
	writeSendFile(t, filepath.Join(e.dir, sendPlanPath), sendPlan+briefPolish)
	sendGit(t, e.dir, "commit", "-qam", "add polish task")
	e.herdrReady(todoCard + "working on Task 7 now")
	code, stdout, stderr := e.send(t, "[fix] rename the helper\n", "--round", "polish", "--note-file", "-")
	if code != exitOK {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
	if !strings.Contains(stdout, "checkpoint: ok\n") {
		t.Fatalf("stdout %q", stdout)
	}
	b := e.readBrief(t)
	if !strings.Contains(b, "NOTE: [fix] rename the helper") {
		t.Fatal("the note did not reach the brief")
	}
	// The checkpoint watches the polish task id, not an empty list.
	if !strings.Contains(b, "task-7") || strings.Contains(b, "task-1") {
		t.Fatalf("polish brief names the wrong tasks\n%s", b)
	}
	if got := e.h.count("agent read dispatch-send --source recent"); got != 1 {
		t.Fatalf("polish checkpoint read %d times, want 1: %q", got, e.h.calls())
	}
}

func TestDispatchSendNoteFromFile(t *testing.T) {
	e := newSendEnv(t)
	e.herdrReady(todoCard + "Task 4")
	note := filepath.Join(t.TempDir(), "note.md")
	writeSendFile(t, note, "mind the cache")
	if code, _, stderr := e.send(t, "", "--round", "r1", "--note-file", note); code != exitOK {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
	if !strings.Contains(e.readBrief(t), "NOTE: mind the cache") {
		t.Fatal("the note did not reach the brief")
	}
}

// Every refusal exits 1 with no herdr call, no record and no brief.
func TestDispatchSendRefusals(t *testing.T) {
	cases := []struct {
		name string
		mut  func(t *testing.T, e *sendEnv)
		args []string // replaces the default flags when set
		want string
	}{
		{"no HERDR_ENV", func(t *testing.T, e *sendEnv) { t.Setenv("HERDR_ENV", "") }, nil, "HERDR_ENV"},
		{"HERDR_ENV not 1", func(t *testing.T, e *sendEnv) { t.Setenv("HERDR_ENV", "0") }, nil, "HERDR_ENV"},
		{"no pane id", func(t *testing.T, e *sendEnv) { t.Setenv("HERDR_PANE_ID", "") }, nil, "HERDR_PANE_ID"},
		{"bad pane id", func(t *testing.T, e *sendEnv) { t.Setenv("HERDR_PANE_ID", "x y") }, nil, "HERDR_PANE_ID"},
		{"not a git repo", func(t *testing.T, e *sendEnv) {
			dir := t.TempDir()
			writeSendFile(t, filepath.Join(dir, sendPlanPath), sendPlan)
			e.dir = dir
		}, nil, "git"},
		{"no plan flag", nil, []string{"--rules", "RULES"}, "--plan"},
		{"plan missing", nil, []string{"--plan", ".acta/plans/nope.md", "--rules", "RULES"}, "plan"},
		{"plan absolute", nil, []string{"--plan", "/etc/passwd", "--rules", "RULES"}, "plan"},
		{"plan outside root", nil, []string{"--plan", "../x.md", "--rules", "RULES"}, "plan"},
		{"bad round", nil, []string{"--plan", sendPlanPath, "--rules", "RULES", "--round", "Bad Round"}, "round"},
		{"round without fix section", func(t *testing.T, e *sendEnv) {
			writeSendFile(t, filepath.Join(e.dir, sendPlanPath), strings.Replace(sendPlan, "## Fix round 1", "## Later", 1))
			sendGit(t, e.dir, "commit", "-qam", "no fix")
		}, []string{"--plan", sendPlanPath, "--rules", "RULES", "--round", "r1"}, "Fix round"},
		{"no rules flag", nil, []string{"--plan", sendPlanPath}, "--rules"},
		{"rules missing", nil, []string{"--plan", sendPlanPath, "--rules", "/nope/house-rules.md"}, "house rules"},
		{"rules relative", nil, []string{"--plan", sendPlanPath, "--rules", "house-rules.md"}, "absolute"},
		{"bad slug", func(t *testing.T, e *sendEnv) { sendGit(t, e.dir, "branch", "-m", "1-start") }, nil, "slug"},
		{"brief error: no verify", func(t *testing.T, e *sendEnv) {
			writeSendFile(t, filepath.Join(e.dir, sendPlanPath), strings.Replace(sendPlan, "**verify:** property two holds on every path", "no claim", 1))
		}, nil, "verify"},
		{"polish without note", nil, []string{"--plan", sendPlanPath, "--rules", "RULES", "--round", "polish"}, "note"},
		{"note file missing", nil, []string{"--plan", sendPlanPath, "--rules", "RULES", "--round", "polish", "--note-file", "/nope/note.md"}, "note"},
		{"unknown flag", nil, []string{"--bogus"}, "bogus"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := newSendEnv(t)
			if c.mut != nil {
				c.mut(t, &e)
			}
			var code int
			var stderr string
			if c.args == nil {
				code, _, stderr = e.send(t, "")
			} else {
				args := make([]string, len(c.args))
				for i, a := range c.args {
					args[i] = strings.Replace(a, "RULES", e.rules, 1)
				}
				code, _, stderr = e.run(t, "", append([]string{"dispatch", "send"}, args...)...)
			}
			if code != exitBadInput {
				t.Fatalf("exit %d, want %d; stderr %q", code, exitBadInput, stderr)
			}
			if !strings.Contains(stderr, c.want) {
				t.Fatalf("stderr %q misses %q", stderr, c.want)
			}
			wantNothingSent(t, e)
		})
	}
}

func sendGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
}

func TestDispatchSendDriftPrintsPaneAndExits4(t *testing.T) {
	e := newSendEnv(t)
	e.herdrReady(todoCard + "Task 1 and Task 2 only")
	code, stdout, stderr := e.send(t, "")
	if code != exitDrift {
		t.Fatalf("exit %d, want %d; stderr %q", code, exitDrift, stderr)
	}
	lines := strings.Split(stdout, "\n")
	if len(lines) < 7 || lines[4] != "checkpoint: drift: missing 3" ||
		lines[5] != "watcher: herdr agent wait dispatch-send --until idle --until done" {
		t.Fatalf("stdout %q", stdout)
	}
	if !strings.Contains(stdout, "Task 1 and Task 2 only") {
		t.Fatalf("the pane text is not printed: %q", stdout)
	}
}

func TestDispatchSendUnconfirmedExits0WithoutPaneText(t *testing.T) {
	e := newSendEnv(t)
	e.herdrReady("nothing on screen yet")
	code, stdout, stderr := e.send(t, "")
	if code != exitOK {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
	if !strings.Contains(stdout, "checkpoint: unconfirmed\n") || strings.Contains(stdout, "nothing on screen") {
		t.Fatalf("stdout %q", stdout)
	}
	if n := strings.Count(stdout, "\n"); n != 6 {
		t.Fatalf("want 6 lines, got %d: %q", n, stdout)
	}
}

// A herdr failure exits 2 with herdr's words. The record and the brief are
// already on disk by then, because nothing may be sent before they are.
func TestDispatchSendHerdrFailureExits2(t *testing.T) {
	e := newSendEnv(t)
	e.h.fail("agent_get", "", "socket exploded")
	code, stdout, stderr := e.send(t, "")
	if code != exitDelivery {
		t.Fatalf("exit %d, want %d; stderr %q", code, exitDelivery, stderr)
	}
	if !strings.Contains(stderr, "socket exploded") || stdout != "" {
		t.Fatalf("stdout %q stderr %q", stdout, stderr)
	}
	if _, err := os.Stat(e.recordFile()); err != nil {
		t.Fatalf("no record before the send: %v", err)
	}
	e.readBrief(t)
}

func TestDispatchSendDeliveryFailuresExit2(t *testing.T) {
	t.Run("agent still working", func(t *testing.T) {
		e := newSendEnv(t)
		e.h.out("agent_get", agentJSON("wM:p5", "working"))
		code, _, stderr := e.send(t, "")
		if code != exitDelivery || !strings.Contains(stderr, "still working") {
			t.Fatalf("exit %d stderr %q", code, stderr)
		}
		if e.h.count("agent prompt") != 0 {
			t.Fatalf("sent a goal: %q", e.h.calls())
		}
	})
	t.Run("goal mark never shows", func(t *testing.T) {
		e := newSendEnv(t)
		fastWaits(t, time.Second, 0, 0, time.Millisecond)
		e.h.out("agent_get", agentJSON("wM:p5", "idle"))
		e.h.out("agent_read_visible", "> ")
		code, _, stderr := e.send(t, "")
		if code != exitDelivery || !strings.Contains(stderr, "goal mark") {
			t.Fatalf("exit %d stderr %q", code, stderr)
		}
	})
}

func TestDispatchCloseFound(t *testing.T) {
	e := newSendEnv(t)
	e.h.out("agent_get", agentJSON("wM:p5", "done"))
	code, stdout, stderr := e.run(t, "", "dispatch", "close")
	if code != exitOK {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
	e.h.wantCalls([]string{slugGet, "pane close wM:p5"}, e.h.calls())
	if !strings.Contains(stdout, "wM:p5") {
		t.Fatalf("stdout %q", stdout)
	}
}

func TestDispatchCloseNotFoundRefuses(t *testing.T) {
	e := newSendEnv(t)
	e.h.fail("agent_get", notFoundJSON, "")
	code, _, stderr := e.run(t, "", "dispatch", "close")
	if code != exitBadInput || !strings.Contains(stderr, "dispatch-send") {
		t.Fatalf("exit %d stderr %q", code, stderr)
	}
	e.h.wantCalls([]string{slugGet}, e.h.calls())
}

func TestDispatchCloseHerdrFailuresExit2(t *testing.T) {
	t.Run("lookup breaks", func(t *testing.T) {
		e := newSendEnv(t)
		e.h.fail("agent_get", "", "socket exploded")
		code, _, stderr := e.run(t, "", "dispatch", "close")
		if code != exitDelivery || !strings.Contains(stderr, "socket exploded") {
			t.Fatalf("exit %d stderr %q", code, stderr)
		}
		if e.h.count("pane close") != 0 {
			t.Fatalf("closed a pane: %q", e.h.calls())
		}
	})
	t.Run("close breaks", func(t *testing.T) {
		e := newSendEnv(t)
		e.h.out("agent_get", agentJSON("wM:p5", "done"))
		e.h.fail("pane_close", "", "pane is busy")
		code, _, stderr := e.run(t, "", "dispatch", "close")
		if code != exitDelivery || !strings.Contains(stderr, "pane is busy") {
			t.Fatalf("exit %d stderr %q", code, stderr)
		}
	})
}

func TestDispatchCloseTakesNoFlags(t *testing.T) {
	e := newSendEnv(t)
	code, _, _ := e.run(t, "", "dispatch", "close", "--plan", "x")
	if code != exitBadInput || len(e.h.calls()) != 0 {
		t.Fatalf("exit %d calls %q", code, e.h.calls())
	}
}

func TestDispatchSlug(t *testing.T) {
	for in, want := range map[string]string{
		"dispatch-send":                  "dispatch-send",
		"feat/My_Thing":                  "feat-my-thing",
		strings.Repeat("a", 31) + "-bbb": strings.Repeat("a", 31),
		strings.Repeat("a", 40):          strings.Repeat("a", 32),
		"1-start":                        "",
		"":                               "",
	} {
		got, ok := dispatchSlug(in)
		if (want == "") == ok || got != want && ok {
			t.Errorf("dispatchSlug(%q) = %q, %v; want %q", in, got, ok, want)
		}
	}
}
