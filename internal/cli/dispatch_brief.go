package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/iyay/acta/internal/board"
)

// polishRound is the round name for the commit that applies the review NOTEs.
// Its tasks come from the "## Polish" section of the plan.
const polishRound = "polish"

var (
	fixHeadRe    = regexp.MustCompile(`^## Fix round(\s|$)`)
	polishHeadRe = regexp.MustCompile(`^## Polish\s*$`)
	verifyRe     = regexp.MustCompile(`^\*\*verify:\*\*\s*(.*\S)\s*$`)
	testsRe      = regexp.MustCompile(`^\*\*Tests:\*\*\s*(.*\S)\s*$`)
)

// briefInput is everything the brief needs besides the plan text.
type briefInput struct {
	Round        string // "" first dispatch, "polish", or any other name for the last fix round
	Note         string
	Rules        string // absolute path of house-rules.md
	Worktree     string
	Branch       string
	Parent       string
	Base         string
	Home         string // user home, where the Claude memory lives
	MainCheckout string // the main checkout, never the worktree
}

// planMarks is what one pass over the plan text finds outside code blocks.
type planMarks struct {
	fixLines   []int // 1-based lines of the "## Fix round" headings
	fixEnd     int   // line after the last fix round section ends, 0 = end of file
	polishLine int   // 1-based line of the "## Polish" heading, 0 = none
	waves      string
	hasWaves   bool
	tests      string // text after "**Tests:**"
}

func scanPlan(src []byte) planMarks {
	var m planMarks
	inFence, inWaves, afterFix := false, false, false
	var waves []string
	for i, ln := range strings.Split(strings.ReplaceAll(string(src), "\r\n", "\n"), "\n") {
		if strings.HasPrefix(strings.TrimSpace(ln), "```") {
			inFence = !inFence
			if inWaves {
				waves = append(waves, ln)
			}
			continue
		}
		if !inFence && strings.HasPrefix(ln, "## ") {
			inWaves = false
			// The polish section ends whatever ran before it, the way a
			// later "## Fix round" ends an earlier one.
			if afterFix && !fixHeadRe.MatchString(ln) && m.fixEnd == 0 {
				m.fixEnd = i + 1
			}
			if fixHeadRe.MatchString(ln) {
				m.fixLines = append(m.fixLines, i+1)
				afterFix, m.fixEnd = true, 0
			}
			// The first "## Polish" heading wins; a second one is a plan
			// typo, and the tasks still sit below the first.
			if m.polishLine == 0 && polishHeadRe.MatchString(strings.TrimSpace(ln)) {
				m.polishLine = i + 1
			}
			if strings.TrimSpace(ln) == "## Waves" {
				inWaves, m.hasWaves = true, true
			}
			continue
		}
		if !inFence && strings.HasPrefix(ln, "### ") {
			// A task heading ends the waves section too.
			inWaves = false
		}
		if inWaves {
			waves = append(waves, ln)
		}
		if !inFence && m.tests == "" {
			if t := testsRe.FindStringSubmatch(ln); t != nil {
				m.tests = t[1]
			}
		}
	}
	m.waves = strings.TrimSpace(strings.Join(waves, "\n"))
	return m
}

// briefTasks picks the tasks one round covers, by heading line. The send
// command also needs their ids for its checkpoint.
func briefTasks(src []byte, round string) ([]board.TaskSec, error) {
	marks := scanPlan(src)
	all := board.Parse(src).Tasks
	lo, hi := 0, 0 // task heading line must be > lo, and < hi when hi > 0
	// The polish round hands over the "## Polish" task, where the review
	// NOTEs land. Without that section there is nothing to hand over.
	if round == polishRound {
		if marks.polishLine == 0 {
			return nil, fmt.Errorf("round %q needs a \"## Polish\" section in the plan, found none", round)
		}
		lo = marks.polishLine
	} else if round == "" {
		// The first dispatch stops at whichever section comes first.
		hi = firstSectionLine(marks)
	} else {
		if !roundPattern.MatchString(round) {
			return nil, fmt.Errorf("round %q is not a valid name", round)
		}
		if len(marks.fixLines) == 0 {
			return nil, fmt.Errorf("round %q needs a \"## Fix round\" section in the plan, found none", round)
		}
		lo, hi = marks.fixLines[len(marks.fixLines)-1], marks.fixEnd
		// A polish section after the last fix round is not a fix task.
		if marks.polishLine > lo && (hi == 0 || marks.polishLine < hi) {
			hi = marks.polishLine
		}
	}
	var out []board.TaskSec
	for _, t := range all {
		if t.Line > lo && (hi == 0 || t.Line < hi) {
			out = append(out, t)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("the plan has no tasks for this round")
	}
	return out, nil
}

// firstSectionLine is the line where the first round stops: the earlier of
// the first fix round and the polish section, 0 when neither exists.
func firstSectionLine(marks planMarks) int {
	if len(marks.fixLines) > 0 && marks.polishLine > 0 {
		if marks.fixLines[0] < marks.polishLine {
			return marks.fixLines[0]
		}
		return marks.polishLine
	}
	if len(marks.fixLines) > 0 {
		return marks.fixLines[0]
	}
	return marks.polishLine
}

// taskVerify returns the text of the task's "**verify:**" line.
func taskVerify(t board.TaskSec) (string, bool) {
	inFence := false
	for _, ln := range strings.Split(t.Body, "\n") {
		if strings.HasPrefix(strings.TrimSpace(ln), "```") {
			inFence = !inFence
		}
		if !inFence {
			if m := verifyRe.FindStringSubmatch(ln); m != nil {
				return m[1], true
			}
		}
	}
	return "", false
}

// projectMemoryDir is how Claude Code names a project folder: every slash and
// dot of the main checkout path becomes a dash.
func projectMemoryDir(main string) string {
	return strings.NewReplacer("/", "-", ".", "-").Replace(main)
}

func (in briefInput) check() error {
	if !filepath.IsAbs(in.Rules) {
		return fmt.Errorf("house rules path %q is not absolute", in.Rules)
	}
	if st, err := os.Stat(in.Rules); err != nil || st.IsDir() {
		return fmt.Errorf("house rules file %q does not exist", in.Rules)
	}
	if !filepath.IsAbs(in.Worktree) {
		return fmt.Errorf("worktree %q is not an absolute path", in.Worktree)
	}
	if in.Branch == "" {
		return fmt.Errorf("branch is empty")
	}
	if in.Parent == "" {
		return fmt.Errorf("parent is empty")
	}
	if !shaPattern.MatchString(in.Base) {
		return fmt.Errorf("base %q is not a full commit id", in.Base)
	}
	if in.Round == polishRound && strings.TrimSpace(in.Note) == "" {
		return fmt.Errorf("round polish needs a note")
	}
	return nil
}

// buildBrief returns the brief text for one round. It writes no file, and on
// any error it returns no text, so nothing half-made gets sent.
func buildBrief(planPath string, src []byte, in briefInput) (string, error) {
	if err := in.check(); err != nil {
		return "", err
	}
	tasks, err := briefTasks(src, in.Round)
	if err != nil {
		return "", err
	}
	doc := board.Parse(src)
	marks := scanPlan(src)
	if marks.tests == "" {
		return "", fmt.Errorf("the plan has no **Tests:** line")
	}
	stem := strings.TrimSuffix(filepath.Base(planPath), ".md")
	var tickets []string
	for _, t := range tasks {
		v, ok := taskVerify(t)
		if !ok {
			return "", fmt.Errorf("task %s has no **verify:** line", t.Num)
		}
		tickets = append(tickets, fmt.Sprintf("  plans/%s#task-%s — %s → verify: %s", stem, t.Num, t.Title, v))
	}
	if len(tasks) > 0 && !marks.hasWaves {
		return "", fmt.Errorf("the plan has no ## Waves section")
	}

	spec := doc.SpecPath
	if spec == "" {
		spec = "none"
	}
	verb := map[bool]string{true: "Apply the review notes on", false: "Build"}[in.Round == polishRound]
	if in.Round != "" && in.Round != polishRound {
		verb = "Fix round " + in.Round + " for"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s plan %s.\n\n", verb, doc.Title)
	fmt.Fprintf(&b, "PLAN: %s (design: %s) — read FIRST, before any todo list. Its tasks listed below are the only tickets; no decomposition of your own.\n", planPath, spec)
	b.WriteString("SKILL: load build (omp: build) and tdd before the todo list, and follow them for every task.\n")
	b.WriteString("STATE: read `acta state <plan>` FIRST, before the tickets: it shows what an earlier session left in the plan's ## State.\n")
	if len(tickets) > 0 {
		fmt.Fprintf(&b, "TICKETS (exactly these %d):\n%s\n", len(tickets), strings.Join(tickets, "\n"))
		fmt.Fprintf(&b, "WAVES (the plan's, as written):\n%s\n", marks.waves)
	}
	fmt.Fprintf(&b, "WORKTREE: %s (branch %s, parent %s, base %s) — cd there FIRST, work ONLY there. The main checkout stays clean. No git checkout, no cd out, no git push.\n", in.Worktree, in.Branch, in.Parent, in.Base)
	b.WriteString("FILES: the plan names the area; find the exact lines yourself. A line number is a hint, never the edge. Surgical: every changed line traces to a ticket.\n")
	fmt.Fprintf(&b, "HOUSE RULES: before the todo list, read %s and AGENTS.md in this worktree.\n", in.Rules)
	if mem := memoryPaths(in); len(mem) > 0 {
		fmt.Fprintf(&b, "MEMORY: before the todo list, read %s. They are indexes: open a linked note only when its hook fits a ticket. Read-only, never write there.\n", strings.Join(mem, " and "))
	}
	fmt.Fprintf(&b, "GATES (from the worktree): %s; typecheck; git diff --stat vs %s shows only plan files. In a repo that has `scripts/test`, never run bare `go test`; the pre-tool hook blocks it, and every implementer shares one machine.\n", marks.tests, in.Base)
	if n := strings.TrimSpace(in.Note); n != "" {
		fmt.Fprintf(&b, "NOTE: %s\n", n)
	}
	b.WriteString("REPLY-BACK: after the last commit the build skill runs `acta reply-back`. Nothing else to hand-fill.\n")
	return b.String(), nil
}

// memoryPaths lists the Claude memory indexes that exist: the user one and the
// one of the main checkout.
func memoryPaths(in briefInput) []string {
	var out []string
	for _, p := range []string{
		filepath.Join(in.Home, ".claude", "memory", "MEMORY.md"),
		filepath.Join(in.Home, ".claude", "projects", projectMemoryDir(in.MainCheckout), "memory", "MEMORY.md"),
	} {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			out = append(out, p)
		}
	}
	return out
}
