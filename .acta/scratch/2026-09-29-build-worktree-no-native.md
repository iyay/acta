---
id: SCR-0023
hash: xjl0g92
title: 'Build: always git worktree add, drop native EnterWorktree'
status: raw
created: "2026-09-29"
---
User: "catet" (after the EnterWorktree discussion, 2026-09-29)

Agent lines:
- acta:build (plugin/skills/build/SKILL.md ~line 69) says prefer a native worktree tool if the harness has one.
- Claude Code EnterWorktree makes the worktree inside the repo (.claude/worktrees/), which breaks the rule "outside the repo, ../<repo>-<slug>" (recursive scanners see both copies).
- Its base ref defaults to origin/<default-branch> (worktree.baseRef: fresh). The user pushes by hand, so a spec+plan just committed on local main is often not on origin: the worktree starts without the plan. This breaks the new "spec and plan on main" flow (SPEC-21 / PLAN-30).
- Proposal: build always uses git worktree add "../$REPO-$SLUG" -b "$SLUG" "$PARENT"; drop the "use native tool" step. Dispatch already does this.
