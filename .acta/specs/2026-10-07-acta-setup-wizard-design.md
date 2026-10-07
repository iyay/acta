---
parent: scratch/2026-10-07-standalone-setup-installer
id: SPC-0100
created: "2026-10-07 09:33:10"
hash: grj78qu
---
Status: Architectural, approved by the user in chat on 2026-10-07.
Why: setup today lives only in the acta:setup skill, so it needs an agent session. A user who installs the acta binary should be able to set everything up from a plain terminal, with a wizard that asks one question at a time.

# acta setup: an interactive setup wizard in the Go CLI

## Scope

In: a new `acta setup` command, reached after `go install ./cmd/acta` (or any later install channel). The acta:setup skill becomes a thin pointer to it.

Out: the curl install script and the release pipeline it needs. They are split to SCR-0052, because the repo has no remote, releases or goreleaser yet. npm/bun and brew come later.

## Flow

1. No TTY on stdin or stdout: print a message that points to `acta config set` and exit non-zero. Nothing is written.
2. Run the same checks as `acta doctor` and show the result. A failed check is shown, not fatal.
3. Ask, with a `charmbracelet/huh` form, one question at a time: chat language, style, tone, build executor, subagent models. Each default is the value in the current user config (`internal/config/user.go`), or the built-in default when there is none.
4. Look for `claude` and `omp` on PATH. For each one found, ask "install the acta plugin into <harness>? [Y/n]". On yes, run its install command. When that command fails, print it so the user can run it by hand, and carry on.
5. When the working directory is inside a git repo, write the acta block. The block is mandatory: no yes is asked and there is no skip. Show the block and the file it goes to. File choice: both CLAUDE.md and AGENTS.md exist, write it to both; one exists, use it; neither exists, create a CLAUDE.md that holds only the block. Never create an AGENTS.md. Write only between `<!-- acta:begin -->` and `<!-- acta:end -->`; a re-run replaces the text inside them and leaves the rest of the file alone. The wizard never runs `/init`, because it runs outside an agent.
6. Print a summary: what was written, which install commands ran or failed, and the next step.

## Code layout

- `internal/setup/plan.go`: a pure function `Plan(answers, env) []Action`. `env` holds what was found (TTY, harnesses on PATH, git repo, which of CLAUDE.md / AGENTS.md exist, current config). An `Action` is one of: write user config, install plugin into a harness, write the acta block to a file. All decisions live here.
- `internal/setup/block.go`: the acta block text as one constant, and the function that writes it between the markers.
- `internal/setup/form.go`: the `huh` layer. It only collects answers and confirms; it holds no rules.
- A runner that carries out actions. Harness install commands go through an interface, so tests use a fake runner and never call the real `claude` or `omp`.
- `internal/cli/cli.go`: `case "setup"`.
- New dependency: `github.com/charmbracelet/huh`.

## Skill change

`plugin/skills/setup/SKILL.md` shrinks to: ask the user to run `! acta setup` (or run it when the harness gives a TTY), and fall back to `acta config set` when it cannot. The block text leaves the skill; `internal/setup/block.go` is the single source. `internal/plugincheck/skill_setup_test.go` checks the block in the skill today, so it changes to match. The skill's fallback path follows the same rule: the acta block is mandatory, written with the same file choice and no yes asked (user ruling 2026-10-07).

## Testing

- `Plan`, table-driven: no TTY, harness found or not, inside a repo or not, CLAUDE.md / AGENTS.md combinations, old config values used as defaults.
- Block writing in a temp dir: first write, re-run replaces only the text between the markers, text outside them stays byte for byte.
- Install actions with a fake runner: success, failure that prints the command.
- `acta setup` with no TTY exits non-zero and writes nothing.

## Close

The plan ends with the version bump task: 0.1.x patch +1 in `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json` and `plugin/package.json`.
