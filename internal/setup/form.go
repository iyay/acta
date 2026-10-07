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
		Questions:      d.Questions,
	}
	return a, nil
}
