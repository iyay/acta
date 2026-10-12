package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/iyay/acta/internal/setup"
	"github.com/iyay/acta/plugin"
)

// manIntro is the DESCRIPTION text of the man page.
const manIntro = "acta shows and changes the planning files of the repo it runs in. " +
	"With no arguments it opens the board."

// manEscape makes one line of table text safe inside a roff page. A
// backslash would start an escape, so it is doubled. A dot or a quote at the
// start of a line would make roff read the line as a macro, so a zero-width
// mark goes in front.
func manEscape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\e`)
	if strings.HasPrefix(s, ".") || strings.HasPrefix(s, "'") {
		s = `\&` + s
	}
	return s
}

// manFlagName escapes a flag name. A dash that opens the flag is a minus
// sign, not a hyphen, so it becomes \- and copies out of the page right.
func manFlagName(s string) string {
	rest := strings.TrimLeft(s, "-")
	dashes := len(s) - len(rest)
	return strings.Repeat(`\-`, dashes) + manEscape(rest)
}

// manBold puts one line of text in bold.
func manBold(s string) string {
	return `\fB` + manEscape(s) + `\fR`
}

// manRoff builds the man page text from the help table.
func manRoff(cmds []helpCmd) string {
	var b strings.Builder
	fmt.Fprintf(&b, ".TH ACTA 1 \"\" \"acta %s\" \"User Commands\"\n", manEscape(plugin.Version()))
	b.WriteString(".SH NAME\n")
	b.WriteString("acta \\- show and change the planning files of a repo\n")
	b.WriteString(".SH SYNOPSIS\n")
	b.WriteString("\\fBacta\\fR [\\fB\\-root\\fR \\fIDIR\\fR]\n.br\n")
	b.WriteString("\\fBacta\\fR \\fIcommand\\fR [\\fIargs\\fR]\n.br\n")
	b.WriteString("\\fBacta help\\fR [\\fIcommand\\fR]\n.br\n")
	b.WriteString("\\fBacta \\-\\-version\\fR\n")
	b.WriteString(".SH DESCRIPTION\n" + manEscape(manIntro) + "\n")
	b.WriteString(".SH OPTIONS\n")
	for _, g := range [][2]string{
		{"-root DIR", "Planning root folder (default .acta; also ACTA_ROOT, PM_ROOT, or root in .acta.yaml or .pm.yaml). Most commands take it too."},
		{"-h, --help", "Show the help."},
		{"--version", "Print the version."},
	} {
		manTag(&b, g[0], g[1])
	}
	for _, group := range helpGroups {
		b.WriteString(".SH " + strings.ToUpper(group) + "\n")
		for _, c := range cmds {
			if c.group == group {
				manCmd(&b, c)
			}
		}
	}
	return b.String()
}

// manTag writes one flag with its meaning.
func manTag(b *strings.Builder, flag, meaning string) {
	b.WriteString(".TP\n")
	b.WriteString(`\fB` + manFlagName(flag) + `\fR` + "\n")
	b.WriteString(manEscape(meaning) + "\n")
}

// manCmd writes the part of the page for one command.
func manCmd(b *strings.Builder, c helpCmd) {
	b.WriteString(".SS " + c.name + "\n")
	b.WriteString(manEscape(c.summary) + "\n")
	b.WriteString(".PP\n")
	for i, u := range c.usage {
		if i > 0 {
			b.WriteString(".br\n")
		}
		b.WriteString(manBold(u) + "\n")
	}
	for _, f := range c.flags {
		flag, meaning, _ := strings.Cut(f, "  ")
		manTag(b, flag, meaning)
	}
	if c.example != "" {
		b.WriteString(".PP\nExample:\n.br\n" + manBold(c.example) + "\n")
	}
}

// renderMan is the whole acta.1 page.
func renderMan() string {
	return manRoff(helpTable)
}

// manPath is where the page goes: next to the binary, in the share folder of
// the same prefix. Links are not followed, so a page for a binary reached
// through a link lands beside the link, where the user's man path looks.
func manPath() (string, error) {
	exe, err := exePath()
	if err != nil {
		return "", err
	}
	return filepath.Clean(filepath.Join(filepath.Dir(exe), "..", "share", "man", "man1", "acta.1")), nil
}

// installMan writes the man page to its place and makes the folders.
func installMan() error {
	path, err := manPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(renderMan()), 0o644)
}

// manWarning is the one line shown when the page could not be written.
func manWarning(err error) string {
	return "could not write the man page: " + err.Error()
}

// setupManStep writes the page during acta setup. A failure shows one line
// in the setup rail and setup goes on.
func setupManStep(stdout io.Writer) {
	if err := installMan(); err != nil {
		fmt.Fprint(stdout, setup.RailProblem(manWarning(err)))
	}
}
