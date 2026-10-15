package doctor

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// codexEnv is env with Codex found and its config.toml holding conf. An empty
// conf means Codex has no config.toml at all.
func codexEnv(t *testing.T, conf string) Env {
	t.Helper()
	e := env(t)
	e.CodexDir = filepath.Join(e.Home, ".codex")
	if conf != "" {
		if err := os.MkdirAll(e.CodexDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(e.CodexDir, "config.toml"), []byte(conf), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return e
}

const codexOn = "[marketplaces.acta-local]\nsource = \"/x\"\n\n[plugins.\"acta@acta-local\"]\nenabled = true\n"

// A user without Codex must not see a line about it.
func TestDoctorSaysNothingAboutCodexWhenItIsMissing(t *testing.T) {
	if got := byName(Run(env(t)), "codex"); got.Name != "" {
		t.Fatalf("codex line without Codex: %+v", got)
	}
}

// Every state of the Codex config. A missing plugin only warns: Codex is an
// extra harness, so it never fails the run.
func TestDoctorCodexStates(t *testing.T) {
	cases := []struct {
		name, conf string
		want       Level
	}{
		{"enabled", codexOn, OK},
		{"enabled under another marketplace", "[plugins.\"acta@iyay-acta\"]\nenabled = true\n", OK},
		{"enabled with a comment after it", "[plugins.\"acta@acta-local\"]\nenabled = true # installed\n", OK},
		{"a tab before true", "[plugins.\"acta@acta-local\"]\nenabled =\ttrue\n", OK},
		{"windows line ends", "[plugins.\"acta@acta-local\"]\r\nenabled = true\r\n", OK},
		{"a quoted enabled key", "[plugins.\"acta@acta-local\"]\n\"enabled\" = true\n", OK},
		{"a single-quoted header with spaces", "[ plugins.'acta@acta-local' ]\nenabled = true\n", OK},
		{"enabled only inside a comment", "[plugins.\"acta@acta-local\"]\n# enabled = true\n", Warn},
		{"enabled is not exactly true", "[plugins.\"acta@acta-local\"]\nenabled = truex\n", Warn},
		{"no config file", "", Warn},
		{"plugin turned off", "[plugins.\"acta@acta-local\"]\nenabled = false\n", Warn},
		{"enabled line under another plugin", "[plugins.\"acta@acta-local\"]\n\n[plugins.\"other@x\"]\nenabled = true\n", Warn},
		{"a plugin whose name only starts with acta", "[plugins.\"actaish@x\"]\nenabled = true\n", Warn},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rs := Run(codexEnv(t, c.conf))
			r := byName(rs, "codex")
			if r.Level != c.want {
				t.Fatalf("level %q want %q: %+v", r.Level, c.want, r)
			}
			if c.want == OK && !strings.Contains(r.Msg, "/hooks") {
				t.Errorf("OK line does not tell the user to trust the hooks in /hooks: %q", r.Msg)
			}
			if c.want == Warn && r.Fix == "" {
				t.Errorf("warn line has no fix: %+v", r)
			}
		})
	}
}

// The codex line comes right after harness, and only when Codex is there.
func TestDoctorCodexLineOrder(t *testing.T) {
	want := []string{"binary", "build", "harness", "codex", "stale-links", "conflicts", "repo", "schema", "files", "wiki", "agents-view", "setup", "theme"}
	if got := names(Run(codexEnv(t, codexOn))); !reflect.DeepEqual(got, want) {
		t.Fatalf("order %v want %v", got, want)
	}
}

// A user who runs acta only in Codex has the plugin installed, so the harness
// check must pass for them.
func TestDoctorHarnessCountsCodex(t *testing.T) {
	e := codexEnv(t, codexOn)
	if r := byName(Run(e), "harness"); r.Level != OK {
		t.Fatalf("harness with acta in Codex only: %+v", r)
	}
	if !PluginInstalled(e, "codex") {
		t.Fatal(`PluginInstalled(e, "codex") = false with acta enabled`)
	}
}
