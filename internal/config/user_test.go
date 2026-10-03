package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefault(t *testing.T) {
	d := UserDefault()
	if d.ChatLanguage != "English" || d.Style != "adhd" || d.RepoLanguage != "English" || d.Tone != "" {
		t.Fatalf("got %+v", d)
	}
}

func TestPath(t *testing.T) {
	t.Setenv("PM_VOICE_FILE", "/tmp/x/voice.yaml")
	if p, err := UserPath(); err != nil || p != "/tmp/x/voice.yaml" {
		t.Fatalf("got %q %v", p, err)
	}
	t.Setenv("PM_VOICE_FILE", "")
	home, _ := os.UserHomeDir()
	if p, _ := UserPath(); p != filepath.Join(home, ".acta", "config.yaml") {
		t.Fatalf("got %q", p)
	}
}

func TestLoadMissing(t *testing.T) {
	v, exists, err := LoadUser(filepath.Join(t.TempDir(), "none.yaml"))
	if err != nil || exists || v != UserDefault() {
		t.Fatalf("got %+v %v %v", v, exists, err)
	}
}

func TestSaveThenLoad(t *testing.T) {
	p := filepath.Join(t.TempDir(), "deep", "voice.yaml")
	want := User{ChatLanguage: "Korean", Style: "plain", Tone: "Short sentences.\nNo jokes.", RepoLanguage: "English"}
	if err := SaveUserFile(p, want); err != nil {
		t.Fatal(err)
	}
	got, exists, err := LoadUser(p)
	if err != nil || !exists || got != want {
		t.Fatalf("got %+v %v %v", got, exists, err)
	}
	if _, err := os.Stat(p + ".tmp"); !os.IsNotExist(err) {
		t.Fatal("temp file left behind")
	}
}

func TestLoadFillsMissingKeys(t *testing.T) {
	p := filepath.Join(t.TempDir(), "voice.yaml")
	if err := os.WriteFile(p, []byte("chat_language: \"  Korean \"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, _, err := LoadUser(p)
	if err != nil || got != (User{ChatLanguage: "Korean", Style: "adhd", RepoLanguage: "English"}) {
		t.Fatalf("got %+v %v", got, err)
	}
}

func TestLoadBroken(t *testing.T) {
	for name, body := range map[string]string{
		"not yaml":  "chat_language: [\n",
		"bad style": "style: loud\n",
	} {
		p := filepath.Join(t.TempDir(), "voice.yaml")
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		v, exists, err := LoadUser(p)
		if err == nil || !exists || v != UserDefault() {
			t.Errorf("%s: got %+v %v %v, want defaults, exists, error", name, v, exists, err)
		}
	}
}

func TestValidate(t *testing.T) {
	long := strings.Repeat("x", 601)
	nine := strings.Repeat("line\n", 8) + "line"
	cases := map[string]User{
		"style":            {ChatLanguage: "English", Style: "loud", RepoLanguage: "English"},
		"language newline": {ChatLanguage: "Eng\nlish", Style: "adhd", RepoLanguage: "English"},
		"language long":    {ChatLanguage: strings.Repeat("a", 41), Style: "adhd", RepoLanguage: "English"},
		"repo empty":       {ChatLanguage: "English", Style: "adhd", RepoLanguage: ""},
		"tone long":        {ChatLanguage: "English", Style: "adhd", RepoLanguage: "English", Tone: long},
		"tone nine lines":  {ChatLanguage: "English", Style: "adhd", RepoLanguage: "English", Tone: nine},
	}
	for name, v := range cases {
		if err := v.Validate(); !errors.Is(err, ErrBadUser) {
			t.Errorf("%s: err = %v, want ErrBadUser", name, err)
		}
	}
	if err := UserDefault().Validate(); err != nil {
		t.Fatalf("defaults invalid: %v", err)
	}
}

func TestSaveRefusesBad(t *testing.T) {
	p := filepath.Join(t.TempDir(), "voice.yaml")
	if err := SaveUserFile(p, User{Style: "loud"}); !errors.Is(err, ErrBadUser) {
		t.Fatalf("err = %v", err)
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Fatal("a bad voice was written")
	}
}

func TestValidateExecutorAndModels(t *testing.T) {
	for _, tc := range []struct {
		exec, models string
		ok           bool
	}{
		{"", "", true},
		{"subagent", "", true},
		{"dispatch", "split", true},
		{"inline", "", true},
		{"Subagent", "", false},
		{"omp", "", false},
		{"", "all", false},
		{"", "default", true},
		{"", "mixed", false},
		{"", "Default", false},
		{"", "split ", true}, // fill trims
	} {
		v := UserDefault()
		v.BuildExecutor, v.SubagentModels = tc.exec, tc.models
		err := fill(v).Validate()
		if (err == nil) != tc.ok {
			t.Errorf("exec %q models %q: err %v, want ok=%v", tc.exec, tc.models, err, tc.ok)
		}
		if err != nil && !errors.Is(err, ErrBadUser) {
			t.Errorf("exec %q models %q: err %v is not ErrBadUser", tc.exec, tc.models, err)
		}
	}
}

func TestLoadRejectsBadExecutor(t *testing.T) {
	p := filepath.Join(t.TempDir(), "voice.yaml")
	os.WriteFile(p, []byte("chat_language: English\nstyle: adhd\nbuild_executor: robot\n"), 0o644)
	if _, exists, err := LoadUser(p); !exists || err == nil {
		t.Fatalf("got exists %v err %v, want a bad-value error", exists, err)
	}
}

// A bad value must never reach the file, so a set is refused whole.
func TestSaveRefusesBadExecutorAndModels(t *testing.T) {
	for name, v := range map[string]User{
		"executor": {ChatLanguage: "English", Style: "adhd", RepoLanguage: "English", BuildExecutor: "omp"},
		"models":   {ChatLanguage: "English", Style: "adhd", RepoLanguage: "English", SubagentModels: "all"},
	} {
		p := filepath.Join(t.TempDir(), "voice.yaml")
		if err := SaveUserFile(p, v); !errors.Is(err, ErrBadUser) {
			t.Errorf("%s: err = %v, want ErrBadUser", name, err)
		}
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("%s: a bad voice was written", name)
		}
	}
}

// coding_guide takes lean or off. Empty is fine, it means lean. Anything else
// is refused by name, and fill trims spaces like it does for the other keys.
func TestValidateCodingGuide(t *testing.T) {
	for _, tc := range []struct {
		guide string
		ok    bool
	}{
		{"", true},
		{"lean", true},
		{"off", true},
		{"lean ", true}, // fill trims
		{"Lean", false},
		{"OFF", false},
		{"on", false},
		{"full", false},
		{"minimal", false},
		{"lean off", false},
	} {
		v := UserDefault()
		v.CodingGuide = tc.guide
		err := fill(v).Validate()
		if (err == nil) != tc.ok {
			t.Errorf("coding_guide %q: err %v, want ok=%v", tc.guide, err, tc.ok)
		}
		if err != nil && !errors.Is(err, ErrBadUser) {
			t.Errorf("coding_guide %q: err %v is not ErrBadUser", tc.guide, err)
		}
		if err != nil && !strings.Contains(err.Error(), "coding_guide") {
			t.Errorf("coding_guide %q: err %v does not name the key", tc.guide, err)
		}
	}
}

// A bad coding_guide never loads from the user file and never lands in it.
func TestCodingGuideBadValueNeverStored(t *testing.T) {
	for _, bad := range []string{"deep", "Lean", "on", "full"} {
		in := filepath.Join(t.TempDir(), "config.yaml")
		body := "chat_language: English\nstyle: adhd\nrepo_language: English\ncoding_guide: " + bad + "\n"
		if err := os.WriteFile(in, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		v, exists, err := LoadUser(in)
		if !exists || !errors.Is(err, ErrBadUser) || v != UserDefault() {
			t.Errorf("load %q: got %+v %v %v, want defaults, exists, ErrBadUser", bad, v, exists, err)
		}
		out := filepath.Join(t.TempDir(), "config.yaml")
		u := UserDefault()
		u.CodingGuide = bad
		if err := SaveUserFile(out, u); !errors.Is(err, ErrBadUser) {
			t.Errorf("save %q: err %v, want ErrBadUser", bad, err)
		}
		if _, err := os.Stat(out); !os.IsNotExist(err) {
			t.Errorf("save %q: a bad value was written", bad)
		}
	}
}

// off comes back on the next read. An unset value is left out of the file and
// stays empty, since empty already means lean.
func TestCodingGuideRoundTripsAndStaysEmpty(t *testing.T) {
	p := filepath.Join(t.TempDir(), "config.yaml")
	v := UserDefault()
	v.CodingGuide = "off"
	if err := SaveUserFile(p, v); err != nil {
		t.Fatal(err)
	}
	got, _, err := LoadUser(p)
	if err != nil || got.CodingGuide != "off" {
		t.Fatalf("off: got %+v, %v", got, err)
	}
	if err := SaveUserFile(p, UserDefault()); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "coding_guide") {
		t.Fatalf("an unset coding_guide was written:\n%s", raw)
	}
	if got, _, err := LoadUser(p); err != nil || got.CodingGuide != "" {
		t.Fatalf("unset: got %+v, %v", got, err)
	}
}

// A saved theme must come back on the next read, or the TUI forgets it on
// every restart.
func TestThemeRoundTrips(t *testing.T) {
	p := filepath.Join(t.TempDir(), "voice.yaml")
	v := UserDefault()
	v.Theme = "dracula"
	if err := SaveUserFile(p, v); err != nil {
		t.Fatal(err)
	}
	got, _, err := LoadUser(p)
	if err != nil || got.Theme != "dracula" {
		t.Fatalf("got %+v, %v", got, err)
	}
}

// An empty theme is not written at all, so a voice file stays as short as it
// was before themes existed.
func TestThemeLeftOutWhenEmpty(t *testing.T) {
	p := filepath.Join(t.TempDir(), "voice.yaml")
	if err := SaveUserFile(p, UserDefault()); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "theme") {
		t.Fatalf("an empty theme was written:\n%s", raw)
	}
}

// The voice file feeds every hook, and a theme file can be deleted or edited
// at any time. So a theme that no longer loads must not break the read; only
// the TUI and the doctor care.
func TestThemeThatDoesNotLoadStillReads(t *testing.T) {
	for _, name := range []string{"gone", "../x", "not a theme", ""} {
		p := filepath.Join(t.TempDir(), "voice.yaml")
		body := "chat_language: English\nstyle: adhd\nrepo_language: English\n"
		if name != "" {
			body += "theme: " + name + "\n"
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		got, _, err := LoadUser(p)
		if err != nil || got.Theme != name {
			t.Fatalf("theme %q: got %+v, %v", name, got, err)
		}
	}
}
