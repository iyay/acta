package setup

import (
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

// Ask collects the answers with one huh field per group. It holds no rules:
// every decision lives in Plan. The block is mandatory, so the form never
// asks about it; it only shows the block and the file it goes to.
func Ask(e Env) (Answers, error) {
	d := FormDefaults(e)
	a := Answers{User: e.Current, Install: map[string]bool{}}

	executors := []huh.Option[string]{
		huh.NewOption("subagent", "subagent"),
		huh.NewOption("inline", "inline"),
	}
	if d.OfferDispatch {
		executors = append(executors, huh.NewOption("dispatch", "dispatch"))
	}

	fields := []huh.Field{
		huh.NewInput().Title("Which language should I use when I talk with you?").
			Value(&d.Language),
		huh.NewSelect[string]().Title("Style: adhd or plain?").
			Options(huh.NewOptions("adhd", "plain")...).Value(&d.Style),
		huh.NewInput().Title("Anything about tone, in your own words? (optional)").
			Value(&d.Tone),
		huh.NewInput().Title("Language for files written to the repo?").
			Value(&d.RepoLanguage),
		huh.NewSelect[string]().Title("Which build executor?").
			Options(executors...).Value(&d.BuildExecutor),
		huh.NewSelect[string]().Title("How should subagent models be picked?").
			Options(
				huh.NewOption("default (leave it to your own config)", "default"),
				huh.NewOption("split", "split"),
			).Value(&d.SubagentModels),
		huh.NewSelect[string]().Title("How much should a plan spell out?").
			Options(
				huh.NewOption("full (real code in every step, plan waits for a yes)", "full"),
				huh.NewOption("minimal (short steps, no code, build starts right away)", "minimal"),
			).Value(&d.PlanDepth),
		huh.NewSelect[string]().Title("How should I ask you things?").
			Options(
				huh.NewOption("one (one question at a time)", "one"),
				huh.NewOption("probe (a round of questions, with a recommended answer each)", "probe"),
			).Value(&d.Questions),
	}
	for _, h := range e.Harnesses {
		yes := true
		fields = append(fields,
			huh.NewConfirm().Title("Install the acta plugin into "+h+"?").
				Value(&yes).Inline(true))
		a.Install[h] = false
	}

	form := huh.NewForm(huh.NewGroup(fields...))
	if err := form.Run(); err != nil {
		return Answers{}, err
	}

	// huh wrote into d through the pointers above; the confirms wrote
	// into their own locals, which are read back here in harness order.
	i := len(fields) - len(e.Harnesses)
	for _, h := range e.Harnesses {
		if c, ok := fields[i].(*huh.Confirm); ok {
			a.Install[h] = c.GetValue().(bool)
		}
		i++
	}

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
