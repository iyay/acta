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

// Voice is one user's setting. It lives in ~/.pm/voice.yaml.
type Voice struct {
	ChatLanguage string `yaml:"chat_language"`
	Style        string `yaml:"style"`
	Tone         string `yaml:"tone,omitempty"`
	RepoLanguage string `yaml:"repo_language"`
}

// ErrBad marks a setting the rules do not allow.
var ErrBad = errors.New("bad voice setting")

// Default is used when there is no voice file.
func Default() Voice {
	return Voice{ChatLanguage: "English", Style: "adhd", RepoLanguage: "English"}
}

// Path is PM_VOICE_FILE when set, else ~/.pm/voice.yaml.
func Path() (string, error) {
	if p := os.Getenv("PM_VOICE_FILE"); p != "" {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".pm", "voice.yaml"), nil
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
	return nil
}

// fill trims values and puts the default in every empty field but tone.
func fill(v Voice) Voice {
	d := Default()
	v.ChatLanguage = strings.TrimSpace(v.ChatLanguage)
	v.RepoLanguage = strings.TrimSpace(v.RepoLanguage)
	v.Style = strings.TrimSpace(v.Style)
	v.Tone = strings.TrimSpace(v.Tone)
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
