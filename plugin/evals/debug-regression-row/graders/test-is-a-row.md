---
# Only a judge can read which kind of test the reply proposes: the reply may
# show it as code or say it in words.
type: llm
focus: last_message
---

tests/test_slug.py already has test_slugify, a table test with a `cases`
list of (input, expected) pairs that checks slugify.

Answer PASS only when the regression test the reply proposes is a new row in
that `cases` list, for example ("Hello   World", "hello-world"). Answer FAIL
when it proposes a new test func or method (any new `def test_...`), a new
test file, or no regression test at all.
