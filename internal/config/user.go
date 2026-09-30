package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// User holds how the agent should talk to the user: the chat language,
// the style, an optional tone, and the language for repo files. It lives in
// ~/.acta/config.yaml. An old ~/.acta/voice.yaml is moved there on the first
// read. The ~/.pm/voice.yaml from before the rename is only read when both
// are missing.
type User struct {
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

// ErrBadUser marks a setting the rules do not allow.
var ErrBadUser = errors.New("bad voice setting")

// UserDefault is used when there is no voice file.
func UserDefault() User {
	return User{ChatLanguage: "English", Style: "adhd", RepoLanguage: "English"}
}

// UserPath is PM_VOICE_FILE when set, else ~/.acta/config.yaml. Writes always
// go here, so one place holds the truth.
func UserPath() (string, error) {
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
// ResolveUser moves it to UserPath once.
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

// ResolveUser reads config.yaml. When it is missing, an old ~/.acta/voice.yaml
// is renamed to config.yaml first, so the user ends up with one file. When
// both are missing, the ~/.pm file from before the acta rename is read.
func ResolveUser() (User, bool, error) {
	path, err := UserPath()
	if err != nil {
		return UserDefault(), false, err
	}
	v, exists, err := LoadUser(path)
	if exists || err != nil || os.Getenv("PM_VOICE_FILE") != "" {
		return v, exists, err
	}
	voiceFile, err := voicePath()
	if err != nil {
		return UserDefault(), false, err
	}
	// A missing voice.yaml is fine: there was none, or another hook moved it
	// a moment ago. Any other failure means the file is still there, so read
	// it where it is and lose nothing. The next read tries the move again.
	if err := os.Rename(voiceFile, path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return LoadUser(voiceFile)
	}
	if v, exists, err := LoadUser(path); exists || err != nil {
		return v, exists, err
	}
	old, err := oldPath()
	if err != nil {
		return UserDefault(), false, err
	}
	return LoadUser(old)
}

// SaveUser writes to the new file, never to the old one.
func SaveUser(v User) error {
	path, err := UserPath()
	if err != nil {
		return err
	}
	return SaveUserFile(path, v)
}

// LoadUser reads the voice file. A missing file gives the defaults and
// exists=false. A broken file gives the defaults, exists=true and the error,
// so a hook can still print something useful.
func LoadUser(path string) (User, bool, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return UserDefault(), false, nil
	}
	if err != nil {
		return UserDefault(), true, err
	}
	var got User
	if err := yaml.Unmarshal(raw, &got); err != nil {
		return UserDefault(), true, fmt.Errorf("%s: %w", path, err)
	}
	v := fill(got)
	if err := v.Validate(); err != nil {
		return UserDefault(), true, fmt.Errorf("%s: %w", path, err)
	}
	return v, true, nil
}

// SaveUserFile checks v and writes it through a temp file, so a crash never
// leaves a half-written voice file.
func SaveUserFile(path string, v User) error {
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
func (v User) Validate() error {
	if v.Style != "adhd" && v.Style != "plain" {
		return fmt.Errorf("%w: style must be adhd or plain, not %q", ErrBadUser, v.Style)
	}
	for _, f := range []struct{ name, value string }{
		{"chat_language", v.ChatLanguage},
		{"repo_language", v.RepoLanguage},
	} {
		if f.value == "" || strings.ContainsAny(f.value, "\r\n") || len(f.value) > 40 {
			return fmt.Errorf("%w: %s must be one short line", ErrBadUser, f.name)
		}
	}
	if len(v.Tone) > 600 || strings.Count(v.Tone, "\n") > 7 {
		return fmt.Errorf("%w: tone must be at most 8 lines and 600 characters", ErrBadUser)
	}
	switch v.BuildExecutor {
	case "", "subagent", "dispatch", "inline":
	default:
		return fmt.Errorf("%w: build_executor must be subagent, dispatch or inline, not %q", ErrBadUser, v.BuildExecutor)
	}
	if v.SubagentModels != "" && v.SubagentModels != "split" {
		return fmt.Errorf("%w: subagent_models must be split or empty, not %q", ErrBadUser, v.SubagentModels)
	}
	return nil
}

// fill trims values and puts the default in every empty field but tone.
func fill(v User) User {
	d := UserDefault()
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
