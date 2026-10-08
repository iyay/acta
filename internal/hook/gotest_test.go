package hook

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// The block text, written out here on purpose: a change to the words an agent
// reads is a change to the product, not to a test.
const goTestWant = "acta: run tests with scripts/test, not go test. It adds -short and the machine-wide lock. Use: scripts/test"

// goTestRepo makes a git repo that keeps its tests in scripts/test, which is
// what turns the block on, plus a subfolder to start the walk up from.
func goTestRepo(t *testing.T, gitIsFile bool) (string, string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "scripts", "test"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	git := filepath.Join(dir, ".git")
	if gitIsFile {
		// A worktree has a .git file pointing at the real folder.
		if err := os.WriteFile(git, []byte("gitdir: /somewhere/.git/worktrees/wt\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	} else if err := os.MkdirAll(git, 0o755); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(dir, "internal", "hook")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	return dir, sub
}

// TestGoTestBlock walks every command shape the block has an opinion about.
// A row with an empty want must let the command run untouched, because a hook
// that stops the wrong work costs the user more than a run that slips past.
func TestGoTestBlock(t *testing.T) {
	dir, sub := goTestRepo(t, false)

	// A repo without scripts/test has no lock to join, so nothing is stopped.
	noScript := t.TempDir()
	if err := os.MkdirAll(filepath.Join(noScript, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	// No .git anywhere above, so there is no repo to speak of.
	plain := t.TempDir()

	for _, c := range []struct{ name, dir, command, want string }{
		{"at the start", dir, "go test ./...", goTestWant + " ./..."},
		{"after &&", dir, "go build ./... && go test ./internal/hook", goTestWant + " ./internal/hook"},
		{"after ||", dir, "go vet ./... || go test ./internal/cli -run TestHook", goTestWant + " ./internal/cli -run TestHook"},
		{"after ;", dir, "echo hi; go test ./...", goTestWant + " ./..."},
		{"after |", dir, "go test ./... | tee out.log", goTestWant + " ./..."},
		{"after (", dir, "(go test ./...)", goTestWant + " ./..."},
		{"after cd and &&", dir, "cd internal/hook && go test ./...", goTestWant + " ./..."},
		{"after rtk", dir, "rtk go test ./...", goTestWant + " ./..."},
		{"after rtk proxy", dir, "rtk proxy go test ./...", goTestWant + " ./..."},
		{"after env", dir, "env FOO=1 go test ./...", goTestWant + " ./..."},
		{"after an assignment", dir, "FOO=bar go test ./...", goTestWant + " ./..."},
		{"after time", dir, "time go test ./...", goTestWant + " ./..."},
		{"from a subfolder", sub, "go test ./...", goTestWant + " ./..."},
		{"no package at all", dir, "go test", goTestWant},
		{"with run-one in front", dir, "env FOO=1 run-one -- go test ./...", ""},
		{"flags that are not packages", dir, "go test -short -count=1 -timeout 1800s ./internal/hook", goTestWant + " ./internal/hook"},
		{"run with an equals sign", dir, "go test ./internal/hook -run=TestHook", goTestWant + " ./internal/hook -run=TestHook"},
		{"a quoted run value", dir, "go test ./internal/hook -run 'TestA|TestB'", goTestWant + " ./internal/hook -run 'TestA|TestB'"},
		{"a dollar quoted run value", dir, "go test ./internal/hook -run $'TestA|TestB'", goTestWant + " ./internal/hook -run $'TestA|TestB'"},

		{"text that only mentions it", dir, `grep "go test" f`, ""},
		{"an echo", dir, "echo go test", ""},
		{"a git log", dir, "git log --grep 'go test'", ""},
		{"scripts/test itself", dir, "scripts/test ./internal/hook -run TestHook", ""},
		{"the run-one wrapper", dir, "acta run-one -- go test -short ./...", ""},
		{"the run-one go run", dir, "go run ./cmd/acta run-one -- go test -short ./...", ""},
		{"a build", dir, "go build ./...", ""},
		{"an ls", dir, "ls -la", ""},
		{"a semicolon inside quotes", dir, `git commit -m "fix; go test passes"`, ""},
		{"an and inside quotes", dir, `echo "run go vet && go test ./x"`, ""},
		{"a pipe and brackets inside quotes", dir, `rg "(go test|scripts/test)" plugin`, ""},
		{"a bracket inside quotes", dir, `acta scratch new "idea: (go test ...) block"`, ""},
		{"a single quote inside double quotes", dir, `echo "don't run go test"`, ""},
		{"a quote that never closes", dir, `echo 'unfinished go test`, ""},
		{"a bracket inside single quotes", dir, `echo 'idea: (go test ./x) block'`, ""},
		{"an escaped quote inside quotes", dir, `git commit -m "a\"; go test ./x"`, ""},
		{"an empty command", dir, "", ""},
		{"a repo with no scripts/test", noScript, "go test ./...", ""},
		{"a folder with no repo above", plain, "go test ./...", ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			block, msg := GoTestBlock(c.dir, c.command)
			if block != (c.want != "") {
				t.Fatalf("block = %v for %q, want %v", block, c.command, c.want != "")
			}
			if msg != c.want {
				t.Errorf("message = %q, want %q", msg, c.want)
			}
		})
	}
}

// TestGoTestBlockWorktreeGitFile covers the worktree shape, where .git is a
// file instead of a folder. The walk must not care which one it finds.
func TestGoTestBlockWorktreeGitFile(t *testing.T) {
	dir, _ := goTestRepo(t, true)
	if block, msg := GoTestBlock(dir, "go test ./..."); !block || msg != goTestWant+" ./..." {
		t.Errorf("block = %v, message %q", block, msg)
	}
}

// TestGoTestBlockBadScriptFolder covers the folders where scripts/test cannot
// be used. A repo that cannot offer the script must keep its own commands.
func TestGoTestBlockBadScriptFolder(t *testing.T) {
	for _, c := range []struct {
		name  string
		make1 func(t *testing.T, dir string)
	}{
		{"scripts is a file", func(t *testing.T, dir string) {
			if err := os.WriteFile(filepath.Join(dir, "scripts"), []byte("x"), 0o644); err != nil {
				t.Fatal(err)
			}
		}},
		{"scripts/test is a folder", func(t *testing.T, dir string) {
			if err := os.MkdirAll(filepath.Join(dir, "scripts", "test"), 0o755); err != nil {
				t.Fatal(err)
			}
		}},
		{"scripts/test cannot be read", func(t *testing.T, dir string) {
			if os.Geteuid() == 0 {
				t.Skip("root reads every file, so this case cannot be built here")
			}
			if err := os.MkdirAll(filepath.Join(dir, "scripts"), 0o755); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, "scripts", "test")
			if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o000); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(path, 0o755) })
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
				t.Fatal(err)
			}
			c.make1(t, dir)
			if block, msg := GoTestBlock(dir, "go test ./..."); block || msg != "" {
				t.Errorf("block = %v, message %q, want no block", block, msg)
			}
		})
	}
}

// TestGoTestBlockNoDir keeps a missing folder from turning into a blocked
// command, because the hook must stay quiet when it cannot even look.
func TestGoTestBlockNoDir(t *testing.T) {
	for _, dir := range []string{"", filepath.Join(t.TempDir(), "gone", "deeper")} {
		if block, msg := GoTestBlock(dir, "go test ./..."); block || msg != "" {
			t.Errorf("dir %q: block = %v, message %q, want no block", dir, block, msg)
		}
	}
}

// TestHookChecksReadPowerShellLikeBash proves the shell checks never look at
// the tool name. On Windows the agent runs commands through PowerShell, and it
// sends the same tool_input.command, so each check must treat both tools alike.
func TestHookChecksReadPowerShellLikeBash(t *testing.T) {
	for _, tool := range []string{"Bash", "PowerShell"} {
		t.Run(tool, func(t *testing.T) {
			// The event goes through ParseEvent, the way the hook gets it.
			parse := func(command string) ToolEvent {
				in, err := json.Marshal(map[string]any{
					"session_id": "s1",
					"tool_name":  tool,
					"tool_input": map[string]string{"command": command},
				})
				if err != nil {
					t.Fatal(err)
				}
				e, ok := ParseEvent(bytes.NewReader(in))
				if !ok || e.ToolName != tool {
					t.Fatalf("ParseEvent gave %+v, ok %v", e, ok)
				}
				return e
			}

			// The go test block.
			dir, _ := goTestRepo(t, false)
			if block, msg := GoTestBlock(dir, parse("go test ./...").ToolInput.Command); !block || msg != goTestWant+" ./..." {
				t.Errorf("go test block = %v, %q", block, msg)
			}

			// The brainstorm record (post-tool) and the second-item block (pre-tool).
			root := t.TempDir()
			scratchFile(t, root, "one", "---\nid: SCRATCH-1\n---\n# One\n")
			scratchFile(t, root, "two", "---\nid: SCRATCH-2\n---\n# Two\n")
			if err := RecordBrainstorm(root, parse("acta set scratch/one status brainstorming")); err != nil {
				t.Fatal(err)
			}
			if block, _ := PreTool(root, parse("acta set scratch/two status brainstorming")); !block {
				t.Error("second brainstorm was not blocked")
			}

			// The wiki hint reads the words of the command.
			wroot, repo := hintWiki(t)
			t.Chdir(repo)
			if got := WikiHints(wroot, repo, parse("cat internal/tui/a.go")); got != tuiLine {
				t.Errorf("wiki hint = %q, want %q", got, tuiLine)
			}
		})
	}
}
