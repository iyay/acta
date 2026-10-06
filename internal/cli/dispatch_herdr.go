package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// Package variables, so a test can shrink the waits to milliseconds. Every
// wait has a limit: nothing in this file can loop forever.
var (
	herdrReadyWait       = 60 * time.Second
	herdrGoalWait        = 10 * time.Second
	herdrCheckpointDelay = 20 * time.Second
	herdrPoll            = time.Second
)

const (
	checkpointOK          = "ok"
	checkpointDrift       = "drift"
	checkpointUnconfirmed = "unconfirmed"
)

// herdrError keeps what herdr printed, so a caller can tell "unknown agent"
// from "herdr itself broke", and the user sees herdr's own words.
type herdrError struct {
	args   []string
	output string
	err    error
}

func (e *herdrError) Error() string {
	cmd := strings.Join(e.args, " ")
	if len(e.args) > 2 {
		cmd = strings.Join(e.args[:2], " ")
	}
	if e.output == "" {
		return fmt.Sprintf("herdr %s: %v", cmd, e.err)
	}
	return fmt.Sprintf("herdr %s: %v: %s", cmd, e.err, e.output)
}

func (e *herdrError) Unwrap() error { return e.err }

// herdr runs one herdr call. The arguments stay separate, so plan text and
// pane text can never turn into shell text.
func herdr(args ...string) (string, error) {
	cmd := exec.Command("herdr", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if err != nil {
		text := strings.TrimSpace(stderr.String())
		if text == "" {
			text = strings.TrimSpace(stdout.String())
		}
		return stdout.String(), &herdrError{args: args, output: text, err: err}
	}
	return stdout.String(), nil
}

// isNotFound is true only when herdr ran, exited non-zero, and said the
// agent is unknown. herdr 0.9 prints the code agent_not_found on stdout. A
// missing herdr binary, a dead socket or any other failure is a real error:
// treating it as "not found" would open a second tab next to a live one.
func isNotFound(err error, stdout string) bool {
	var he *herdrError
	var exit *exec.ExitError
	if !errors.As(err, &he) || !errors.As(err, &exit) {
		return false
	}
	text := strings.ToLower(stdout + " " + he.output)
	return strings.Contains(text, "agent_not_found")
}

type agentInfo struct {
	Pane   string
	Status string
}

// parseAgent reads the answer of `herdr agent get`.
func parseAgent(out string) (agentInfo, error) {
	var doc struct {
		Result struct {
			Agent struct {
				PaneID string `json:"pane_id"`
				Status string `json:"agent_status"`
			} `json:"agent"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		return agentInfo{}, fmt.Errorf("herdr agent get printed text I cannot read: %w", err)
	}
	a := doc.Result.Agent
	if a.PaneID == "" {
		return agentInfo{}, errors.New("herdr agent get named no pane")
	}
	return agentInfo{Pane: a.PaneID, Status: a.Status}, nil
}

// waitFor runs check until it says yes, says error, or the limit passes. The
// check always runs once, so a limit of zero still looks one time.
func waitFor(limit time.Duration, check func() (bool, error)) (bool, error) {
	deadline := time.Now().Add(limit)
	for {
		ok, err := check()
		if ok || err != nil {
			return ok, err
		}
		if !time.Now().Before(deadline) {
			return false, nil
		}
		time.Sleep(herdrPoll)
	}
}

// findOrMakeTab returns the pane of the agent named slug. A working agent is
// refused, because its round is not over. An unknown agent gets a new tab in
// the orchestrator's workspace, with omp running in the worktree.
func findOrMakeTab(slug, worktree string) (string, error) {
	out, err := herdr("agent", "get", slug)
	switch {
	case err == nil:
		a, perr := parseAgent(out)
		if perr != nil {
			return "", perr
		}
		if a.Status == "working" {
			return "", fmt.Errorf("agent %s is still working: a previous round still runs, nothing sent", slug)
		}
		return a.Pane, nil
	case !isNotFound(err, out):
		return "", err
	}
	return makeTab(slug, worktree)
}

func makeTab(slug, worktree string) (string, error) {
	ws, _, _ := strings.Cut(os.Getenv("HERDR_PANE_ID"), ":")
	if ws == "" {
		return "", errors.New("HERDR_PANE_ID is not set, so I do not know which workspace to open the tab in")
	}
	out, err := herdr("tab", "create", "--cwd", worktree, "--workspace", ws, "--label", slug, "--no-focus")
	if err != nil {
		return "", err
	}
	var doc struct {
		Result struct {
			Root struct {
				PaneID string `json:"pane_id"`
			} `json:"root_pane"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &doc); err != nil || doc.Result.Root.PaneID == "" {
		return "", fmt.Errorf("herdr tab create named no pane: %q", strings.TrimSpace(out))
	}
	pane := doc.Result.Root.PaneID
	if _, err := herdr("pane", "run", pane, "omp"); err != nil {
		return "", fmt.Errorf("tab made in pane %s but omp did not start: %w", pane, err)
	}
	var lastErr error
	up, _ := waitFor(herdrReadyWait, func() (bool, error) {
		out, err := herdr("agent", "get", pane)
		if err != nil {
			lastErr = err
			return false, nil
		}
		_, lastErr = parseAgent(out)
		return lastErr == nil, nil
	})
	if !up {
		return "", fmt.Errorf("omp did not show up in pane %s within %s: %v", pane, herdrReadyWait, lastErr)
	}
	if _, err := herdr("agent", "rename", pane, slug); err != nil {
		return "", err
	}
	return pane, nil
}

// heldGoalMarks are what omp shows when an old goal is paused. A new goal
// sent on top of it is ignored, so it has to be dropped first.
var heldGoalMarks = []string{"⏸ Goal", "Resume the current goal first"}

const goalMark = "🎯 Goal"

// deliverGoal waits for omp, clears a held goal, sends the goal and checks
// that goal mode started. The goal text must start with "/goal ", so nothing
// else, "/new" included, can go through this door.
func deliverGoal(slug, goal string) error {
	if !strings.HasPrefix(goal, "/goal ") {
		return fmt.Errorf("goal text must start with \"/goal \", got %q", goal)
	}
	if err := waitIdle(slug); err != nil {
		return err
	}
	screen, err := herdr("agent", "read", slug, "--source", "visible", "--lines", "10")
	if err != nil {
		return err
	}
	for _, mark := range heldGoalMarks {
		if strings.Contains(screen, mark) {
			if err := dropGoal(slug); err != nil {
				return err
			}
			break
		}
	}
	// One send, one check, and one more send and check. Then stop.
	for try := 0; try < 2; try++ {
		if _, err := herdr("agent", "prompt", slug, goal); err != nil {
			return err
		}
		shown, err := waitFor(herdrGoalWait, func() (bool, error) {
			out, err := herdr("agent", "read", slug, "--source", "visible", "--lines", "10")
			return strings.Contains(out, goalMark), err
		})
		if err != nil {
			return err
		}
		if shown {
			return nil
		}
	}
	return fmt.Errorf("goal mark %q not shown for %s after one retry", goalMark, slug)
}

// waitIdle waits until omp can take input. A done agent has finished its
// last turn and is waiting too, so it counts. Working and blocked do not.
func waitIdle(slug string) error {
	var last string
	ready, _ := waitFor(herdrReadyWait, func() (bool, error) {
		out, err := herdr("agent", "get", slug)
		if err != nil {
			last = err.Error()
			return false, nil
		}
		a, err := parseAgent(out)
		if err != nil {
			last = err.Error()
			return false, nil
		}
		last = "status " + a.Status
		return a.Status == "idle" || a.Status == "done", nil
	})
	if !ready {
		return fmt.Errorf("omp not ready in %s after %s (%s)", slug, herdrReadyWait, last)
	}
	return nil
}

// dropGoal clears a held goal: the drop command, then two enters for the
// prompts omp asks next.
func dropGoal(slug string) error {
	if _, err := herdr("agent", "prompt", slug, "/goal drop"); err != nil {
		return err
	}
	for i := 0; i < 2; i++ {
		if _, err := herdr("agent", "send-keys", slug, "enter"); err != nil {
			return err
		}
	}
	return nil
}

// todoCardHeader finds the header line of omp's todo list card. omp 18 draws the
// card with a header whose title is the word Todo, after a status glyph or a
// frame edge, and the glyphs change with the theme. So the header is the word
// Todo on its own, with only symbols before it. A word like Todos or mytodo,
// or a sentence about a todo list, is not the card. The ascii symbol preset
// draws the icon as "[x]", and x is a letter, so that form is allowed first.
var todoCardHeader = regexp.MustCompile(`(?m)^(?:\[x\])?[^\p{L}\p{N}\n]*Todo(?:\s|$)`)

// checkpoint waits, then reads the pane once and looks for every task id.
// Only the newest round counts: omp leaves the old round's card on screen,
// so everything up to the last goal line is cut away. With no goal line the
// goal never showed, which is "unconfirmed", not a failure. With no todo
// card on screen the list is not written yet, which is
// "unconfirmed", not a failure. A card that names none of the ids is the
// same: the real list is still not written. Only a card that names some ids
// but skips others is "drift". The pane text comes back so the caller can
// show it on drift. A failed read is "unconfirmed" too, with the failure in
// place of the text. With no ids there is nothing to check: ok, and herdr
// is not called.
func checkpoint(slug string, ids []string) (string, []string, string) {
	if len(ids) == 0 {
		return checkpointOK, nil, ""
	}
	time.Sleep(herdrCheckpointDelay)
	text, err := herdr("agent", "read", slug, "--source", "recent-unwrapped", "--lines", "60")
	if err != nil {
		return checkpointUnconfirmed, nil, err.Error()
	}
	at := strings.LastIndex(text, goalMark)
	if at < 0 {
		return checkpointUnconfirmed, nil, text
	}
	if i := strings.IndexByte(text[at:], '\n'); i >= 0 {
		text = text[at+i+1:]
	} else {
		text = ""
	}
	if !todoCardHeader.MatchString(text) {
		return checkpointUnconfirmed, nil, text
	}
	var missing []string
	found := 0
	for _, id := range ids {
		if !regexp.MustCompile(`(?i)\b(?:task[ -]?|t-)` + regexp.QuoteMeta(id) + `\b`).MatchString(text) {
			missing = append(missing, id)
		} else {
			found++
		}
	}
	if found == 0 {
		return checkpointUnconfirmed, nil, text
	}
	if len(missing) > 0 {
		return checkpointDrift, missing, text
	}
	return checkpointOK, nil, text
}
