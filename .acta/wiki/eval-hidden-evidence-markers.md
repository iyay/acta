---
type: Gotcha
title: Keep eval execution evidence out of workspace cleanup
description: Root marker files can be deleted as stray outputs; keep execution evidence under .git instead
paths: [plugin/evals/build-close-no-suite/]
---
An agent can run a check successfully and then delete its root marker as a
stray file. A `file_exists` grader then reports a skipped check even though
the native trace shows execution.

Keep execution markers under `.git/` so normal workspace cleanup does not
target them. Anchor marker paths to the generated script, not its caller's
working directory. Keep the same assertions and verify that an unchanged
skill still fails with the revised fixture; moving evidence must not hide
the behavior the eval is meant to catch.
