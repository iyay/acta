// Package trees finds the other worktrees of a repo, so the board can show
// what agents are doing in them before their branches land.
package trees

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/gitc"
)

// Others gives every other worktree of cfg's repo, each read in its own
// root folder. A worktree keeps its old folder until its branch merges,
// so pushing this repo's folder name onto it would read it as empty.
// It never fails: outside git, or when git cannot list the
// worktrees, there are none.
func Others(cfg config.Config) []board.Tree {
	if !cfg.IsGit {
		return nil
	}
	wts, err := gitc.Worktrees(cfg.RepoRoot)
	if err != nil {
		return nil
	}
	root := cfg.Root
	if rel, err := filepath.Rel(cfg.RepoRoot, cfg.Root); err == nil && !strings.HasPrefix(rel, "..") {
		root = rel
	}
	var out []board.Tree
	onDisk := map[string]bool{}
	for _, w := range wts {
		onDisk[w.Branch] = true
		if same(w.Path, cfg.RepoRoot) {
			continue
		}
		// Only keep our folder name when the other worktree really has
		// it. Otherwise let it pick its own, so a worktree that still
		// keeps the old folder keeps showing up.
		forced := root
		if st, err := os.Stat(filepath.Join(w.Path, root)); err != nil || !st.IsDir() {
			forced = ""
		}
		c, err := config.Load(w.Path, forced)
		if err != nil {
			continue
		}
		out = append(out, board.Tree{Cfg: c, Branch: w.Branch})
	}
	return append(out, branches(cfg, onDisk)...)
}

// branches reads the root folder of every unmerged local branch that is not
// checked out anywhere, straight from git. A branch that still keeps the
// old folder name is read from there instead.
func branches(cfg config.Config, onDisk map[string]bool) []board.Tree {
	if cfg.BranchesOff {
		return nil
	}
	names, err := gitc.UnmergedBranches(cfg.RepoRoot)
	if err != nil {
		return nil
	}
	rootRel, err := filepath.Rel(cfg.RepoRoot, cfg.Root)
	if err != nil || strings.HasPrefix(rootRel, "..") {
		return nil
	}
	var out []board.Tree
	for _, name := range names {
		if onDisk[name] || !wanted(name, cfg.Branches) {
			continue
		}
		files, err := gitc.BranchFiles(cfg.RepoRoot, name, filepath.ToSlash(rootRel))
		branchCfg := cfg
		if err != nil || len(files) == 0 {
			// The branch may never have moved folders, so look under the
			// other name before giving up on it.
			if other := otherRoot(filepath.ToSlash(rootRel)); other != "" {
				if f, ferr := gitc.BranchFiles(cfg.RepoRoot, name, other); ferr == nil && len(f) > 0 {
					files, err = f, nil
					// Point the tree at the folder the branch really
					// keeps, or its files are skipped as foreign.
					branchCfg.Root = filepath.Join(cfg.RepoRoot, filepath.FromSlash(other))
				}
			}
		}
		if err != nil || len(files) == 0 {
			continue
		}
		out = append(out, board.Tree{Cfg: branchCfg, Branch: name, Files: files})
	}
	return out
}

// otherRoot swaps the last folder between the old and the new name, so a
// branch that never moved folders is still found. It returns "" when the
// root is neither name and there is nothing else to try.
func otherRoot(slashed string) string {
	switch path.Base(slashed) {
	case ".acta":
		return path.Join(path.Dir(slashed), ".pm")
	case ".pm":
		return path.Join(path.Dir(slashed), ".acta")
	}
	return ""
}

func wanted(name string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}
	for _, p := range patterns {
		if ok, _ := path.Match(p, name); ok {
			return true
		}
	}
	return false
}

// Load reads the main tree and every other worktree into one board.
func Load(cfg config.Config) (*board.Board, error) {
	return board.LoadTrees(cfg, Others(cfg))
}

// WatchDirs lists the folders to watch in the other worktrees, plus git's
// worktrees folder so a new worktree causes a reload. dirsOf gives the folders
// of one tree; the caller watches the main tree itself.
func WatchDirs(cfg config.Config, dirsOf func(config.Config) []string) []string {
	var out []string
	for _, t := range Others(cfg) {
		out = append(out, dirsOf(t.Cfg)...)
	}
	if cfg.IsGit {
		if common, err := gitc.CommonDir(cfg.RepoRoot); err == nil {
			out = append(out, filepath.Join(common, "worktrees"), filepath.Join(common, "refs", "heads"))
		}
	}
	return out
}

func same(a, b string) bool {
	ra, err1 := filepath.EvalSymlinks(a)
	rb, err2 := filepath.EvalSymlinks(b)
	if err1 != nil || err2 != nil {
		return filepath.Clean(a) == filepath.Clean(b)
	}
	return ra == rb
}
