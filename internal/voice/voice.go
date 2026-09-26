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

// Voice is one user's setting. It now lives in ~/.acta/voice.yaml; the old
// ~/.pm/voice.yaml is only read when the new file is missing.
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

// Path is PM_VOICE_FILE when set, else ~/.acta/voice.yaml. Writes always go
// here, so one place holds the truth.
func Path() (string, error) {
	if p := os.Getenv("PM_VOICE_FILE"); p != "" {
		return p, nil
	}
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

// Resolve reads the new file first and the old one when the new file is
// missing, so existing users keep their setting after the rename.
func Resolve() (Voice, bool, error) {
	path, err := Path()
	if err != nil {
		return Default(), false, err
	}
	v, exists, err := Load(path)
	if exists || err != nil || os.Getenv("PM_VOICE_FILE") != "" {
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
