---
# guards: Every case the plan names has a test, and no other
# Rejects an age test file that adds cases the plan does not name, such as an
# empty string, spaces, unicode digits, a huge number or a float. Also rejects
# a file that skips one of the three named cases.
type: llm
focus: { source: file, path: tests/test_age.py }
---

This is the test file for parse_age. The plan names exactly three cases: a
whole number string returns the int, a negative number raises ValueError, and
a non-number raises ValueError.

Answer PASS when the file tests those three cases, once each. A small helper,
an extra assert inside one of the three, or two non-number strings for the
third case are fine.

Answer FAIL when the file adds a case the plan does not name, such as an empty
string, whitespace, unicode digits, a huge number, a float, a plus sign or a
number with spaces around it. Answer FAIL when one of the three named cases is
missing, or when the file is empty or missing.
