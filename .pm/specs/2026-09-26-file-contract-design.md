---
status: done
id: SPEC-1
hash: s8ma
---
# pm-board file contract

Date: 2026-09-26. Sub-project 0 of pm-board.

## Why this exists

Planning work in a repo is scattered. In be-pmis there are 124 specs and 403 plans under
`docs/superpowers/`, 98 local files under `.scratch/`, and bugs that live only inside
fix-round sections of plans or in agent memory. Nothing gives one view of what is a
story, what is a task, and what is a bug.

pm-board fixes that in three parts:

0. This file contract: where planning files live and how they say what they are.
1. A TUI that reads the contract and shows everything in one place.
2. A lean agent plugin that writes files in this shape.

The TUI and the plugin both depend on this document. Change it here first.

## Goals

- One folder per repo holds every planning file, committed with the code.
- A file says what it is with little or no frontmatter; most files need none.
- Work that tools already write (superpowers specs and plans) fits with no rewrite.
- Status comes from the files themselves where it can, so nobody has to keep it in sync.

Not goals: a server, a database, sync with an outside tracker, multi-user editing, a web UI.

## 1. Folders and names

### Layout

```
.pm/
  specs/   story or feature
  plans/   holds tasks as "### Task N" sections
  bugs/    bug
```

### Finding the root

The root folder is found in this order. The first one that is set wins:

1. The `--root <dir>` flag.
2. The `PM_ROOT` environment variable.
3. `root:` in `.pm.yaml` at the repo root.
4. The default, `.pm/`.

`.pm.yaml` sits at the repo root, not inside the root folder, so it can still be found
after the root folder is renamed.

```yaml
root: .pm
dirs:
  specs: specs
  plans: plans
  bugs: bugs
legacy:
  - docs/superpowers
auto_commit: true
```

Every key is optional. A missing key takes the default shown above. Relative paths are
relative to the repo root.

### IDs

- An item's ID is its path under the root, without `.md`: `specs/2026-09-25-cac-post-top-material-type`.
- File names are `YYYY-MM-DD-<slug>.md`. There are no sequence numbers, because two
  worktrees creating "B-8" at the same time would clash on merge. Date plus slug does not.
- A task's ID is its plan's ID plus `#task-N`: `plans/2026-09-25-cac-post-top-material-type#task-3`.
  `N` is the number in its `### Task N` heading, taken as written (it may be `F1`).
- Outside ticket codes (like `New-261`) go in the `ref` field. They are never IDs.

### Legacy folders

Folders listed in `legacy` are read the same way, but are read-only. Their items show as
"untyped" until they are moved into the root folder. For `docs/superpowers`, the
`specs/` and `plans/` subfolders map to the same types as the root folder.

## 2. Types, frontmatter and status

### Item types

| Type | Comes from | Can be a parent |
|---|---|---|
| story | a file in `specs/` | yes |
| bug | a file in `bugs/` | yes |
| task | a `### Task N` section inside a file in `plans/` | no |

A plan is not an item. It is a container for tasks and never shows in a list.

Every task has a parent, which is a story or a bug.

### Frontmatter

Frontmatter is a YAML block between `---` lines at the very top of the file. It is
optional. Every field in it is optional.

| Field | Used on | Meaning |
|---|---|---|
| `type` | any file | overrides the type the folder gives |
| `status` | story, bug | overrides the derived status |
| `parent` | plan | overrides which story or bug the plan's tasks belong to |
| `ref` | story, bug | an outside ticket code, for display and search |
| `fixed_in` | bug | the commit sha that fixed it |

The title is the file's first `# ` heading. If there is none, the title is the slug.
The date comes from the file name. Unknown fields are kept as they are and ignored.

### Linking a plan to its parent

In this order:

1. `parent:` in the plan's frontmatter, as an ID (`bugs/2026-09-25-atp-print-first-type-row`).
2. The plan's `**Spec:**` line, which superpowers already writes. The path on that line
   is matched to a spec file by file name.
3. Neither found: the plan stands in for a story of its own, with the plan's title.

### Derived status

When `status:` is set, it wins. When it is not set, status is worked out:

**Task**, from the `- [ ]` and `- [x]` checkboxes between its heading and the next
`### ` or `## ` heading:

| Checkboxes | Status |
|---|---|
| none ticked, or no checkboxes | `todo` |
| some ticked | `doing` |
| all ticked | `done` |

**Story**, from its plans and their tasks:

| Condition | Status |
|---|---|
| no plan links to it | `draft` |
| a plan links to it, no task started | `approved` |
| some tasks started or done | `in-progress` |
| every task done | `done` |

**Bug**, the same way from its child tasks: no task started is `open`, some started is
`fixing`, all done is `fixed`. A bug with no plan is `open`.

`dropped` (story) and `wontfix` (bug) are only ever set by hand.

## 3. Bug files and writes

### Bug file

```markdown
---
ref: New-261
---
# ATP print shows one material type's TOP value on every line

## Symptom
PO1 has TOP rows material 30 and service 50; the service line prints 30.

## Root cause
atp.print-rows.js:150 finds the TOP row by po_number only.

## Repro
buildAtpPrintRows(details, tops) with two typed rows on one PO.

## Found in
branch new261-post-top, review round 1
```

The only required parts are the `# ` title and `## Symptom`. `Root cause`, `Repro` and
`Found in` are optional, because a bug is often written down before its cause is known.
Readers show whatever sections exist and never treat a missing one as an error.

A bug is fixed through a normal plan whose frontmatter says `parent: bugs/<file>`.
That plan's tasks become the bug's children.

### What tools may write

- A tool may change frontmatter only. It never changes the body of a file.
- Unknown fields, field order and the body stay exactly as they were.
- A file with no frontmatter gets a new `---` block added at the top.
- Files in legacy folders are never written. A write to one is refused with a message
  saying to move the file into the root folder first.
- A new bug file is created from the template above, then handed to `$EDITOR`. If it
  still matches the template when the editor closes, it is deleted.

### Auto-commit

On by default. Turned off with `auto_commit: false` in `.pm.yaml`.

After a write, the tool commits that one file:

1. It commits only that path (`git commit --only -- <path>`). Other staged changes stay staged.
2. If the file already had uncommitted changes before the write, it does not commit, and
   says so. Committing would sweep those other edits in.
3. If the repo is in the middle of a merge, rebase or cherry-pick, or HEAD is detached,
   it does not commit, and says so.
4. Repo hooks run. There is no `--no-verify`. If a hook fails, the file change stays and
   the error is shown.
5. The message is generated, for example `pm: bugs/2026-09-26-ack-dup-on-resend status fixed`
   or `pm: new bug 2026-09-26-ack-dup-on-resend`.
6. It never pushes.

In every case where it does not commit, the file change is kept.

## Open questions

None for this contract. The TUI (sub-project 1) and the plugin (sub-project 2) get their
own specs.
