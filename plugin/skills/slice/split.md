Read this when `plan_split` is on, or the user ran `/slice split`: it cuts one spec into several plans that different sessions build at once.

# Split a spec into plans

Cut only when the work splits cleanly. A spec that fits one plan stays one plan. More than 4 plans means the spec is too big: say so and send it back to `acta:shape`.

1. Cut the spec into 2 to 4 plans by file area. Two plans that do not wait on each other never touch the same file, so their lands never clash.
2. A plan that needs another plan's work lists it in its frontmatter: `after: [PLN-0160]`. Only plan ids go there, never the plan itself, and never a loop.
3. Every plan sets the same `parent:` (the spec), follows the rest of this skill, and gets its own `acta id`.
4. Right after `acta id`, before anything else, check the after lists: run `acta ready <plan id> --parent <parent>` for each plan. Exit 2 names a bad after list (unknown id, not a plan, the plan itself, a dropped plan, or a cycle with its chain): fix the list with `acta set <plan id> after <ids>` (it commits and folds into the plan's own commit), and run the check again until no plan exits 2.
5. Give each plan a `## Build brief` that names only what differs from the default: the code repo, when it is not the repo that holds the plan; the parent branch, when it is not the default branch, and whether it is checked out anywhere; the land method, when it is not the usual land; and the land order, when a sibling plan lands on the same branch. Leave the section out when all four are default. Rules the skills already hold stay out of it. Commit the brief with `acta commit`, like any plan edit, before the next step.
6. Print one line per plan, in this form, and do not start a build yourself. The plan carries the rest, so the line stays short. The user hands each line to a session of their choice.

```
build PLN-0160            (ready now)
build PLN-0161            (ready now)
build PLN-0162            (wait for PLN-0160 to land)
```

Each plan lands on its own. Land writes the spec's dates once the plans are on the parent branch; a tick in a worktree does not.
