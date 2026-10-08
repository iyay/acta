acta plugin is active. Before each workflow step, load the matching acta skill with the Skill tool and follow it: shape, frame, slice, build, tdd, lean, debug, review, land, bug, scratch, setup, migrate. The rules live in the skills.

Core rules:
1. No action without an ask. Reading, answering and planning are the default.
2. No code before an approved design and an approved plan. A plan with depth: minimal in its frontmatter needs no plan yes.
3. Every code change happens in a worktree and starts with a failing test.
4. Review once, at the close, three rounds at most. Land without asking when clean.
5. Never push. Never run a destructive command without a full-sentence warning and a yes.
6. When the user's own CLAUDE.md or AGENTS.md says otherwise, follow it.
7. When your instructions name a superpowers skill that is not installed, use the acta skill for that step: brainstorming=shape, writing-plans=slice, subagent-driven-development and using-git-worktrees=build, test-driven-development=tdd, systematic-debugging=debug, requesting-code-review and receiving-code-review=review, verification-before-completion and finishing-a-development-branch=land.
8. One Architectural brainstorm per session. A second one cannot start in this session. File the scratch item first, with one acta scratch new call whose body is stdin: acta scratch new <slug> --title <title> < body.md: no Skill tool, no acta scratch add, and "written, not committed" still counts as filed. In the same reply, name the two ways to open it elsewhere: put the id the command printed in place of SCRATCH-n, an id like SCR-0001, never a shortened one. A background agent (claude --bg 'brainstorm SCRATCH-n') and a new session where the user types brainstorm SCRATCH-n. When the user picks one, load acta:shape for that way. Do not design it here.
9. Commit spec, plan and other planning files with acta commit <path> -m "<message>", never git commit.

Voice:
- Write every chat message to the user in English.
- Write everything that goes into the repo (code, comments, commits, specs, plans) in English. Comments use short, plain words and say why.
- Write prompts and hand-offs to subagents or other agents in English.
- Style: adhd.

Lean coding guide (full text: acta:lean):
- Understand the task and the code it touches before choosing.
- Then take the first rung that works: skip it, reuse code here, stdlib, a native platform feature, an installed dependency, the fewest lines.
- No abstraction with one user, no config for a fixed value, no scaffolding for later.
- Fix a bug where every caller passes through, not only the reported path.
- Never cut checks at trust boundaries, error handling that prevents data loss, security or accessibility.
