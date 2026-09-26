// Package config finds where the planning files of one repo live.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Dirs names the three folders inside the root folder.
type Dirs struct {
	Specs string `yaml:"specs"`
	Plans string `yaml:"plans"`
	Bugs  string `yaml:"bugs"`
}

// Config says where the planning files of one repo live. Paths are absolute.
type Config struct {
	RepoRoot    string
	Root        string
	Dirs        Dirs
	Legacy      []string
	AutoCommit  bool
	IsGit       bool
	Branches    []string // glob patterns; empty means every unmerged branch
	BranchesOff bool
	Links       Links
}

// Links holds the footer URLs: ko-fi for Donate, the tracker for Feedback.
type Links struct {
	Donate   string `yaml:"donate"`
	Feedback string `yaml:"feedback"`
}

type fileConfig struct {
	Root       string    `yaml:"root"`
	Dirs       Dirs      `yaml:"dirs"`
	Legacy     *[]string `yaml:"legacy"`
	AutoCommit *bool     `yaml:"auto_commit"`
	Branches   *[]string `yaml:"branches"`
	Links      Links     `yaml:"links"`
}

// Default is the starting config. Load overwrites the Root here: it picks
// .acta/ for a new repo and .pm/ when the repo still has the old folder.
func Default(repoRoot string) Config {
	return Config{
		RepoRoot:   repoRoot,
		Root:       filepath.Join(repoRoot, ".pm"),
		Dirs:       Dirs{Specs: "specs", Plans: "plans", Bugs: "bugs"},
		Legacy:     []string{filepath.Join(repoRoot, "docs", "superpowers")},
		AutoCommit: true,
		Links:      Links{Feedback: "https://github.com/iyay/acta/issues"},
	}
}

// Load finds the repo from cwd, then picks the root folder in this order:
// flagRoot, the ACTA_ROOT env var, the PM_ROOT env var, root in .acta.yaml,
// root in .pm.yaml, then ".acta/" if it exists, else ".pm/" if it exists,
// else ".acta/". Other settings come from .acta.yaml when it exists,
// otherwise from .pm.yaml. A missing root in the winning file falls back
// to the other file, so an .acta.yaml with only dirs still honours root
// in .pm.yaml.
func Load(cwd, flagRoot string) (Config, error) {
	repo, isGit := findRepo(cwd)
	cfg := Default(repo)
	cfg.IsGit = isGit

	var actaFC, pmFC fileConfig
	var haveActa bool
	raw, err := os.ReadFile(filepath.Join(repo, ".acta.yaml"))
	switch {
	case err == nil:
		haveActa = true
		if err := yaml.Unmarshal(raw, &actaFC); err != nil {
			return Config{}, fmt.Errorf(".acta.yaml: %w", err)
		}
	case !errors.Is(err, os.ErrNotExist):
		return Config{}, err
	}

	raw, err = os.ReadFile(filepath.Join(repo, ".pm.yaml"))
	switch {
	case err == nil:
		if err := yaml.Unmarshal(raw, &pmFC); err != nil {
			return Config{}, fmt.Errorf(".pm.yaml: %w", err)
		}
	case !errors.Is(err, os.ErrNotExist):
		return Config{}, err
	}

	// The new file wins when it exists; the old file only fills a missing
	// root, so mixed pairs never surprise.
	fc := pmFC
	if haveActa {
		fc = actaFC
		if fc.Root == "" {
			fc.Root = pmFC.Root
		}
	}

	root := ".acta"
	switch {
	case flagRoot != "":
		root = flagRoot
	case os.Getenv("ACTA_ROOT") != "":
		root = os.Getenv("ACTA_ROOT")
	case os.Getenv("PM_ROOT") != "":
		root = os.Getenv("PM_ROOT")
	case fc.Root != "":
		root = fc.Root
	case dirExists(filepath.Join(repo, ".acta")):
		root = ".acta"
	case dirExists(filepath.Join(repo, ".pm")):
		root = ".pm"
	}
	cfg.Root = inRepo(repo, root)

	if fc.Dirs.Specs != "" {
		cfg.Dirs.Specs = fc.Dirs.Specs
	}
	if fc.Dirs.Plans != "" {
		cfg.Dirs.Plans = fc.Dirs.Plans
	}
	if fc.Dirs.Bugs != "" {
		cfg.Dirs.Bugs = fc.Dirs.Bugs
	}
	// A present but empty legacy list means "no legacy folders".
	if fc.Legacy != nil {
		cfg.Legacy = nil
		for _, l := range *fc.Legacy {
			cfg.Legacy = append(cfg.Legacy, inRepo(repo, l))
		}
	}
	if fc.AutoCommit != nil {
		cfg.AutoCommit = *fc.AutoCommit
	}
	// branches: missing reads every unmerged branch; [] reads none; a list
	// keeps only names matching one of its patterns.
	if fc.Branches != nil {
		if len(*fc.Branches) == 0 {
			cfg.BranchesOff = true
		} else {
			cfg.Branches = append([]string(nil), (*fc.Branches)...)
		}
	}
	// Links: anything but a real http(s) url never reaches the opener, so a
	// cloned .pm.yaml cannot turn a click into a local program.
	cfg.Links = Links{Donate: httpLink(fc.Links.Donate), Feedback: httpLink(fc.Links.Feedback)}
	if cfg.Links.Feedback == "" {
		cfg.Links.Feedback = "https://github.com/iyay/acta/issues"
	}
	return cfg, nil
}

// httpLink keeps a footer url only when it is a real http(s) address with a
// host behind it. Anything else comes back empty, so a bad value can never
// reach the browser opener.
func httpLink(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return ""
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https":
		return raw
	}
	return ""
}

// findRepo asks git for the top folder. Outside git the start folder is used.
func findRepo(cwd string) (string, bool) {
	out, err := exec.Command("git", "-C", cwd, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		abs, _ := filepath.Abs(cwd)
		return abs, false
	}
	return strings.TrimSpace(string(out)), true
}

// inRepo reads a relative path as relative to the repo root.
func inRepo(repo, p string) string {
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	return filepath.Join(repo, p)
}

// dirExists is true for a real folder, so an old .pm/ still full of plans
// keeps working until the user runs migrate-root.
func dirExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}
