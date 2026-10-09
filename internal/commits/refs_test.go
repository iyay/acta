package commits

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/iyay/acta/internal/config"
)

func TestRefsFor(t *testing.T) {
	d := "2026-01-02T00:00:00Z"
	t.Run("main plus a worktree on feat", func(t *testing.T) {
		dir := newRepo(t)
		commit(t, dir, d, "feat: base")
		wt := filepath.Join(t.TempDir(), "wt")
		git(t, dir, d, "worktree", "add", "-q", "-b", "feat", wt)
		cfg := config.Config{RepoRoot: dir, Root: filepath.Join(dir, ".acta"), IsGit: true}
		want := []Ref{{Name: "main"}, {Name: "feat"}}
		if got := RefsFor(cfg); !reflect.DeepEqual(got, want) {
			t.Fatalf("RefsFor = %v, want %v", got, want)
		}
	})
	t.Run("detached HEAD gives HEAD first", func(t *testing.T) {
		dir := newRepo(t)
		sha := commit(t, dir, d, "feat: base")
		git(t, dir, d, "checkout", "-q", "--detach", sha)
		cfg := config.Config{RepoRoot: dir, Root: filepath.Join(dir, ".acta"), IsGit: true}
		want := []Ref{{Name: "HEAD"}}
		if got := RefsFor(cfg); !reflect.DeepEqual(got, want) {
			t.Fatalf("RefsFor = %v, want %v", got, want)
		}
	})
	t.Run("not git gives nil", func(t *testing.T) {
		cfg := config.Config{RepoRoot: t.TempDir(), IsGit: false}
		if got := RefsFor(cfg); got != nil {
			t.Fatalf("RefsFor = %v, want nil", got)
		}
	})
}
