package setup_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/setup"
)

func testUser() config.User {
	return config.User{
		ChatLanguage:   "Korean",
		Style:          "adhd",
		Tone:           "Casual, short sentences.",
		RepoLanguage:   "English",
		BuildExecutor:  "subagent",
		SubagentModels: "split",
		PlanDepth:      "full",
		CommitHistory:  "full",
		Questions:      "probe",
	}
}

func TestPlan(t *testing.T) {
	u := testUser()
	other := config.User{ChatLanguage: "English", Style: "plain", RepoLanguage: "English"}
	claudeInstall := [][]string{
		{"claude", "plugin", "marketplace", "add", "/p"},
		{"claude", "plugin", "install", "acta@acta-local"},
	}
	ompInstall := [][]string{{"omp", "plugin", "link", "/p"}}
	noTTY := setup.Action{Kind: "print", Path: "acta setup needs a terminal: nothing was written. Set each value by hand with `acta config set`, or re-run `acta setup` in a terminal."}

	rows := []struct {
		name string
		a    setup.Answers
		e    setup.Env
		want []setup.Action
	}{
		{
			name: "no tty with everything set returns only the print",
			a:    setup.Answers{User: u, Install: map[string]bool{"claude": true, "omp": true}},
			e: setup.Env{TTY: false, Harnesses: []string{"claude", "omp"},
				PluginDir: "/p", RepoRoot: "/r", HasClaudeMD: true, HasAgentsMD: true, Current: other},
			want: []setup.Action{noTTY},
		},
		{
			name: "no tty outside repo without harness returns only the print",
			a:    setup.Answers{User: u},
			e:    setup.Env{TTY: false, Current: other},
			want: []setup.Action{noTTY},
		},
		{
			name: "full tty run with both files writes two blocks",
			a:    setup.Answers{User: u, Install: map[string]bool{"claude": true, "omp": true}},
			e: setup.Env{TTY: true, Harnesses: []string{"claude", "omp"},
				PluginDir: "/p", RepoRoot: "/r", HasClaudeMD: true, HasAgentsMD: true, Current: other},
			want: []setup.Action{
				{Kind: "config", User: u},
				{Kind: "install", Harness: "claude", Argv: claudeInstall},
				{Kind: "install", Harness: "omp", Argv: ompInstall},
				{Kind: "block", Path: "/r/CLAUDE.md"},
				{Kind: "block", Path: "/r/AGENTS.md"},
			},
		},
		{
			name: "claude only with dir writes claude block",
			a:    setup.Answers{User: u, Install: map[string]bool{"claude": true}},
			e: setup.Env{TTY: true, Harnesses: []string{"claude"},
				PluginDir: "/p", RepoRoot: "/r", HasClaudeMD: true, Current: other},
			want: []setup.Action{
				{Kind: "config", User: u},
				{Kind: "install", Harness: "claude", Argv: claudeInstall},
				{Kind: "block", Path: "/r/CLAUDE.md"},
			},
		},
		{
			name: "omp only without dir prints the placeholder command",
			a:    setup.Answers{User: u, Install: map[string]bool{"omp": true}},
			e: setup.Env{TTY: true, Harnesses: []string{"omp"},
				RepoRoot: "/r", HasAgentsMD: true, Current: other},
			want: []setup.Action{
				{Kind: "config", User: u},
				{Kind: "print", Harness: "omp", Path: "omp plugin link <plugin dir>"},
				{Kind: "block", Path: "/r/AGENTS.md"},
			},
		},
		{
			name: "both harnesses without dir print both placeholder commands",
			a:    setup.Answers{User: u, Install: map[string]bool{"claude": true, "omp": true}},
			e: setup.Env{TTY: true, Harnesses: []string{"claude", "omp"},
				RepoRoot: "/r", HasClaudeMD: true, HasAgentsMD: true, Current: other},
			want: []setup.Action{
				{Kind: "config", User: u},
				{Kind: "print", Harness: "claude", Path: "claude plugin marketplace add <plugin dir>\nclaude plugin install acta@acta-local"},
				{Kind: "print", Harness: "omp", Path: "omp plugin link <plugin dir>"},
				{Kind: "block", Path: "/r/CLAUDE.md"},
				{Kind: "block", Path: "/r/AGENTS.md"},
			},
		},
		{
			name: "harness found but user said no writes nothing for it",
			a:    setup.Answers{User: u, Install: map[string]bool{"claude": false}},
			e: setup.Env{TTY: true, Harnesses: []string{"claude", "omp"},
				PluginDir: "/p", RepoRoot: "/r", Current: other},
			want: []setup.Action{
				{Kind: "config", User: u},
				{Kind: "block", Path: "/r/CLAUDE.md"},
			},
		},
		{
			name: "no file at all creates claude md never agents md",
			a:    setup.Answers{User: u},
			e:    setup.Env{TTY: true, RepoRoot: "/r", Current: other},
			want: []setup.Action{
				{Kind: "config", User: u},
				{Kind: "block", Path: "/r/CLAUDE.md"},
			},
		},
		{
			name: "outside repo writes no block even when file flags are set",
			a:    setup.Answers{User: u, Install: map[string]bool{"claude": true}},
			e: setup.Env{TTY: true, Harnesses: []string{"claude"},
				PluginDir: "/p", HasClaudeMD: true, HasAgentsMD: true, Current: other},
			want: []setup.Action{
				{Kind: "config", User: u},
				{Kind: "install", Harness: "claude", Argv: claudeInstall},
			},
		},
		{
			name: "no harness and no repo leaves only the config write",
			a:    setup.Answers{User: u},
			e:    setup.Env{TTY: true, Current: other},
			want: []setup.Action{{Kind: "config", User: u}},
		},
		{
			name: "unknown harness yields no install",
			a:    setup.Answers{User: u, Install: map[string]bool{"weird": true}},
			e:    setup.Env{TTY: true, Harnesses: []string{"weird"}, PluginDir: "/p", Current: other},
			want: []setup.Action{{Kind: "config", User: u}},
		},
		{
			name: "no dir prints by hand even when the answer map says no",
			a:    setup.Answers{User: u, Install: map[string]bool{"claude": false}},
			e: setup.Env{TTY: true, Harnesses: []string{"claude"},
				RepoRoot: "/r", HasClaudeMD: true, Current: other},
			want: []setup.Action{
				{Kind: "config", User: u},
				{Kind: "print", Harness: "claude", Path: "claude plugin marketplace add <plugin dir>\nclaude plugin install acta@acta-local"},
				{Kind: "block", Path: "/r/CLAUDE.md"},
			},
		},
		{
			name: "installed tool is noted, never installed or printed, with a dir",
			a:    setup.Answers{User: u, Install: map[string]bool{}},
			e: setup.Env{TTY: true, Harnesses: []string{"claude"}, Installed: map[string]bool{"claude": true},
				PluginDir: "/p", Current: other},
			want: []setup.Action{
				{Kind: "config", User: u},
				{Kind: "installed", Harness: "claude"},
			},
		},
		{
			name: "installed tool is noted, never printed, without a dir",
			a:    setup.Answers{User: u},
			e: setup.Env{TTY: true, Harnesses: []string{"claude"}, Installed: map[string]bool{"claude": true},
				Current: other},
			want: []setup.Action{
				{Kind: "config", User: u},
				{Kind: "installed", Harness: "claude"},
			},
		},
		{
			name: "mix with a dir: claude installed, omp not and said yes",
			a:    setup.Answers{User: u, Install: map[string]bool{"omp": true}},
			e: setup.Env{TTY: true, Harnesses: []string{"claude", "omp"}, Installed: map[string]bool{"claude": true},
				PluginDir: "/p", Current: other},
			want: []setup.Action{
				{Kind: "config", User: u},
				{Kind: "installed", Harness: "claude"},
				{Kind: "install", Harness: "omp", Argv: ompInstall},
			},
		},
		{
			name: "mix without a dir: omp installed, claude printed by hand",
			a:    setup.Answers{User: u},
			e: setup.Env{TTY: true, Harnesses: []string{"claude", "omp"}, Installed: map[string]bool{"omp": true},
				Current: other},
			want: []setup.Action{
				{Kind: "config", User: u},
				{Kind: "print", Harness: "claude", Path: "claude plugin marketplace add <plugin dir>\nclaude plugin install acta@acta-local"},
				{Kind: "installed", Harness: "omp"},
			},
		},
		{
			name: "nil install map installs nothing",
			a:    setup.Answers{User: u},
			e: setup.Env{TTY: true, Harnesses: []string{"claude"},
				PluginDir: "/p", RepoRoot: "/r", HasAgentsMD: true, Current: other},
			want: []setup.Action{
				{Kind: "config", User: u},
				{Kind: "block", Path: "/r/AGENTS.md"},
			},
		},
	}

	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			got := setup.Plan(r.a, r.e)
			if !reflect.DeepEqual(got, r.want) {
				t.Fatalf("Plan() = %#v, want %#v", got, r.want)
			}
		})
	}
}

// TestPlanMatrix covers every combination of TTY, harness set, plugin dir,
// repo presence and CLAUDE.md/AGENTS.md presence: 2*4*2*2*4 = 128 cases.
func TestPlanMatrix(t *testing.T) {
	u := testUser()
	harnessSets := map[string][]string{
		"none":   nil,
		"claude": {"claude"},
		"omp":    {"omp"},
		"both":   {"claude", "omp"},
	}
	fileSets := map[string][2]bool{
		"both":   {true, true},
		"claude": {true, false},
		"agents": {false, true},
		"none":   {false, false},
	}

	for _, tty := range []bool{false, true} {
		for hName, hs := range harnessSets {
			for _, dir := range []string{"", "/p"} {
				for _, repo := range []string{"", "/r"} {
					for fName, fs := range fileSets {
						name := strings.Join([]string{b(tty), hName, d(dir), r(repo), fName}, "/")
						t.Run(name, func(t *testing.T) {
							a := setup.Answers{User: u,
								Install: map[string]bool{"claude": true, "omp": true}}
							e := setup.Env{TTY: tty, Harnesses: hs,
								PluginDir: dir, RepoRoot: repo,
								HasClaudeMD: fs[0], HasAgentsMD: fs[1]}
							got := setup.Plan(a, e)

							if !tty {
								if len(got) != 1 {
									t.Fatalf("no-TTY Plan() = %d actions, want exactly 1 print", len(got))
								}
								if got[0].Kind != "print" || !strings.Contains(got[0].Path, "acta config set") {
									t.Fatalf("no-TTY Plan() = %#v, want one print naming acta config set", got)
								}
								return
							}

							var cfgs, installs, prints, blocks []setup.Action
							for _, act := range got {
								switch act.Kind {
								case "config":
									cfgs = append(cfgs, act)
								case "install":
									installs = append(installs, act)
								case "print":
									prints = append(prints, act)
								case "block":
									blocks = append(blocks, act)
								default:
									t.Fatalf("unknown action kind %q", act.Kind)
								}
							}

							if len(cfgs) != 1 || !reflect.DeepEqual(cfgs[0].User, u) {
								t.Fatalf("want exactly one config action with the answers user, got %#v", cfgs)
							}

							wantInst, wantPrint := 0, 0
							for range hs {
								if dir == "" {
									wantPrint++
								} else {
									wantInst++
								}
							}
							if len(installs) != wantInst {
								t.Fatalf("want %d install actions, got %#v", wantInst, installs)
							}
							if len(prints) != wantPrint {
								t.Fatalf("want %d print actions, got %#v", wantPrint, prints)
							}
							for _, p := range prints {
								if !strings.Contains(p.Path, "<plugin dir>") {
									t.Fatalf("placeholder print lost its placeholder: %#v", p)
								}
							}
							for i, act := range installs {
								if act.Harness != hs[i] || len(act.Argv) == 0 {
									t.Fatalf("install out of order or empty: %#v", installs)
								}
							}

							wantBlocks := 0
							if repo != "" {
								wantBlocks = 1
								if fs[0] && fs[1] {
									wantBlocks = 2
								}
							}
							if len(blocks) != wantBlocks {
								t.Fatalf("want %d block actions, got %#v", wantBlocks, blocks)
							}
							for _, b := range blocks {
								if strings.HasSuffix(b.Path, "AGENTS.md") && !fs[1] {
									t.Fatalf("must never create AGENTS.md, got %#v", blocks)
								}
							}
							if repo != "" {
								if !(fs[0] || fs[1]) && (len(blocks) != 1 || !strings.HasSuffix(blocks[0].Path, "CLAUDE.md")) {
									t.Fatalf("no file yet must create CLAUDE.md only, got %#v", blocks)
								}
								if fs[0] && fs[1] {
									if !strings.HasSuffix(blocks[0].Path, "CLAUDE.md") || !strings.HasSuffix(blocks[1].Path, "AGENTS.md") {
										t.Fatalf("both files must each get a block, CLAUDE.md first, got %#v", blocks)
									}
								}
							}

							if got[0].Kind != "config" {
								t.Fatalf("config must come first, got %#v", got)
							}
						})
					}
				}
			}
		}
	}
}

func b(tty bool) string {
	if tty {
		return "tty"
	}
	return "notty"
}

func d(dir string) string {
	if dir == "" {
		return "nodir"
	}
	return "dir"
}

func r(repo string) string {
	if repo == "" {
		return "norepo"
	}
	return "repo"
}
