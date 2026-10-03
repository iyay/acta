package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// scriptedHerdr is a stand-in herdr on PATH. It logs every call, one line per
// call, and answers from numbered reply files, so a test can say "the first
// agent get fails, the second one works". When the replies run out, the last
// one repeats. Reads are keyed by their --source, so a visible read and a
// checkpoint read do not share replies.
type scriptedHerdr struct {
	t    *testing.T
	dir  string
	sent map[string]int
}

const scriptedHerdrScript = `#!/bin/sh
d="$HERDR_FAKE_DIR"
printf '%s\n' "$*" >> "$d/calls.log"
key="$1_$2"
if [ "$1 $2" = "agent read" ]; then key="agent_read_$5"; fi
n=0
if [ -f "$d/$key.count" ]; then n=$(cat "$d/$key.count"); fi
n=$((n+1))
echo $n > "$d/$key.count"
while [ ! -f "$d/$key.$n.out" ] && [ $n -gt 1 ]; do n=$((n-1)); done
if [ -f "$d/$key.$n.err" ]; then cat "$d/$key.$n.err" >&2; fi
if [ -f "$d/$key.$n.out" ]; then cat "$d/$key.$n.out"; fi
code=0
if [ -f "$d/$key.$n.exit" ]; then code=$(cat "$d/$key.$n.exit"); fi
exit $code
`

func newScriptedHerdr(t *testing.T) *scriptedHerdr {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "herdr"), []byte(scriptedHerdrScript), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("HERDR_FAKE_DIR", dir)
	t.Setenv("HERDR_PANE_ID", "wM:p9")
	return &scriptedHerdr{t: t, dir: dir, sent: map[string]int{}}
}

func (h *scriptedHerdr) write(key, ext, text string) {
	h.t.Helper()
	name := fmt.Sprintf("%s.%d.%s", key, h.sent[key], ext)
	if err := os.WriteFile(filepath.Join(h.dir, name), []byte(text), 0o644); err != nil {
		h.t.Fatal(err)
	}
}

// out queues the next reply for key (for example "agent_get"): stdout, exit 0.
func (h *scriptedHerdr) out(key, stdout string) {
	h.t.Helper()
	h.sent[key]++
	h.write(key, "out", stdout)
}

// fail queues the next reply for key as a non-zero exit with the given text.
func (h *scriptedHerdr) fail(key, stdout, stderr string) {
	h.t.Helper()
	h.sent[key]++
	h.write(key, "out", stdout)
	h.write(key, "err", stderr)
	h.write(key, "exit", "1")
}

// calls returns every call in order, each as its arguments joined by spaces.
func (h *scriptedHerdr) calls() []string {
	raw, err := os.ReadFile(filepath.Join(h.dir, "calls.log"))
	if err != nil {
		return nil
	}
	return strings.Split(strings.TrimRight(string(raw), "\n"), "\n")
}

// collapsed drops a call that repeats the one before it, so a test about
// order does not depend on how many times a wait polled.
func (h *scriptedHerdr) collapsed() []string {
	var out []string
	for _, c := range h.calls() {
		if len(out) == 0 || out[len(out)-1] != c {
			out = append(out, c)
		}
	}
	return out
}

func (h *scriptedHerdr) count(prefix string) int {
	n := 0
	for _, c := range h.calls() {
		if strings.HasPrefix(c, prefix) {
			n++
		}
	}
	return n
}

// wantCalls compares the full call list and also proves /new was never sent.
func (h *scriptedHerdr) wantCalls(want []string, got []string) {
	h.t.Helper()
	if !reflect.DeepEqual(got, want) {
		h.t.Fatalf("herdr calls\n got: %q\nwant: %q", got, want)
	}
	for _, c := range h.calls() {
		if strings.Contains(c, "/new") {
			h.t.Fatalf("sent /new: %q", c)
		}
	}
}

// fastWaits sets the waits for one test and puts the old ones back after it.
func fastWaits(t *testing.T, ready, goal, delay, poll time.Duration) {
	t.Helper()
	r, g, d, p := herdrReadyWait, herdrGoalWait, herdrCheckpointDelay, herdrPoll
	herdrReadyWait, herdrGoalWait, herdrCheckpointDelay, herdrPoll = ready, goal, delay, poll
	t.Cleanup(func() {
		herdrReadyWait, herdrGoalWait, herdrCheckpointDelay, herdrPoll = r, g, d, p
	})
}

func agentJSON(pane, status string) string {
	return `{"result":{"agent":{"pane_id":"` + pane + `","agent":"omp","agent_status":"` + status + `","cwd":"/w"}}}`
}

func tabJSON(pane string) string {
	return `{"result":{"root_pane":{"pane_id":"` + pane + `"},"tab":{"tab_id":"t1"}}}`
}

const notFoundJSON = `{"error":{"code":"agent_not_found","message":"agent target round-1 not found"},"id":"cli:agent:get"}`

const (
	readVisible = "agent read round-1 --source visible --lines 10"
	readRecent  = "agent read round-1 --source recent-unwrapped --lines 60"
	getSlug     = "agent get round-1"
	goalLine    = "/goal do the plan, brief at /w/.claude/dispatch/round-1-brief.md"
)

// ---- findOrMakeTab ----

func TestHerdrFindFoundIdleReusesPane(t *testing.T) {
	fastWaits(t, 50*time.Millisecond, 0, 0, time.Millisecond)
	h := newScriptedHerdr(t)
	h.out("agent_get", agentJSON("wM:p3", "idle"))
	pane, err := findOrMakeTab("round-1", "/w")
	if err != nil {
		t.Fatal(err)
	}
	if pane != "wM:p3" {
		t.Fatalf("pane %q, want wM:p3", pane)
	}
	h.wantCalls([]string{getSlug}, h.calls())
}

func TestHerdrFindFoundDoneReusesPane(t *testing.T) {
	fastWaits(t, 50*time.Millisecond, 0, 0, time.Millisecond)
	h := newScriptedHerdr(t)
	h.out("agent_get", agentJSON("wM:p3", "done"))
	pane, err := findOrMakeTab("round-1", "/w")
	if err != nil || pane != "wM:p3" {
		t.Fatalf("pane %q err %v, want wM:p3", pane, err)
	}
	h.wantCalls([]string{getSlug}, h.calls())
}

func TestHerdrFindFoundWorkingRefusesAndSendsNothing(t *testing.T) {
	fastWaits(t, 50*time.Millisecond, 0, 0, time.Millisecond)
	h := newScriptedHerdr(t)
	h.out("agent_get", agentJSON("wM:p3", "working"))
	pane, err := findOrMakeTab("round-1", "/w")
	if err == nil || !strings.Contains(err.Error(), "still") {
		t.Fatalf("want an error that a round still runs, got pane %q err %v", pane, err)
	}
	if pane != "" {
		t.Fatalf("pane %q returned with an error", pane)
	}
	h.wantCalls([]string{getSlug}, h.calls())
}

func TestHerdrFindNotFoundMakesOneTab(t *testing.T) {
	fastWaits(t, 2*time.Second, 0, 0, time.Millisecond)
	h := newScriptedHerdr(t)
	h.fail("agent_get", notFoundJSON, "")      // the slug is unknown
	h.fail("agent_get", notFoundJSON, "")      // the new pane has no agent yet
	h.out("agent_get", agentJSON("wM:p5", "")) // omp is up
	h.out("tab_create", tabJSON("wM:p5"))
	pane, err := findOrMakeTab("round-1", "/w")
	if err != nil {
		t.Fatal(err)
	}
	if pane != "wM:p5" {
		t.Fatalf("pane %q, want wM:p5", pane)
	}
	h.wantCalls([]string{
		getSlug,
		"tab create --cwd /w --workspace wM --label round-1 --no-focus",
		"pane run wM:p5 omp",
		"agent get wM:p5",
		"agent get wM:p5",
		"agent rename wM:p5 round-1",
	}, h.calls())
}

func TestHerdrFindHerdrErrorIsNotNotFound(t *testing.T) {
	fastWaits(t, 50*time.Millisecond, 0, 0, time.Millisecond)
	h := newScriptedHerdr(t)
	h.fail("agent_get", "", "socket exploded")
	_, err := findOrMakeTab("round-1", "/w")
	if err == nil || !strings.Contains(err.Error(), "socket exploded") {
		t.Fatalf("want herdr's stderr in the error, got %v", err)
	}
	h.wantCalls([]string{getSlug}, h.calls())
}

// "not found" in a herdr failure that is not the agent_not_found answer is a
// herdr error: acting on it would open a second tab next to a live one.
func TestHerdrIsNotFoundOnlyForAgentNotFound(t *testing.T) {
	fastWaits(t, 50*time.Millisecond, 0, 0, time.Millisecond)
	h := newScriptedHerdr(t)
	h.fail("agent_get", "", "socket not found")
	_, err := findOrMakeTab("round-1", "/w")
	if err == nil || !strings.Contains(err.Error(), "socket not found") {
		t.Fatalf("want herdr's stderr in the error, got %v", err)
	}
	h.wantCalls([]string{getSlug}, h.calls())
}

func TestHerdrFindHerdrMissingIsAnError(t *testing.T) {
	fastWaits(t, 50*time.Millisecond, 0, 0, time.Millisecond)
	t.Setenv("PATH", t.TempDir())
	if _, err := findOrMakeTab("round-1", "/w"); err == nil {
		t.Fatal("want an error when herdr is not installed")
	}
}

func TestHerdrFindNeedsPaneIDBeforeMakingATab(t *testing.T) {
	fastWaits(t, 50*time.Millisecond, 0, 0, time.Millisecond)
	h := newScriptedHerdr(t)
	t.Setenv("HERDR_PANE_ID", "")
	h.fail("agent_get", notFoundJSON, "")
	if _, err := findOrMakeTab("round-1", "/w"); err == nil || !strings.Contains(err.Error(), "HERDR_PANE_ID") {
		t.Fatalf("want an error naming HERDR_PANE_ID, got %v", err)
	}
	h.wantCalls([]string{getSlug}, h.calls())
}

func TestHerdrFindOmpNeverStartsEndsAtTheLimit(t *testing.T) {
	// The limit is wider than the poll needs: on a busy machine each call starts a shell and can be slow.
	fastWaits(t, 500*time.Millisecond, 0, 0, 5*time.Millisecond)
	h := newScriptedHerdr(t)
	h.fail("agent_get", notFoundJSON, "")
	h.out("tab_create", tabJSON("wM:p5"))
	start := time.Now()
	_, err := findOrMakeTab("round-1", "/w")
	if err == nil || !strings.Contains(err.Error(), "wM:p5") {
		t.Fatalf("want an error naming the pane, got %v", err)
	}
	if time.Since(start) > 2*time.Second {
		t.Fatal("the wait did not stop at its limit")
	}
	if h.count("agent get wM:p5") < 2 {
		t.Fatalf("the wait did not poll: %q", h.calls())
	}
	if h.count("agent rename") != 0 {
		t.Fatalf("renamed a pane that never got an agent: %q", h.calls())
	}
}

// ---- deliverGoal ----

func TestHerdrDeliverGoalClean(t *testing.T) {
	fastWaits(t, time.Second, time.Second, 0, time.Millisecond)
	h := newScriptedHerdr(t)
	h.out("agent_get", agentJSON("wM:p3", "idle"))
	h.out("agent_read_visible", "\n> \n")
	h.out("agent_read_visible", "status 🎯 Goal active")
	if err := deliverGoal("round-1", goalLine); err != nil {
		t.Fatal(err)
	}
	h.wantCalls([]string{
		getSlug,
		readVisible,
		"agent prompt round-1 " + goalLine,
		readVisible,
	}, h.calls())
}

func TestHerdrDeliverGoalWaitsForIdle(t *testing.T) {
	fastWaits(t, 2*time.Second, time.Second, 0, time.Millisecond)
	h := newScriptedHerdr(t)
	h.out("agent_get", agentJSON("wM:p3", "working"))
	h.out("agent_get", agentJSON("wM:p3", "working"))
	h.out("agent_get", agentJSON("wM:p3", "idle"))
	h.out("agent_read_visible", "> ")
	h.out("agent_read_visible", "🎯 Goal")
	if err := deliverGoal("round-1", goalLine); err != nil {
		t.Fatal(err)
	}
	if h.count(getSlug) != 3 {
		t.Fatalf("want 3 status polls, got %q", h.calls())
	}
	if h.count("agent prompt") != 1 {
		t.Fatalf("want one prompt, got %q", h.calls())
	}
}

func TestHerdrDeliverGoalOmpNeverReady(t *testing.T) {
	fastWaits(t, 30*time.Millisecond, time.Second, 0, 5*time.Millisecond)
	h := newScriptedHerdr(t)
	h.out("agent_get", agentJSON("wM:p3", "working"))
	start := time.Now()
	err := deliverGoal("round-1", goalLine)
	if err == nil || !strings.Contains(err.Error(), "not ready") {
		t.Fatalf("want an omp not ready error, got %v", err)
	}
	if time.Since(start) > 2*time.Second {
		t.Fatal("the wait did not stop at its limit")
	}
	if h.count("agent prompt") != 0 || h.count("agent read") != 0 {
		t.Fatalf("sent or read something before omp was ready: %q", h.calls())
	}
}

func TestHerdrDeliverGoalClearsHeldGoal(t *testing.T) {
	for name, screen := range map[string]string{
		"paused mark":  "status ⏸ Goal paused",
		"resume first": "warning: Resume the current goal first",
	} {
		t.Run(name, func(t *testing.T) {
			fastWaits(t, time.Second, time.Second, 0, time.Millisecond)
			h := newScriptedHerdr(t)
			h.out("agent_get", agentJSON("wM:p3", "idle"))
			h.out("agent_read_visible", screen)
			h.out("agent_read_visible", "🎯 Goal")
			if err := deliverGoal("round-1", goalLine); err != nil {
				t.Fatal(err)
			}
			h.wantCalls([]string{
				getSlug,
				readVisible,
				"agent prompt round-1 /goal drop",
				"agent send-keys round-1 enter",
				"agent send-keys round-1 enter",
				"agent prompt round-1 " + goalLine,
				readVisible,
			}, h.calls())
		})
	}
}

func TestHerdrDeliverGoalMarkNeverShowsFailsAfterOneRetry(t *testing.T) {
	fastWaits(t, time.Second, 0, 0, time.Millisecond)
	h := newScriptedHerdr(t)
	h.out("agent_get", agentJSON("wM:p3", "idle"))
	h.out("agent_read_visible", "> ")
	h.out("agent_read_visible", "> still nothing")
	err := deliverGoal("round-1", goalLine)
	if err == nil || !strings.Contains(err.Error(), "goal mark") {
		t.Fatalf("want a goal mark error, got %v", err)
	}
	h.wantCalls([]string{
		getSlug,
		readVisible,
		"agent prompt round-1 " + goalLine,
		readVisible,
		"agent prompt round-1 " + goalLine,
		readVisible,
	}, h.collapsed())
	if h.count("agent prompt") != 2 {
		t.Fatalf("want the goal sent twice, got %q", h.calls())
	}
}

func TestHerdrDeliverGoalMarkShowsAfterRetry(t *testing.T) {
	fastWaits(t, time.Second, 0, 0, time.Millisecond)
	h := newScriptedHerdr(t)
	h.out("agent_get", agentJSON("wM:p3", "idle"))
	h.out("agent_read_visible", "> ")
	h.out("agent_read_visible", "> not yet")
	h.out("agent_read_visible", "🎯 Goal")
	if err := deliverGoal("round-1", goalLine); err != nil {
		t.Fatal(err)
	}
	h.wantCalls([]string{
		getSlug,
		readVisible,
		"agent prompt round-1 " + goalLine,
		readVisible,
		"agent prompt round-1 " + goalLine,
		readVisible,
	}, h.calls())
}

func TestHerdrDeliverGoalRefusesTextThatIsNotAGoal(t *testing.T) {
	fastWaits(t, time.Second, time.Second, 0, time.Millisecond)
	h := newScriptedHerdr(t)
	for _, bad := range []string{"/new", "hello", "/goalish", ""} {
		if err := deliverGoal("round-1", bad); err == nil {
			t.Fatalf("goal %q was accepted", bad)
		}
	}
	if len(h.calls()) != 0 {
		t.Fatalf("herdr was called: %q", h.calls())
	}
}

func TestHerdrDeliverGoalReadFailureStops(t *testing.T) {
	fastWaits(t, time.Second, time.Second, 0, time.Millisecond)
	h := newScriptedHerdr(t)
	h.out("agent_get", agentJSON("wM:p3", "idle"))
	h.fail("agent_read_visible", "", "pane gone")
	err := deliverGoal("round-1", goalLine)
	if err == nil || !strings.Contains(err.Error(), "pane gone") {
		t.Fatalf("want herdr's stderr in the error, got %v", err)
	}
	if h.count("agent prompt") != 0 {
		t.Fatalf("sent a goal after a failed read: %q", h.calls())
	}
}

// ---- checkpoint ----

func TestHerdrCheckpointOK(t *testing.T) {
	fastWaits(t, time.Second, time.Second, 20*time.Millisecond, time.Millisecond)
	h := newScriptedHerdr(t)
	h.out("agent_read_recent-unwrapped", todoCard+"- Task 1: a\n- task-2 b\n- TASK 4 c\n")
	start := time.Now()
	verdict, missing, pane := checkpoint("round-1", []string{"1", "2", "4"})
	if verdict != "ok" || len(missing) != 0 {
		t.Fatalf("verdict %q missing %q", verdict, missing)
	}
	if time.Since(start) < 20*time.Millisecond {
		t.Fatal("did not wait for the delay before reading")
	}
	if !strings.Contains(pane, "Task 1") {
		t.Fatalf("pane text not returned: %q", pane)
	}
	h.wantCalls([]string{readRecent}, h.calls())
}

func TestHerdrCheckpointDriftNamesMissingIDs(t *testing.T) {
	fastWaits(t, time.Second, time.Second, 0, time.Millisecond)
	h := newScriptedHerdr(t)
	// Task 12 must not count as task 1, and task 3 is absent.
	h.out("agent_read_recent-unwrapped", todoCard+"- Task 12: other\n- Task 2: b\n")
	verdict, missing, pane := checkpoint("round-1", []string{"1", "2", "3"})
	if verdict != "drift" || !reflect.DeepEqual(missing, []string{"1", "3"}) {
		t.Fatalf("verdict %q missing %q", verdict, missing)
	}
	if !strings.Contains(pane, "Task 12") {
		t.Fatalf("pane text not returned: %q", pane)
	}
	h.wantCalls([]string{readRecent}, h.calls())
}

// What omp draws for its todo list: a header line titled Todo. The glyph
// and the colors depend on the theme, so the tests never rely on them.
const todoCard = "\u23fa Todo 3 tasks\n"

func TestHerdrCheckpointUnconfirmedWithNoTodoList(t *testing.T) {
	fastWaits(t, time.Second, time.Second, 0, time.Millisecond)
	h := newScriptedHerdr(t)
	h.out("agent_read_recent-unwrapped", "> thinking about the brief\n")
	verdict, missing, _ := checkpoint("round-1", []string{"1", "2"})
	if verdict != "unconfirmed" || len(missing) != 0 {
		t.Fatalf("verdict %q missing %q", verdict, missing)
	}
	h.wantCalls([]string{readRecent}, h.calls())
}

// A made-up todo list is the case the checkpoint exists for: a card is on
// screen and none of the ids is in it. That is drift, not "not written yet".
func TestHerdrCheckpointCardWithNoIDsIsDrift(t *testing.T) {
	fastWaits(t, time.Second, time.Second, 0, time.Millisecond)
	h := newScriptedHerdr(t)
	h.out("agent_read_recent-unwrapped", "\u23fa Todo 2 tasks\n\u2610 Build the auth module\n\u2610 Add the REST endpoints\n")
	verdict, missing, pane := checkpoint("round-1", []string{"1", "2"})
	if verdict != "drift" || !reflect.DeepEqual(missing, []string{"1", "2"}) {
		t.Fatalf("verdict %q missing %q", verdict, missing)
	}
	if !strings.Contains(pane, "auth module") {
		t.Fatalf("pane text not returned: %q", pane)
	}
}

// Ids that sit in plain text with no Todo card are not a todo list. The
// words Todos, Todo-list and mytodo are not the card title either.
func TestHerdrCheckpointIDsWithoutCardAreUnconfirmed(t *testing.T) {
	for _, pane := range []string{
		"Task 1: a\nTask 2: b\n",
		"Todos\n- Task 1: a\n- Task 2: b\n",
		"Todo-list\n- Task 1: a\n- Task 2: b\n",
		"mytodo 3 tasks\n- Task 1: a\n- Task 2: b\n",
		"I will make a todo list: Task 1, Task 2\n",
	} {
		fastWaits(t, time.Second, time.Second, 0, time.Millisecond)
		h := newScriptedHerdr(t)
		h.out("agent_read_recent-unwrapped", pane)
		verdict, missing, _ := checkpoint("round-1", []string{"1", "2"})
		if verdict != "unconfirmed" || len(missing) != 0 {
			t.Fatalf("pane %q: verdict %q missing %q", pane, verdict, missing)
		}
	}
}

// The card header may sit inside a frame, so glyphs and spaces before the
// word Todo are allowed.
func TestHerdrCheckpointCardInsideAFrame(t *testing.T) {
	fastWaits(t, time.Second, time.Second, 0, time.Millisecond)
	h := newScriptedHerdr(t)
	h.out("agent_read_recent-unwrapped", "\u256d\u2500 \u23fa Todo 2/2\n\u2502 Task 1: a\n\u2502 Task 2: b\n")
	verdict, missing, _ := checkpoint("round-1", []string{"1", "2"})
	if verdict != "ok" || len(missing) != 0 {
		t.Fatalf("verdict %q missing %q", verdict, missing)
	}
}

// omp's ascii symbol preset draws the todo icon as "[x]", and x is a letter.
func TestHerdrCheckpointAsciiHeader(t *testing.T) {
	fastWaits(t, time.Second, time.Second, 0, time.Millisecond)
	h := newScriptedHerdr(t)
	h.out("agent_read_recent-unwrapped", "[x] Todo 2 tasks\n- Task 1: a\n- Task 2: b\n")
	verdict, missing, _ := checkpoint("round-1", []string{"1", "2"})
	if verdict != "ok" || len(missing) != 0 {
		t.Fatalf("verdict %q missing %q", verdict, missing)
	}
}

// The house rules show "T-1, T-3, T-4", so that form counts as naming the id.
func TestHerdrCheckpointAcceptsTDashForm(t *testing.T) {
	fastWaits(t, time.Second, time.Second, 0, time.Millisecond)
	h := newScriptedHerdr(t)
	h.out("agent_read_recent-unwrapped", todoCard+"- T-1 / T-2 / t-3\n")
	verdict, missing, _ := checkpoint("round-1", []string{"1", "2", "3"})
	if verdict != "ok" || len(missing) != 0 {
		t.Fatalf("verdict %q missing %q", verdict, missing)
	}
}

func TestHerdrCheckpointTDashTenIsNotOne(t *testing.T) {
	fastWaits(t, time.Second, time.Second, 0, time.Millisecond)
	h := newScriptedHerdr(t)
	h.out("agent_read_recent-unwrapped", todoCard+"- T-10: other\n")
	verdict, missing, _ := checkpoint("round-1", []string{"1"})
	if verdict != "drift" || !reflect.DeepEqual(missing, []string{"1"}) {
		t.Fatalf("verdict %q missing %q", verdict, missing)
	}
}

func TestHerdrCheckpointReadFailureIsUnconfirmed(t *testing.T) {
	fastWaits(t, time.Second, time.Second, 0, time.Millisecond)
	h := newScriptedHerdr(t)
	h.fail("agent_read_recent-unwrapped", "", "pane gone")
	verdict, _, pane := checkpoint("round-1", []string{"1"})
	if verdict != "unconfirmed" {
		t.Fatalf("verdict %q, want unconfirmed", verdict)
	}
	if !strings.Contains(pane, "pane gone") {
		t.Fatalf("want the failure in the text, got %q", pane)
	}
	if n := h.count("agent read"); n != 1 {
		t.Fatalf("want exactly one read, got %d", n)
	}
}

func TestHerdrCheckpointNoTasksCallsNothing(t *testing.T) {
	fastWaits(t, time.Second, time.Second, time.Hour, time.Millisecond)
	h := newScriptedHerdr(t)
	verdict, missing, _ := checkpoint("round-1", nil)
	if verdict != "ok" || len(missing) != 0 {
		t.Fatalf("verdict %q missing %q", verdict, missing)
	}
	if len(h.calls()) != 0 {
		t.Fatalf("herdr was called: %q", h.calls())
	}
}

// The helper's own contract: a polling wait reaches its limit instead of
// looping forever, even when every answer is "not yet".
func TestHerdrWaitStopsAtTheLimit(t *testing.T) {
	fastWaits(t, time.Second, time.Second, 0, 2*time.Millisecond)
	n := 0
	start := time.Now()
	// The limit is wider than the poll needs: a busy machine can stall one poll.
	ok, err := waitFor(200*time.Millisecond, func() (bool, error) { n++; return false, nil })
	if ok || err != nil {
		t.Fatalf("ok %v err %v", ok, err)
	}
	if n < 2 || time.Since(start) > time.Second {
		t.Fatalf("polled %d times in %v", n, time.Since(start))
	}
	zero := 0
	waitFor(0, func() (bool, error) { zero++; return false, nil })
	if zero != 1 {
		t.Fatalf("a zero limit must check exactly once, got %d", zero)
	}
}

// A finished omp waits for input too, so a done agent is ready.
func TestHerdrDeliverGoalDoneAgentIsReady(t *testing.T) {
	fastWaits(t, 0, time.Second, 0, time.Millisecond)
	h := newScriptedHerdr(t)
	h.out("agent_get", agentJSON("wM:p3", "done"))
	h.out("agent_read_visible", "> ")
	h.out("agent_read_visible", "🎯 Goal")
	if err := deliverGoal("round-1", goalLine); err != nil {
		t.Fatal(err)
	}
	if h.count("agent prompt") != 1 {
		t.Fatalf("want one prompt, got %q", h.calls())
	}
}

// Blocked is not ready: sending a goal into a blocked prompt could answer
// the wrong question.
func TestHerdrDeliverGoalBlockedAgentIsNotReady(t *testing.T) {
	fastWaits(t, 0, time.Second, 0, time.Millisecond)
	h := newScriptedHerdr(t)
	h.out("agent_get", agentJSON("wM:p3", "blocked"))
	if err := deliverGoal("round-1", goalLine); err == nil || !strings.Contains(err.Error(), "not ready") {
		t.Fatalf("want not ready, got %v", err)
	}
	h.wantCalls([]string{getSlug}, h.calls())
}
