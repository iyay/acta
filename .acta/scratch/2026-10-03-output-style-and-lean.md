---
id: SCR-0036
hash: peat1k3
title: Merged output style, builtin coding guide lean, session-start diet
status: brainstorming
created: "2026-10-03 14:39:48"
schema: "1"
started: "2026-10-03 14:41:39"
finished: "2026-10-03 15:03:55"
---
# Merged output style, builtin coding guide lean, session-start diet

## Words

### 2026-10-03

Step 2 of the 2026-10-03 roadmap (step 1 landed as 1428e20). User rulings so far:

1. One builtin output style, forced: plugin/output-styles/ with force-for-plugin: true and keep-coding-instructions: true. It merges Claude Code's Concise style, caveman (JuliusBrussee/caveman, MIT except its engine folders) and the acta adhd rules into one text: overlaps written once, conflicts settled. Proposed settlements: a one-line status only during multi-turn work; errors quote the deciding lines, then cause and fix, full log on request; an explicit request for detail beats every length cap; the tone from config beats compression; the chat language comes from config. A core for every user plus an ADHD block that applies when the session note says adhd, so `style: plain` still works. One level, no lite/full/ultra. The hook stops printing adhdRules under Claude Code; omp keeps it. Target about 300 words.
2. Builtin coding guide `lean`, from ponytail (MIT), rewritten: a skill of about 400 words plus a five-line summary in session start, on by default, with a config switch. Drop the `ponytail:` marker comments, the levels, the test rule (acta:tdd owns tests) and the output rules (the style owns them). Rename the "ponytail-lazy" Global Constraints line. acta doctor warns when the ponytail or caveman plugins are also on (hook.EnabledPlugins and Conflicts exist).
3. Session start text diet: now 3591 bytes (about 900 tokens); target about 600 tokens.
4. Every adapted text is rewritten, never copied 1:1, and costs fewer tokens without losing quality (evals stay green, byte caps go down). Credit MIT sources in plugin/NOTICE. Claude Code's Concise text is not MIT: ideas only.
5. Open: should every landed plan bump the patch (iOS build number style), which also closes the go.mod and scripts/ gap in the CLAUDE.md version rule?

After it lands, the user can turn off the caveman and ponytail plugins.

## Context

Facts found 2026-10-03 (brainstorm session):

- Claude Code docs confirm `force-for-plugin: true` for plugin output styles in `output-styles/`: the style applies whenever the plugin is on and overrides the user's `outputStyle`; if several plugins force one, the first loaded wins. Style files are read at start, so edits need a restart.
- An output style is a static file: it cannot read the acta config. Language, tone and the adhd/plain switch must still come from the hook text.
- Session start text for the default config is 3591 bytes: header plus skill index 1159, core rules 1685 (rule 7 superpowers map 487, rule 8 second brainstorm 659), voice 297 (destructive-warning line 97), adhd block 451.
- The same `acta hook session-start` serves Claude Code (plugin/hooks/session-start) and omp (plugin/omp/index.ts). There is no harness flag today. plugin/hooks/default-rules.md is a golden copy of SessionStart for the default config (TestDefaultRulesFile).
- Implementers get ponytail today through two lines: the "ponytail-lazy" Global Constraints line in plugin/skills/slice/SKILL.md (lines 29 and 98) and the PONYTAIL line in plugin/references/house-rules.md (line 17).
- Clash detection exists: hook.EnabledPlugins + hook.Conflicts + plugin/hooks/workflow-plugins.txt (superpowers, gstack, mattpocock) feed both the session start note and doctor checkConflicts.
- Config: RepoKeys = repo_language, build_executor, plan_depth; personal keys = chat_language, style, tone, theme, subagent_models.
- Sources: ponytail skill 1079 words (MIT, Copyright (c) 2026 DietrichGebert); caveman skill 942 words (MIT outside its engine folders, Copyright (c) 2026 Julius Brussee).

## Log

### 2026-10-03

Q1 (how omp gets the style): user picked (a). The omp extension reads output-styles/acta.md and adds its body to the session start text. The hook stops printing the adhd block in both harnesses, so the style text has one source and no harness flag is needed. omp gets the core rules too, not only the adhd block. This replaces the older note "omp keeps adhdRules".

### 2026-10-03

Q2 (proof with new evals): user picked (c), one new eval case for the style only (style-short-answer: a simple question; graders check the first line answers it, with no opener, recap or closing pleasantry). No lean case, since lean is hard to grade. The first plan task measures whether the eval sandbox loads plugin output styles; if it does not, the style case is dropped and the fact goes into plugin/evals/FACTS.md. Lean is proven by byte caps and the 6 existing evals staying green.

### 2026-10-03

Q3 (version rule): user picked (a). Every landed plan adds 1 to the patch, with no exceptions, like an iOS build number. The CLAUDE.md "## Version" section gets this one-sentence rule in place of the folder list, which also closes the scripts/ and go.mod gap. This plan bumps 0.1.1 to 0.1.2.

### 2026-10-03

Approach: user answered "a" to options numbered 1-3; read as option 1 (stated back to the user). The style file covers chat only (core block plus ADHD block). The five-line lean summary is printed by the session start hook only when coding_guide is lean, so turning it off costs 0 bytes. Implementers get lean through the plan Global Constraints line and house-rules.md, both renamed from ponytail to acta:lean.

### 2026-10-03

Design section 1 approved (output style file): plugin/output-styles/acta.md with name acta, keep-coding-instructions: true, force-for-plugin: true; body about 283 words in two blocks, "Every reply" (core) and "ADHD reader (only when the acta session note says style: adhd)". Settled conflicts: the status line lives only in the ADHD block and only during multi-turn work; errors quote the deciding line, then cause and fix, full log on request; a request for detail beats every length rule; tone beats cutting words; chat language comes from config. The hook prints "- Style: adhd" or "- Style: plain" in Voice and drops its destructive-warning Voice line. The omp extension reads the file, drops the frontmatter and appends the body to the session start text, skipping it quietly when unreadable. Dropped from caveman: levels, wenyan, stats, the drop-articles rule. Concise contributes ideas only, no copied sentences.

### 2026-10-03

Design section 2 approved (session start diet): default config goes from 3591 to about 2508 bytes (about 627 tokens). Header plus skill index becomes one line with skill names only (the Skill listing already shows each description); rules 1-6 stay; rule 7 keeps the same superpowers map in fewer words; rule 8 is not touched (evals lock its wording); Voice drops the destructive-warning line and adds "- Style: adhd."; the adhd block moves to the style file; a five-line lean summary is printed only when coding_guide is lean. New config key coding_guide (lean or off, default lean) joins RepoKeys like plan_depth; setup does not ask; acta config set coding_guide off turns it off. caveman and ponytail join workflow-plugins.txt, and the session note and doctor warning say "a plugin that overlaps acta" instead of "workflow plugin". default-rules.md is regenerated by the golden test. Risk: names-only index may shift routing; the existing routing evals gate it, and a failing skill gets its "when" hint back.

### 2026-10-03

Design section 3 approved (lean skill and ponytail lines): new plugin/skills/lean/SKILL.md, at most 400 words, rewritten from ponytail (MIT): stance (lazy means efficient, not careless), understand first, the ladder (need it at all, already in the repo, stdlib, native feature, installed dependency, one line, minimum code), rules (no one-user abstraction, no config for a fixed value, no scaffolding for later, deletion over addition, fewest files, the edge-case-correct option when sizes tie), root-cause bug fixes at the point every caller passes, never cut trust-boundary checks, data-loss error handling, security, accessibility or anything the user asked for, and skipped items named in chat or the plan, never as marker comments. Dropped: levels, ponytail: markers, the test rule, output rules, examples, the hardware paragraph. slice/SKILL.md:98 line becomes "Follow acta:lean: ..." and is written only when acta config show says coding_guide: lean; slice/SKILL.md:29 says "the lean line"; the PONYTAIL line in references/house-rules.md is deleted (the plan line already reaches every dispatch and follows the config); dispatch.md:175 drops "ponytail" from its list; hook.Skills gains lean and loses the When field, unused once the index is names only. NOTICE gains ponytail (DietrichGebert) for skills/lean/ and caveman (Julius Brussee) for output-styles/acta.md; the i-have-adhd entry points at the ADHD block of the style file. Lean as a review axis stays out (SCR-0038).

### 2026-10-03

Design section 4 approved (tests, eval, version): budget caps for output-styles/acta.md, skills/lean/SKILL.md and the lean description at their real sizes; sessionStartCap drops from 3591 to the new size; the budget walk adds output-styles/. A contract test ties the style file (force-for-plugin, keep-coding-instructions, both block headings, the exact phrase "style: adhd") to what the hook prints. hook_test covers no adhd block, the Style line, the lean summary only when lean, the names-only index; default-rules.md is regenerated. Config tests: coding_guide accepts only lean or off, defaults to lean, is a repo key, config show marks (repo). Conflicts flag caveman and ponytail. "ponytail" may appear only in NOTICE and workflow-plugins.txt. omp index.test.ts covers the appended style body without frontmatter and a missing file without a crash, run with (cd plugin && bun test). The first task measures whether the claude plugin eval sandbox loads plugin output styles and records it in plugin/evals/FACTS.md; if it does, the style-short-answer case is added (scaffold writes an English adhd config, 2 graders), otherwise it is dropped. CLAUDE.md "## Version" becomes one sentence: every landed plan bumps the patch; the last task bumps 0.1.1 to 0.1.2 in the three manifests. Land gates: scripts/test --full, go vet ./..., gofmt -l ., bun test, scripts/eval with the branch binary on PATH. After land the user runs go install ./cmd/acta, restarts Claude Code and turns off the caveman and ponytail plugins.

## Open questions
