package theme

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func writeUserTheme(t *testing.T, home, name, body string) {
	t.Helper()
	dir := filepath.Join(home, ".acta", "themes")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

const goodBody = `bg: "#101010"
fg: "#eeeeee"
ansi: ["#000000","#111111","#222222","#333333","#444444","#555555","#666666","#777777",
       "#888888","#999999","#aaaaaa","#bbbbbb","#cccccc","#dddddd","#eeeeee","#ffffff"]
`

func TestLoadEmptyNameGivesTokyoNight(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	got, err := Load("")
	if err != nil || got.Name != "tokyo-night" || got.BG != "#1a1b26" {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestEveryBuiltinLoads(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	want := []string{"catppuccin-latte", "catppuccin-mocha", "dracula", "gruvbox-dark", "terminal", "tokyo-night", "tokyo-night-day"}
	if got := strings.Join(Names(), ","); got != strings.Join(want, ",") {
		t.Fatalf("Names() = %s", got)
	}
	for _, n := range want {
		th, err := Load(n)
		if err != nil || th.Name != n {
			t.Fatalf("%s: %+v, %v", n, th, err)
		}
	}
}

// Every theme Load hands out has to be paintable, so each built-in color has
// to be a real hex value. The terminal theme is the one allowed to be empty.
func TestEveryBuiltinCarriesSixteenGoodColors(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	for _, n := range Names() {
		th, err := Load(n)
		if err != nil {
			t.Fatalf("%s: %v", n, err)
		}
		if n == "terminal" {
			continue
		}
		for _, c := range slices.Concat([]string{th.BG, th.FG, th.SelectionBG, th.SelectionFG}, th.ANSI[:]) {
			if !hexRE.MatchString(c) {
				t.Errorf("%s: color %q is not #rrggbb", n, c)
			}
		}
		if th.ANSI[15] == "" {
			t.Errorf("%s: ansi slot 15 is empty", n)
		}
	}
}

func TestTerminalThemeHasNoHex(t *testing.T) {
	th, _ := Builtin("terminal")
	if th.BG != "" || th.FG != "" || th.ANSI != [16]string{} {
		t.Fatalf("terminal carries colors: %+v", th)
	}
}

func TestUserFileWinsOverBuiltin(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeUserTheme(t, home, "dracula", goodBody)
	got, err := Load("dracula")
	if err != nil || got.BG != "#101010" {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestUserFileWithNewName(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeUserTheme(t, home, "mine", goodBody)
	got, err := Load("mine")
	if err != nil || got.Name != "mine" || got.ANSI[15] != "#ffffff" {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestUnknownNameNamesIt(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	_, err := Load("nope")
	if !errors.Is(err, ErrUnknown) || !strings.Contains(err.Error(), `"nope"`) {
		t.Fatalf("err = %v", err)
	}
}

func TestBadNamesRefused(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	// A file one folder up must never be reachable by name.
	if err := os.MkdirAll(filepath.Join(home, ".acta"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".acta", "x.yaml"), []byte(goodBody), 0o644); err != nil {
		t.Fatal(err)
	}
	// A bad name has to fail on the name rule itself, not slip into a file
	// lookup and come back as an unknown theme.
	for _, n := range []string{"../x", "a/b", "A", "-x", "x y", "x.yaml", "Dracula", "日本"} {
		_, err := Load(n)
		if err == nil {
			t.Errorf("Load(%q) gave no error", n)
			continue
		}
		if errors.Is(err, ErrUnknown) {
			t.Errorf("Load(%q) reached the file lookup: %v", n, err)
		}
		if !strings.Contains(err.Error(), `"`+n+`"`) {
			t.Errorf("Load(%q) does not name itself: %v", n, err)
		}
	}
}

// A name that holds only spaces is not the same as no name, so it is refused
// instead of quietly becoming the default theme.
func TestBlankNameAfterTrimIsRefused(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	_, err := Load("   ")
	if err == nil {
		t.Fatal("a blank name was accepted")
	}
	if errors.Is(err, ErrUnknown) {
		t.Fatalf("a bad name should be refused on its own rule, got %v", err)
	}
	if !strings.Contains(err.Error(), `"   "`) {
		t.Fatalf("err = %v", err)
	}
}

// A folder where the theme file should be is not a theme; the read error has
// to reach the user instead of looking like an unknown name.
func TestUnreadableUserFileReportsTheReadError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.MkdirAll(filepath.Join(home, ".acta", "themes", "mine.yaml"), 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := Load("mine")
	if err == nil || errors.Is(err, ErrUnknown) {
		t.Fatalf("err = %v", err)
	}
}

func TestParseRejects(t *testing.T) {
	cases := []struct{ name, body string }{
		{"bad yaml", "bg: ["},
		{"bad bg", strings.Replace(goodBody, `"#101010"`, `"#10101"`, 1)},
		{"bad fg", strings.Replace(goodBody, `"#eeeeee"`, `"white"`, 1)},
		{"bad slot", strings.Replace(goodBody, `"#000000"`, `"#00000g"`, 1)},
		{"bad last slot", strings.Replace(goodBody, `"#ffffff"`, `"#fff"`, 1)},
		{"15 slots", strings.Replace(goodBody, `"#000000",`, ``, 1)},
		{"17 slots", strings.Replace(goodBody, `"#000000",`, `"#000000","#000000",`, 1)},
		{"bad sel bg", goodBody + "selection_bg: \"#12\"\n"},
		{"bad sel fg", goodBody + "selection_fg: \"red\"\n"},
		{"missing bg", strings.Replace(goodBody, `bg: "#101010"`, ``, 1)},
		{"missing fg", strings.Replace(goodBody, `fg: "#eeeeee"`, ``, 1)},
		{"no ansi", strings.Replace(goodBody, "ansi: [", "other: [", 1)},
	}
	for _, c := range cases {
		if _, err := Parse("x", []byte(c.body)); err == nil {
			t.Errorf("%s: no error", c.name)
		}
	}
}

func TestParseAcceptsUppercaseHex(t *testing.T) {
	th, err := Parse("x", []byte(strings.Replace(goodBody, `"#101010"`, `"#AABBCC"`, 1)))
	if err != nil || th.BG != "#AABBCC" {
		t.Fatalf("got %+v, %v", th, err)
	}
}

func TestMissingSelectionKeysStayEmpty(t *testing.T) {
	th, err := Parse("x", []byte(goodBody))
	if err != nil || th.SelectionBG != "" || th.SelectionFG != "" {
		t.Fatalf("%+v, %v", th, err)
	}
}

// A HOME that is not a full path must not be turned into a path at all, or a
// theme file sitting next to acta would be read by accident. The test runs
// from a folder that really holds such a file.
func unusableHome(t *testing.T, home string) {
	t.Helper()
	dir := t.TempDir()
	writeUserTheme(t, dir, "mine", goodBody)
	t.Chdir(dir)
	t.Setenv("HOME", home)
	if got, err := Load("tokyo-night"); err != nil || got.BG != "#1a1b26" {
		t.Fatalf("built-in broke: %+v, %v", got, err)
	}
	if th, err := Load("mine"); !errors.Is(err, ErrUnknown) {
		t.Fatalf("HOME=%q read a user file: %+v, %v", home, th, err)
	}
}

func TestRelativeHomeSkipsUserFiles(t *testing.T) {
	unusableHome(t, "rel")
	unusableHome(t, ".")
}

func TestEmptyHomeSkipsUserFiles(t *testing.T) {
	unusableHome(t, "")
}

func TestDark(t *testing.T) {
	tn, _ := Builtin("tokyo-night")
	day, _ := Builtin("tokyo-night-day")
	term, _ := Builtin("terminal")
	if !tn.Dark(false) || day.Dark(true) || !term.Dark(true) || term.Dark(false) {
		t.Fatal("Dark gave the wrong answer")
	}
}
