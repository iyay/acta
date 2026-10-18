Read this when `plan_split` is on, or the user ran `/slice split`: it cuts one spec into several plans that different sessions build at once.

# Split a spec into plans

Cut only when the work splits cleanly. A spec that fits one plan stays one plan. More than 4 plans means the spec is too big: say so and send it back to `acta:shape`.

1. Cut the spec into 2 to 4 plans by file area. Two plans that do not wait on each other never touch the same file, so their lands never clash.
2. A plan that needs another plan's work lists it in its frontmatter: `after: [PLN-0160]`. Only plan ids go there, never the plan itself, and never a loop.
3. Every plan sets the same `parent:` (the spec), follows the rest of this skill, and gets its own `acta id`.
4. Right after `acta id`, before anything else, check the after lists: run `acta ready <plan id> --parent <parent>` for each plan. Exit 2 names a bad after list (unknown id, not a plan, the plan itself, a dropped plan, or a cycle with its chain): fix the plan, commit it with `acta commit` (it folds into the plan's own commit), and run the check again until no plan exits 2.
5. Print one line per plan, in this form, and do not start a build yourself, even for a minimal plan. The user hands each line to a session of their choice.

```
build PLN-0160            (ready now)
build PLN-0161            (ready now)
build PLN-0162            (wait for PLN-0160 to land)
```

Each plan lands on its own. Land writes the spec's dates once the plans are on the parent branch; a tick in a worktree does not.
