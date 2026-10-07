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
	t.Run("begin without end errors and leaves the file alone", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "CLAUDE.md")
		old := "# My project\nKeep this.\n<!-- acta:begin -->\nStale text.\n"
		write(t, path, old)
		// Both runs must refuse: the old code ate "Stale text." on run 2.
		for run := 1; run <= 2; run++ {
			if err := setup.WriteBlock(path); err == nil {
				t.Fatalf("run %d: WriteBlock with a begin but no end = nil, want an error", run)
			}
			if got := read(t, path); got != old {
				t.Fatalf("run %d: file = %q, want the old bytes untouched", run, got)
			}
		}
	})

	t.Run("end without begin errors and leaves the file alone", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "CLAUDE.md")
		old := "# My project\nKeep this.\n<!-- acta:end -->\n"
		write(t, path, old)
		for run := 1; run <= 2; run++ {
			if err := setup.WriteBlock(path); err == nil {
				t.Fatalf("run %d: WriteBlock with an end but no begin = nil, want an error", run)
			}
			if got := read(t, path); got != old {
				t.Fatalf("run %d: file = %q, want the old bytes untouched", run, got)
			}
		}
	})

	t.Run("end before begin errors and leaves the file alone", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "CLAUDE.md")
		old := "<!-- acta:end -->\nMiddle.\n<!-- acta:begin -->\n"
		write(t, path, old)
		for run := 1; run <= 2; run++ {
			if err := setup.WriteBlock(path); err == nil {
				t.Fatalf("run %d: WriteBlock with the end before the begin = nil, want an error", run)
			}
			if got := read(t, path); got != old {
				t.Fatalf("run %d: file = %q, want the old bytes untouched", run, got)
			}
		}
	})

	t.Run("two begins error and leave the file alone", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "CLAUDE.md")
		old := "# My project\n<!-- acta:begin -->\nFirst.\n<!-- acta:begin -->\nSecond.\n<!-- acta:end -->\nTail.\n"
		write(t, path, old)
		for run := 1; run <= 2; run++ {
			if err := setup.WriteBlock(path); err == nil {
				t.Fatalf("run %d: WriteBlock with two begins = nil, want an error", run)
			}
			if got := read(t, path); got != old {
				t.Fatalf("run %d: file = %q, want the old bytes untouched", run, got)
			}
		}
	})

	t.Run("CRLF file keeps its line endings on every re-run", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "CLAUDE.md")
		head := "# My project\r\n\r\n"
		tail := "\r\nMore notes.\r\n"
		oldInner := "<!-- acta:begin -->\r\nStale text.\r\n<!-- acta:end -->\r\n"
		write(t, path, head+oldInner+tail)
		crlfBlock := strings.ReplaceAll(wantBlock, "\n", "\r\n")
		want := head + crlfBlock + tail
		// Three runs: the fix must hold however often the wizard rewrites.
		for run := 1; run <= 3; run++ {
			if err := setup.WriteBlock(path); err != nil {
				t.Fatal(err)
			}
			got := read(t, path)
			if got != want {
				t.Fatalf("run %d: file = %q, want head plus a CRLF block plus tail", run, got)
			}
			if strings.Contains(strings.ReplaceAll(got, "\r\n", ""), "\r") {
				t.Fatalf("run %d: stray carriage return in %q", run, got)
			}
		}
	})

	t.Run("CRLF file without markers gains a CRLF block", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "CLAUDE.md")
		old := "# My project\r\nSome notes.\r\n"
		write(t, path, old)
		crlfBlock := strings.ReplaceAll(wantBlock, "\n", "\r\n")
		for run := 1; run <= 2; run++ {
			if err := setup.WriteBlock(path); err != nil {
				t.Fatal(err)
			}
			got := read(t, path)
			if got != old+crlfBlock {
				t.Fatalf("run %d: file = %q, want the old bytes then a CRLF block", run, got)
			}
			if strings.Contains(strings.ReplaceAll(got, "\r\n", ""), "\r") {
				t.Fatalf("run %d: stray carriage return in %q", run, got)
			}
		}
	})

	t.Run("failed write leaves the old file whole", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "CLAUDE.md")
		old := "# My project\n"
		write(t, path, old)
		// A locked folder stops the temp file, so the rename never runs.
		if err := os.Chmod(dir, 0o555); err != nil {
			t.Fatal(err)
		}
		defer os.Chmod(dir, 0o755)
		if err := setup.WriteBlock(path); err == nil {
			t.Fatalf("WriteBlock in a read-only folder = nil, want an error")
		}
		if got := read(t, path); got != old {
			t.Fatalf("after a failed write = %q, want the old bytes untouched", got)
		}
	})

	t.Run("successful write swaps through a rename and leaves no temp file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "CLAUDE.md")
		write(t, path, "# My project\n")
		before, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := setup.WriteBlock(path); err != nil {
			t.Fatal(err)
		}
		after, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		// A rename swaps the file; writing in place would keep it.
		if os.SameFile(before, after) {
			t.Fatalf("WriteBlock rewrote the file in place, want a temp file plus rename")
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), ".acta-block-") {
				t.Fatalf("leftover temp file %s in %s", e.Name(), dir)
			}
		}
	})
	t.Run("symlink to a file in the same dir keeps the link", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "TARGET.md")
		link := filepath.Join(dir, "CLAUDE.md")
		old := "# My project\nSome notes.\n"
		write(t, target, old)
		if err := os.Symlink(target, link); err != nil {
			t.Fatal(err)
		}
		// Two runs: the link must survive every rewrite.
		for run := 1; run <= 2; run++ {
			if err := setup.WriteBlock(link); err != nil {
				t.Fatal(err)
			}
			wantLink(t, link)
			if got := read(t, target); got != old+wantBlock {
				t.Fatalf("run %d: target = %q, want old bytes then the block", run, got)
			}
		}
	})

	t.Run("symlink to a file in another dir keeps the link", func(t *testing.T) {
		base := t.TempDir()
		realDir := filepath.Join(base, "real")
		linkDir := filepath.Join(base, "links")
		if err := os.Mkdir(realDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(linkDir, 0o755); err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(realDir, "NOTES.md")
		link := filepath.Join(linkDir, "CLAUDE.md")
		old := "# Notes\n"
		write(t, target, old)
		// A relative link: the target lives one folder over.
		if err := os.Symlink(filepath.Join("..", "real", "NOTES.md"), link); err != nil {
			t.Fatal(err)
		}
		for run := 1; run <= 2; run++ {
			if err := setup.WriteBlock(link); err != nil {
				t.Fatal(err)
			}
			wantLink(t, link)
			if got := read(t, target); got != old+wantBlock {
				t.Fatalf("run %d: target = %q, want old bytes then the block", run, got)
			}
		}
		// No temp file may leak into either folder.
		for _, dir := range []string{realDir, linkDir} {
			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatal(err)
			}
			for _, e := range entries {
				if strings.HasPrefix(e.Name(), ".acta-block-") {
					t.Fatalf("leftover temp file %s in %s", e.Name(), dir)
				}
			}
		}
	})

	t.Run("chained symlinks keep every link", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "REAL.md")
		mid := filepath.Join(dir, "MID.md")
		link := filepath.Join(dir, "CLAUDE.md")
		old := "# Real\n"
		write(t, target, old)
		if err := os.Symlink(target, mid); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(mid, link); err != nil {
			t.Fatal(err)
		}
		for run := 1; run <= 2; run++ {
			if err := setup.WriteBlock(link); err != nil {
				t.Fatal(err)
			}
			wantLink(t, link)
			wantLink(t, mid)
			if got := read(t, target); got != old+wantBlock {
				t.Fatalf("run %d: target = %q, want old bytes then the block", run, got)
			}
		}
	})

	t.Run("CLAUDE.md linked to AGENTS.md takes writes to both names", func(t *testing.T) {
		dir := t.TempDir()
		agents := filepath.Join(dir, "AGENTS.md")
		claude := filepath.Join(dir, "CLAUDE.md")
		old := "# Agents\n"
		write(t, agents, old)
		if err := os.Symlink(agents, claude); err != nil {
			t.Fatal(err)
		}
		// Both names point at one file: write through each name twice.
		for run := 1; run <= 2; run++ {
			if err := setup.WriteBlock(claude); err != nil {
				t.Fatal(err)
			}
			if err := setup.WriteBlock(agents); err != nil {
				t.Fatal(err)
			}
		}
		wantLink(t, claude)
		if got := read(t, agents); got != old+wantBlock {
			t.Fatalf("AGENTS.md = %q, want old bytes then one block", got)
		}
		if got := read(t, claude); got != old+wantBlock {
			t.Fatalf("CLAUDE.md = %q, want old bytes then one block", got)
		}
		if n := strings.Count(read(t, agents), "<!-- acta:begin -->"); n != 1 {
			t.Fatalf("target holds %d begin markers, want exactly 1", n)
		}
	})

	t.Run("failed write through a symlink leaves the target whole", func(t *testing.T) {
		base := t.TempDir()
		realDir := filepath.Join(base, "real")
		linkDir := filepath.Join(base, "links")
		if err := os.Mkdir(realDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(linkDir, 0o755); err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(realDir, "CLAUDE.md")
		link := filepath.Join(linkDir, "CLAUDE.md")
		old := "# My project\n"
		write(t, target, old)
		if err := os.Symlink(target, link); err != nil {
			t.Fatal(err)
		}
		// A locked target folder stops the temp file, so the rename never runs.
		if err := os.Chmod(realDir, 0o555); err != nil {
			t.Fatal(err)
		}
		defer os.Chmod(realDir, 0o755)
		if err := setup.WriteBlock(link); err == nil {
			t.Fatalf("WriteBlock with a read-only target folder = nil, want an error")
		}
		wantLink(t, link)
		if got := read(t, target); got != old {
			t.Fatalf("after a failed write = %q, want the old bytes untouched", got)
		}
	})

	t.Run("symlink write lands its temp file in the target dir", func(t *testing.T) {
		base := t.TempDir()
		realDir := filepath.Join(base, "real")
		linkDir := filepath.Join(base, "links")
		if err := os.Mkdir(realDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(linkDir, 0o755); err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(realDir, "NOTES.md")
		link := filepath.Join(linkDir, "CLAUDE.md")
		old := "# Notes\n"
		write(t, target, old)
		if err := os.Symlink(target, link); err != nil {
			t.Fatal(err)
		}
		// A locked link folder must not stop the write: the temp file
		// belongs in the target's folder, not the link's.
		if err := os.Chmod(linkDir, 0o555); err != nil {
			t.Fatal(err)
		}
		defer os.Chmod(linkDir, 0o755)
		if err := setup.WriteBlock(link); err != nil {
			t.Fatalf("WriteBlock with a read-only link folder = %v, want nil", err)
		}
		wantLink(t, link)
		if got := read(t, target); got != old+wantBlock {
			t.Fatalf("target = %q, want old bytes then the block", got)
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

func wantLink(t *testing.T, path string) {
	t.Helper()
	fi, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("%s is no longer a symlink after WriteBlock", path)
	}
}
