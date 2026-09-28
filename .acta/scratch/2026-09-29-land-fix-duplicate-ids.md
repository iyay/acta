---
id: SCRATCH-14
hash: v5er
title: acta:land fixes duplicate ids after merge
status: raw
created: "2026-09-29"
---
User words: "catet" (after asking acta:land to run acta id --fix-duplicates on its own).

Branches cut before another branch lands pick the same next short id, since AssignIDs takes max+1 of what the branch sees. The merge has no git conflict because the file names differ, so both ids land. Hit twice on 2026-09-29: DEBT-14 (pane-sort vs tui-top-tabs) and DEBT-15 (reply-back vs the first fix).

Idea: acta:land runs `acta id --fix-duplicates` right after the merge and before the post-merge gates, and names any renumbered id in the landing report.
