package main

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/iyay/acta/internal/voice"
)

const voiceUsage = "usage: pmb voice show [--json] | pmb voice set [--language L] [--style adhd|plain] [--tone T] [--clear-tone] [--repo-language L]"

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
		v, exists, err := voice.Load(path)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return exitBadInput
		}
		if *asJSON {
			return printJSON(stdout, stderr, map[string]any{
				"path": path, "exists": exists, "chat_language": v.ChatLanguage,
				"style": v.Style, "tone": v.Tone, "repo_language": v.RepoLanguage,
			})
		}
		fmt.Fprintf(stdout, "file: %s (exists: %v)\nchat_language: %s\nstyle: %s\nrepo_language: %s\n",
			path, exists, v.ChatLanguage, v.Style, v.RepoLanguage)
		if v.Tone != "" {
			fmt.Fprintf(stdout, "tone: %s\n", v.Tone)
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
		if err := fs.Parse(args[1:]); err != nil || fs.NArg() != 0 {
			fmt.Fprintln(stderr, voiceUsage)
			return exitBadInput
		}
		if *lang == "" && *style == "" && *tone == "" && *repo == "" && !*clearTone {
			fmt.Fprintln(stderr, voiceUsage)
			return exitBadInput
		}
		v, _, err := voice.Load(path)
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
		if err := voice.Save(path, v); err != nil {
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
