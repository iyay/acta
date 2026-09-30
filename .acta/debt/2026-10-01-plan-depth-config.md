---
id: DBT-0058
hash: cg0rznz
parent: plans/2026-10-01-plan-depth-config
---
# Review NOTEs: Plan Depth Setting and Repo-Level Config Implementation Plan

- [ ] internal/cli/hook.go loadVoice: an .acta.yaml that does not parse makes config.Load fail first, so RepoErr stays nil and session-start does not name the file; the voice stays global, nothing crashes.
- [ ] internal/config/repo_user.go MergeRepo: a non-string repo value (plan_depth: 3, a list, a bool) is skipped silently, while the same value in the global file fails with ErrBadUser.
- [ ] internal/config/repo_user.go SaveRepoUser: IndexFunc matches any node value, so a value equal to a repo key name (a: plan_depth) makes the key append twice; look at even indexes only.
- [ ] internal/config/repo_user.go SaveRepoUser: a comment on a replaced value line is dropped; a file holding only comments loses them all on save.
- [ ] internal/config/repo_user.go SaveRepoUser: a multi-document .acta.yaml keeps only the first document on save.
- [ ] internal/config/repo_user.go SaveRepoUser: a failed Rename leaves .acta.yaml.tmp behind; the write resets the mode to 0644 and turns a symlinked .acta.yaml into a regular file.
- [ ] internal/cli/config_cmd.go: no test covers --language with --repo; dropping that check keeps the tests green.
- [ ] internal/cli/config_cmd.go: config show --json prints from_repo: null, not [], when the repo set nothing.
- [ ] internal/cli/config_cmd.go: config show now exits 1 when .acta.yaml has a personal key or does not parse; it used to exit 0.
- [ ] internal/cli/config_cmd.go setRepo: Fprintf with no format verbs where the code around uses Fprintln.
- [ ] internal/cli/hook.go loadVoice: os.Getwd error is dropped, same as hookRoot.
