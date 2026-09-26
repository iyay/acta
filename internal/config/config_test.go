package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

func gitInit(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "init", "-q", dir).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	return dir
}

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadDefaults(t *testing.T) {
	repo := gitInit(t)
	t.Setenv("ACTA_ROOT", "")
	t.Setenv("PM_ROOT", "")
	got, err := Load(repo, "")
	if err != nil {
		t.Fatal(err)
	}
	// A fresh folder with no yaml and no root folders is a new repo, so
	// Load resolves to .acta/, same as the Default base.
	if want := filepath.Join(repo, ".acta"); got.Root != want {
		t.Fatalf("Root = %s, want %s", got.Root, want)
	}
	base := Default(repo)
	if base.Root != filepath.Join(repo, ".acta") || !base.AutoCommit ||
		!reflect.DeepEqual(base.Legacy, []string{filepath.Join(repo, "docs", "superpowers")}) ||
		base.Dirs != (Dirs{Specs: "specs", Plans: "plans", Bugs: "bugs", Debt: "debt"}) ||
		base.Links.Feedback != "https://github.com/iyay/acta/issues" || base.Links.Donate != "" {
		t.Fatalf("bad defaults %+v", base)
	}
}

func TestLoadFindsRepoFromSubfolder(t *testing.T) {
	repo := gitInit(t)
	sub := filepath.Join(repo, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := Load(sub, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.RepoRoot != repo {
		t.Fatalf("RepoRoot = %s, want %s", got.RepoRoot, repo)
	}
}

func TestLoadRootOrder(t *testing.T) {
	repo := gitInit(t)
	write(t, filepath.Join(repo, ".pm.yaml"), "root: from-file\n")
	cases := []struct{ name, env, flag, want string }{
		{"file beats default", "", "", "from-file"},
		{"env beats file", "from-env", "", "from-env"},
		{"flag beats env", "from-env", "from-flag", "from-flag"},
		{"absolute flag", "", "/abs/root", "/abs/root"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv("PM_ROOT", c.env)
			got, err := Load(repo, c.flag)
			if err != nil {
				t.Fatal(err)
			}
			want := c.want
			if !filepath.IsAbs(want) {
				want = filepath.Join(repo, want)
			}
			if got.Root != want {
				t.Fatalf("Root = %s, want %s", got.Root, want)
			}
		})
	}
}

func TestLoadFileOverrides(t *testing.T) {
	repo := gitInit(t)
	t.Setenv("PM_ROOT", "")
	write(t, filepath.Join(repo, ".pm.yaml"),
		"dirs:\n  specs: s\nlegacy: []\nauto_commit: false\n")
	got, err := Load(repo, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Dirs != (Dirs{Specs: "s", Plans: "plans", Bugs: "bugs", Debt: "debt"}) {
		t.Fatalf("Dirs = %+v", got.Dirs)
	}
	if len(got.Legacy) != 0 {
		t.Fatalf("Legacy = %v, want none", got.Legacy)
	}
	if got.AutoCommit {
		t.Fatal("AutoCommit should be off")
	}
}

func TestLoadDebtDirOverride(t *testing.T) {
	repo := gitInit(t)
	t.Setenv("PM_ROOT", "")
	write(t, filepath.Join(repo, ".acta.yaml"), "dirs:\n  debt: backlog\n")
	got, err := Load(repo, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Dirs.Debt != "backlog" {
		t.Fatalf("Dirs.Debt = %q, want backlog", got.Dirs.Debt)
	}
}

func TestLoadBadYAML(t *testing.T) {
	repo := gitInit(t)
	write(t, filepath.Join(repo, ".pm.yaml"), "root: [\n")
	if _, err := Load(repo, ""); err == nil {
		t.Fatal("want an error for broken .pm.yaml")
	}
}

func TestLoadOutsideGit(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PM_ROOT", "")
	got, err := Load(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.IsGit || got.RepoRoot != dir {
		t.Fatalf("got %+v", got)
	}
}

func TestLoadBranches(t *testing.T) {
	repo := gitInit(t)
	t.Setenv("PM_ROOT", "")
	got, _ := Load(repo, "")
	if got.BranchesOff || got.Branches != nil {
		t.Fatalf("default: %+v", got)
	}
	write(t, filepath.Join(repo, ".pm.yaml"), "branches: []\n")
	if got, _ := Load(repo, ""); !got.BranchesOff {
		t.Fatal("branches: [] must turn branch reading off")
	}
	write(t, filepath.Join(repo, ".pm.yaml"), "branches: [\"feat/*\"]\n")
	if got, _ := Load(repo, ""); got.BranchesOff || !reflect.DeepEqual(got.Branches, []string{"feat/*"}) {
		t.Fatalf("patterns: %+v", got)
	}
}

func TestLoadLinks(t *testing.T) {
	repo := gitInit(t)
	t.Setenv("PM_ROOT", "")
	write(t, filepath.Join(repo, ".pm.yaml"),
		"links:\n  donate: https://ko-fi.com/someone\n  feedback: https://example.com/bugs\n")
	got, err := Load(repo, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Links.Donate != "https://ko-fi.com/someone" {
		t.Fatalf("Donate = %q, want the ko-fi url", got.Links.Donate)
	}
	if got.Links.Feedback != "https://example.com/bugs" {
		t.Fatalf("Feedback = %q, want the custom url", got.Links.Feedback)
	}
}

func TestLoadLinksDefaults(t *testing.T) {
	repo := gitInit(t)
	t.Setenv("PM_ROOT", "")
	got, err := Load(repo, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Links.Feedback != "https://github.com/iyay/acta/issues" {
		t.Fatalf("Feedback = %q, want the issue tracker", got.Links.Feedback)
	}
	if got.Links.Donate != "" {
		t.Fatalf("Donate = %q, want empty when unset", got.Links.Donate)
	}
}

func TestLoadLinksKeepOnlyHTTP(t *testing.T) {
	for _, tc := range []struct {
		name         string
		donate       string
		feedback     string
		wantDonate   string
		wantFeedback string
	}{
		{"good https", "https://ko-fi.com/someone", "https://example.com/bugs", "https://ko-fi.com/someone", "https://example.com/bugs"},
		{"uppercase scheme", "HTTPS://Example.COM/ok", "HTTP://Example.COM/bugs", "HTTPS://Example.COM/ok", "HTTP://Example.COM/bugs"},
		{"file url", "file:///tmp/Donate.app", "file:///tmp/bugs", "", "https://github.com/iyay/acta/issues"},
		{"javascript", "javascript:alert(1)", "javascript:alert(1)", "", "https://github.com/iyay/acta/issues"},
		{"flag", "-aTerminal", "-aTerminal", "", "https://github.com/iyay/acta/issues"},
		{"hostless http", "http://", "http://", "", "https://github.com/iyay/acta/issues"},
		{"hostless https", "https://", "https://", "", "https://github.com/iyay/acta/issues"},
		{"ftp", "ftp://host/x", "ftp://host/x", "", "https://github.com/iyay/acta/issues"},
		{"leading space", " https://ko-fi.com/someone", " https://example.com/bugs", "", "https://github.com/iyay/acta/issues"},
		{"empty", "", "", "", "https://github.com/iyay/acta/issues"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := gitInit(t)
			t.Setenv("PM_ROOT", "")
			write(t, filepath.Join(repo, ".pm.yaml"),
				fmt.Sprintf("links:\n  donate: %q\n  feedback: %q\n", tc.donate, tc.feedback))
			got, err := Load(repo, "")
			if err != nil {
				t.Fatal(err)
			}
			if got.Links.Donate != tc.wantDonate {
				t.Errorf("Donate = %q, want %q", got.Links.Donate, tc.wantDonate)
			}
			if got.Links.Feedback != tc.wantFeedback {
				t.Errorf("Feedback = %q, want %q", got.Links.Feedback, tc.wantFeedback)
			}
		})
	}
}
