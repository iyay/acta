---
# guards: ## Finding bar: BLOCKER or NOTE
# The branch adds tests/test_discount.py, which reads src/discount.py and
# asserts it holds one line. A text test is the one test finding that is a
# BLOCKER, so its path and a line of the test (the class, the def, the read
# or the assert, lines 5 to 8) must show before any line that starts the
# NOTEs, in any case such as "Notes:". A reply that names it only as a NOTE
# fails.
type: regex
pattern: '^(?:(?!\n[#*\s-]*NOTEs?\b)[\s\S])*?tests/test_discount\.py:[5-8]\b'
target: last_message
flags: i
---
