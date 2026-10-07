package evalomp

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Outcome is one grader's verdict. Why is set on every FAIL.
type Outcome struct {
	Grader string
	Pass   bool
	Why    string
}

// Judge asks a model to grade a reply and returns its raw answer.
type Judge func(prompt string) (string, error)

// Unsupported gives the reason this runner cannot grade g, or "" when it can.
// It covers only what the suite uses; the rest is named, never passed.
func Unsupported(g Grader) string {
	switch g.Type {
	case "file_exists", "tool_used", "llm":
		return ""
	case "regex":
		switch g.Target.Kind {
		case "", "last_message", "file":
		default:
			return fmt.Sprintf("regex target %q is not supported in omp", g.Target.Kind)
		}
		switch g.Match {
		case "", "contains", "not_contains":
		default:
			return fmt.Sprintf("regex match %q is not supported in omp", g.Match)
		}
		return ""
	}
	return fmt.Sprintf("grader type %q is not supported in omp", g.Type)
}

// Grade runs one grader over one finished run.
func Grade(g Grader, w Workspace, judge Judge) Outcome {
	if why := Unsupported(g); why != "" {
		return fail(g, "%s", why)
	}
	switch g.Type {
	case "file_exists":
		return gradeFile(g, w)
	case "regex":
		return gradeRegex(g, w)
	case "tool_used":
		return gradeTool(g, w)
	}
	return gradeLLM(g, w, judge)
}

func fail(g Grader, format string, args ...any) Outcome {
	return Outcome{Grader: g.Name, Why: fmt.Sprintf(format, args...)}
}

func pass(g Grader) Outcome {
	return Outcome{Grader: g.Name, Pass: true}
}

// gradeFile counts only files the agent made. A file the scaffold left there
// was in the before list, and claude plugin eval does not count those either.
func gradeFile(g Grader, w Workspace) Outcome {
	now, err := ListFiles(w.Dir)
	if err != nil {
		return fail(g, "cannot list the workspace: %v", err)
	}
	found := false
	for f := range now {
		if w.Before[f] {
			continue
		}
		if ok, _ := path.Match(g.Path, f); ok {
			found = true
			break
		}
	}
	want := g.Exists == nil || *g.Exists
	if found != want {
		return fail(g, "a new file matching %q: want %v, got %v", g.Path, want, found)
	}
	return pass(g)
}

func gradeRegex(g Grader, w Workspace) Outcome {
	text := w.Result.Reply
	if g.Target.Kind == "file" {
		data, err := os.ReadFile(filepath.Join(w.Dir, filepath.FromSlash(g.Target.Path)))
		if err != nil {
			return fail(g, "cannot read %s: %v", g.Target.Path, err)
		}
		text = string(data)
	}
	re, err := CompilePattern(g.Pattern, g.Flags)
	if err != nil {
		return fail(g, "bad pattern %q: %v", g.Pattern, err)
	}
	found, err := re.MatchString(text)
	if err != nil {
		return fail(g, "pattern %q could not be matched: %v", g.Pattern, err)
	}
	want := g.Match != "not_contains"
	if found != want {
		return fail(g, "pattern %q: want found %v, got %v", g.Pattern, want, found)
	}
	return pass(g)
}

// gradeTool matches the tool name without case, because the suite says Bash
// and omp calls the same tool bash.
func gradeTool(g Grader, w Workspace) Outcome {
	re, err := CompilePattern(g.InputMatch, "")
	if err != nil {
		return fail(g, "bad input_match %q: %v", g.InputMatch, err)
	}
	count := 0
	for _, c := range w.Result.Calls {
		if !strings.EqualFold(c.Tool, g.Tool) {
			continue
		}
		hit, err := re.MatchString(c.Input)
		if err != nil {
			return fail(g, "input_match %q could not be matched: %v", g.InputMatch, err)
		}
		if hit {
			count++
		}
	}
	lo := 1
	if g.Min != nil {
		lo = *g.Min
	}
	if count < lo || (g.Max != nil && count > *g.Max) {
		return fail(g, "%s calls matching %q: got %d", g.Tool, g.InputMatch, count)
	}
	return pass(g)
}

const judgePrompt = `You grade one reply from an AI agent against a rubric.

Rubric:
%s

Reply:
%s

Answer with PASS or FAIL as your very first word, then one short sentence why.`

func gradeLLM(g Grader, w Workspace, judge Judge) Outcome {
	ans, err := judge(fmt.Sprintf(judgePrompt, g.Body, w.Result.Reply))
	if err != nil {
		return fail(g, "judge failed: %v", err)
	}
	words := strings.Fields(ans)
	verdict := ""
	if len(words) > 0 {
		verdict = strings.ToUpper(strings.Trim(words[0], "*:.,!"))
	}
	switch verdict {
	case "PASS":
		return pass(g)
	case "FAIL":
		return fail(g, "judge: %s", strings.TrimSpace(ans))
	}
	return fail(g, "judge gave no PASS or FAIL: %q", strings.TrimSpace(ans))
}
