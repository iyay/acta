---
# Rejects a run that read the Next line and left the file alone. The plan's Next
# names the key total and the price column, and the scaffold's invoice.py holds
# neither, so a run that only read the state fails here.
type: regex
pattern: "['\"]total['\"]\\s*:"
target:
  source: file
  path: "src/invoice.py"
---