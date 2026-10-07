package setup

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/huh"

	"github.com/iyay/acta/internal/config"
)

// Defaults holds the starting values the form shows: the current config
// where set, else the built-in defaults. FormDefaults keeps that rule in
// one place so Ask and tests read the same values.
type Defaults struct {
	Language       string
	Style          string
	Tone           string
	RepoLanguage   string
	BuildExecutor  string
	SubagentModels string
	PlanDepth      string
	Questions      string
	OfferDispatch  bool
}

// FormDefaults reads the defaults out of the current config. Dispatch is
// offered as an executor only inside a herdr pane, which HERDR_ENV=1 says.
func FormDefaults(e Env) Defaults {
	herdr := isHerdr()
	c := e.Current
	if c.ChatLanguage == "" {
		c.ChatLanguage = "English"
	}
	if c.Style == "" {
		c.Style = "adhd"
	}
	if c.RepoLanguage == "" {
		c.RepoLanguage = "English"
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
		Questions:      c.Questions,
		OfferDispatch:  herdr,
	}
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
// title, its step count and a one-line plain description saying what the
// question means. Dispatch is offered only when HERDR_ENV=1 says so.
func buildFormState(e Env) *formState {
	d := FormDefaults(e)

	executors := []huh.Option[string]{
		huh.NewOption("subagent", "subagent"),
		huh.NewOption("inline", "inline"),
	}
	if d.OfferDispatch {
		executors = append(executors, huh.NewOption("dispatch", "dispatch"))
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
		{"Which language should I use when I talk with you?",
			"The language I use when I talk with you.",
			huh.NewInput().Value(&d.Language), text(&d.Language)},
		{"Style: adhd or plain?",
			"Short replies for speed, or full sentences.",
			huh.NewSelect[string]().
				Options(huh.NewOptions("adhd", "plain")...).Value(&d.Style), text(&d.Style)},
		{"Anything about tone, in your own words? (optional)",
			"Anything about tone, in your own words. Optional.",
			huh.NewInput().Value(&d.Tone), text(&d.Tone)},
		{"Language for files written to the repo?",
			"The language for files written to the repo.",
			huh.NewInput().Value(&d.RepoLanguage), text(&d.RepoLanguage)},
		{"Which build executor?",
			"Who writes the code when a plan runs.",
			huh.NewSelect[string]().Options(executors...).Value(&d.BuildExecutor), text(&d.BuildExecutor)},
		{"How should subagent models be picked?",
			"Who picks the model for background work.",
			huh.NewSelect[string]().
				Options(
					huh.NewOption("default (leave it to your own config)", "default"),
					huh.NewOption("split", "split"),
				).Value(&d.SubagentModels), text(&d.SubagentModels)},
		{"How much should a plan spell out?",
			"How much a plan spells out before it runs.",
			huh.NewSelect[string]().
				Options(
					huh.NewOption("full (real code in every step, plan waits for a yes)", "full"),
					huh.NewOption("minimal (short steps, no code, build starts right away)", "minimal"),
				).Value(&d.PlanDepth), text(&d.PlanDepth)},
		{"How should I ask you things?",
			"How I ask you things while working.",
			huh.NewSelect[string]().
				Options(
					huh.NewOption("one (one question at a time)", "one"),
					huh.NewOption("probe (a round of questions, with a recommended answer each)", "probe"),
				).Value(&d.Questions), text(&d.Questions)},
	}

	n := len(specs)
	if len(e.Harnesses) > 0 {
		n++
	}
	st := &formState{answers: &d, install: map[string]*bool{}}
	add := func(title, desc string, answer func() string, fields ...huh.Field) {
		k := len(st.questions) + 1
		g := huh.NewGroup(fields...).
			WithTheme(Theme()).
			Title(ActiveTitle(title, k, n)).
			Description(ActiveDescription(desc))
		st.groups = append(st.groups, g)
		st.questions = append(st.questions, question{title: title, group: g, answer: answer})
	}
	for _, q := range specs {
		add(q.title, q.desc, q.answer, q.field)
	}

	if len(e.Harnesses) > 0 {
		var harnessFields []huh.Field
		for _, h := range e.Harnesses {
			v := true
			st.install[h] = &v
			harnessFields = append(harnessFields,
				huh.NewConfirm().Title(h).Value(st.install[h]).Inline(true))
		}
		add("Install the acta plugin?", "Install the acta plugin into each tool found.",
			func() string {
				var parts []string
				for _, h := range e.Harnesses {
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

	for _, h := range e.Harnesses {
		a.Install[h] = *st.install[h]
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
		Questions:      d.Questions,
	}
	return a, nil
}
