---
id: SPEC-19
hash: eqjj
---
# Private Lock Folder and Locked Writes

Status: approved by the user on 2026-09-29 (Bounded, security). Debt group A1 from the triage of 2026-09-29: DEBT-1.1, DEBT-8.1, DEBT-5.3, DEBT-6.2. Group A2 (root link containment) gets its own spec.

## Why

- DEBT-1.1 (security): the lock folder is `/tmp/pmb-<uid>`. On a shared host another user can create it first. `safeDir` (`internal/write/tick.go`) then refuses it on every run, and because of the sticky bit only root can remove it, so every `acta tick` fails.
- DEBT-8.1: `appendDebt` (`internal/write/ops.go`) writes the debt file in place with no lock, while `TickLine` locks and renames. A review adding NOTEs while land ticks the same file can lose one side.
- DEBT-5.3: `RecordAgent` (`internal/write/agents.go`) reads, changes and writes `.agents.json` with no lock. Two ticks on different plans can drop one agent tag.
- DEBT-6.2: a `--start` with no agent name writes an empty agent over the one an earlier tick recorded.

## 1. Lock folder in the user's cache folder

- The lock folder becomes `<os.UserCacheDir()>/acta/locks` (macOS `~/Library/Caches/acta/locks`, Linux `$XDG_CACHE_HOME/acta/locks` or `~/.cache/acta/locks`). It sits inside the user's own home, so no other user can create it first. Processes of one user still share it whatever their `TMPDIR`.
- The parent folders are created with mode `0700`. `safeDir` keeps all three checks on the lock folder: a real folder (not a link), owned by this user, closed to everyone else.
- Lock files are named `acta-<hash>.lock`; the `pmb-` names go.
- When `os.UserCacheDir()` fails (for example no `HOME`), the command fails with `cannot find a cache folder for the lock: <err>`. There is no fallback to `/tmp`.
- The `safeDir` error names the check that failed: `lock folder <dir> is not a folder`, `... is not owned by you`, or `... is open to other users`.
- Tests keep a way to point the lock folder at a temp folder, so no test touches the real home.

## 2. Locked writes

- `appendDebt` takes `lock(path)` for the debt file and writes through a temp file and a rename, the same way `TickLine` does.
- `RecordAgent` takes `lock(path)` for `.agents.json` around the whole read, change and write.

## 3. A start with no name keeps the agent

When `RecordAgent` gets an empty agent, it keeps the agent already recorded for that task and only updates `Started` and `At`. A named tick still replaces the name.

## Testing

Each behaviour red first.

- `safeDir` refuses a folder with mode `0755`, a link to a folder, and a plain file, each with its own message. The owner check is read in code only, since a test cannot create a folder as another user.
- An empty `HOME` (and no cache variable) makes a tick fail with the cache-folder message and write nothing.
- A lock file is created under the cache folder and named `acta-<hash>.lock`.
- `appendDebt` and `TickLine` run at the same time on one debt file: both changes are in the file afterwards.
- Two `RecordAgent` calls at the same time for different tasks: both records are in `.agents.json` afterwards.
- A named tick, then a `--start` with no name: the agent name is still there and `Started` is true.
- `gofmt -l .` empty, `go vet ./...` ok, `go test ./...` green.

## Out of scope

- Root link containment (DEBT-13.2, 13.5, 13.6), which is group A2.
- Cleaning up old `/tmp/pmb-<uid>` folders; they are left alone.
