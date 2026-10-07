---
parent: specs/2026-10-07-acta-setup-wizard-design
id: SPC-0101
created: "2026-10-07 11:26:37"
hash: c875bpf
---
Status: Bounded, approved by the user in chat on 2026-10-07.
Why: the first `acta setup` (SPC-0100) shows every question in one long form, prints ten doctor lines and the whole acta block, and has no title or progress. SPC-0100 asked for one question at a time like an `npx` wizard. The user saw it in a real terminal and asked for it to look much better.

# acta setup: one question per screen, and a nicer look

## Changes

1. **One question per screen.** Each question is its own `huh` group. The screen clears between questions. Questions already answered stay above as one short line each, for example `✓ Language  Indonesian`.
2. **Title and progress.** A title line `acta setup` sits on top, with a step count such as `3/8` beside the current question. Each question carries a one-line plain description so a new user knows what it means (for example, build executor: "Who writes the code when a plan runs.").
3. **Minimal mono look (user pick, 2026-10-07).** Almost everything is grey: question text in the terminal's normal fg, descriptions and answered lines in dim grey. One accent color only, a calm blue, used for the focus marker, the selected option and the `✓` marks. Title in bold with no color. No filled color blocks: yes/no reads `● Yes  ○ No`, select options read `› adhd` for the focused one. Errors use the terminal's red. The look does not follow the acta theme. Every dim grey keeps a contrast ratio of at least 1.6 against both a dark and a light terminal background (wiki `ghostty-minimum-contrast`), so use adaptive colors (one value for dark, one for light).
4. **Doctor in one line when clean.** All checks ok: one line, `✓ install checks ok (10)`. Any check not ok: show only those lines, then carry on.
5. **Harness install as its own step.** After the config questions, one screen lists the harnesses found, each with a yes/no. Install output is one line per harness (`✓ claude` or `✗ omp: <command to run by hand>`).
6. **Block shown short.** Print `acta block → CLAUDE.md` (one line per file) instead of the whole block text.
7. **Closing summary.** A short box at the end: config file path, what was installed, which files got the block, then the next step.

Nothing changes in `Plan`, `Apply`, `WriteBlock`, the no-TTY guard or the config values written. Only the look of `form.go` and the printing in `internal/cli/setup_cmd.go` change.

## Testing

- Unit tests on the printing helpers: doctor summary (all ok, one failing), block lines, install lines, summary box, using the setup theme.
- A test that the form builds one group per question in the fixed order, with the step count right.
- A contrast test: every dim color the form uses has ratio 1.6 or more against a dark and a light background.
- After merge, the user runs `acta setup` in a real terminal to judge the look.
