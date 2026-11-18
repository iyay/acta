---
# guards: Read only until the cause is proven
# Importing src from the repo leaves src/__pycache__ behind.
type: file_exists
path: "src/__pycache__/*"
exists: false
---
