package hook

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// brainStems covers every command shape the hook has to tell apart. The
// quoted ones matter most: a brainstorm written inside an echo is a message,
// not a command, so it must not start a session's one brainstorm.
var brainStems = []struct {
	name, cmd, stem string
	ok              bool
}{
	{"plain", "acta set scratch/2026-09-28-harness-spec status brainstorming", "2026-09-28-harness-spec", true},
	{"chained and", "cd x && acta set scratch/a-b status brainstorming", "a-b", true},
	{"chained semicolon", "cd x; acta set scratch/a-b status brainstorming", "a-b", true},
	{"chained or", "false || acta set scratch/a-b status brainstorming", "a-b", true},
	{"chained after", "acta set scratch/a-b status brainstorming && acta build .", "a-b", true},
	{"unquoted in echo", "cd x && echo acta set scratch/a-b status brainstorming", "", false},
	{"unquoted in cat", `cat <<EOF | tee /tmp/note
acta set scratch/a-b status brainstorming
EOF`, "", false},
	{"extra space", "acta  set scratch/a-b status brainstorming", "", false},
	{"other status", "acta set scratch/a-b status raw", "", false},
	{"other kind", "acta set specs/a-b status brainstorming", "", false},
	{"quoted in echo", `echo "acta set scratch/a-b status brainstorming"`, "", false},
	{"quoted after and", `cd x && echo "acta set scratch/a-b status brainstorming"`, "", false},
	{"single quoted in echo", `echo 'acta set scratch/a-b status brainstorming'`, "", false},
	{"trailing words", "acta set scratch/a-b status brainstorming now", "", false},
	{"leading words", "yes acta set scratch/a-b status brainstorming", "", false},
	{"wrong binary", "acta-sc set scratch/a-b status brainstorming", "", false},
	{"empty", "", "", false},
}

func TestBrainstormStem(t *testing.T) {
	for _, c := range brainStems {
		stem, ok := BrainstormStem(c.cmd)
		if stem != c.stem || ok != c.ok {
			t.Errorf("%s: BrainstormStem(%q) = %q,%v want %q,%v", c.name, c.cmd, stem, ok, c.stem, c.ok)
		}
	}
}

func ev(session, cmd string) ToolEvent {
	var e ToolEvent
	e.SessionID, e.ToolInput.Command = session, cmd
	return e
}

// scratchFile writes a scratch item file, the way `acta scratch new` leaves it.
func scratchFile(t *testing.T, root, stem, body string) {
	t.Helper()
	dir := filepath.Join(root, "scratch")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, stem+".md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// putState writes the state file, for the broken and hand-made cases.
func putState(t *testing.T, root, body string) {
	t.Helper()
	dir := filepath.Join(root, "state")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, stateFile), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestPreToolBlocksSecondItemOnly(t *testing.T) {
	root := t.TempDir()
	scratchFile(t, root, "one", "---\nid: SCRATCH-1\n---\n# One\n")
	scratchFile(t, root, "two", "---\nid: SCRATCH-2\n---\n# Two\n")
	first := ev("s1", "acta set scratch/one status brainstorming")
	if block, _ := PreTool(root, first); block {
		t.Fatal("first brainstorm must pass")
	}
	if err := RecordBrainstorm(root, first); err != nil {
		t.Fatal(err)
	}
	if block, _ := PreTool(root, first); block {
		t.Fatal("same item again must pass")
	}
	block, msg := PreTool(root, ev("s1", "acta set scratch/two status brainstorming"))
	if !block || !strings.Contains(msg, "one") || !strings.Contains(msg, "the choices rule 8 names") {
		t.Fatalf("second item: block=%v msg=%q", block, msg)
	}
	if block, _ := PreTool(root, ev("s2", "acta set scratch/two status brainstorming")); block {
		t.Fatal("another session must pass")
	}
}

// A stem with no scratch file is a mistype, not a brainstorm. It is never
// recorded and never blocks, so a failed set cannot block the real one.
func TestMistypedStemCountsForNothing(t *testing.T) {
	makeRoot := func(t *testing.T) string {
		t.Helper()
		root := t.TempDir()
		scratchFile(t, root, "2026-10-05-x", "---\nid: SCRATCH-1\n---\n# X\n")
		scratchFile(t, root, "2026-10-05-y", "---\nid: SCRATCH-2\n---\n# Y\n")
		return root
	}
	brain := func(stem string) string { return "acta set scratch/" + stem + " status brainstorming" }
	// Mistyped then real: neither recorded nor blocked.
	t.Run("mistyped then real", func(t *testing.T) {
		root := makeRoot(t)
		if err := RecordBrainstorm(root, ev("s1", brain("x"))); err != nil {
			t.Fatal(err)
		}
		if got := Reminder(root, "s1"); got != "" {
			t.Fatalf("mistyped stem was recorded: reminder=%q", got)
		}
		if block, msg := PreTool(root, ev("s1", brain("2026-10-05-x"))); block {
			t.Fatalf("real item after a mistype blocked: msg=%q", msg)
		}
	})
	// Real then mistyped: the late mistype changes nothing.
	t.Run("real then mistyped", func(t *testing.T) {
		root := makeRoot(t)
		if err := RecordBrainstorm(root, ev("s1", brain("2026-10-05-x"))); err != nil {
			t.Fatal(err)
		}
		if err := RecordBrainstorm(root, ev("s1", brain("x"))); err != nil {
			t.Fatal(err)
		}
		if got, want := Reminder(root, "s1"), "SCRATCH-1"; !strings.Contains(got, want) {
			t.Fatalf("reminder=%q, want it to still name %s", got, want)
		}
		if block, _ := PreTool(root, ev("s1", brain("x"))); block {
			t.Fatal("mistype after the real item must pass")
		}
	})
	// Same real item twice: still one brainstorm.
	t.Run("same real twice", func(t *testing.T) {
		root := makeRoot(t)
		if err := RecordBrainstorm(root, ev("s1", brain("2026-10-05-x"))); err != nil {
			t.Fatal(err)
		}
		if block, _ := PreTool(root, ev("s1", brain("2026-10-05-x"))); block {
			t.Fatal("same item again must pass")
		}
	})
	// Two different real items: still blocked.
	t.Run("two different real items", func(t *testing.T) {
		root := makeRoot(t)
		if err := RecordBrainstorm(root, ev("s1", brain("2026-10-05-x"))); err != nil {
			t.Fatal(err)
		}
		if block, _ := PreTool(root, ev("s1", brain("2026-10-05-y"))); !block {
			t.Fatal("two different real items must still block")
		}
	})
	// A link is not a file: following it would read outside scratch.
	t.Run("link is not a file", func(t *testing.T) {
		root := makeRoot(t)
		if err := os.Symlink("2026-10-05-x.md", filepath.Join(root, "scratch", "linked.md")); err != nil {
			t.Fatal(err)
		}
		if err := RecordBrainstorm(root, ev("s1", brain("linked"))); err != nil {
			t.Fatal(err)
		}
		if got := Reminder(root, "s1"); got != "" {
			t.Fatalf("linked stem was recorded: reminder=%q", got)
		}
		if block, _ := PreTool(root, ev("s1", brain("linked"))); block {
			t.Fatal("linked stem must pass")
		}
	})
}

func TestBlockAndReminderNameScratchID(t *testing.T) {
	root := t.TempDir()
	scratchFile(t, root, "one", "---\nid: SCRATCH-6\nhash: a1b2\nstatus: brainstorming\n---\n# One\n")
	scratchFile(t, root, "two", "---\nid: SCRATCH-7\n---\n# Two\n")
	one := ev("s1", "acta set scratch/one status brainstorming")
	if err := RecordBrainstorm(root, one); err != nil {
		t.Fatal(err)
	}
	_, msg := PreTool(root, ev("s1", "acta set scratch/two status brainstorming"))
	want := "acta: this session already brainstormed SCRATCH-6. One Architectural brainstorm per session: " +
		"file this one as a scratch item and offer the user the choices rule 8 names."
	if msg != want {
		t.Errorf("block message = %q, want %q", msg, want)
	}
	wantReminder := "acta: this session already brainstormed SCRATCH-6; a second brainstorm gets the choices rule 8 names."
	if got := Reminder(root, "s1"); got != wantReminder {
		t.Errorf("Reminder = %q, want %q", got, wantReminder)
	}
}

func TestLabelFallsBackToStem(t *testing.T) {
	root := t.TempDir()
	// The file is gone: the state still names the item, by its stem.
	scratchFile(t, root, "one", "---\nid: SCRATCH-6\n---\n# One\n")
	if err := RecordBrainstorm(root, ev("s1", "acta set scratch/one status brainstorming")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, "scratch", "one.md")); err != nil {
		t.Fatal(err)
	}
	// A file whose frontmatter has no id names the item by its stem, even
	// when the body below the frontmatter talks about an id.
	scratchFile(t, root, "two", "---\nhash: c3d4\n---\n# Two\n\nA pasted example of the shape:\nid: SCRATCH-9\nstatus: raw\n")
	if err := RecordBrainstorm(root, ev("s2", "acta set scratch/two status brainstorming")); err != nil {
		t.Fatal(err)
	}
	scratchFile(t, root, "three", "---\nid: SCRATCH-8\n---\n# Three\n")
	_, msg := PreTool(root, ev("s1", "acta set scratch/three status brainstorming"))
	if !strings.Contains(msg, "brainstormed one.") {
		t.Errorf("missing file: block message = %q, want the stem 'one'", msg)
	}
	if got := Reminder(root, "s2"); !strings.Contains(got, "brainstormed two;") {
		t.Errorf("no id: Reminder = %q, want the stem 'two'", got)
	}
}

// A state file that is not a usable map must read as a session that has
// brainstormed nothing. The type-error case matters most: json keeps the
// entries it could read before it hit the bad one, so a reader that ignored
// the error would still block a session.
func TestBrokenStateNeverBlocks(t *testing.T) {
	for _, broken := range []string{"{not json", "", "null", `{"s1":42,"s2":"two"}`, `["one"]`} {
		t.Run(broken, func(t *testing.T) {
			root := t.TempDir()
			scratchFile(t, root, "one", "---\nid: SCRATCH-1\n---\n# One\n")
			scratchFile(t, root, "two", "---\nid: SCRATCH-2\n---\n# Two\n")
			putState(t, root, broken)
			if block, msg := PreTool(root, ev("s1", "acta set scratch/two status brainstorming")); block {
				t.Errorf("blocked on a broken state file: msg=%q", msg)
			}
			if block, msg := PreTool(root, ev("s2", "acta set scratch/two status brainstorming")); block {
				t.Errorf("blocked on a broken state file: msg=%q", msg)
			}
			if r := Reminder(root, "s1") + Reminder(root, "s2"); r != "" {
				t.Errorf("broken state file gave the reminder %q", r)
			}
			// Writing on top of a broken file must not crash: a hook that
			// panics would take the whole session with it.
			if err := RecordBrainstorm(root, ev("s9", "acta set scratch/one status brainstorming")); err != nil {
				t.Errorf("RecordBrainstorm on a broken state file: %v", err)
			}
			if got := mustRead(t, filepath.Join(root, stateFile)); got != `{"s9":"one"}` {
				t.Errorf("state file after the repair = %s, want only the new session", got)
			}
		})
	}
}

func TestMissingStateNeverBlocks(t *testing.T) {
	root := t.TempDir()
	if block, msg := PreTool(root, ev("s1", "acta set scratch/one status brainstorming")); block {
		t.Fatalf("no state file at all: block=%v msg=%q", block, msg)
	}
	if Reminder(root, "s1") != "" {
		t.Fatal("no state file must give no reminder")
	}
	if _, err := os.Stat(filepath.Join(root, stateFile)); !os.IsNotExist(err) {
		t.Fatal("reading must not create the state file")
	}
}

// TestNonBrainstormTouchesNothing runs every input that only looks like a
// brainstorm against all three entry points. None may block, remind, or leave
// a key in the state file.
func TestNonBrainstormTouchesNothing(t *testing.T) {
	for _, c := range brainStems {
		if c.ok {
			continue
		}
		t.Run(c.name, func(t *testing.T) {
			root := t.TempDir()
			scratchFile(t, root, "one", "---\nid: SCRATCH-6\n---\n# One\n")
			putState(t, root, `{"s1":"one"}`)
			before := mustRead(t, filepath.Join(root, stateFile))

			if block, msg := PreTool(root, ev("s1", c.cmd)); block {
				t.Errorf("blocked a command that is not a brainstorm: msg=%q", msg)
			}
			if err := RecordBrainstorm(root, ev("s1", c.cmd)); err != nil {
				t.Errorf("RecordBrainstorm: %v", err)
			}
			if got := mustRead(t, filepath.Join(root, stateFile)); got != before {
				t.Errorf("state file changed to %s, want %s", got, before)
			}
			if got := Reminder(root, "s2"); got != "" {
				t.Errorf("Reminder for a session that never brainstormed = %q", got)
			}
		})
	}
}

func TestEmptyEventNeverBlocks(t *testing.T) {
	root := t.TempDir()
	putState(t, root, `{"":"one"}`)
	if block, _ := PreTool(root, ToolEvent{}); block {
		t.Error("an empty event must not block")
	}
	// The state file is the whole truth: a command with no session id writes
	// nothing, so no dead key can pile up in it.
	for _, cmd := range []string{"acta set scratch/one status brainstorming", "acta set scratch/two status brainstorming"} {
		if err := RecordBrainstorm(root, ev("", cmd)); err != nil {
			t.Fatal(err)
		}
		if got := mustRead(t, filepath.Join(root, stateFile)); got != `{"":"one"}` {
			t.Errorf("an event with no session id wrote %s", got)
		}
	}
}

func TestRecordBrainstormKeepsOtherSessions(t *testing.T) {
	root := t.TempDir()
	scratchFile(t, root, "one", "---\nid: SCRATCH-1\n---\n# One\n")
	scratchFile(t, root, "two", "---\nid: SCRATCH-2\n---\n# Two\n")
	scratchFile(t, root, "three", "---\nid: SCRATCH-3\n---\n# Three\n")
	if err := RecordBrainstorm(root, ev("s1", "acta set scratch/one status brainstorming")); err != nil {
		t.Fatal(err)
	}
	if err := RecordBrainstorm(root, ev("s2", "acta set scratch/two status brainstorming")); err != nil {
		t.Fatal(err)
	}
	if err := RecordBrainstorm(root, ev("s1", "acta set scratch/three status brainstorming")); err != nil {
		t.Fatal(err)
	}
	if got := mustRead(t, filepath.Join(root, stateFile)); got != `{"s1":"three","s2":"two"}` {
		t.Errorf("state = %s, want both sessions with the last stem each", got)
	}
	// The rename leaves no temp file behind in state/.
	entries, err := os.ReadDir(filepath.Join(root, "state"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != filepath.Base(stateFile) {
		t.Errorf("state folder holds %v, want only sessions.json", entries)
	}
}

// brokenReader stands in for a stdin that cannot be read.
type brokenReader struct{}

func (brokenReader) Read([]byte) (int, error) { return 0, errors.New("stdin is gone") }

func TestParseEvent(t *testing.T) {
	for _, in := range []string{"", "{bad", `{"tool_input":{"command":"x"}}`, "null", "[]", `{"session_id":""}`} {
		if _, ok := ParseEvent(strings.NewReader(in)); ok {
			t.Errorf("ParseEvent(%q) ok, want false", in)
		}
	}
	if _, ok := ParseEvent(brokenReader{}); ok {
		t.Error("a read error must not parse")
	}
	e, ok := ParseEvent(strings.NewReader(`{"session_id":"s1","tool_input":{"command":"ls"}}`))
	if !ok || e.SessionID != "s1" || e.ToolInput.Command != "ls" {
		t.Fatalf("got %+v %v", e, ok)
	}
	// A missing tool_input is still an event: it just has no command.
	e, ok = ParseEvent(strings.NewReader(`{"session_id":"s1"}`))
	if !ok || e.ToolInput.Command != "" {
		t.Fatalf("got %+v %v", e, ok)
	}
}

func TestReminderNamesItem(t *testing.T) {
	root := t.TempDir()
	scratchFile(t, root, "one", "---\nid: SCRATCH-1\n---\n# One\n")
	if err := RecordBrainstorm(root, ev("s1", "acta set scratch/one status brainstorming")); err != nil {
		t.Fatal(err)
	}
	if r := Reminder(root, "s1"); !strings.Contains(r, "one") && !strings.Contains(r, "SCRATCH-1") {
		t.Fatalf("Reminder = %q", r)
	}
	if Reminder(root, "s2") != "" {
		t.Fatal("other session must get no reminder")
	}
	if err := RecordBrainstorm(root, ev("s3", "ls -la")); err != nil {
		t.Fatal(err)
	}
	if Reminder(root, "s3") != "" {
		t.Fatal("non-matching command must not be recorded")
	}
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
