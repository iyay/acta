package setup

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/huh"

	"github.com/iyay/acta/internal/config"
)

// Defaults holds the starting values the form shows: the current config
// where set. Text answers start empty on a first run; selects start on
// their first option. FormDefaults keeps that rule in
// one place so Ask and tests read the same values.
type Defaults struct {
	Language       string
	Style          string
	Tone           string
	RepoLanguage   string
	BuildExecutor  string
	SubagentModels string
	PlanDepth      string
	CommitHistory  string
	Questions      string
	OfferDispatch  bool
}

// FormDefaults reads the defaults out of the current config. Dispatch is
// offered as an executor only inside a herdr pane, which HERDR_ENV=1 says.
func FormDefaults(e Env) Defaults {
	herdr := isHerdr()
	c := e.Current
	if c.Style == "" {
		c.Style = "adhd"
	}
	if c.BuildExecutor == "" {
		c.BuildExecutor = "subagent"
	}
	if c.SubagentModels == "" {
		c.SubagentModels = "default"
	}
	if c.PlanDepth == "" {
		c.PlanDepth = "full"
	}
	if c.CommitHistory == "" {
		c.CommitHistory = "tidy"
	}
	if c.Questions == "" {
		c.Questions = "one"
	}
	return Defaults{
		Language:       c.ChatLanguage,
		Style:          c.Style,
		Tone:           c.Tone,
		RepoLanguage:   c.RepoLanguage,
		BuildExecutor:  c.BuildExecutor,
		SubagentModels: c.SubagentModels,
		PlanDepth:      c.PlanDepth,
		CommitHistory:  c.CommitHistory,
		Questions:      c.Questions,
		OfferDispatch:  herdr,
	}
}

// required refuses an empty answer. The two language answers have no
// built-in value, so the user must type one.
func required(s string) error {
	if strings.TrimSpace(s) == "" {
		return errors.New("required")
	}
	return nil
}

// question is one screen: its title, the group that draws it, and how to
// read the answer back as one short line once it is answered.
type question struct {
	title  string
	group  *huh.Group
	answer func() string
}

// formState holds the questions with the answer slots their fields write
// into: one bool per harness found, in harness order. Ask reads them back
// after each run; tests render the groups without running.
type formState struct {
	questions []question
	groups    []*huh.Group
	answers   *Defaults
	install   map[string]*bool
}

// buildFormState puts one question on each screen in fixed order, then one
// harness screen with a yes/no per harness found. Every group carries the
// title, its step count and, when it has something to add, a one-line
// plain description. Dispatch is offered only when HERDR_ENV=1 says so.
func buildFormState(e Env) *formState {
	d := FormDefaults(e)

	executors := []huh.Option[string]{
		huh.NewOption("subagent: helper agents in the same session", "subagent"),
		huh.NewOption("inline: the main agent writes the code", "inline"),
	}
	if d.OfferDispatch {
		executors = append(executors, huh.NewOption("dispatch: an omp agent in its own herdr tab", "dispatch"))
	}

	// The field itself carries no title: the group header above it is the
	// title line, so the title shows once, on the rail.
	text := func(v *string) func() string { return func() string { return *v } }
	specs := []struct {
		title  string
		desc   string
		field  huh.Field
		answer func() string
	}{
		{"Chat language",
			"The language the agent chats in. Code and files use the repo language.",
			huh.NewInput().Placeholder("English").Validate(required).Value(&d.Language), text(&d.Language)},
		{"Reply style", "",
			huh.NewSelect[string]().
				Options(
					huh.NewOption("adhd: short, one next step at the end", "adhd"),
					huh.NewOption("plain: normal paragraphs", "plain"),
				).Value(&d.Style), text(&d.Style)},
		{"Tone",
			"Optional. Your own words, like \"casual, no jargon\".",
			huh.NewInput().Value(&d.Tone), text(&d.Tone)},
		{"Repo language",
			"Code, comments, commits, specs and plans.",
			huh.NewInput().Placeholder("English").Validate(required).Value(&d.RepoLanguage), text(&d.RepoLanguage)},
		{"Build executor", "",
			huh.NewSelect[string]().Options(executors...).Value(&d.BuildExecutor), text(&d.BuildExecutor)},
		{"Subagent models",
			"Claude Code only.",
			huh.NewSelect[string]().
				Options(
					huh.NewOption("default: your own config decides", "default"),
					huh.NewOption("split: sonnet writes code, the rest use a stronger model", "split"),
				).Value(&d.SubagentModels), text(&d.SubagentModels)},
		{"Plan detail", "",
			huh.NewSelect[string]().
				Options(
					huh.NewOption("full: real code in every step, waits for your yes", "full"),
					huh.NewOption("minimal: short steps, build starts at once", "minimal"),
				).Value(&d.PlanDepth), text(&d.PlanDepth)},
		{"Commit history", "",
			huh.NewSelect[string]().
				Options(
					huh.NewOption("tidy: one commit per task, straight history", "tidy"),
					huh.NewOption("full: every commit plus a merge commit", "full"),
				).Value(&d.CommitHistory), text(&d.CommitHistory)},
		{"Questions", "",
			huh.NewSelect[string]().
				Options(
					huh.NewOption("one: one question at a time", "one"),
					huh.NewOption("probe: a batch, each with the agent's pick", "probe"),
				).Value(&d.Questions), text(&d.Questions)},
	}

	// Only a tool that still needs the plugin gets a row, and only when a
	// plugin dir lets the wizard install it. The rest are shown by Apply.
	var askable []string
	if e.PluginDir != "" {
		for _, h := range e.Harnesses {
			if !e.Installed[h] {
				askable = append(askable, h)
			}
		}
	}
	n := len(specs)
	if len(askable) > 0 {
		n++
	}
	st := &formState{answers: &d, install: map[string]*bool{}}
	add := func(title, desc string, answer func() string, fields ...huh.Field) {
		k := len(st.questions) + 1
		g := huh.NewGroup(fields...).
			WithTheme(Theme()).
			Title(ActiveTitle(title, k, n))
		// A question with nothing to add gets no description line.
		if desc != "" {
			g = g.Description(ActiveDescription(desc))
		}
		st.groups = append(st.groups, g)
		st.questions = append(st.questions, question{title: title, group: g, answer: answer})
	}
	for _, q := range specs {
		add(q.title, q.desc, q.answer, q.field)
	}

	if len(askable) > 0 {
		// Pad every title to the longest name plus two spaces so the
		// Yes/No marks start in one column with a gap after the name.
		width := 0
		for _, h := range askable {
			width = max(width, len(h))
		}
		var harnessFields []huh.Field
		for _, h := range askable {
			v := true
			st.install[h] = &v
			harnessFields = append(harnessFields,
				huh.NewConfirm().Title(fmt.Sprintf("%-*s", width+2, h)).Value(st.install[h]).Inline(true))
		}
		add("Plugin install", "One row per tool found.",
			func() string {
				var parts []string
				for _, h := range askable {
					yn := "no"
					if *st.install[h] {
						yn = "yes"
					}
					parts = append(parts, h+": "+yn)
				}
				return strings.Join(parts, ", ")
			}, harnessFields...)
	}
	return st
}

// FormGroups returns the wizard groups for tests: same order as the form
// shows, without running anything.
func FormGroups(e Env) []*huh.Group {
	return buildFormState(e).groups
}

// Ask collects the answers with one question per screen. It holds no rules:
// every decision lives in Plan. The block is mandatory, so the form never
// asks about it; it only shows the block and the file it goes to.
func Ask(e Env, out io.Writer) (Answers, error) {
	a := Answers{User: e.Current, Install: map[string]bool{}}
	st := buildFormState(e)
	// One form per question: huh wipes the question when it is answered, and
	// the collapsed line takes its place, so the answers pile up on the rail.
	for _, q := range st.questions {
		if err := huh.NewForm(q.group).WithTheme(Theme()).Run(); err != nil {
			return Answers{}, err
		}
		fmt.Fprint(out, Collapsed(q.title, q.answer()))
	}

	for h, v := range st.install {
		a.Install[h] = *v
	}
	d := st.answers
	a.User = config.User{
		ChatLanguage:   d.Language,
		Style:          d.Style,
		Tone:           d.Tone,
		RepoLanguage:   d.RepoLanguage,
		BuildExecutor:  d.BuildExecutor,
		SubagentModels: d.SubagentModels,
		PlanDepth:      d.PlanDepth,
		CommitHistory:  d.CommitHistory,
		Questions:      d.Questions,
	}
	// The override questions depend on the answers just given, so they come
	// as a second step after the fixed screens.
	var err error
	a.Overrides, a.Repo, err = askOverrides(a.User, e.RepoRoot, out, runGroup)
	if err != nil {
		return Answers{}, err
	}
	return a, nil
}

// runGroup shows one override screen and waits for the answer.
func runGroup(q *overrideQuestion) error {
	return huh.NewForm(q.group).WithTheme(Theme()).Run()
}

// overrideQuestion is one select about one repo key: the override it asks
// about, the screen, and the slot the answer is written into.
type overrideQuestion struct {
	override config.Override
	group    *huh.Group
	choice   *string
}

// overrideSentence tells the user what the repo file does to this key and
// what their own value is.
func overrideSentence(o config.Override) string {
	yours := fmt.Sprintf("Yours is %q.", o.Yours)
	if o.Yours == "" {
		yours = "Yours is not set."
	}
	return fmt.Sprintf("%s is %q in this repo's .acta.yaml (shared with everyone who clones it). %s", o.Key, o.Repo, yours)
}

// overrideQuestions builds one select per key the repo file changes. It
// builds none outside a repo. A repo file that cannot be read is an error.
func overrideQuestions(u config.User, repoRoot string) ([]*overrideQuestion, error) {
	if repoRoot == "" {
		return nil, nil
	}
	list, err := config.Overrides(u, repoRoot)
	if err != nil {
		return nil, err
	}
	var out []*overrideQuestion
	for i, o := range list {
		choice := RepoKeep
		opts := []huh.Option[string]{huh.NewOption("keep: this repo keeps "+o.Repo, RepoKeep)}
		// With no value of its own there is nothing to write for yours.
		if o.Yours != "" {
			opts = append(opts, huh.NewOption("yours: write "+o.Yours+" into .acta.yaml", RepoYours))
		}
		remove := "remove: drop it from .acta.yaml"
		if o.Yours != "" {
			remove += ", your own value applies"
		}
		opts = append(opts, huh.NewOption(remove, RepoRemove))
		g := huh.NewGroup(huh.NewSelect[string]().Options(opts...).Value(&choice)).
			WithTheme(Theme()).
			Title(ActiveTitle(o.Key, i+1, len(list))).
			Description(ActiveDescription(overrideSentence(o)))
		out = append(out, &overrideQuestion{override: o, group: g, choice: &choice})
	}
	return out, nil
}

// OverrideGroups returns the override screens for tests, without running
// anything.
func OverrideGroups(u config.User, repoRoot string) ([]*huh.Group, error) {
	qs, err := overrideQuestions(u, repoRoot)
	var out []*huh.Group
	for _, q := range qs {
		out = append(out, q.group)
	}
	return out, err
}

// askOverrides runs the override step with the user values just chosen. A
// repo file that cannot be read costs one line and no questions: the rest of
// setup goes on. run shows one screen; tests swap it for a fake.
func askOverrides(u config.User, repoRoot string, out io.Writer, run func(*overrideQuestion) error) ([]config.Override, map[string]string, error) {
	qs, err := overrideQuestions(u, repoRoot)
	if err != nil {
		fmt.Fprint(out, RailProblem("skipped the repo override questions, .acta.yaml could not be read: "+err.Error()))
		return nil, nil, nil
	}
	var list []config.Override
	choices := map[string]string{}
	for _, q := range qs {
		if err := run(q); err != nil {
			return nil, nil, err
		}
		list = append(list, q.override)
		choices[q.override.Key] = *q.choice
		fmt.Fprint(out, Collapsed(q.override.Key+" in .acta.yaml", *q.choice))
	}
	return list, choices, nil
}
