package setup_test

import (
	"errors"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/iyay/acta/internal/doctor"
	"github.com/iyay/acta/internal/setup"
)

// collectStyles walks every struct under v and returns each lipgloss.Style
// it finds, so the theme tests see the same styles the form renders with.
func collectStyles(t *testing.T, v reflect.Value) []lipgloss.Style {
	t.Helper()
	styleType := reflect.TypeOf(lipgloss.Style{})
	var out []lipgloss.Style
	var walk func(reflect.Value)
	walk = func(v reflect.Value) {
		if !v.IsValid() {
			return
		}
		if v.Type() == styleType {
			if v.CanInterface() {
				out = append(out, v.Interface().(lipgloss.Style))
			}
			return
		}
		switch v.Kind() {
		case reflect.Ptr:
			if !v.IsNil() {
				walk(v.Elem())
			}
		case reflect.Struct:
			for i := range v.NumField() {
				walk(v.Field(i))
			}
		}
	}
	walk(v)
	return out
}

// hexLuminance is the WCAG relative luminance of a #rrggbb color.
func hexLuminance(t *testing.T, hex string) float64 {
	t.Helper()
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		t.Fatalf("color %q is not #rrggbb", hex)
	}
	var lin [3]float64
	for i := range 3 {
		n, err := strconv.ParseUint(hex[2*i:2*i+2], 16, 8)
		if err != nil {
			t.Fatalf("color %q is not hex: %v", hex, err)
		}
		c := float64(n) / 255
		if c <= 0.03928 {
			lin[i] = c / 12.92
		} else {
			lin[i] = math.Pow((c+0.055)/1.055, 2.4)
		}
	}
	return 0.2126*lin[0] + 0.7152*lin[1] + 0.0722*lin[2]
}

// contrastRatio is the WCAG contrast ratio of two #rrggbb colors.
func contrastRatio(t *testing.T, a, b string) float64 {
	t.Helper()
	la, lb := hexLuminance(t, a), hexLuminance(t, b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// TestDimContrast keeps dim text readable where the terminal lifts dim
// colors: every color the theme uses holds ratio 1.6 or more against both a
// dark and a light background (wiki ghostty-minimum-contrast).
func TestDimContrast(t *testing.T) {
	styles := collectStyles(t, reflect.ValueOf(setup.Theme()))
	if len(styles) < 40 {
		t.Fatalf("found %d styles, the walker misses theme fields", len(styles))
	}
	seen := map[lipgloss.AdaptiveColor]bool{}
	var pairs []lipgloss.AdaptiveColor
	for _, s := range styles {
		fg := s.GetForeground()
		if fg == nil {
			continue
		}
		if _, ok := fg.(lipgloss.NoColor); ok {
			continue
		}
		ac, ok := fg.(lipgloss.AdaptiveColor)
		if !ok {
			t.Fatalf("foreground %T (%v) is not AdaptiveColor: one dark and one light value per color", fg, fg)
		}
		if !seen[ac] {
			seen[ac] = true
			pairs = append(pairs, ac)
		}
	}
	if len(pairs) == 0 {
		t.Fatal("theme sets no colors at all")
	}
	if len(pairs) > 3 {
		t.Errorf("found %d distinct colors %v, want at most 3 from the one palette in look.go", len(pairs), pairs)
	}
	for _, ac := range pairs {
		for _, hex := range []string{ac.Light, ac.Dark} {
			for _, bg := range []string{"#1a1b26", "#ffffff"} {
				if r := contrastRatio(t, hex, bg); r < 1.6 {
					t.Errorf("color %s has ratio %.2f against %s, want >= 1.6", hex, r, bg)
				}
			}
		}
	}
}

// TestNoBackgrounds keeps the mono look flat: no style the form renders
// with may carry a filled background color.
func TestNoBackgrounds(t *testing.T) {
	styles := collectStyles(t, reflect.ValueOf(setup.Theme()))
	if len(styles) < 40 {
		t.Fatalf("found %d styles, the walker misses theme fields", len(styles))
	}
	for i, s := range styles {
		bg := s.GetBackground()
		if bg == nil {
			continue
		}
		if _, ok := bg.(lipgloss.NoColor); !ok {
			t.Errorf("style %d has background %T (%v), want none", i, bg, bg)
		}
	}
}

func TestDoctorSummary(t *testing.T) {
	t.Run("all ok prints one line with the count", func(t *testing.T) {
		rs := []doctor.Result{
			{Name: "binary", Level: doctor.OK, Msg: "/tmp/acta"},
			{Name: "harness", Level: doctor.OK, Msg: "claude enabled"},
			{Name: "files", Level: doctor.OK, Msg: "every file has an id"},
		}
		if got := setup.DoctorSummary(rs); got != "✓ install checks ok (3)\n" {
			t.Errorf("got %q", got)
		}
	})
	t.Run("empty counts as ok", func(t *testing.T) {
		if got := setup.DoctorSummary(nil); got != "✓ install checks ok (0)\n" {
			t.Errorf("got %q", got)
		}
	})
	t.Run("one failing shows only the not-ok results", func(t *testing.T) {
		rs := []doctor.Result{
			{Name: "binary", Level: doctor.OK, Msg: "/tmp/acta"},
			{Name: "harness", Level: doctor.Warn, Msg: "omp link missing", Fix: "omp plugin install"},
			{Name: "files", Level: doctor.Fail, Msg: "no id in plans/foo.md"},
		}
		want := "warn harness: omp link missing\nfix: omp plugin install\nfail files: no id in plans/foo.md\n"
		if got := setup.DoctorSummary(rs); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestBlockLines(t *testing.T) {
	got := setup.BlockLines([]string{"CLAUDE.md", "AGENTS.md"})
	want := "acta block → CLAUDE.md\nacta block → AGENTS.md\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got := setup.BlockLines(nil); got != "" {
		t.Errorf("no block files want no lines, got %q", got)
	}
}

func TestInstallLines(t *testing.T) {
	t.Run("ok install names the harness", func(t *testing.T) {
		got := setup.InstallLine("claude", []string{"claude", "plugin", "add", "<dir>"}, nil)
		if got != "✓ claude\n" {
			t.Errorf("got %q", got)
		}
	})
	t.Run("failed install names the harness and the command", func(t *testing.T) {
		got := setup.InstallLine("omp", []string{"omp", "plugin", "install", "<dir>"}, errors.New("exit 1"))
		want := "✗ omp: omp plugin install <dir>\n"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestSummaryBox(t *testing.T) {
	t.Run("names config, installs, block files and next step", func(t *testing.T) {
		box := setup.SummaryBox("/home/u/.acta/config.yaml",
			[]string{"claude", "omp"}, []string{"CLAUDE.md"}, "run `acta --help`")
		for _, want := range []string{
			"config: /home/u/.acta/config.yaml", "installed: claude, omp",
			"block: CLAUDE.md", "next: run `acta --help`",
		} {
			if !strings.Contains(box, want) {
				t.Errorf("summary box misses %q:\n%s", want, box)
			}
		}
	})
	t.Run("empty installs and blocks still print a line each", func(t *testing.T) {
		box := setup.SummaryBox("cfg", nil, nil, "next")
		for _, want := range []string{"config: cfg", "installed: none", "block: none", "next: next"} {
			if !strings.Contains(box, want) {
				t.Errorf("summary box misses %q:\n%s", want, box)
			}
		}
	})
}
