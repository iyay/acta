#!/bin/sh
set -eu
[ -z "$(ls -A)" ] || exit 1
cat > records.md <<'RECORDS'
# Completed review histories

## mixed
Round 1 input: 1111111..2222222.
Spec report: complete Coverage, one reproduced BLOCKER.
Standards report: complete Coverage, two listed BLOCKERs rejected after
executed repros, plus "2 more BLOCKERs not listed" due to the report cap.
The reproduced Spec BLOCKER was fixed and committed: 2222222..3333333.
No further review has happened. There are no polish items or other fixes.

## incomplete-only
Round 1 input: 4444444..5555555. No committed fixes followed it.
Round 2 input: 4444444..5555555.
Spec report: complete Coverage, all listed BLOCKERs rejected after their
executed repros, plus "3 more BLOCKERs not listed".
Standards report: no Coverage; returned once to that same reviewer, whose
second report still has no Coverage. Both supplied final labels say CLEAN.
No proven BLOCKER or fix NOTE remains to implement. No commit followed round 2.

## third
Round 1 input: 6666666..7777777; Standards cut its BLOCKER list.
First fixes: 7777777..8888888.
Round 2 input: both 6666666..7777777 and 7777777..8888888; the original range
remained incomplete because the Standards list was still cut.
Second fixes: 8888888..9999999.
Round 3 input: both 6666666..7777777 and 8888888..9999999.
All listed BLOCKERs were rejected after executed repros; no fix remains.
Spec has Coverage. Standards has no Coverage even after its one return to
that same reviewer in round 3. Its supplied final label says CLEAN.
RECORDS
