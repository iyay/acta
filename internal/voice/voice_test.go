package voice

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefault(t *testing.T) {
	d := Default()
	if d.ChatLanguage != "English" || d.Style != "adhd" || d.RepoLanguage != "English" || d.Tone != "" {
		t.Fatalf("got %+v", d)
	}
}

func TestPath(t *testing.T) {
	t.Setenv("PM_VOICE_FILE", "/tmp/x/voice.yaml")
	if p, err := Path(); err != nil || p != "/tmp/x/voice.yaml" {
		t.Fatalf("got %q %v", p, err)
	}
	t.Setenv("PM_VOICE_FILE", "")
	home, _ := os.UserHomeDir()
	if p, _ := Path(); p != filepath.Join(home, ".acta", "voice.yaml") {
		t.Fatalf("got %q", p)
	}
}

func TestLoadMissing(t *testing.T) {
	v, exists, err := Load(filepath.Join(t.TempDir(), "none.yaml"))
	if err != nil || exists || v != Default() {
		t.Fatalf("got %+v %v %v", v, exists, err)
	}
}

func TestSaveThenLoad(t *testing.T) {
	p := filepath.Join(t.TempDir(), "deep", "voice.yaml")
	want := Voice{ChatLanguage: "Korean", Style: "plain", Tone: "Short sentences.\nNo jokes.", RepoLanguage: "English"}
	if err := Save(p, want); err != nil {
		t.Fatal(err)
	}
	got, exists, err := Load(p)
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
	got, _, err := Load(p)
	if err != nil || got != (Voice{ChatLanguage: "Korean", Style: "adhd", RepoLanguage: "English"}) {
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
		v, exists, err := Load(p)
		if err == nil || !exists || v != Default() {
			t.Errorf("%s: got %+v %v %v, want defaults, exists, error", name, v, exists, err)
		}
	}
}

func TestValidate(t *testing.T) {
	long := strings.Repeat("x", 601)
	nine := strings.Repeat("line\n", 8) + "line"
	cases := map[string]Voice{
		"style":            {ChatLanguage: "English", Style: "loud", RepoLanguage: "English"},
		"language newline": {ChatLanguage: "Eng\nlish", Style: "adhd", RepoLanguage: "English"},
		"language long":    {ChatLanguage: strings.Repeat("a", 41), Style: "adhd", RepoLanguage: "English"},
		"repo empty":       {ChatLanguage: "English", Style: "adhd", RepoLanguage: ""},
		"tone long":        {ChatLanguage: "English", Style: "adhd", RepoLanguage: "English", Tone: long},
		"tone nine lines":  {ChatLanguage: "English", Style: "adhd", RepoLanguage: "English", Tone: nine},
	}
	for name, v := range cases {
		if err := v.Validate(); !errors.Is(err, ErrBad) {
			t.Errorf("%s: err = %v, want ErrBad", name, err)
		}
	}
	if err := Default().Validate(); err != nil {
		t.Fatalf("defaults invalid: %v", err)
	}
}

func TestSaveRefusesBad(t *testing.T) {
	p := filepath.Join(t.TempDir(), "voice.yaml")
	if err := Save(p, Voice{Style: "loud"}); !errors.Is(err, ErrBad) {
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
		{"", "split ", true}, // fill trims
	} {
		v := Default()
		v.BuildExecutor, v.SubagentModels = tc.exec, tc.models
		err := fill(v).Validate()
		if (err == nil) != tc.ok {
			t.Errorf("exec %q models %q: err %v, want ok=%v", tc.exec, tc.models, err, tc.ok)
		}
		if err != nil && !errors.Is(err, ErrBad) {
			t.Errorf("exec %q models %q: err %v is not ErrBad", tc.exec, tc.models, err)
		}
	}
}

func TestLoadRejectsBadExecutor(t *testing.T) {
	p := filepath.Join(t.TempDir(), "voice.yaml")
	os.WriteFile(p, []byte("chat_language: English\nstyle: adhd\nbuild_executor: robot\n"), 0o644)
	if _, exists, err := Load(p); !exists || err == nil {
		t.Fatalf("got exists %v err %v, want a bad-value error", exists, err)
	}
}

// A bad value must never reach the file, so a set is refused whole.
func TestSaveRefusesBadExecutorAndModels(t *testing.T) {
	for name, v := range map[string]Voice{
		"executor": {ChatLanguage: "English", Style: "adhd", RepoLanguage: "English", BuildExecutor: "omp"},
		"models":   {ChatLanguage: "English", Style: "adhd", RepoLanguage: "English", SubagentModels: "all"},
	} {
		p := filepath.Join(t.TempDir(), "voice.yaml")
		if err := Save(p, v); !errors.Is(err, ErrBad) {
			t.Errorf("%s: err = %v, want ErrBad", name, err)
		}
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("%s: a bad voice was written", name)
		}
	}
}
