# Test audit

Finds tests that prove nothing or prove the same thing twice. It changes no test and deletes no file: removing a test is a code change, so it goes through a plan.

1. Scan, read-only, for two kinds. A text test asserts that a code, script, skill or config file contains an exact line or phrase; see "Behavior, not text" in `acta:tdd`. Near duplicates are tests in one file or one unit whose names or asserts are almost the same. A big repo gets one read-only subagent per test folder.
2. Report at most 10 numbered items per message. Each item has the `file:line`, the kind, the evidence and a proposal: delete, fold into a table, or rewrite as a behavior test. A sound test is no item.
3. Ask which items to approve. The approved items go into one debt item with `acta debt new`, one line per test, only after the user's yes. Nothing is deleted here. The work then goes through `acta:slice` and `acta:build` like any debt.
