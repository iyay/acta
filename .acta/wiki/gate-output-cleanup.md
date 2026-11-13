---
type: Decision
title: Temporary gate outputs need typed list approval
description: Phone users can approve proven temporary gate outputs; ordinary build worktrees keep their no-delete policy
paths: [plugin/skills/land/tidy.md, plugin/hooks/default-rules.md, internal/hook/hook.go]
---

The user chose this policy while shaping BUG-0080: requiring terminal cleanup leaves a phone user unable to finish it.

Only temporary gate worktrees created by the current session qualify. A before-gate inventory and completed producer evidence establish provenance; saved file identities and content fingerprints bind the displayed list to typed approval. WhatsApp and terminal replies have the same authority. Revalidate the whole batch before deleting any file; missing or stale proof preserves the entire batch.

This classifies proven single-file cleanup as non-destructive without loosening the destructive-command hook. It never permits force, recursive deletion or cleanup in main/build/other-session worktrees. Pending cleanup does not block an otherwise green approved land.

The full procedure is in `plugin/skills/land/tidy.md`. It must be read through a file-reading tool when shell access to plugin files is sandboxed. Compare trace judgments with filesystem audits, rather than treating a claimed check as executed proof.
