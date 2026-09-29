// Package voice holds how the agent should talk to the user: the chat
// language, the style, an optional tone, and the language for repo files.
package voice

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Voice is one user's setting. It lives in ~/.acta/config.yaml. An old
// ~/.acta/voice.yaml is moved there on the first read. The ~/.pm/voice.yaml
// from before the rename is only read when both are missing.
type Voice struct {
	ChatLanguage   string `yaml:"chat_language"`
	Style          string `yaml:"style"`
	Tone           string `yaml:"tone,omitempty"`
	RepoLanguage   string `yaml:"repo_language"`
	BuildExecutor  string `yaml:"build_executor,omitempty"`
	SubagentModels string `yaml:"subagent_models,omitempty"`
	// Theme names the TUI colors. Empty means the default theme. It is not
	// checked here: a theme file can be deleted or edited at any time, and
	// the hooks read this file on every call, so they must keep working.
	Theme string `yaml:"theme,omitempty"`
}

// ErrBad marks a setting the rules do not allow.
var ErrBad = errors.New("bad voice setting")

// Default is used when there is no voice file.
func Default() Voice {
	return Voice{ChatLanguage: "English", Style: "adhd", RepoLanguage: "English"}
}

// Path is PM_VOICE_FILE when set, else ~/.acta/config.yaml. Writes always go
// here, so one place holds the truth.
func Path() (string, error) {
	if p := os.Getenv("PM_VOICE_FILE"); p != "" {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".acta", "config.yaml"), nil
}

// voicePath is the name the file had before it held more than the voice.
// Resolve moves it to Path once.
func voicePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".acta", "voice.yaml"), nil
}

// oldPath is where the setting lived before the rename. Reads fall back to
// it, writes never touch it.
func oldPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".pm", "voice.yaml"), nil
}

// Resolve reads config.yaml. When it is missing, an old ~/.acta/voice.yaml
// is renamed to config.yaml first, so the user ends up with one file. When
// both are missing, the ~/.pm file from before the acta rename is read.
func Resolve() (Voice, bool, error) {
	path, err := Path()
	if err != nil {
		return Default(), false, err
	}
	v, exists, err := Load(path)
	if exists || err != nil || os.Getenv("PM_VOICE_FILE") != "" {
		return v, exists, err
	}
	voiceFile, err := voicePath()
	if err != nil {
		return Default(), false, err
	}
	// A missing voice.yaml is fine: there was none, or another hook moved it
	// a moment ago. Any other failure means the file is still there, so read
	// it where it is and lose nothing. The next read tries the move again.
	if err := os.Rename(voiceFile, path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Load(voiceFile)
	}
	if v, exists, err := Load(path); exists || err != nil {
		return v, exists, err
	}
	old, err := oldPath()
	if err != nil {
		return Default(), false, err
	}
	return Load(old)
}

// SaveResolved writes to the new file, never to the old one.
func SaveResolved(v Voice) error {
	path, err := Path()
	if err != nil {
		return err
	}
	return Save(path, v)
}

// Load reads the voice file. A missing file gives the defaults and
// exists=false. A broken file gives the defaults, exists=true and the error,
// so a hook can still print something useful.
func Load(path string) (Voice, bool, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Default(), false, nil
	}
	if err != nil {
		return Default(), true, err
	}
	var got Voice
	if err := yaml.Unmarshal(raw, &got); err != nil {
		return Default(), true, fmt.Errorf("%s: %w", path, err)
	}
	v := fill(got)
	if err := v.Validate(); err != nil {
		return Default(), true, fmt.Errorf("%s: %w", path, err)
	}
	return v, true, nil
}

// Save checks v and writes it through a temp file, so a crash never leaves a
// half-written voice file.
func Save(path string, v Voice) error {
	v = fill(v)
	if err := v.Validate(); err != nil {
		return err
	}
	out, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, out, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Validate keeps the values short enough to fit the session rules.
func (v Voice) Validate() error {
	if v.Style != "adhd" && v.Style != "plain" {
		return fmt.Errorf("%w: style must be adhd or plain, not %q", ErrBad, v.Style)
	}
	for _, f := range []struct{ name, value string }{
		{"chat_language", v.ChatLanguage},
		{"repo_language", v.RepoLanguage},
	} {
		if f.value == "" || strings.ContainsAny(f.value, "\r\n") || len(f.value) > 40 {
			return fmt.Errorf("%w: %s must be one short line", ErrBad, f.name)
		}
	}
	if len(v.Tone) > 600 || strings.Count(v.Tone, "\n") > 7 {
		return fmt.Errorf("%w: tone must be at most 8 lines and 600 characters", ErrBad)
	}
	switch v.BuildExecutor {
	case "", "subagent", "dispatch", "inline":
	default:
		return fmt.Errorf("%w: build_executor must be subagent, dispatch or inline, not %q", ErrBad, v.BuildExecutor)
	}
	if v.SubagentModels != "" && v.SubagentModels != "split" {
		return fmt.Errorf("%w: subagent_models must be split or empty, not %q", ErrBad, v.SubagentModels)
	}
	return nil
}

// fill trims values and puts the default in every empty field but tone.
func fill(v Voice) Voice {
	d := Default()
	v.ChatLanguage = strings.TrimSpace(v.ChatLanguage)
	v.RepoLanguage = strings.TrimSpace(v.RepoLanguage)
	v.Style = strings.TrimSpace(v.Style)
	v.Tone = strings.TrimSpace(v.Tone)
	v.BuildExecutor = strings.TrimSpace(v.BuildExecutor)
	v.SubagentModels = strings.TrimSpace(v.SubagentModels)
	if v.ChatLanguage == "" {
		v.ChatLanguage = d.ChatLanguage
	}
	if v.RepoLanguage == "" {
		v.RepoLanguage = d.RepoLanguage
	}
	if v.Style == "" {
		v.Style = d.Style
	}
	return v
}
