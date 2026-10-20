---
parent: bugs/2026-10-20-omp-startup-skill-name
created: "2026-10-20 07:58:43"
id: SPC-0168
hash: kflycsg
---
# Load acta skills by their OMP names

## Problem

OMP exposes acta skills by their bare frontmatter names, such as `build`.
The shared startup context asks for the Skill tool and does not explain OMP's
Read URI. The user sees `Read skill://acta-build` fail before work starts,
although `build` is available.

## Design

Bounded fix approved in chat on 2026-10-20.

- Add an OMP-only instruction to the session context in `plugin/omp/index.ts`:
  load acta skills with Read using `skill://<bare-name>`; a reference to
  `acta:build` means `skill://build`, never `skill://acta-build`.
- Apply the mapping to every acta workflow skill, including references from
  other skills and prompts. Use each installed skill's frontmatter name;
  do not rename skills, add aliases, or rewrite arbitrary user input.
- Give the instruction with session rules on startup, session switch, and
  after compaction. Include it when the acta binary is missing or fails and
  the extension falls back to default rules. Do not repeat it on every prompt.
- Keep the shared hook text, Claude Code and Codex instructions, skill
  registration, user links, and existing OMP event behavior unchanged.

## Files and verification

Change `plugin/omp/index.ts` and its existing tests in
`plugin/omp/index.test.ts`. Update `.acta/wiki/omp-skill-names.md` with the
runtime instruction and preserve its rule about user-owned links.

Use the existing `createState` and extension-handler test seams. Require
the bare-name Read URI and namespaced-reference mapping on the first context,
after session switch and compaction, and on binary failure; ordinary later
prompts must carry only their usual reminder. Watch the regression fail before
the change and pass afterward with `bun test plugin/omp/index.test.ts`.
Verify actual bare skill registration against the installed skill frontmatter.
Report a live OMP agent run as unverified unless it is actually exercised.

No new CLI flag, dependency, skill alias, or global configuration change.
