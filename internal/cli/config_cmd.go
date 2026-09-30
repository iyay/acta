package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/iyay/acta/internal/theme"
	"github.com/iyay/acta/internal/voice"
)

const voiceUsage = "usage: acta voice show [--json] | acta voice set [--language L] [--style adhd|plain] [--tone T] [--clear-tone] [--repo-language L] [--executor subagent|dispatch|inline] [--subagent-models split] [--clear-subagent-models] [--theme NAME] [--clear-theme]"

func cmdVoice(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, voiceUsage)
		return exitBadInput
	}
	path, err := voice.Path()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	switch args[0] {
	case "show":
		fs := flag.NewFlagSet("voice show", flag.ContinueOnError)
		fs.SetOutput(stderr)
		asJSON := fs.Bool("json", false, "print JSON")
		if err := fs.Parse(args[1:]); err != nil || fs.NArg() != 0 {
			fmt.Fprintln(stderr, voiceUsage)
			return exitBadInput
		}
		v, exists, err := voice.Resolve()
		if err != nil {
			fmt.Fprintln(stderr, err)
			return exitBadInput
		}
		if *asJSON {
			return printJSON(stdout, stderr, map[string]any{
				"path": path, "exists": exists, "chat_language": v.ChatLanguage,
				"style": v.Style, "tone": v.Tone, "repo_language": v.RepoLanguage,
				"build_executor": v.BuildExecutor, "subagent_models": v.SubagentModels,
				"theme": v.Theme,
			})
		}
		fmt.Fprintf(stdout, "file: %s (exists: %v)\nchat_language: %s\nstyle: %s\nrepo_language: %s\n",
			path, exists, v.ChatLanguage, v.Style, v.RepoLanguage)
		if v.Tone != "" {
			fmt.Fprintf(stdout, "tone: %s\n", v.Tone)
		}
		if v.BuildExecutor != "" {
			fmt.Fprintf(stdout, "build_executor: %s\n", v.BuildExecutor)
		}
		if v.SubagentModels != "" {
			fmt.Fprintf(stdout, "subagent_models: %s\n", v.SubagentModels)
		}
		if v.Theme != "" {
			fmt.Fprintf(stdout, "theme: %s\n", v.Theme)
		}
		return exitOK
	case "set":
		fs := flag.NewFlagSet("voice set", flag.ContinueOnError)
		fs.SetOutput(stderr)
		lang := fs.String("language", "", "chat language, as a full name (Korean)")
		style := fs.String("style", "", "adhd or plain")
		tone := fs.String("tone", "", "how you want to be spoken to, in your own words")
		clearTone := fs.Bool("clear-tone", false, "remove the tone")
		repo := fs.String("repo-language", "", "language for files written to the repo")
		executor := fs.String("executor", "", "which executor runs the plan: subagent, dispatch or inline")
		models := fs.String("subagent-models", "", "how models are picked for subagents: split")
		clearModels := fs.Bool("clear-subagent-models", false, "remove the subagent_models setting")
		themeName := fs.String("theme", "", "the TUI color theme")
		clearTheme := fs.Bool("clear-theme", false, "go back to the default theme")
		if err := fs.Parse(args[1:]); err != nil || fs.NArg() != 0 {
			fmt.Fprintln(stderr, voiceUsage)
			return exitBadInput
		}
		if *lang == "" && *style == "" && *tone == "" && *repo == "" && *executor == "" && *models == "" && *themeName == "" && !*clearTone && !*clearModels && !*clearTheme {
			fmt.Fprintln(stderr, voiceUsage)
			return exitBadInput
		}
		v, _, err := voice.Resolve()
		if err != nil {
			// Never overwrite a file the user may still want to repair by hand.
			fmt.Fprintf(stderr, "%v\nfix or delete %s first\n", err, path)
			return exitBadInput
		}
		if *lang != "" {
			v.ChatLanguage = *lang
		}
		if *style != "" {
			v.Style = *style
		}
		if *repo != "" {
			v.RepoLanguage = *repo
		}
		if *clearTone {
			v.Tone = ""
		}
		if *tone != "" {
			v.Tone = *tone
		}
		if *executor != "" {
			v.BuildExecutor = *executor
		}
		// Clear first, set second, so one call can replace the value.
		if *clearModels {
			v.SubagentModels = ""
		}
		if *models != "" {
			v.SubagentModels = *models
		}
		// Clear first, set second, so one call can replace the value.
		if *clearTheme {
			v.Theme = ""
		}
		if *themeName != "" {
			// Check the name here, so a typo is answered now and not as a
			// silent fallback the next time the TUI opens.
			if _, err := theme.Load(*themeName); err != nil {
				// The whole run stops before the save, so the file keeps the
				// theme it had and the other flags in this call change
				// nothing either.
				fmt.Fprintln(stderr, err)
				return exitBadInput
			}
			v.Theme = *themeName
		}
		if err := voice.SaveResolved(v); err != nil {
			fmt.Fprintln(stderr, err)
			if errors.Is(err, voice.ErrBad) {
				return exitBadInput
			}
			return exitOther
		}
		fmt.Fprintln(stdout, path)
		return exitOK
	default:
		fmt.Fprintln(stderr, voiceUsage)
		return exitBadInput
	}
}
