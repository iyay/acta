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
	button := lipgloss.NewStyle().MarginRight(2)

	paint := func(f *huh.FieldStyles) {
		f.Title = title
		f.NoteTitle = title
		f.Description = dim
		f.ErrorIndicator = danger
		f.ErrorMessage = danger
		f.SelectSelector = accent.SetString("› ")
		f.Option = lipgloss.NewStyle()
		f.MultiSelectSelector = accent.SetString("› ")
		f.SelectedOption = accent.SetString("●")
		f.SelectedPrefix = accent.SetString("✓ ")
		f.UnselectedOption = lipgloss.NewStyle().SetString("○")
		f.UnselectedPrefix = dim.SetString("• ")
		f.FocusedButton = button.Foreground(setupAccent).Bold(true).SetString("●")
		f.BlurredButton = button.Foreground(setupDim).SetString("○")
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
	t.Focused.Base = railBase()
	// Fields stack with one line between them, so the rail never breaks.
	t.FieldSeparator = lipgloss.NewStyle().SetString("\n")
	t.Blurred.Base = railBase()
	t.Blurred.NextIndicator = lipgloss.NewStyle()
	t.Blurred.PrevIndicator = lipgloss.NewStyle()
	t.Blurred.MultiSelectSelector = lipgloss.NewStyle().SetString("  ")

	// The group header carries the diamond title and the rail description
	// already styled, so the group styles stay plain except the dim grey.
	t.Group.Title = lipgloss.NewStyle()
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

// railBase puts a grey rail on the left of a field: "│" then two spaces, so
// every field line lines up under the question title.
func railBase() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.Border{Left: "│"}, false, false, false, true).
		BorderForeground(setupDim).
		PaddingLeft(2)
}

func dimText(s string) string {
	return lipgloss.NewStyle().Foreground(setupDim).Render(s)
}

// RailOpen starts the run: the title line and one empty rail line.
func RailOpen() string {
	return dimText("┌") + "  " + lipgloss.NewStyle().Bold(true).Render("acta setup") + "\n" + dimText("│") + "\n"
}

// RailLine prints one line of text on the rail.
func RailLine(s string) string {
	return dimText("│") + "  " + s + "\n"
}

// RailProblem prints a line that needs the user's eye: a red ▲ on the rail.
func RailProblem(s string) string {
	mark := lipgloss.NewStyle().Foreground(setupDanger).Render("▲")
	return dimText("│") + "  " + mark + " " + s + "\n"
}

// ActiveTitle is the first line of the question on screen: a blue diamond,
// the bold title and the step count.
func ActiveTitle(title string, k, n int) string {
	diamond := lipgloss.NewStyle().Foreground(setupAccent).Render("◆")
	return fmt.Sprintf("%s  %s  (%d/%d)", diamond, lipgloss.NewStyle().Bold(true).Render(title), k, n)
}

// ActiveDescription is the dim one-line help under the title, on the rail.
func ActiveDescription(desc string) string {
	return dimText("│") + "  " + desc
}

// Collapsed is an answered question that stays on screen: a hollow diamond
// with the title, the answer in dim grey on the rail, and a spacer line.
func Collapsed(title, answer string) string {
	if answer == "" {
		answer = "(none)"
	}
	return dimText("◇") + "  " + title + "\n" +
		dimText("│  "+answer) + "\n" +
		dimText("│") + "\n"
}

// DoctorSummary compresses the install checks onto the rail: one line with
// the count when every result is ok, else one problem line per result that
// is not ok with its fix. The wizard carries on either way.
func DoctorSummary(rs []doctor.Result) string {
	bad := false
	for _, r := range rs {
		if r.Level != doctor.OK {
			bad = true
			break
		}
	}
	if !bad {
		return RailLine(fmt.Sprintf("✓ install checks ok (%d)", len(rs))) + dimText("│") + "\n"
	}
	var b strings.Builder
	for _, r := range rs {
		if r.Level == doctor.OK {
			continue
		}
		b.WriteString(RailProblem(fmt.Sprintf("%s: %s", r.Name, r.Msg)))
		if r.Fix != "" {
			b.WriteString(RailLine("fix: " + r.Fix))
		}
	}
	return b.String() + dimText("│") + "\n"
}

// BlockLines names each file the acta block goes to, one rail line per
// file, so the wizard never prints the whole block text.
func BlockLines(paths []string) string {
	var b strings.Builder
	for _, p := range paths {
		b.WriteString(RailLine("acta block → " + p))
	}
	return b.String()
}

// InstallLine prints one harness install result on the rail: a ✓ with the
// harness name when it ran, else a red ▲ with the harness name and the
// command to run by hand.
func InstallLine(harness string, argv []string, err error) string {
	if err == nil {
		return RailLine("✓ " + harness)
	}
	return RailProblem(harness + ": " + strings.Join(argv, " "))
}

// SummaryBox closes the wizard: where the config landed, then the closing
// corner of the rail with the hint to open the TUI.
func SummaryBox(configPath string) string {
	if configPath == "" {
		configPath = "(unknown)"
	}
	return dimText("│") + "\n" +
		dimText("◇") + "  Setup done\n" +
		RailLine("config: "+configPath) +
		dimText("└") + "  Run acta in a repo to browse specs, plans and bugs.\n"
}
