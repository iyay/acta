Read this before every full suite run in a land: step 1, a step-6 rerun on the merge result, or a tidy-tip rerun.

# Full suite baseline

The run keeps its complete output and its true exit status. Save the result in the absolute Git common directory, not the worktree folder: `git rev-parse --path-format=absolute --git-common-dir`. Linked worktrees share it. Quote the path, since a checkout may sit under a folder with spaces.

1. Make a unique log next to the record, with the time and process id in its name: `acta-baseline-<utc-time>-<pid>.log`. An old run never overwrites a new log.
2. Run the full command exactly as step 1 names it; `scripts/test --full` already takes the run-one lock. Keep both streams in the log, then read the status at once:

```bash
common="$(git rev-parse --path-format=absolute --git-common-dir)"
log="$common/acta-baseline-$(date -u +%Y%m%dT%H%M%SZ)-$$.log"
<full command> >"$log" 2>&1
rc=$?
cat "$log"
echo "exit status: $rc"
```

Never pipe the suite into `tee`, `head` or a filter: the pipeline hides the status.
3. A non-zero status stops the land, as step 1 says. Keep the failed log and report its absolute path. Leave `acta-baseline.md` and the log it names untouched: the last PASS stays.
4. A zero status: read the log for the summary. Never rerun the suite to recover a count.
5. Write the record to a temp file in the same folder, then rename it over `acta-baseline.md`. A reader sees the old record or the new one, never half a record. A rename that fails leaves the old record in place: report it and do not merge.

The record is one line per field:
- `commit:` the tested `git rev-parse HEAD` in the checkout where the suite ran
- `tree:` the tested `git rev-parse HEAD^{tree}` in that checkout
- `command:` the exact command
- `when:` UTC time from `date -u +%Y-%m-%dT%H:%M:%SZ`
- `runtime:` the versions the suite used, such as `go version` or `node --version`, and a lockfile hash when the repo has one
- `exit:` `0`
- `log:` the absolute path of the kept log
- `summary:` the counts the log shows, such as `ok 689, fail 0`

A missing, unreadable or other-tree record changes nothing: run the suite as step 1 says. A step-6 or tidy-tip rerun that passes replaces the record the same way, with the identity of the checkout it tested. A gate reused under step 6 or the tidy path writes no record and copies nothing: the record still names the commit and tree of the run that passed.
