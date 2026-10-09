package commits

import (
	"strings"

	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/trees"
)

// RefsFor lists the refs Find should search for a repo: the checked-out
// branch first (HEAD when detached), then the branch of every other
// worktree. Outside git there is nothing to search.
func RefsFor(cfg config.Config) []Ref {
	if !cfg.IsGit {
		return nil
	}
	first := "HEAD"
	if out, err := run(cfg.RepoRoot, "branch", "--show-current"); err == nil {
		if name := strings.TrimSpace(out); name != "" {
			first = name
		}
	}
	refs := []Ref{{Name: first}}
	seen := map[string]bool{first: true}
	for _, tr := range trees.Others(cfg) {
		// A detached worktree reports "(detached)". That is no ref name.
		if tr.Branch == "" || tr.Branch == "(detached)" || seen[tr.Branch] {
			continue
		}
		seen[tr.Branch] = true
		refs = append(refs, Ref{Name: tr.Branch})
	}
	return refs
}
