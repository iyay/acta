---
name: setup
description: Use when the session rules say the voice is not set up yet, or when the user asks to change the chat language, the style (adhd or plain), the tone, or the language used for files in the repo.
---

# Setup

The chat language, style and tone live in `~/.pm/voice.yaml` (or the file `PM_VOICE_FILE` names). `pmb` reads it at the start of every session and before every message.

## First run

When the session rules say the voice is not set up, ask these once, in English, before other work:

1. Which language should I use when I talk with you? (default: English)
2. Style: `adhd` (the answer or next action first, short numbered steps) or `plain`? (default: adhd)
3. Anything about tone, in your own words? (optional)

Save the answers, writing the language as its full English name (Korean, not ko):

```bash
pmb voice set --language Korean --style adhd --tone "Casual, short sentences."
```

From the next message on, talk in the chosen language.

## Change later

Pass only the flags that change: `pmb voice set --style plain`, `pmb voice set --clear-tone`, `pmb voice set --repo-language English`. `pmb voice show` prints the current setting.

## Limits

- The user's own CLAUDE.md or AGENTS.md wins when it names a language or style.
- Tone is at most 8 lines and 600 characters.
- This skill writes only the voice file. It never edits CLAUDE.md, AGENTS.md or settings.
