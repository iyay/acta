---
created: "2026-10-01 05:07:10"
parent: scratch/2026-09-30-id-lookup-command
id: SPC-0054
hash: az0vfpu
---
# Find any item's file fast with acta show --path

Status: design approved by the user in chat on 2026-10-01. Bounded: `acta show <id>` (`cmdShow` in `internal/cli/cli.go`) already finds any kind by id, hash, old id, sub-item or stem.

## Why

Agents need one default way to go from an id to its file. `acta show` already resolves every id form, but it loads the whole board and every worktree first, which takes about 0.6 s. A plain `grep -rlx 'id: <ID>' .acta` takes about 0.02 s. No skill tells agents how to find an item by id.

## Design

1. **New flag.** `acta show <id> --path` prints one line: the item's file path, in the same form as the `path:` line of `acta show`. Plain `acta show` does not change. Its status is derived from the whole board (parent links, `closes:`), so a fast path there could print a wrong status.
2. **Fast lookup first.** For an id in the new format (`SCR-0028`) or a hash (`iivf3gy`), read the frontmatter `id:` and `hash:` of the files in the planning dirs of the main tree. The root comes from `loadConfig`, so `.acta.yaml`, `ACTA_ROOT` and `--root` still apply. No git call and no board load on a hit.
3. **Sub-items.** For `PLN-0040.01` or `DBT-0031.01`, drop the `.NN` part and print the parent's path.
4. **Fallback.** When the fast lookup finds nothing (old ids like `PLAN-30`, stems, items that live only in a worktree), fall back to today's `loadAllTrees` and `b.Get`, then print that item's path.
5. **Unknown.** Still not found: print `unknown id X` and exit with bad input, the same as today.
6. **Skills.** The build and brainstorm skills gain one line: to find an item's file, run `acta show <id|hash> --path`.

## Testing

A failing test first for each input form: new id, hash, sub-item, old id through the fallback, a moved root, and an unknown id. One test proves a hit runs no git: it runs with a PATH that has no git on it. Target: a hit takes under 0.05 s on this repo.

## Out of scope

A new command name, ids or hashes in file names, making plain `acta show` faster, and a `#task-N` anchor in the `--path` output.
