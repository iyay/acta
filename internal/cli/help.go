package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/iyay/acta/plugin"
)

// helpCmd is one command in the help table. The man page is built from the
// same table, so a line changed here changes both.
type helpCmd struct {
	name    string
	group   string
	summary string   // one sentence
	usage   []string // one line per way to call it
	flags   []string // one line each: the flag, two spaces, what it does
	example string   // one line; only commands that write files have one
}

// helpGroups is the order the top help shows the groups in.
var helpGroups = []string{"Board", "Write", "Setup", "Agent plumbing"}

// helpTable has one entry for every case in Run's switch. A test reads
// cli.go and fails when a case has no entry here. The old voice alias is
// not a command any more, so it has no entry.
var helpTable = []helpCmd{
	// Board
	{
		name:    "list",
		group:   "Board",
		summary: "List the open planning items of this repo and its worktrees.",
		usage:   []string{"acta list [--type story|task|bug] [--all] [--json]"},
		flags: []string{
			"--type T  show only story, task or bug items",
			"--all  include closed and legacy items",
			"--json  print JSON",
		},
	},
	{
		name:    "show",
		group:   "Board",
		summary: "Show one planning item by its id, short id or hash.",
		usage:   []string{"acta show <id> [--json] [--path]"},
		flags: []string{
			"--json  print JSON",
			"--path  print the file path only",
		},
	},
	{
		name:    "set",
		group:   "Board",
		summary: "Change one field of a planning item and commit the change.",
		usage:   []string{"acta set <id> status|type|title|fixed_in|ref|priority <value>"},
		example: "acta set SPC-0128 status done",
	},
	{
		name:    "state",
		group:   "Board",
		summary: "Show the running plans, or read and write the State part of one plan.",
		usage: []string{
			"acta state [plan]",
			"acta state set <plan> next|findings|rulings [--clear] < text.md",
		},
		flags:   []string{"--clear  empty the State part instead of writing lines (state set only)"},
		example: `echo "task 2 next" | acta state set PLN-0138 next`,
	},
	{
		name:    "tick",
		group:   "Board",
		summary: "Tick the checkboxes of a plan task so the board shows live progress.",
		usage:   []string{"acta tick <id> [--step N | --all | --start | --wontfix | --undo] [--agent NAME]"},
		flags: []string{
			"--step N  tick this checkbox, counting from 1",
			"--all  tick every checkbox of the task",
			"--start  mark the task started without ticking a box",
			"--wontfix  mark a debt line wontfix instead of done",
			"--undo  put the task or debt line back to open",
			"--agent NAME  name the agent running this tick (default: the AI_AGENT variable)",
		},
		example: "acta tick plans/2026-10-12-acta-help-and-man-page#task-01 --step 1",
	},
	{
		name:    "commits",
		group:   "Board",
		summary: "List the commits that name a plan or one of its tasks in a Task trailer.",
		usage:   []string{"acta commits <plan> [task] [--all] [--json]"},
		flags: []string{
			"--all  include chore commits",
			"--json  print JSON",
		},
	},

	// Write
	{
		name:    "bug",
		group:   "Write",
		summary: "File a confirmed bug in the planning folder.",
		usage:   []string{"acta bug new <slug> [--ref X] [--title T] [--priority P] < body.md"},
		flags: []string{
			"--ref X  outside ticket code",
			"--title T  bug title (default: the # line of the body, or the slug)",
			"--priority P  high, medium or low",
		},
		example: "acta bug new tidy-folds-late --priority high < body.md",
	},
	{
		name:    "debt",
		group:   "Write",
		summary: "File the review notes of a plan as a debt item. Every note line must start with (medium) or (high). drop-low closes every open low or untagged line in one commit.",
		usage:   []string{"acta debt new <plan id> [--title T] < notes.md", "acta debt drop-low"},
		flags:   []string{"--title T  debt heading (default: Review NOTEs: <plan title>)"},
		example: "printf '(medium) retry loop has no cap\\n' | acta debt new PLN-0138",
	},
	{
		name:    "scratch",
		group:   "Write",
		summary: "Drop a raw idea in the Scratchpad, or add text to one that is there.",
		usage: []string{
			"acta scratch new <slug> [--title T] < body.md",
			"acta scratch add <SCRATCH-n> [--section words|context|log|questions] < text.md",
		},
		flags: []string{
			"--title T  scratch title (default: the slug)",
			"--section S  words, context, log or questions (default: words); scratch add only",
		},
		example: `echo "idea for later" | acta scratch new later-idea`,
	},
	{
		name:    "commit",
		group:   "Write",
		summary: "Commit one planning file; repeat edits of the same file fold into its last commit.",
		usage:   []string{`acta commit <path> -m "<message>"`},
		flags:   []string{"-m MESSAGE  commit message (required)"},
		example: `acta commit plans/2026-10-12-acta-help-and-man-page.md -m "plan: help"`,
	},
	{
		name:    "id",
		group:   "Write",
		summary: "Give short ids and hashes to items that lack them, or repair duplicate numbers after a merge.",
		usage:   []string{"acta id [--fix-duplicates] [<path-id>...]"},
		flags:   []string{"--fix-duplicates  give the later of two items with one number the next free number"},
		example: "acta id --fix-duplicates",
	},
	{
		name:    "wiki",
		group:   "Write",
		summary: "List or match wiki pages, check errors and stale warnings, or optionally record a review.",
		usage: []string{
			"acta wiki ls [--type T]",
			"acta wiki match <file>...",
			"acta wiki check [<range>]",
			"acta wiki ok <page id>...",
		},
		flags:   []string{"--type T  list only the pages of this type (wiki ls only)"},
		example: "acta wiki match internal/cli/cli.go",
	},

	// Setup
	{
		name:    "setup",
		group:   "Setup",
		summary: "Run the setup wizard in a terminal: your settings, the plugin install and the doctor checks.",
		usage:   []string{"acta setup [--plugin-dir <path>]"},
		flags:   []string{"--plugin-dir PATH  folder holding the acta plugin, used for harness installs"},
		example: "acta setup",
	},
	{
		name:    "config",
		group:   "Setup",
		summary: "Show or change your acta settings, or the ones saved in .acta.yaml of this repo.",
		usage: []string{
			"acta config show [--json]",
			"acta config set [--language L] [--style adhd|plain] [--tone T] [--clear-tone] [--executor subagent|dispatch|inline] [--plan-depth minimal|full] [--plan-split on|off] [--review-depth quick|standard|deep] [--commit-history tidy|full] [--questions one|probe] [--coding-guide lean|off] [--review-understanding true|false] [--theme NAME] [--clear-theme]",
			"acta config set --repo [--executor E] [--plan-depth D] [--plan-split S] [--review-depth R] [--commit-history H] [--coding-guide G] [--review-understanding B] [--unset KEY]...",
		},
		flags: []string{
			"--json  print JSON (config show)",
			"--language L  chat language, as a full name (Korean)",
			"--style S  adhd or plain",
			"--tone T  how you want to be spoken to, in your own words",
			"--clear-tone  remove the tone",
			"--executor E  which executor runs the plan: subagent, dispatch or inline",
			"--plan-depth D  how much a plan spells out: minimal or full",
			"--review-depth R  how much review a plan gets: quick, standard or deep",
			"--commit-history H  how a landed branch reads: tidy or full",
			"--questions Q  how the agent asks: one question at a time, or probe",
			"--coding-guide G  the coding guide: lean or off",
			"--review-understanding B  true or false: after a landing, explain the change",
			"--theme NAME  the TUI color theme",
			"--clear-theme  go back to the default theme",
			"--repo  save to .acta.yaml in this repo instead of your own config",
			"--unset KEY  with --repo: remove this key (review_depth, review_understanding, ...) from .acta.yaml so your own value applies (repeatable)",
		},
		example: "acta config set --executor subagent",
	},
	{
		name:    "doctor",
		group:   "Setup",
		summary: "Check the install and the repo, and say what to fix.",
		usage:   []string{"acta doctor [--fix] [--known <file>]"},
		flags: []string{
			"--fix  fix repo problems (.acta folder, .gitignore)",
			"--known FILE  file listing plugins that overlap acta",
		},
		example: "acta doctor --fix",
	},
	{
		name:    "update",
		group:   "Setup",
		summary: "Replace this binary with the latest release, then refresh the plugin copies.",
		usage:   []string{"acta update [--check]"},
		flags:   []string{"--check  show the current and latest version, change nothing"},
		example: "acta update --check",
	},
	{
		name:    "migrate-root",
		group:   "Setup",
		summary: "Move the old .pm folder of this repo to .acta in one commit.",
		usage:   []string{"acta migrate-root"},
		example: "acta migrate-root",
	},
	{
		name:    "tidy",
		group:   "Setup",
		summary: "Fold a branch into one commit per code task, leaving every ref as it was when a step fails.",
		usage:   []string{"acta tidy <base> <branch> [--onto <ref>]"},
		flags:   []string{"--onto REF  ref whose chore commits fold forward into the first new commit"},
		example: "acta tidy main acta-help-man",
	},
	{
		name:    "ready",
		group:   "Setup",
		summary: "Say whether a plan may start: every plan in its after list has landed on the parent. Exit 1 waits, exit 2 means it can never be ready (bad after list or parent), exit 3 is a usage or load error.",
		usage:   []string{"acta ready <plan> --parent <ref> [--head <rev>]"},
		flags: []string{
			"--parent REF  the branch the build will land on",
			"--head REV    a reused worktree head that must already hold each landed after plan",
		},
		example: "acta ready PLN-0160 --parent main",
	},
	{
		name:    "spec-dates",
		group:   "Setup",
		summary: "After a plan lands, write and commit the dates of its spec.",
		usage:   []string{"acta spec-dates <plan>"},
		example: "acta spec-dates PLN-0160",
	},

	// Agent plumbing
	{
		name:    "hook",
		group:   "Agent plumbing",
		summary: "Print the text a plugin hook injects. The plugin hook scripts call this.",
		usage: []string{
			"acta hook session-start [--known <file>]",
			"acta hook prompt",
			"acta hook pre-tool",
			"acta hook post-tool",
		},
		flags: []string{"--known FILE  file listing plugins that overlap acta (session-start only)"},
	},
	{
		name:    "dispatch",
		group:   "Agent plumbing",
		summary: "Hand a plan round to an agent in another pane, and close that pane at land time.",
		usage: []string{
			"acta dispatch init --pane <id> --plan <path> [--round <slug>]",
			"acta dispatch send --plan <path> --rules <path> [--round <slug>] [--note-file <path>|-]",
			"acta dispatch close",
		},
		flags: []string{
			"--pane ID  orchestrator pane id, for example wM:pH (init)",
			"--plan PATH  plan path relative to the repo root",
			"--round SLUG  short slug for this round (init default: the branch name; send: empty means first dispatch, polish, or the name of a fix round)",
			"--rules PATH  absolute path of references/house-rules.md (send)",
			"--note-file PATH  file with the note for the recipient, or - for stdin (send)",
		},
	},
	{
		name:    "reply-back",
		group:   "Agent plumbing",
		summary: "Tell the orchestrator pane that a dispatched round is done, or why it is stuck.",
		usage:   []string{"acta reply-back [--blocked <reason>]"},
		flags:   []string{"--blocked REASON  say the round is blocked and why"},
	},
	{
		name:    "run-one",
		group:   "Agent plumbing",
		summary: "Run one command while holding the machine-wide lock, so two full test runs never fight over the CPU.",
		usage:   []string{"acta run-one -- <command> [args...]"},
		flags:   []string{"--  everything after this is the command and its arguments; -h and --help after it go to the command"},
	},
	{
		name:    "eval-omp",
		group:   "Agent plumbing",
		summary: "Run the plugin eval cases through omp. Each case costs model quota.",
		usage:   []string{"acta eval-omp [--eval-dir <dir>] [--case <glob>] [plugin-dir]"},
		flags: []string{
			"--case GLOB  run only the cases whose name matches this glob",
			"--eval-dir DIR  eval case folder below the plugin dir (default: evals)",
		},
	},
}

// findHelpCmd looks a command up by name.
func findHelpCmd(name string) (helpCmd, bool) {
	for _, c := range helpTable {
		if c.name == name {
			return c, true
		}
	}
	return helpCmd{}, false
}

// usageLine is the short usage a bad call prints on stderr. It reads the
// help table, so it can never differ from acta help. With no sub it joins
// every usage line of the command. With a sub it gives only the line that
// starts with "acta <name> <sub>". A name or sub the table lacks is a
// mistake in the code, so it panics and the tests show it at once.
func usageLine(name string, sub ...string) string {
	c, ok := findHelpCmd(name)
	if !ok {
		panic("usageLine: no help entry for " + name)
	}
	if len(sub) == 0 {
		return "usage: " + strings.Join(c.usage, " | ")
	}
	prefix := "acta " + name + " " + sub[0]
	for _, u := range c.usage {
		if strings.HasPrefix(u, prefix) {
			return "usage: " + u
		}
	}
	panic("usageLine: no usage line of " + name + " starts with " + prefix)
}

// renderHelp is the text of acta help: every group and every command.
func renderHelp() string {
	var b strings.Builder
	b.WriteString("acta shows and changes the planning files of the repo it runs in.\n\n")
	b.WriteString("Usage:\n")
	b.WriteString("  acta                      open the board\n")
	b.WriteString("  acta <command> [args]     run one command\n")
	b.WriteString("  acta help [command]       show this help, or the help of one command\n")
	b.WriteString("  acta <command> --help     same as acta help <command>\n\n")
	b.WriteString("Global flags:\n")
	b.WriteString("  -root DIR   planning root folder (default .acta; also ACTA_ROOT, PM_ROOT,\n")
	b.WriteString("              or root in .acta.yaml or .pm.yaml); most commands take it too\n")
	b.WriteString("  -h, --help  show this help\n")
	b.WriteString("  --version   print the version\n")
	for _, g := range helpGroups {
		b.WriteString("\n" + g + ":\n")
		for _, c := range helpTable {
			if c.group == g {
				fmt.Fprintf(&b, "  %-13s %s\n", c.name, c.summary)
			}
		}
	}
	b.WriteString("\nRun acta help <command> for the usage and flags of one command.\n")
	return b.String()
}

// renderCmdHelp is the text of acta help <command>.
func renderCmdHelp(c helpCmd) string {
	var b strings.Builder
	fmt.Fprintf(&b, "acta %s: %s\n\nUsage:\n", c.name, c.summary)
	for _, u := range c.usage {
		b.WriteString("  " + u + "\n")
	}
	if len(c.flags) > 0 {
		b.WriteString("\nFlags:\n")
		for _, f := range c.flags {
			b.WriteString("  " + f + "\n")
		}
	}
	if c.example != "" {
		b.WriteString("\nExample:\n  " + c.example + "\n")
	}
	return b.String()
}

// versionLine is what acta --version prints. A binary that goreleaser did
// not build says so, so a go install copy is easy to tell from the release.
func versionLine() string {
	// Name the build, so a user can tell the small one apart.
	build := "full"
	if Variant == "cli" {
		build = "cli, no TUI"
	}
	if isRelease() {
		return "acta " + plugin.Version() + " (" + build + ")"
	}
	return "acta " + plugin.Version() + " (" + build + ", source build)"
}

// isHelpFlag says whether one argument asks for help.
func isHelpFlag(a string) bool {
	return a == "-h" || a == "-help" || a == "--help"
}

// helpAsked says whether a help flag shows up before a bare --. What comes
// after -- belongs to another command, as in run-one.
func helpAsked(args []string) bool {
	for _, a := range args {
		if a == "--" {
			return false
		}
		if isHelpFlag(a) {
			return true
		}
	}
	return false
}

func unknownCommand(name string, stderr io.Writer) int {
	fmt.Fprintf(stderr, "unknown command %q; run acta help\n", name)
	return exitBadInput
}

// helpRoute answers the help and version asks before Run dispatches. The
// bool says whether it answered; when false, Run goes on as before.
func helpRoute(args []string, stdout, stderr io.Writer) (int, bool) {
	if len(args) == 0 {
		return 0, false
	}
	if args[0] == "help" {
		switch {
		case len(args) == 1 || (len(args) == 2 && isHelpFlag(args[1])):
			fmt.Fprint(stdout, renderHelp())
		case len(args) == 2:
			c, ok := findHelpCmd(args[1])
			if !ok {
				return unknownCommand(args[1], stderr), true
			}
			fmt.Fprint(stdout, renderCmdHelp(c))
		default:
			fmt.Fprintln(stderr, "usage: acta help [command]")
			return exitBadInput, true
		}
		return exitOK, true
	}
	if strings.HasPrefix(args[0], "-") {
		for _, a := range args {
			switch {
			case a == "--":
				return 0, false
			case isHelpFlag(a):
				fmt.Fprint(stdout, renderHelp())
				return exitOK, true
			case a == "--version":
				fmt.Fprintln(stdout, versionLine())
				return exitOK, true
			}
		}
		return 0, false
	}
	if c, ok := findHelpCmd(args[0]); ok && helpAsked(args[1:]) {
		fmt.Fprint(stdout, renderCmdHelp(c))
		return exitOK, true
	}
	return 0, false
}
