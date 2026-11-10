Read this only when slice hands over two or more ready plans for a parallel build, or when `acta state` of a plan says `parallel run`. Building one plan never reads it.

# Many plans: one subagent per plan

You are the orchestrator. You write no code; each plan gets its own subagent, and you review and land. The worktree, task loop and close rules are in SKILL.md.

1. Each plan's `## Build brief` wins for its code repo, parent branch, land method and land order. Make each plan's worktree where its brief says, the usual way, and run `acta ready` for each.
2. Start one subagent per plan, all in one message, so they run at once. Name no model; on omp use `agent()` with `agent="task"`.
3. Brief each one from [implementer-prompt.md](implementer-prompt.md), with every task of the plan, in wave order, in place of one task. The rest stays: the `pwd` and `git rev-parse --show-toplevel` check, the path of `house-rules.md`, red test first, one commit per task with its `Task:` trailer, `acta tick` per step and task, the `acta state set` step, and a first report line that names its model. Add: it never starts a subagent of its own, never touches untracked files it did not make, and never `git reset`, rebase, amend or squash. It reports DONE with its commits, or BLOCKED with the task number.
4. When a plan reports DONE, run the close from SKILL.md for that plan: `progress.total` and `progress.done` must match, and you handle its untracked files. Then `acta:review` over that plan's range. A BLOCKER round goes to a new subagent for that plan, briefed with the findings. Three rounds at most, as usual.
5. Lands go one at a time through `acta:land`. A land order in a Build brief wins; otherwise land in the order plans finish. After each land, start every plan whose `after:` list has all landed, as in step 2.
6. A stuck plan (BLOCKED, three review rounds with a BLOCKER left, or a red land) keeps its worktree, and you tell the user. Plans that wait on it stop too. The rest go on.
7. After each step, write `acta state set <plan id> next` for every plan in the run, from that plan's worktree. The text holds the spec id, the line `parallel run`, and each sibling plan with its status: running, waiting on which ids, review round n, landed, or stuck. A new session that runs `/build` on any plan of the run reads that state, opens this file, and picks up every sibling from its status.
