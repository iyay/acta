package setup_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/setup"
)

// wantBlock is the block text byte for byte as in
// git show main:plugin/skills/setup/SKILL.md (the fenced block).
const wantBlock = `<!-- acta:begin -->
## acta
This repo uses the acta plugin. Before each workflow step, load the matching acta skill and follow it.
Specs, plans, bugs, debt, scratch items and the wiki live in ` + "`.acta/`" + `.
Project knowledge (gotchas, runbooks, decisions with their why) goes to ` + "`.acta/wiki/`" + `, never to agent memory. Write a page only when a fresh agent would lose time or repeat a mistake without it. When a fact changes, rewrite its page.
Before changing a file, run ` + "`acta wiki match <file>`" + ` and read each page it names.
Work in flight goes to ` + "`acta state set <plan id> next`" + ` with the text on stdin (read it back with ` + "`acta state <plan id>`" + `), not to agent memory. Agent memory keeps only the user's own setup.
Raw ideas go to Scratchpad with ` + "`acta scratch new`" + `. A finished scratch item is specced, never dropped; dropped means not done or not valid.
<!-- acta:end -->
`

func TestBlockMatchesSkill(t *testing.T) {
	if setup.Block != wantBlock {
		t.Fatalf("Block differs from the skill block:\ngot:\n%s\nwant:\n%s", setup.Block, wantBlock)
	}
}

func TestWriteBlock(t *testing.T) {
	// Never touch the real home: every path below lives in temp dirs.
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PM_VOICE_FILE", filepath.Join(home, "config.yaml"))
	t.Setenv("TMPDIR", home)

	t.Run("new file holds only the block", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "CLAUDE.md")
		if err := setup.WriteBlock(path); err != nil {
			t.Fatal(err)
		}
		got := read(t, path)
		if got != wantBlock {
			t.Fatalf("new file = %q, want the block only", got)
		}
	})

	t.Run("file without markers keeps old bytes then the block", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "CLAUDE.md")
		old := "# My project\nSome notes.\n"
		write(t, path, old)
		if err := setup.WriteBlock(path); err != nil {
			t.Fatal(err)
		}
		got := read(t, path)
		if got != old+wantBlock {
			t.Fatalf("appended file = %q, want old bytes then the block", got)
		}
	})

	t.Run("file without trailing newline still keeps old bytes", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "CLAUDE.md")
		write(t, path, "# My project")
		if err := setup.WriteBlock(path); err != nil {
			t.Fatal(err)
		}
		got := read(t, path)
		if got != "# My project\n"+wantBlock {
			t.Fatalf("appended file = %q, want old bytes plus a newline then the block", got)
		}
	})

	t.Run("empty file holds only the block", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "CLAUDE.md")
		write(t, path, "")
		if err := setup.WriteBlock(path); err != nil {
			t.Fatal(err)
		}
		if got := read(t, path); got != wantBlock {
			t.Fatalf("empty file = %q, want the block only", got)
		}
	})

	t.Run("re-run replaces only the bytes between the markers", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "CLAUDE.md")
		head := "# My project\n\n"
		tail := "\nMore notes.\n"
		oldInner := "<!-- acta:begin -->\nStale text.\n<!-- acta:end -->\n"
		write(t, path, head+oldInner+tail)
		if err := setup.WriteBlock(path); err != nil {
			t.Fatal(err)
		}
		got := read(t, path)
		if got != head+wantBlock+tail {
			t.Fatalf("rewritten file = %q, want head + block + tail", got)
		}
		if !strings.HasPrefix(got, head) || !strings.HasSuffix(got, tail) {
			t.Fatalf("surrounding text moved: %q", got)
		}
		// A second run changes nothing.
		if err := setup.WriteBlock(path); err != nil {
			t.Fatal(err)
		}
		if again := read(t, path); again != got {
			t.Fatalf("second run changed bytes:\n%s\nwas:\n%s", again, got)
		}
	})

	t.Run("file holding only the block stays put", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "CLAUDE.md")
		write(t, path, wantBlock)
		if err := setup.WriteBlock(path); err != nil {
			t.Fatal(err)
		}
		if got := read(t, path); got != wantBlock {
			t.Fatalf("block-only file = %q, want it unchanged", got)
		}
	})
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}
