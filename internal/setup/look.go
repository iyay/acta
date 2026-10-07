package setup

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/iyay/acta/internal/doctor"
)

// The setup palette holds every color the wizard screens use: dim grey for
// descriptions and answered lines, one calm blue for the focus marker, the
// picked option and the ✓ marks, and red for errors. Each color is adaptive
// (one value for light terminals, one for dark) and every value keeps a
// WCAG contrast ratio of 1.6 or more against both #1a1b26 and #ffffff, so a
// terminal that lifts dim colors keeps the text readable.
var (
	setupDim    = lipgloss.AdaptiveColor{Light: "#6F6F6F", Dark: "#A0A0A0"}
	setupAccent = lipgloss.AdaptiveColor{Light: "#2F6FED", Dark: "#7AA2F7"}
	setupDanger = lipgloss.AdaptiveColor{Light: "#C5362B", Dark: "#F26D5F"}
)

// Theme is the setup wizard look: almost all grey, question text in the
// terminal's normal color, one calm blue accent, bold colorless titles and
// no filled backgrounds. It starts from huh.ThemeBase so the huh layout
// stays stock; only colors, markers and the title weight change.
func Theme() *huh.Theme {
	t := huh.ThemeBase()

	dim := lipgloss.NewStyle().Foreground(setupDim)
	accent := lipgloss.NewStyle().Foreground(setupAccent)
	danger := lipgloss.NewStyle().Foreground(setupDanger)
	title := lipgloss.NewStyle().Bold(true)
	button := lipgloss.NewStyle().Padding(0, 2).MarginRight(1)

	paint := func(f *huh.FieldStyles) {
		f.Title = title
		f.NoteTitle = title
		f.Description = dim
		f.ErrorIndicator = danger
		f.ErrorMessage = danger
		f.SelectSelector = accent.SetString("› ")
		f.Option = lipgloss.NewStyle()
		f.MultiSelectSelector = accent.SetString("› ")
		f.SelectedOption = accent
		f.SelectedPrefix = accent.SetString("✓ ")
		f.UnselectedOption = lipgloss.NewStyle()
		f.UnselectedPrefix = dim.SetString("• ")
		f.FocusedButton = button.Foreground(setupAccent).Bold(true)
		f.BlurredButton = button.Foreground(setupDim)
		f.Next = f.FocusedButton
		f.Directory = accent
		f.File = lipgloss.NewStyle()
		f.TextInput.Cursor = accent
		f.TextInput.Placeholder = dim
		f.TextInput.Prompt = accent
		f.TextInput.Text = lipgloss.NewStyle()
	}
	paint(&t.Focused)
	paint(&t.Blurred)
	t.Blurred.Base = t.Blurred.Base.BorderStyle(lipgloss.HiddenBorder())
	t.Blurred.NextIndicator = lipgloss.NewStyle()
	t.Blurred.PrevIndicator = lipgloss.NewStyle()
	t.Blurred.MultiSelectSelector = lipgloss.NewStyle().SetString("  ")

	t.Group.Title = title
	t.Group.Description = dim

	t.Help.ShortKey = accent
	t.Help.FullKey = accent
	t.Help.ShortDesc = dim
	t.Help.FullDesc = dim
	t.Help.ShortSeparator = dim
	t.Help.FullSeparator = dim
	t.Help.Ellipsis = dim

	return t
}

// DoctorSummary compresses the install checks: one line with the count when
// every result is ok, else one line per result that is not ok with its fix.
// The wizard carries on either way.
func DoctorSummary(rs []doctor.Result) string {
	bad := false
	for _, r := range rs {
		if r.Level != doctor.OK {
			bad = true
			break
		}
	}
	if !bad {
		return fmt.Sprintf("✓ install checks ok (%d)\n", len(rs))
	}
	var b strings.Builder
	for _, r := range rs {
		if r.Level == doctor.OK {
			continue
		}
		fmt.Fprintf(&b, "%s %s: %s\n", r.Level, r.Name, r.Msg)
		if r.Fix != "" {
			fmt.Fprintf(&b, "fix: %s\n", r.Fix)
		}
	}
	return b.String()
}

// BlockLines names each file the acta block went to, one line per file, so
// the wizard never prints the whole block text.
func BlockLines(paths []string) string {
	var b strings.Builder
	for _, p := range paths {
		fmt.Fprintf(&b, "acta block → %s\n", p)
	}
	return b.String()
}

// InstallLine prints one harness install result: a ✓ with the harness name
// when it ran, else a ✗ with the harness name and the command to run by
// hand.
func InstallLine(harness string, argv []string, err error) string {
	if err == nil {
		return fmt.Sprintf("✓ %s\n", harness)
	}
	return fmt.Sprintf("✗ %s: %s\n", harness, strings.Join(argv, " "))
}

// SummaryBox closes the wizard: where the config landed, what got
// installed, which files got the block, and what to run next.
func SummaryBox(configPath string, installed []string, blockFiles []string, next string) string {
	join := func(ss []string) string {
		if len(ss) == 0 {
			return "none"
		}
		return strings.Join(ss, ", ")
	}
	body := strings.Join([]string{
		"Setup done.",
		"config: " + configPath,
		"installed: " + join(installed),
		"block: " + join(blockFiles),
		"next: " + next,
	}, "\n")
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(setupDim).
		Padding(0, 1)
	return box.Render(body) + "\n"
}
