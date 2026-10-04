---
name: shape
description: "acta: Use to brainstorm before any creative or change work. Classifies the work as Spike, Bounded or Architectural, refines it with the user, and writes the approved design to .acta/specs/. No code until the user says yes."
---

# Brainstorming Ideas Into Designs

Classify the request first, then work your path: understand the
context, refine the idea, present the design, get approval.

<HARD-GATE>
Do NOT write code, invoke an implementation skill or scaffold
anything until you have told your human partner what you intend and
they have approved it. Every path ends with that approval; only the
ceremony around it scales with the task.
</HARD-GATE>

## Three paths

Say the classification out loud before your first question, so your
partner can override it.

- **Spike** — a feasibility question ("can we...", "is it possible...",
  "quick and dirty is fine") whose output is an answer, not code you
  keep. Present the question and what you'll try in 2-3 sentences, get
  a nod, then find out as cheaply as correctness allows. No spec file.
  Report findings as a recommendation; anything you built stays
  labeled throwaway.
- **Bounded** — a change to a flow that already exists in this repo: a
  new flag, a small endpoint, a one-file fix. Knowing the kind of app
  is not enough. If there is no existing flow to change, the task is
  not bounded. Ask the questions that matter, present ONE recommended
  design in chat, and STOP until your partner says yes. Then write a
  short spec and go to acta:slice.
- **Architectural** — new projects, new subsystems, changes that
  restructure how components fit together or alter interfaces others
  depend on. The full process: questions, approaches, sectioned
  design, written spec, then acta:slice.

When in doubt, take the heavier path. The ratchet is one-way: hidden
complexity found mid-task upgrades the path. Stop and say so. Nothing
downgrades mid-task.

**Sensitive changes take the heavier path.** A change touching a trust boundary, auth, money or a data migration gets a written spec in `.acta/specs/` even when it is Bounded.

A new product with no Context yet? Offer `acta:frame` in one line: it
answers why, for whom and what, then hands the how back here. Never
force it.

**Probe mode.** When the user says "probe" or "grill me", read `probe.md` and use it in place of asking one question at a time. The same fires when the session note says `Questions: probe`. "one at a time" turns probe off.

## Architectural only

Spike and Bounded need no scratch item and are free in number.

**Step 0: scratch item.** Find the scratch item the request names. If
there is none, file the user's words with `acta scratch new`, then run
`acta set scratch/<stem> status brainstorming`. The brainstorm hangs
off that item, so a dead session loses nothing.

Append as you go: every answer and every approved section goes in with
`acta scratch add SCRATCH-n --section log`.

**One Architectural brainstorm per session.** When a second one comes
up, file it as a scratch item and ask the user to pick one of:

- (a) **Background agent** — `claude --bg 'brainstorm SCR-0001'` from the repo, then open the agents view (press ←, or `claude agents`).
- (b) **New herdr tab** — offer this only when the session text has the `herdr:` line. With no such line, do not mention herdr.
- (c) **Manual new session** — put `brainstorm SCR-0001` on the clipboard (`pbcopy`, `wl-copy` or `xclip`, OSC 52 when none of those work) and print the prompt too, so a failed clipboard still has a way in.

Plan, build, review and land may continue in this session. Only the
brainstorm itself is one per session.

**Subagent models.** Only when `acta config show` lists `subagent_models: split` and you run in Claude Code: subagents that write code use `model: "sonnet"`; all other subagents (mapping, explore, planning help, debug investigation, spikes) use `model: "opus"`; reviewers use your own model alias. Otherwise name no model and follow the user's own config.

## Checklist

Classify, announce the path, then do these in order.

**Spike:**
1. Explore project context — enough to frame the probe
2. Present question + probe plan — 2-3 sentences
3. Get approval — a nod is enough
4. Investigate — as cheaply as correctness allows
5. Report findings — a recommendation; label anything built throwaway

**Bounded:**
1. Explore project context — check files, docs, recent commits
2. Ask clarifying questions — one at a time, the ones that matter
3. Present short design in chat — approach, files touched, testing
4. Get approval — STOP and wait for an explicit yes; presenting the design and starting in the same breath skips the gate
5. Write short spec — short spec (about half a page) in `.acta/specs/`, run `acta id`, commit on the checked-out branch
6. User reviews spec — ask them to read the file, then wait. On changes requested, make them in the short spec and ask for the review again; never the Architectural flow or its design doc
7. Transition — invoke acta:slice

**Architectural:**
0. Scratch item — as above, before any question
1. Explore project context — check files, docs, recent commits
2. Ask clarifying questions — one at a time; purpose, constraints, success criteria. Append each answer with `acta scratch add SCRATCH-n --section log`
3. Propose 2-3 approaches — trade-offs and your recommendation
4. Present design — in sections scaled to their complexity, approval after each section, each approved section appended to the Log
5. Write design doc — `.acta/specs/YYYY-MM-DD-<topic>-design.md`, run `acta id` right after so the spec gets its SPC number and hash, commit
6. Spec self-review — the four checks below
7. User reviews written spec — ask, then wait for the yes
8. Transition — invoke acta:slice

**Terminal states are path-bound.** After approval the only skill any
path invokes is acta:slice. A Spike ends at the reported
recommendation instead.

## Spec rules

**Where it goes.** `.acta/specs/` is the default root; `.acta.yaml`,
`ACTA_ROOT` or `acta --root` move it. `acta list --json` shows the
specs and plans that already exist. Write the spec plainly and commit it on the branch that is checked out (usually main). No worktree yet: `acta:build` makes the worktree once the plan is approved.

**Finding a file.** With only an id or hash, `acta show <id|hash> --path` prints the path.

**Parents.** A spec from one scratch item sets `parent: scratch/<stem>`.
A spec from several sets the first as `parent:` and lists the rest in
`closes:`, so no item is left dangling. A plan that works a debt item
sets `parent: debt/<stem>` and lists every DBT id it closes in
`closes:`.

**Shared language.** While refining, keep `CONTEXT.md` (one domain term
per line, in English) and `docs/adr/` (one file per hard-to-explain
decision: the decision, why, the alternatives rejected) up to date.
Propose a new term and wait for a yes. Both change only while the
session is brainstorming.

**Self-review, four checks.** Placeholders left in; sections that
contradict each other; scope too wide for one plan; any requirement a
reader could take two ways. Fix them inline, no second pass.

**User review gate.** "Spec written and committed to `<path>`. Please review it and let me know if you want changes before we write the plan." Then wait. On changes requested, make them and run the four checks again. Approval of the design does not approve the plan; each gets its own yes.

**Next.** Invoke acta:slice to write the implementation plan.
