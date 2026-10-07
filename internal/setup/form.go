package setup

import (
	"strconv"

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

// formState holds the built form with the answer slots its fields write
// into: one bool per harness found, in harness order. Ask reads them back
// after the run; tests render the groups without running.
type formState struct {
	form    *huh.Form
	groups  []*huh.Group
	answers *Defaults
	install map[string]*bool
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

	questions := []struct {
		desc  string
		field huh.Field
	}{
		{"The language I use when I talk with you.",
			huh.NewInput().Title("Which language should I use when I talk with you?").
				Value(&d.Language)},
		{"Short replies for speed, or full sentences.",
			huh.NewSelect[string]().Title("Style: adhd or plain?").
				Options(huh.NewOptions("adhd", "plain")...).Value(&d.Style)},
		{"Anything about tone, in your own words. Optional.",
			huh.NewInput().Title("Anything about tone, in your own words? (optional)").
				Value(&d.Tone)},
		{"The language for files written to the repo.",
			huh.NewInput().Title("Language for files written to the repo?").
				Value(&d.RepoLanguage)},
		{"Who writes the code when a plan runs.",
			huh.NewSelect[string]().Title("Which build executor?").
				Options(executors...).Value(&d.BuildExecutor)},
		{"Who picks the model for background work.",
			huh.NewSelect[string]().Title("How should subagent models be picked?").
				Options(
					huh.NewOption("default (leave it to your own config)", "default"),
					huh.NewOption("split", "split"),
				).Value(&d.SubagentModels)},
		{"How much a plan spells out before it runs.",
			huh.NewSelect[string]().Title("How should a plan spell out?").
				Options(
					huh.NewOption("full (real code in every step, plan waits for a yes)", "full"),
					huh.NewOption("minimal (short steps, no code, build starts right away)", "minimal"),
				).Value(&d.PlanDepth)},
		{"How I ask you things while working.",
			huh.NewSelect[string]().Title("How should I ask you things?").
				Options(
					huh.NewOption("one (one question at a time)", "one"),
					huh.NewOption("probe (a round of questions, with a recommended answer each)", "probe"),
				).Value(&d.Questions)},
	}

	n := len(questions)
	if len(e.Harnesses) > 0 {
		n++
	}
	groups := make([]*huh.Group, 0, n)
	for i, q := range questions {
		groups = append(groups, huh.NewGroup(q.field).
			Title("acta setup").
			Description(stepLine(i+1, n, q.desc)))
	}

	install := map[string]*bool{}
	if len(e.Harnesses) > 0 {
		var harnessFields []huh.Field
		for _, h := range e.Harnesses {
			v := true
			install[h] = &v
			harnessFields = append(harnessFields,
				huh.NewConfirm().Title("Install the acta plugin into "+h+"?").
					Value(install[h]).Inline(true))
		}
		groups = append(groups, huh.NewGroup(harnessFields...).
			Title("acta setup").
			Description(stepLine(n, n, "Install the acta plugin into each tool found.")))
	}

	return &formState{
		form:    huh.NewForm(groups...).WithTheme(Theme()),
		groups:  groups,
		answers: &d,
		install: install,
	}
}

// stepLine joins the step count and the one-line description: "3/9. What
// this question means." huh shows the group description as one block, so
// both ride in it and every screen shows title, count and help together.
func stepLine(k, n int, desc string) string {
	return strconv.Itoa(k) + "/" + strconv.Itoa(n) + ". " + desc
}

// BuildForm returns the wizard form: one group per question in fixed order,
// then one harness group. Tests render its groups without running it.
func BuildForm(e Env) *huh.Form {
	return buildFormState(e).form
}

// FormGroups returns the wizard groups for tests: same order as the form
// shows, without running anything.
func FormGroups(e Env) []*huh.Group {
	return buildFormState(e).groups
}

// Ask collects the answers with one question per screen. It holds no rules:
// every decision lives in Plan. The block is mandatory, so the form never
// asks about it; it only shows the block and the file it goes to.
func Ask(e Env) (Answers, error) {
	a := Answers{User: e.Current, Install: map[string]bool{}}
	st := buildFormState(e)
	if err := st.form.Run(); err != nil {
		return Answers{}, err
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
