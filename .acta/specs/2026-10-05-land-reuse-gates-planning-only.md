---
id: SPC-0081
created: "2026-10-05 15:52:56"
hash: m3t6jn9
---
# Land reuses the gates when only planning files differ

Status: Bounded, approved by the user in chat on 2026-10-05.

Why: `acta:land` step 6 re-runs every gate when the merge tree differs from the branch tree. On a busy main the difference is often only new specs and plans committed after the branch was cut, so the full suite runs again for nothing.

Design:
- `plugin/skills/land/SKILL.md` step 6: after the merge, list what differs with `git diff --name-only <branch> HEAD`. When every path is under the acta planning root (`.acta/` by default), reuse the branch gates and write "only planning files differ, gates reused" in the report. Any other path, including Markdown outside the planning root (skill text is tested), re-runs the gates. Same tree stays as today.
- The last task adds 1 to the patch version in the three plugin files.

Tests: `internal/plugincheck` passes; the land step names the diff command and both outcomes.
