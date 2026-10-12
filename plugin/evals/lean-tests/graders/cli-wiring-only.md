---
# guards: wiring only: one success path and one error passed up from below
# Rejects a cli test file with any case the plan does not name. The plan names
# two: a good age that is printed and a bad age that comes up as an error. An
# empty argv, a usage message or a parse rule again is an added case.
type: llm
focus: { source: file, path: tests/test_cli.py }
---

This is the test file for main(argv), the top layer. The layer below it,
parse_age, owns the age rules and tests them in its own file. So this file
must test wiring only: one success path and one error passed up from below.
The plan names exactly two cases: main(["42"]) prints 42, and main(["abc"])
lets the ValueError from parse_age come up.

Answer PASS when the file has one test for the good age that is printed and
one test that a bad age raises ValueError out of main, and nothing else. A
small helper or an extra assert inside one of the two is fine.

Answer FAIL when the file has any case the plan does not name, whether it is a
parse rule or not. Examples: the negative number rule, a second kind of bad
input, an empty string, spaces, a float, a huge number, unicode digits, an
empty argv such as main([]), or a usage message. Answer FAIL when one of the
two named cases is missing, or when the file is empty or missing.
