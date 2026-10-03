package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/theme"
)

const configUsage = "usage: acta config show [--json] | acta config set [--language L] [--style adhd|plain] [--tone T] [--clear-tone] [--repo-language L] [--executor subagent|dispatch|inline] [--plan-depth minimal|full] [--coding-guide lean|off] [--subagent-models split|default] [--clear-subagent-models] [--theme NAME] [--clear-theme] | acta config set --repo [--repo-language L] [--executor E] [--plan-depth D] [--coding-guide G]"

func cmdConfig(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, configUsage)
		return exitBadInput
	}
	path, err := config.UserPath()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	switch args[0] {
	case "show":
		fs := flag.NewFlagSet("config show", flag.ContinueOnError)
		fs.SetOutput(stderr)
		asJSON := fs.Bool("json", false, "print JSON")
		if err := fs.Parse(args[1:]); err != nil || fs.NArg() != 0 {
			fmt.Fprintln(stderr, configUsage)
			return exitBadInput
		}
		v, read, exists, err := config.ResolveUserFile()
		if err != nil {
			fmt.Fprintln(stderr, err)
			return exitBadInput
		}
		cwd, _ := os.Getwd()
		cfg, err := config.Load(cwd, "")
		if err != nil {
			fmt.Fprintln(stderr, err)
			return exitBadInput
		}
		v, from, err := config.MergeRepo(v, cfg.RepoRoot)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return exitBadInput
		}
		depth, depthMark := v.PlanDepth, ""
		if depth == "" {
			// No file sets it. Say so, or setup thinks full was chosen.
			depth, depthMark = "full", " (default)"
		}
		guide, guideMark := v.CodingGuide, ""
		if guide == "" {
			// Same here: an unset guide is lean, and show says so.
			guide, guideMark = "lean", " (default)"
		}
		// Name the values the repo set, so the user knows which file to edit.
		mark := func(k string) string {
			if from[k] {
				return " (repo)"
			}
			return ""
		}
		var fromRepo []string
		for _, k := range config.RepoKeys {
			if from[k] {
				fromRepo = append(fromRepo, k)
			}
		}
		if *asJSON {
			return printJSON(stdout, stderr, map[string]any{
				"path": read, "writes": path, "exists": exists, "chat_language": v.ChatLanguage,
				"style": v.Style, "tone": v.Tone, "repo_language": v.RepoLanguage,
				"build_executor": v.BuildExecutor, "subagent_models": v.SubagentModels,
				"theme": v.Theme, "plan_depth": depth, "coding_guide": guide, "from_repo": fromRepo,
			})
		}
		// The values can come from an old file. Name it, and say where the
		// next config set goes, so the user edits the right file.
		state := fmt.Sprintf("exists: %v", exists)
		if read != path {
			state += ", old file; config set writes " + path
		}
		fmt.Fprintf(stdout, "file: %s (%s)\nchat_language: %s\nstyle: %s\nrepo_language: %s%s\n",
			read, state, v.ChatLanguage, v.Style, v.RepoLanguage, mark("repo_language"))
		if v.Tone != "" {
			fmt.Fprintf(stdout, "tone: %s\n", v.Tone)
		}
		if v.BuildExecutor != "" {
			fmt.Fprintf(stdout, "build_executor: %s%s\n", v.BuildExecutor, mark("build_executor"))
		}
		fmt.Fprintf(stdout, "plan_depth: %s%s%s\n", depth, mark("plan_depth"), depthMark)
		fmt.Fprintf(stdout, "coding_guide: %s%s%s\n", guide, mark("coding_guide"), guideMark)
		if v.SubagentModels != "" {
			fmt.Fprintf(stdout, "subagent_models: %s\n", v.SubagentModels)
		}
		if v.Theme != "" {
			fmt.Fprintf(stdout, "theme: %s\n", v.Theme)
		}
		return exitOK
	case "set":
		fs := flag.NewFlagSet("config set", flag.ContinueOnError)
		fs.SetOutput(stderr)
		lang := fs.String("language", "", "chat language, as a full name (Korean)")
		style := fs.String("style", "", "adhd or plain")
		tone := fs.String("tone", "", "how you want to be spoken to, in your own words")
		clearTone := fs.Bool("clear-tone", false, "remove the tone")
		repo := fs.String("repo-language", "", "language for files written to the repo")
		executor := fs.String("executor", "", "which executor runs the plan: subagent, dispatch or inline")
		models := fs.String("subagent-models", "", "how models are picked for subagents: split, or default to leave it to your own config")
		clearModels := fs.Bool("clear-subagent-models", false, "remove the subagent_models setting")
		themeName := fs.String("theme", "", "the TUI color theme")
		clearTheme := fs.Bool("clear-theme", false, "go back to the default theme")
		depth := fs.String("plan-depth", "", "how much a plan spells out: minimal or full")
		guide := fs.String("coding-guide", "", "the coding guide: lean or off")
		repoOnly := fs.Bool("repo", false, "save to .acta.yaml in this repo instead of your own config")
		if err := fs.Parse(args[1:]); err != nil || fs.NArg() != 0 {
			fmt.Fprintln(stderr, configUsage)
			return exitBadInput
		}
		if *lang == "" && *style == "" && *tone == "" && *repo == "" && *executor == "" && *depth == "" && *guide == "" && *models == "" && *themeName == "" && !*clearTone && !*clearModels && !*clearTheme {
			fmt.Fprintln(stderr, configUsage)
			return exitBadInput
		}
		if *repoOnly {
			return setRepo(stdout, stderr, map[string]string{"repo_language": *repo, "build_executor": *executor, "plan_depth": *depth, "coding_guide": *guide},
				*lang != "" || *style != "" || *tone != "" || *models != "" || *themeName != "" || *clearTone || *clearModels || *clearTheme)
		}
		v, read, _, err := config.ResolveUserFile()
		if err != nil {
			// Never overwrite a file the user may still want to repair by hand.
			// Name the file that failed: it can be an old one, not config.yaml.
			fmt.Fprintf(stderr, "%v\nfix or delete %s first\n", err, read)
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
		if *depth != "" {
			v.PlanDepth = *depth
		}
		if *guide != "" {
			v.CodingGuide = *guide
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
		if err := config.SaveUser(v); err != nil {
			fmt.Fprintln(stderr, err)
			if errors.Is(err, config.ErrBadUser) {
				return exitBadInput
			}
			return exitOther
		}
		fmt.Fprintln(stdout, path)
		return exitOK
	default:
		fmt.Fprintln(stderr, configUsage)
		return exitBadInput
	}
}

// setRepo saves repo keys to .acta.yaml. Personal flags are refused, since
// that file is committed and would change how the agent talks to everyone
// who clones the repo.
func setRepo(stdout, stderr io.Writer, want map[string]string, personal bool) int {
	if personal {
		fmt.Fprintf(stderr, "--repo takes only --repo-language, --executor, --plan-depth and --coding-guide; set the others without --repo\n")
		return exitBadInput
	}
	set := map[string]string{}
	for k, val := range want {
		if val != "" {
			set[k] = val
		}
	}
	if len(set) == 0 {
		fmt.Fprintln(stderr, configUsage)
		return exitBadInput
	}
	cwd, _ := os.Getwd()
	cfg, err := config.Load(cwd, "")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitBadInput
	}
	// With only .pm.yaml, a new .acta.yaml would win and hide every other
	// setting in the old file.
	_, errActa := os.Stat(filepath.Join(cfg.RepoRoot, ".acta.yaml"))
	if _, errPM := os.Stat(filepath.Join(cfg.RepoRoot, ".pm.yaml")); errPM == nil && errActa != nil {
		fmt.Fprintf(stderr, "this repo still uses .pm.yaml; rename it to .acta.yaml first\n")
		return exitBadInput
	}
	// Check the values before the write, so a typo never lands in the file.
	check := config.UserDefault()
	for _, k := range config.RepoKeys {
		switch val := set[k]; {
		case val == "":
		case k == "repo_language":
			check.RepoLanguage = val
		case k == "build_executor":
			check.BuildExecutor = val
		case k == "plan_depth":
			check.PlanDepth = val
		case k == "coding_guide":
			check.CodingGuide = val
		}
	}
	if err := check.Validate(); err != nil {
		fmt.Fprintln(stderr, err)
		return exitBadInput
	}
	path, err := config.SaveRepoUser(cfg.RepoRoot, set)
	if err != nil {
		fmt.Fprintln(stderr, err)
		if errors.Is(err, config.ErrBadUser) {
			return exitBadInput
		}
		return exitOther
	}
	fmt.Fprintln(stdout, path)
	return exitOK
}
