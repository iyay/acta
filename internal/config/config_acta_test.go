package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// rootCase is one row of the new root order: flag, ACTA_ROOT, PM_ROOT,
// .acta.yaml, .pm.yaml, and which folders exist.
type rootCase struct {
	name     string
	flag     string
	actaRoot string
	pmRoot   string
	actaYAML string // "" means the file is missing
	pmYAML   string // "" means the file is missing
	actaDir  bool
	pmDir    bool
	want     string // relative to repo, or absolute
}

func TestLoadRootActaOrder(t *testing.T) {
	cases := []rootCase{
		{name: "flag wins over everything", flag: "from-flag", actaRoot: "from-acta-env", pmRoot: "from-pm-env", actaYAML: "root: from-acta-file\n", pmYAML: "root: from-pm-file\n", want: "from-flag"},
		{name: "ACTA_ROOT wins over PM_ROOT", actaRoot: "from-acta-env", pmRoot: "from-pm-env", actaYAML: "root: from-acta-file\n", pmYAML: "root: from-pm-file\n", want: "from-acta-env"},
		{name: "PM_ROOT honoured when ACTA_ROOT empty", pmRoot: "from-pm-env", actaYAML: "root: from-acta-file\n", pmYAML: "root: from-pm-file\n", want: "from-pm-env"},
		{name: "root in .acta.yaml used", actaYAML: "root: from-acta-file\n", pmYAML: "root: from-pm-file\n", want: "from-acta-file"},
		{name: "root in .pm.yaml when .acta.yaml missing", pmYAML: "root: from-pm-file\n", want: "from-pm-file"},
		{name: "root in .pm.yaml when .acta.yaml has no root", actaYAML: "dirs:\n  specs: s\n", pmYAML: "root: from-pm-file\n", want: "from-pm-file"},
		{name: "both files with root: .acta.yaml wins", actaYAML: "root: from-acta-file\n", pmYAML: "root: from-pm-file\n", want: "from-acta-file"},
		{name: "no yaml, only .pm exists", pmDir: true, want: ".pm"},
		{name: "no yaml, only .acta exists", actaDir: true, want: ".acta"},
		{name: "both folders, no yaml: .acta wins", actaDir: true, pmDir: true, want: ".acta"},
		{name: "neither folder, no yaml: .acta for a new repo", want: ".acta"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			repo := gitInit(t)
			t.Setenv("ACTA_ROOT", c.actaRoot)
			t.Setenv("PM_ROOT", c.pmRoot)
			if c.actaYAML != "" {
				write(t, filepath.Join(repo, ".acta.yaml"), c.actaYAML)
			}
			if c.pmYAML != "" {
				write(t, filepath.Join(repo, ".pm.yaml"), c.pmYAML)
			}
			if c.actaDir {
				if err := os.MkdirAll(filepath.Join(repo, ".acta"), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			if c.pmDir {
				if err := os.MkdirAll(filepath.Join(repo, ".pm"), 0o755); err != nil {
					t.Fatal(err)
				}
			}
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

func TestLoadBadActaYAML(t *testing.T) {
	repo := gitInit(t)
	t.Setenv("ACTA_ROOT", "")
	t.Setenv("PM_ROOT", "")
	write(t, filepath.Join(repo, ".acta.yaml"), "root: [\n")
	_, err := Load(repo, "")
	if err == nil {
		t.Fatal("want an error for broken .acta.yaml")
	}
	if !strings.Contains(err.Error(), ".acta.yaml") {
		t.Fatalf("error %q must name .acta.yaml", err)
	}
}

func TestLoadBadPmYAMLNamesFile(t *testing.T) {
	repo := gitInit(t)
	t.Setenv("ACTA_ROOT", "")
	t.Setenv("PM_ROOT", "")
	write(t, filepath.Join(repo, ".pm.yaml"), "root: [\n")
	_, err := Load(repo, "")
	if err == nil {
		t.Fatal("want an error for broken .pm.yaml")
	}
	if !strings.Contains(err.Error(), ".pm.yaml") {
		t.Fatalf("error %q must name .pm.yaml", err)
	}
}

func TestLoadOldFileSettingsStillLoad(t *testing.T) {
	repo := gitInit(t)
	t.Setenv("ACTA_ROOT", "")
	t.Setenv("PM_ROOT", "")
	write(t, filepath.Join(repo, ".pm.yaml"),
		"dirs:\n  specs: s\nlegacy: []\nauto_commit: false\n")
	got, err := Load(repo, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Dirs.Specs != "s" {
		t.Fatalf("Dirs.Specs = %q, want s", got.Dirs.Specs)
	}
	if len(got.Legacy) != 0 {
		t.Fatalf("Legacy = %v, want none", got.Legacy)
	}
	if got.AutoCommit {
		t.Fatal("AutoCommit should be off")
	}
}

func TestLoadActaFileSettingsWin(t *testing.T) {
	repo := gitInit(t)
	t.Setenv("ACTA_ROOT", "")
	t.Setenv("PM_ROOT", "")
	write(t, filepath.Join(repo, ".acta.yaml"), "dirs:\n  specs: new\n")
	write(t, filepath.Join(repo, ".pm.yaml"), "dirs:\n  specs: old\n")
	got, err := Load(repo, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Dirs.Specs != "new" {
		t.Fatalf("Dirs.Specs = %q, want new", got.Dirs.Specs)
	}
}
