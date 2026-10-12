package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// useExe points exePath at a binary path under a temp bin folder, so no
// test touches the real home or the real ~/.local.
func useExe(t *testing.T, exe string, err error) {
	t.Helper()
	old := exePath
	t.Cleanup(func() { exePath = old })
	exePath = func() (string, error) { return exe, err }
}

// unroff undoes the escapes the man page uses, so a test can compare the
// text with the table.
func unroff(s string) string {
	s = strings.ReplaceAll(s, `\&`, "")
	s = strings.ReplaceAll(s, `\-`, "-")
	s = strings.ReplaceAll(s, `\fB`, "")
	s = strings.ReplaceAll(s, `\fI`, "")
	s = strings.ReplaceAll(s, `\fR`, "")
	return strings.ReplaceAll(s, `\e`, `\`)
}

// manDoc is the man page cut into its parts: the commands (.SS), the
// flag tags of each section and the global option tags.
type manDoc struct {
	commands []string
	flags    []string // tags under a command
	global   []string // tags under OPTIONS
	text     string
}

func parseMan(roff string) manDoc {
	d := manDoc{text: roff}
	lines := strings.Split(roff, "\n")
	section := ""
	for i, l := range lines {
		switch {
		case strings.HasPrefix(l, ".SH "):
			section = strings.TrimPrefix(l, ".SH ")
		case strings.HasPrefix(l, ".SS "):
			d.commands = append(d.commands, strings.TrimPrefix(l, ".SS "))
		case l == ".TP" && i+1 < len(lines):
			tag := unroff(lines[i+1])
			if section == "OPTIONS" {
				d.global = append(d.global, tag)
			} else {
				d.flags = append(d.flags, tag)
			}
		}
	}
	return d
}

func TestManRoffMatchesTable(t *testing.T) {
	doc := parseMan(renderMan())

	var wantCmds, wantFlags []string
	for _, c := range helpTable {
		wantCmds = append(wantCmds, c.name)
		for _, f := range c.flags {
			head, _, _ := strings.Cut(f, "  ")
			wantFlags = append(wantFlags, head)
		}
	}
	sort.Strings(wantCmds)
	sort.Strings(wantFlags)
	gotCmds := append([]string(nil), doc.commands...)
	gotFlags := append([]string(nil), doc.flags...)
	sort.Strings(gotCmds)
	sort.Strings(gotFlags)
	if strings.Join(gotCmds, "|") != strings.Join(wantCmds, "|") {
		t.Errorf("commands in man page\n got %v\nwant %v", gotCmds, wantCmds)
	}
	if strings.Join(gotFlags, "|") != strings.Join(wantFlags, "|") {
		t.Errorf("flags in man page\n got %v\nwant %v", gotFlags, wantFlags)
	}

	// Every usage line, summary, flag meaning and example is in the text.
	text := unroff(doc.text)
	for _, c := range helpTable {
		for _, s := range append([]string{c.summary, c.example}, c.usage...) {
			if s != "" && !strings.Contains(text, s) {
				t.Errorf("%s: %q missing from man page", c.name, s)
			}
		}
		for _, f := range c.flags {
			_, meaning, _ := strings.Cut(f, "  ")
			if !strings.Contains(text, meaning) {
				t.Errorf("%s: flag meaning %q missing", c.name, meaning)
			}
		}
	}

	// The same global flags the top help lists.
	for _, want := range []string{"-root DIR", "-h, --help", "--version"} {
		found := false
		for _, g := range doc.global {
			if g == want {
				found = true
			}
		}
		if !found {
			t.Errorf("global flag %q missing, have %v", want, doc.global)
		}
	}
	if len(doc.global) != 3 {
		t.Errorf("global flags = %v, want exactly three", doc.global)
	}
}

func TestManRoffShape(t *testing.T) {
	roff := renderMan()
	for _, want := range []string{".TH ACTA 1", ".SH NAME", ".SH SYNOPSIS", ".SH OPTIONS"} {
		if !strings.Contains(roff, want) {
			t.Errorf("man page lacks %q", want)
		}
	}
	for _, g := range helpGroups {
		if !strings.Contains(roff, ".SH "+strings.ToUpper(g)+"\n") {
			t.Errorf("man page lacks a section for group %q", g)
		}
	}
	if strings.Contains(roff, ".SS voice") {
		t.Error("voice is not a command and must not be in the man page")
	}
}

func TestManEscape(t *testing.T) {
	cases := []struct{ in, want string }{
		{`a\b`, `a\eb`},
		{`.hidden`, `\&.hidden`},
		{`'quoted`, `\&'quoted`},
		{`plain .dot 'mid`, `plain .dot 'mid`},
		{`\.x`, `\e.x`},
	}
	for _, c := range cases {
		if got := manEscape(c.in); got != c.want {
			t.Errorf("manEscape(%q) = %q, want %q", c.in, got, c.want)
		}
	}
	if got := manFlagName("--type T"); got != `\-\-type T` {
		t.Errorf("manFlagName = %q", got)
	}
	if got := manFlagName("-m MESSAGE"); got != `\-m MESSAGE` {
		t.Errorf("manFlagName = %q", got)
	}
}

func TestManEscapeInRoff(t *testing.T) {
	roff := manRoff([]helpCmd{{
		name:    "odd",
		group:   "Board",
		summary: ".starts with a dot",
		usage:   []string{`'starts with a quote`, `has \ backslash`},
		flags:   []string{`--x  .flag text`, `-y  back\slash`},
		example: `.dot example`,
	}})
	for _, l := range strings.Split(roff, "\n") {
		if strings.HasPrefix(l, ".") || strings.HasPrefix(l, "'") {
			ok := false
			for _, m := range []string{".TH ", ".SH ", ".SS ", ".TP", ".PP", ".br", ".RS", ".RE", ".nf", ".fi"} {
				if strings.HasPrefix(l, m) {
					ok = true
				}
			}
			if !ok {
				t.Errorf("text line starts like a macro: %q", l)
			}
		}
	}
	for _, want := range []string{`\&.starts with a dot`, `\&'starts with a quote`, `has \e backslash`, `back\eslash`, `\-\-x`, `\-y`} {
		if !strings.Contains(roff, want) {
			t.Errorf("roff lacks %q", want)
		}
	}
}

func TestManInstallPath(t *testing.T) {
	root := t.TempDir()
	exe := filepath.Join(root, "prefix", "bin", "acta")
	useExe(t, exe, nil)

	got, err := manPath()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(root, "prefix", "share", "man", "man1", "acta.1")
	if got != want {
		t.Errorf("manPath = %q, want %q", got, want)
	}

	if err := installMan(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(want)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != renderMan() {
		t.Error("installed file is not the rendered man page")
	}
	info, _ := os.Stat(want)
	if info.Mode().Perm() != 0o644 {
		t.Errorf("file mode = %v, want 0644", info.Mode().Perm())
	}
	dir, _ := os.Stat(filepath.Dir(want))
	if dir.Mode().Perm() != 0o755 {
		t.Errorf("folder mode = %v, want 0755", dir.Mode().Perm())
	}

	// A second run replaces the file.
	if err := installMan(); err != nil {
		t.Errorf("second install: %v", err)
	}
}

func TestManInstallThroughSymlinkedBin(t *testing.T) {
	// The page goes next to the link, not next to the real binary.
	root := t.TempDir()
	real := filepath.Join(root, "real", "bin")
	link := filepath.Join(root, "link", "bin")
	for _, d := range []string{real, link} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	useExe(t, filepath.Join(link, "acta"), nil)
	if err := installMan(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "link", "share", "man", "man1", "acta.1")); err != nil {
		t.Errorf("page not next to the link: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "real", "share")); err == nil {
		t.Error("page went next to the real binary")
	}
}

// brokenInstalls has one case per way the man page write can fail. Each
// case sets up its own broken world on the test it is given.
var brokenInstalls = map[string]func(t *testing.T){
	"exePath error": func(t *testing.T) {
		useExe(t, "", errors.New("no executable"))
	},
	"file where the folder should be": func(t *testing.T) {
		root := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, "bin"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "share"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		useExe(t, filepath.Join(root, "bin", "acta"), nil)
	},
	"unwritable folder": func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("root can write anywhere")
		}
		root := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, "bin"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(root, 0o500); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.Chmod(root, 0o755) })
		useExe(t, filepath.Join(root, "bin", "acta"), nil)
	},
	"page path is a folder": func(t *testing.T) {
		root := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, "share", "man", "man1", "acta.1"), 0o755); err != nil {
			t.Fatal(err)
		}
		useExe(t, filepath.Join(root, "bin", "acta"), nil)
	},
}

func TestManInstallFailurePaths(t *testing.T) {
	for name, broken := range brokenInstalls {
		t.Run(name, func(t *testing.T) {
			broken(t)
			if err := installMan(); err == nil {
				t.Error("installMan returned nil for a broken install")
			}
		})
	}
}

// refreshEnv gives the refresh a temp home and no claude on PATH, so it
// touches nothing outside the test.
func refreshEnv(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("PATH", t.TempDir())
}

func TestManRefreshWarnsAndKeepsExit(t *testing.T) {
	for name, broken := range brokenInstalls {
		t.Run(name, func(t *testing.T) {
			refreshEnv(t)
			broken(t)
			var stderr bytes.Buffer
			if code := cmdRefreshPlugin(&stderr); code != exitOK {
				t.Errorf("exit = %d, want %d; stderr %q", code, exitOK, stderr.String())
			}
			out := strings.TrimRight(stderr.String(), "\n")
			if strings.Count(out, "\n") != 0 || !strings.Contains(out, "man page") {
				t.Errorf("want one man warning line, got %q", stderr.String())
			}
		})
	}
}

func TestManRefreshKeepsFailingExit(t *testing.T) {
	// The plugin refresh fails (home is a file) and the man page fails too:
	// the exit code stays the refresh one.
	t.Setenv("PATH", t.TempDir())
	homeFile := filepath.Join(t.TempDir(), "home")
	if err := os.WriteFile(homeFile, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", homeFile)
	useExe(t, "", errors.New("no executable"))
	var stderr bytes.Buffer
	if code := cmdRefreshPlugin(&stderr); code != exitBadInput {
		t.Errorf("exit = %d, want %d", code, exitBadInput)
	}
	if !strings.Contains(stderr.String(), refreshFailedMsg) {
		t.Errorf("refresh message missing: %q", stderr.String())
	}
}

func TestManRefreshWritesPage(t *testing.T) {
	refreshEnv(t)
	root := t.TempDir()
	useExe(t, filepath.Join(root, "bin", "acta"), nil)
	var stderr bytes.Buffer
	if code := cmdRefreshPlugin(&stderr); code != exitOK {
		t.Fatalf("exit = %d; stderr %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("unexpected stderr %q", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(root, "share", "man", "man1", "acta.1")); err != nil {
		t.Errorf("page missing after refresh: %v", err)
	}
}

func TestManSetupStep(t *testing.T) {
	// Good install: nothing printed. Broken install: one rail warning.
	root := t.TempDir()
	useExe(t, filepath.Join(root, "bin", "acta"), nil)
	var out bytes.Buffer
	setupManStep(&out)
	if out.Len() != 0 {
		t.Errorf("good install printed %q", out.String())
	}
	if _, err := os.Stat(filepath.Join(root, "share", "man", "man1", "acta.1")); err != nil {
		t.Errorf("page missing after setup step: %v", err)
	}

	for name, broken := range brokenInstalls {
		t.Run(name, func(t *testing.T) {
			broken(t)
			var out bytes.Buffer
			setupManStep(&out)
			s := out.String()
			if strings.Count(s, "\n") != 1 || !strings.Contains(s, "man page") || !strings.Contains(s, "│") {
				t.Errorf("want one rail warning line, got %q", s)
			}
		})
	}
}
