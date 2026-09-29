// Package theme holds the colors the TUI paints with. A theme has the same
// shape as a terminal color scheme, so a user can copy one straight from a
// site like terminalcolors.com.
package theme

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Default is the theme used when the user has not picked one.
const Default = "tokyo-night"

// ErrUnknown marks a name that is neither a user file nor a built-in.
var ErrUnknown = errors.New("unknown theme")

// Theme is one set of colors. An empty BG means the terminal theme: every
// slot is then the terminal's own ANSI color and nothing is hex.
type Theme struct {
	Name                     string
	BG, FG                   string
	SelectionBG, SelectionFG string
	ANSI                     [16]string
}

// A name becomes part of a file path, so it may only hold safe letters.
var nameRE = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

var hexRE = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

type file struct {
	BG          string   `yaml:"bg"`
	FG          string   `yaml:"fg"`
	SelectionBG string   `yaml:"selection_bg"`
	SelectionFG string   `yaml:"selection_fg"`
	ANSI        []string `yaml:"ansi"`
}

// Load finds a theme by name. A user file wins over a built-in, so a user
// can change a built-in by copying it.
func Load(name string) (Theme, error) {
	if name == "" {
		name = Default
	}
	if !nameRE.MatchString(name) {
		return Theme{}, fmt.Errorf("theme %q: name may only use a-z, 0-9 and -", name)
	}
	if raw, err := os.ReadFile(userPath(name)); err == nil {
		return Parse(name, raw)
	} else if !errors.Is(err, os.ErrNotExist) {
		return Theme{}, fmt.Errorf("theme %q: %w", name, err)
	}
	if t, ok := Builtin(name); ok {
		return t, nil
	}
	return Theme{}, fmt.Errorf("theme %q: %w", name, ErrUnknown)
}

// userPath is where a user theme lives. A HOME that is not a full path would
// read from wherever acta was started, so it gives a path that is never there.
func userPath(name string) string {
	home, err := os.UserHomeDir()
	if err != nil || !filepath.IsAbs(home) {
		return ""
	}
	return filepath.Join(home, ".acta", "themes", name+".yaml")
}

// Parse reads one theme file and checks every color in it.
func Parse(name string, raw []byte) (Theme, error) {
	var f file
	if err := yaml.Unmarshal(raw, &f); err != nil {
		return Theme{}, fmt.Errorf("theme %q: %w", name, err)
	}
	if len(f.ANSI) != 16 {
		return Theme{}, fmt.Errorf("theme %q: ansi needs 16 colors, has %d", name, len(f.ANSI))
	}
	t := Theme{Name: name, BG: f.BG, FG: f.FG, SelectionBG: f.SelectionBG, SelectionFG: f.SelectionFG}
	copy(t.ANSI[:], f.ANSI)
	for _, c := range []struct{ key, value string }{{"bg", t.BG}, {"fg", t.FG}} {
		if !hexRE.MatchString(c.value) {
			return Theme{}, fmt.Errorf("theme %q: %s %q is not #rrggbb", name, c.key, c.value)
		}
	}
	for _, c := range []struct{ key, value string }{{"selection_bg", t.SelectionBG}, {"selection_fg", t.SelectionFG}} {
		if c.value != "" && !hexRE.MatchString(c.value) {
			return Theme{}, fmt.Errorf("theme %q: %s %q is not #rrggbb", name, c.key, c.value)
		}
	}
	for i, c := range t.ANSI {
		if !hexRE.MatchString(c) {
			return Theme{}, fmt.Errorf("theme %q: ansi %d %q is not #rrggbb", name, i, c)
		}
	}
	return t, nil
}

// Builtin gives a theme that ships with acta.
func Builtin(name string) (Theme, bool) {
	t, ok := builtins[name]
	if ok {
		t.Name = name
	}
	return t, ok
}

// Names lists the built-in themes in a fixed order, for help text.
func Names() []string {
	out := make([]string, 0, len(builtins))
	for n := range builtins {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// Dark says if the theme is a dark one. The terminal theme cannot know, so
// it takes the answer the terminal gave.
func (t Theme) Dark(fallback bool) bool {
	if t.BG == "" {
		return fallback
	}
	r, _ := strconv.ParseUint(t.BG[1:3], 16, 8)
	g, _ := strconv.ParseUint(t.BG[3:5], 16, 8)
	b, _ := strconv.ParseUint(t.BG[5:7], 16, 8)
	return 0.299*float64(r)+0.587*float64(g)+0.114*float64(b) < 128
}
