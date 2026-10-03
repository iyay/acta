---
parent: scratch/2026-10-03-output-style-and-lean
id: SPC-0068
created: "2026-10-03 15:03:55"
hash: zxxug8f
started: "2026-10-03 15:21:15"
---
# Merged output style, lean coding guide, session start diet

Status: design approved by the user in chat on 2026-10-03, section by section. Architectural: it adds a new plugin part (an output style), a new skill and a new config key, and it changes the session text that both harnesses read.

Step 2 of five (step 1 was SPC-0067). The goal of all five: acta costs fewer tokens than superpowers, mattpocock/skills and gstack for the same guarantees. This step folds three always-on texts into one forced output style, brings a lean coding guide into acta, and cuts the session start text. After it lands, the user can turn off the caveman and ponytail plugins.

## 1. Output style

- New file `plugin/output-styles/acta.md`. Claude Code finds plugin styles in `output-styles/` by itself, so `plugin.json` needs no new field.
- Frontmatter: `name: acta`, a one-line `description`, `keep-coding-instructions: true` (keeps Claude Code's built-in coding rules) and `force-for-plugin: true` (the style is on whenever the plugin is on).
- What forcing does, from the Claude Code docs (checked 2026-10-03): it overrides the user's `outputStyle` setting; when two plugins force a style, the first one loaded wins; style files are read at start, so a change needs a restart.
- The body is about 300 words, in two blocks:
  - `## Every reply` is the core for every user. Answer first. No opener, no plan narration, no recap, no sign-off. Tools run without being announced. Cut filler, hedges and pleasantries. Keep negations, numbers, code, paths and error text exact. A simple question gets one to three sentences. Use structure only when the content has that shape. Errors quote the deciding line, then give cause and fix. Full sentences for security findings, warnings before destructive or irreversible actions, steps whose order could be misread, and requests to explain. A request for detail beats every length rule. Chat language and tone come from the acta session note, and tone beats cutting words. Never name the style unless asked.
  - `## ADHD reader (only when the acta session note says Style: adhd)`: one status line during multi-turn work; numbered steps, one action each; five list items at most; time in concrete units; what works now and how to see it; end with one next action that takes under two minutes.
- Conflicts between the sources are settled once, in the text:
  1. The status line lives only in the ADHD block, and only during multi-turn work.
  2. Errors get the deciding line plus cause and fix. The full log comes only on request.
  3. A request for detail beats every length cap.
  4. Tone beats compression.
  5. The chat language comes from the config, not from the language the user writes in.
- Sources: caveman (MIT, rewritten), the i-have-adhd rules acta already has (MIT, rewritten), and Claude Code's Concise style (not MIT: ideas only, no copied sentence). Left out from caveman: its levels, wenyan modes, stats and the drop-articles rule.
- omp has no output styles. `plugin/omp/index.ts` reads `output-styles/acta.md`, drops the frontmatter and appends the body to the session start text, both the normal text and the fallback text. When it cannot read the file it skips it without a word, because a hook must never stop a session.

## 2. Session start text

The hook text for the default config drops from 3591 to about 2508 bytes, about 627 tokens:

| Part | Now | New | Change |
|---|---|---|---|
| Header and skill index | 1159 | ~201 | One line with the skill names only. Claude Code and omp already show each skill's description in their skill list. |
| Rules 1-6 | 537 | 537 | Same. |
| Rule 7 (superpowers map) | 487 | ~379 | Same map, fewer words. |
| Rule 8 (second brainstorm) | 659 | 659 | Not touched. Evals lock its wording. |
| Voice | 297 | ~214 | The destructive-warning line moves to the style. New line `- Style: adhd.` or `- Style: plain.` |
| adhd block | 451 | 0 | Moves to the style. |
| Lean summary (new) | 0 | ~518 | Only when `coding_guide` is `lean`. |

- `hook.Skills` gains `lean` and loses its `When` field. Nothing reads `When` once the index shows names only.
- The first-run text and the broken-config text print the style line too, so the ADHD block applies there as well.
- The lean summary, five lines:

  ```
  Lean coding guide (full text: acta:lean):
  - Understand the task and the code it touches before choosing.
  - Then take the first rung that works: skip it, reuse code here, stdlib, a native platform feature, an installed dependency, the fewest lines.
  - No abstraction with one user, no config for a fixed value, no scaffolding for later.
  - Fix a bug where every caller passes through, not only the reported path.
  - Never cut checks at trust boundaries, error handling that prevents data loss, security or accessibility.
  ```

- `plugin/hooks/default-rules.md` is made again from the golden test: `go test ./internal/hook -run TestDefaultRulesFile -update`.
- The reminder on every message (`acta config: reply in X, Y style.`) stays as it is.
- Risk: with names only, the agent may pick a different skill. The routing evals guard this. A skill whose eval goes red gets its "when" hint back, and only that skill.

## 3. Config key `coding_guide`

- Values: `lean` (the default) and `off`. Any other value fails validation.
- It joins `RepoKeys`, next to `plan_depth`. It shapes code that lands in the repo, so a repo can pin it in `.acta.yaml`, and a user can set it for all their repos.
- `acta config set --coding-guide off` turns it off, and `acta config set --repo --coding-guide off` pins it for one repo, the same way `--plan-depth` works. `acta config show` lists it, shows `coding_guide: lean (default)` when nothing sets it, and marks a value that comes from the repo with `(repo)`.
- `acta:setup` does not ask about it.
- Off means: no lean summary in the session text, and no lean line in new plans. The skill stays installed.

## 4. Skill `lean`

- New `plugin/skills/lean/SKILL.md`, at most 400 words, rewritten from ponytail (MIT). Ponytail's skill is 1079 words. The new one holds:
  1. Stance: lazy means efficient, not careless. The best code is the code nobody writes.
  2. Understand first: read the task and every file it touches, and trace the real flow. Reading is never cut short.
  3. The ladder, where the first rung that works wins: is it needed at all, is it already in this repo, the standard library, a native platform feature, an installed dependency, one line, and only then the minimum code.
  4. Rules: no abstraction with one user, no config for a fixed value, no scaffolding for later, deletion over addition, fewest files. When two options are the same size, take the one that is right on edge cases.
  5. Bug fixes go to the root: find every caller, and fix it once where all of them pass through.
  6. Never cut: checks at trust boundaries, error handling that prevents data loss, security, accessibility, or anything the user asked for.
  7. Name what you skipped in chat or in the plan, never as a marker comment in a file.
- Left out: levels, `ponytail:` marker comments, the test rule (acta:tdd owns tests), the output rules (the style owns them), the examples and the hardware paragraph.
- The description stays short. SCR-0039 aims at 40 tokens or less for each description.

## 5. Where ponytail is named today

- `plugin/skills/slice/SKILL.md:98`: the Global Constraints line becomes "Follow acta:lean: ..." with the same ladder. Slice writes it only when `acta config show` says `coding_guide: lean`. Slice already runs that command for `plan_depth`.
- `plugin/skills/slice/SKILL.md:29`: "the ponytail-lazy line" becomes "the lean line".
- `plugin/references/house-rules.md:17`: the PONYTAIL line is deleted. The plan's line already reaches every dispatch, and it follows the config.
- `plugin/skills/build/dispatch.md:175`: "ponytail" leaves the list of fixed rules.
- A test fails when "ponytail" shows up anywhere in `plugin/` outside `NOTICE` and `hooks/workflow-plugins.txt`.

## 6. Plugins that overlap acta

- `caveman` and `ponytail` join `plugin/hooks/workflow-plugins.txt`. The comment at the top of the file now says "plugins that overlap acta" and no longer names pm.
- The session start note and the doctor `conflicts` warning say "a plugin that overlaps acta" instead of "another workflow plugin". The fix they give stays the same: set the plugin to false in `.claude/settings.local.json`.

## 7. Credits

`plugin/NOTICE` adds two entries:

- ponytail, https://github.com/DietrichGebert/ponytail, Copyright (c) 2026 DietrichGebert. Adapted into `skills/lean/` and the lean summary the session start hook prints.
- caveman, https://github.com/JuliusBrussee/caveman, Copyright (c) 2026 Julius Brussee (MIT outside its engine folders; acta uses only its skill text). Adapted into `output-styles/acta.md`.

The i-have-adhd entry now points at the ADHD block of `output-styles/acta.md`.

## 8. Version rule

- The `## Version` section of this repo's `CLAUDE.md` becomes one rule: every plan that lands ends with a task that adds 1 to the patch in the three manifests, with no exceptions, like a build number. This also closes the gap where `scripts/` and `go.mod` were not named.
- This plan's last task bumps `0.1.1` to `0.1.2`.

## Testing

1. Budgets in `internal/plugincheck/budget_test.go`: caps for `output-styles/acta.md`, `skills/lean/SKILL.md` and the lean description, each at its real size. `sessionStartCap` drops from 3591 to the new size. The file walk adds `output-styles/`.
2. A contract test: the style file has `force-for-plugin: true`, `keep-coding-instructions: true` and both block headings, and the exact text `Style: adhd` in its ADHD heading is also in the line the hook prints (`- Style: adhd.`).
3. `internal/hook`: no adhd block; the style line for adhd and for plain; the lean summary only when lean; skill names only; the golden `default-rules.md`.
4. `internal/config`: `coding_guide` takes only `lean` and `off`, defaults to `lean`, is a repo key, and `config show` marks a repo value.
5. Conflicts: caveman and ponytail are flagged when they are enabled.
6. omp, run with `(cd plugin && bun test)`: the style body is appended without its frontmatter, and a missing style file causes no crash.
7. Eval. The first task checks whether the `claude plugin eval` sandbox loads plugin output styles, with a throwaway case, never committed, that asks the agent to quote any instruction that mentions "ADHD reader". The answer goes into `plugin/evals/FACTS.md`. If the style loads, add the case `style-short-answer`: the scaffold writes an English, adhd config, the prompt is a simple question, and two graders check that the first line answers it and that there is no opener and no closing pleasantry. If the style does not load, the case is dropped and FACTS.md says why.
8. Land gates: `scripts/test --full`, `go vet ./...`, `gofmt -l .`, `(cd plugin && bun test)`, and `scripts/eval` with the branch binary on PATH (it uses quota).

## After land

The user runs `go install ./cmd/acta`, restarts Claude Code (styles and skills are read at start), and turns off the caveman and ponytail plugins.

## Out of scope

- Lean as a review axis: SCR-0038.
- Shorter skill descriptions: SCR-0039.
- The wording of rule 8.

## Appendix: approved style text

The user approved this text in chat on 2026-10-03 (design section 1). The only change since is `Style: adhd` with a capital S in the ADHD heading, so it matches the line the hook prints. `plugin/output-styles/acta.md` holds exactly this:

````markdown
---
name: acta
description: Short, direct replies. Adds ADHD-friendly rules when the acta session note says style adhd.
keep-coding-instructions: true
force-for-plugin: true
---
# acta style

These rules shape chat replies only. Files, code, comments, commits and PR text follow the repo's own rules.

## Every reply
- Open with the answer or the result. No opener, no restated request, no plan narration, no recap, no sign-off.
- Run tools without announcing them. Write between tool calls only to warn, to ask, or to settle something unclear.
- Cut filler, hedges and pleasantries. Fragments are fine. Pick the short common word. No made-up abbreviations, no arrows for "leads to".
- Keep exact: not, never, no, only, except; numbers and units; code, commands, paths, names and error text.
- Simple question: one to three sentences. Headings, tables and lists only when the content has that shape.
- Errors: quote the line that decides it, then the cause and the fix. The full log only when asked.
- Full, plain sentences for security findings, warnings before anything destructive or irreversible, steps whose order could be misread, and whenever the user asks you to explain.
- A request for detail, a report or a walkthrough gets it in full. That beats every length rule here.
- Chat language and tone come from the acta session note; the tone wins over cutting words. Never name or describe this style unless asked.

## ADHD reader (only when the acta session note says Style: adhd)
- During multi-turn work, one line on where the work stands.
- Steps numbered, one action each, as few as work.
- Lists: five items at most.
- Time estimates in concrete units.
- Say what works now and how to see it.
- End with one next action the reader can do in under two minutes.
````
