package config

import (
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
	t.Setenv("PM_ROOT", "")
	got, err := Load(repo, "")
	if err != nil {
		t.Fatal(err)
	}
	want := Default(repo)
	want.IsGit = true
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v\nwant %+v", got, want)
	}
	if want.Root != filepath.Join(repo, ".pm") || !want.AutoCommit ||
		!reflect.DeepEqual(want.Legacy, []string{filepath.Join(repo, "docs", "superpowers")}) ||
		want.Dirs != (Dirs{Specs: "specs", Plans: "plans", Bugs: "bugs"}) {
		t.Fatalf("bad defaults %+v", want)
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
	if got.Dirs != (Dirs{Specs: "s", Plans: "plans", Bugs: "bugs"}) {
		t.Fatalf("Dirs = %+v", got.Dirs)
	}
	if len(got.Legacy) != 0 {
		t.Fatalf("Legacy = %v, want none", got.Legacy)
	}
	if got.AutoCommit {
		t.Fatal("AutoCommit should be off")
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
