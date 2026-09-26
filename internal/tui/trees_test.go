package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pm-board/internal/board"
	"pm-board/internal/config"
)

func treeCfg(t *testing.T, files map[string]string) config.Config {
	t.Helper()
	dir := t.TempDir()
	for p, body := range files {
		full := filepath.Join(dir, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return config.Default(dir)
}

func worktreeModel(t *testing.T) Model {
	t.Helper()
	main := treeCfg(t, map[string]string{".pm/bugs/2026-09-20-main.md": "# Main bug\n\n## Symptom\nx\n"})
	wt := treeCfg(t, map[string]string{".pm/bugs/2026-09-24-wt.md": "# Worktree bug\n\n## Symptom\ny\n"})
	b, err := board.LoadTrees(main, []board.Tree{{Cfg: wt, Branch: "feat"}})
	if err != nil {
		t.Fatal(err)
	}
	m := New(main, b, true)
	m.render = func(md string, _ int) string { return md }
	return sized(m, 120, 40)
}

func TestWorktreeItemsAreLabelled(t *testing.T) {
	m := press(worktreeModel(t), "]", "]", "]") // the Bugs tab
	v := m.View()
	if !strings.Contains(v, "Worktree bug") || !strings.Contains(v, "· feat") || !strings.Contains(v, "Main bug") {
		t.Fatalf("list does not label the worktree item:\n%s", v)
	}
	if !strings.Contains(v, "worktree feat") {
		t.Fatalf("detail pane does not name the worktree:\n%s", v)
	}
}

func TestWorktreeItemsAreReadOnly(t *testing.T) {
	m := press(worktreeModel(t), "]", "]", "]") // the worktree bug is newest, so it is selected
	for _, k := range []string{"s", "t"} {
		m = press(m, k)
		if m.popup != nil || !strings.Contains(m.status, "worktree feat") {
			t.Fatalf("%s on a worktree item: popup %v status %q", k, m.popup, m.status)
		}
	}
	m = press(m, "j", "s")
	if m.popup == nil {
		t.Fatal("the main tree's bug should still open the popup")
	}
}

func TestBranchItemsAreNotOpened(t *testing.T) {
	main := treeCfg(t, map[string]string{".pm/bugs/2026-09-20-main.md": "# Main bug\n\n## Symptom\nx\n"})
	branch := board.Tree{Cfg: main, Branch: "feat-x", Files: map[string][]byte{
		".pm/bugs/2026-09-25-branch.md": []byte("# Branch bug\n\n## Symptom\ny\n"),
	}}
	b, err := board.LoadTrees(main, []board.Tree{branch})
	if err != nil {
		t.Fatal(err)
	}
	m := New(main, b, true)
	m.render = func(md string, _ int) string { return md }
	m = press(sized(m, 120, 40), "]", "]", "]")
	if v := m.View(); !strings.Contains(v, "branch feat-x (not checked out)") || !strings.Contains(v, "· feat-x") {
		t.Fatalf("branch item not labelled:\n%s", v)
	}
	next, cmd := m.Update(key("enter"))
	m = next.(Model)
	if cmd != nil || !strings.Contains(m.status, "git worktree add") {
		t.Fatalf("enter on a branch item: cmd %v status %q", cmd != nil, m.status)
	}
	m = press(m, "s")
	if m.popup != nil {
		t.Fatal("s must not open a popup for a branch item")
	}
}

func TestWithLoad(t *testing.T) {
	called := false
	m := worktreeModel(t).WithLoad(func() (*board.Board, error) {
		called = true
		return &board.Board{}, nil
	})
	if msg := m.reloadCmd()(); msg == nil || !called {
		t.Fatal("WithLoad did not replace the loader")
	}
}
