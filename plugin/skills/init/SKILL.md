---
name: init
description: "acta: Use when the user wants acta set up in the repo they are in, or asks to init a new repo for acta. Runs acta init, which makes the .acta folder and writes the acta block in one commit, and asks nothing."
---

# Init

Run `acta init` at the repo, then report what it printed. Ask nothing first: the settings come from the user's own config.

- It refuses a folder outside git and a CLAUDE.md or AGENTS.md with broken acta markers. Pass its error on as it is and stop. Fix nothing yourself.
- `written, not committed:` means a file already had the user's own changes. Name the reason and let the user commit.
- A `fail` line in the doctor report below it is not an init failure. Report each one with its `fix:` line.

To change a setting, use acta:setup instead.
