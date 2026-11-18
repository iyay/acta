---
# guards: Read only until the cause is proven
# The checker writes .cache/check.json wherever it runs. A file there means it
# ran in the repo, not in the private copy.
type: file_exists
path: ".cache/*"
exists: false
---
