// Package config finds where the planning files of one repo live.
package config

import (
	"errors"
	"fmt"
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
}

type fileConfig struct {
	Root       string    `yaml:"root"`
	Dirs       Dirs      `yaml:"dirs"`
	Legacy     *[]string `yaml:"legacy"`
	AutoCommit *bool     `yaml:"auto_commit"`
	Branches   *[]string `yaml:"branches"`
}

// Default is what a repo gets with no .pm.yaml, no flag and no env var.
func Default(repoRoot string) Config {
	return Config{
		RepoRoot:   repoRoot,
		Root:       filepath.Join(repoRoot, ".pm"),
		Dirs:       Dirs{Specs: "specs", Plans: "plans", Bugs: "bugs"},
		Legacy:     []string{filepath.Join(repoRoot, "docs", "superpowers")},
		AutoCommit: true,
	}
}

// Load finds the repo from cwd, then picks the root folder in this order:
// flagRoot, the PM_ROOT env var, root in .pm.yaml, then ".pm".
func Load(cwd, flagRoot string) (Config, error) {
	repo, isGit := findRepo(cwd)
	cfg := Default(repo)
	cfg.IsGit = isGit

	var fc fileConfig
	raw, err := os.ReadFile(filepath.Join(repo, ".pm.yaml"))
	switch {
	case err == nil:
		if err := yaml.Unmarshal(raw, &fc); err != nil {
			return Config{}, fmt.Errorf(".pm.yaml: %w", err)
		}
	case !errors.Is(err, os.ErrNotExist):
		return Config{}, err
	}

	root := ".pm"
	switch {
	case flagRoot != "":
		root = flagRoot
	case os.Getenv("PM_ROOT") != "":
		root = os.Getenv("PM_ROOT")
	case fc.Root != "":
		root = fc.Root
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
	return cfg, nil
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
