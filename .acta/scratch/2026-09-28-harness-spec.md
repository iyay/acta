---
id: SCRATCH-6
hash: gc5a
title: harness-spec
status: brainstorming
created: "2026-09-28"
---
# Spec 3: harness so agents do not skip the rules

Third of three specs, after spec 2 (skill rules).

User, 2026-09-27: skills and the main agent must be hardened so they do not misjudge or skip the scratch and brainstorm rules. Proposed:
1. Hook state: acta records each brainstorm start per session (session id), and the prompt-submit hook reminds "this session already brainstormed SCRATCH-n" when true.
2. plugincheck text guards per skill.
3. A behaviour eval suite (claude plugin eval) with scenarios: side idea mid-build goes to scratch, not memory; "catet aja" goes to scratch; second brainstorm asks the three choices; a one-file obvious fix gets no brainstorm; brainstorm start files a scratch item first; answers get appended.
4. Run the evals greenfield: a clean config and a scratch repo, because a global CLAUDE.md or AGENTS.md hides plugin gaps.

Q1 scope, user 2026-09-29: (a) all three - hook state, fill gaps in plugincheck guards, greenfield eval suite.

Q2 session tracking, user 2026-09-29: (a) PostToolUse hook on Bash watches 'acta set scratch/... status brainstorming' and records session_id -> SCRATCH-n in a state file. No skill change needed.

Q3 state location, user 2026-09-29: (a) .acta/state/sessions.json in the repo, gitignored.

Q4 eval timing, user 2026-09-29: conditional land gate. acta:land runs the full eval suite (about 6 scenarios, sonnet) only when the diff touches plugin/skills/ or plugin/hooks/. Red eval stops the land like a red test. make eval for manual runs. User worried manual runs get forgotten, and about cost.

Q5 second brainstorm, user 2026-09-29: (b) both. UserPromptSubmit reminder names the SCRATCH-n already brainstormed this session; PreToolUse blocks 'acta set scratch/... status brainstorming' for a different item in the same session, with a message to offer the three choices. Gap: a brainstorm that never runs acta set is caught only by evals.

Approach, user 2026-09-29: A - claude plugin eval built-in suite. omp is out of scope: omp does not run Claude Code shell hooks, so the brainstorm block and state do not work there, and evals cover Claude Code only. omp eval + block filed as a new scratch item.

Design section 1 approved 2026-09-29 (hook state): new subcommands acta hook post-tool (records session_id -> SCRATCH-n in .acta/state/sessions.json when a Bash command is `acta set scratch/<stem> status brainstorming`) and acta hook pre-tool (exit 2 with a three-choices message when the session already has a different SCRATCH; the same item passes). acta hook prompt adds a reminder line when the session has brainstormed. hooks.json gains PreToolUse and PostToolUse on Bash through never-failing shell wrappers. .acta/state/ gitignored through EnsureGitignore. Any error: silent, exit 0.

Design section 2 approved 2026-09-29 (guards): existing brainstorm/scratch guards stay. New guards: hooks.json has PreToolUse and PostToolUse on Bash, wrappers exist, are executable, end in exit 0; land skill names the eval gate, plugin/skills/, plugin/hooks/ and make eval; each eval scenario maps to one skill phrase that plugincheck requires.

Design section 3 approved 2026-09-29 (eval suite): plugin/evals/ with six cases (side idea to scratch, "catet aja" to scratch, second brainstorm offers three choices, obvious one-file fix skips Architectural, brainstorm start files scratch then status brainstorming, answers appended). Free file/command graders first; haiku LLM grader only for cases 3 and 4. scripts/eval wraps claude plugin eval with --model sonnet --ablation none --max-cost-usd 2 --no-publish --allow-tools Bash (no Makefile in repo, so make eval became scripts/eval). Plan task 1 is a spike proving the eval sandbox ignores the global CLAUDE.md. Cost cap default $2, user gave no number.

Cost ruling 2026-09-29: user is on a subscription, so no --max-cost-usd; each case caps max_turns and timeout_seconds.

Land gate ruling 2026-09-29: gate also needs scripts/eval in the repo, since the land skill is shared by every project. Hooks ship to every user; evals stay in the acta repo.
