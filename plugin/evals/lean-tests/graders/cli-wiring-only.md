---
# guards: wiring only: one success path and one error passed up from below
# Rejects a cli test file that checks the age rules again. The negative rule
# and the non-number rule belong to parse_age, which has its own tests, so the
# cli tests may only show one success and one error coming up from below.
type: llm
focus: { source: file, path: tests/test_cli.py }
---

This is the test file for main(argv), the top layer. The layer below it,
parse_age, owns the age rules and tests them in its own file. So this file
must test wiring only: one success path and one error passed up from below.

Answer PASS when the file has at most one test for a good age that is printed
and at most one test that a bad age raises ValueError out of main. Two to four
small test methods in all is fine.

Answer FAIL when the file tests the negative number rule and the non-number
rule both, as separate cases, or adds other parse cases such as an empty
string, spaces, a float, a huge number or unicode digits. Answer FAIL when the
file is empty or missing.
