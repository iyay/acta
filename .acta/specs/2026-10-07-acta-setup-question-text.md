---
parent: specs/2026-10-07-acta-setup-wizard-look
id: SPC-0102
created: "2026-10-07 12:59:28"
hash: z0zopnh
---
Status: Bounded, approved by the user in chat on 2026-10-07.
Why: in `acta setup` most descriptions repeat their question title, while the options that need a word of help (executor, subagent models, style) have none. The wizard also speaks as "I", but the wizard is not the agent; text names the agent instead.

# acta setup: short titles, useful help, no "I"

## Changes

Only the text in `internal/setup/form.go` changes. Titles are short nouns. A description says only what the title does not; when it has nothing to add, the question has no description line. Each option of a select carries a short note. No text in the wizard says "I", "me" or "my"; it says "the agent" or "the main agent".

| # | Title | Description | Options |
|---|---|---|---|
| 1 | Chat language | The language the agent chats in. Code and files use the repo language. | (input) |
| 2 | Reply style | (none) | `adhd`: short, one next step at the end · `plain`: normal paragraphs |
| 3 | Tone | Optional. Your own words, like "casual, no jargon". | (input) |
| 4 | Repo language | Code, comments, commits, specs and plans. | (input) |
| 5 | Build executor | (none) | `subagent`: helper agents in the same session · `inline`: the main agent writes the code · `dispatch`: an omp agent in its own herdr tab |
| 6 | Subagent models | Claude Code only. | `default`: your own config decides · `split`: sonnet writes code, the rest use a stronger model |
| 7 | Plan detail | (none) | `full`: real code in every step, waits for your yes · `minimal`: short steps, build starts at once |
| 8 | Questions | (none) | `one`: one question at a time · `probe`: a batch, each with the agent's pick |

The answered line uses the short title (`◇ Chat language`, `│ Indonesian`). The answer shown is the option value, without its note. `dispatch` is still offered only when `HERDR_ENV=1`.

Nothing else changes: values written, order, step count, layout, colors, `Plan`, `Apply`, `WriteBlock`, the no-TTY guard.

## Testing

A test pins every title, description (or its absence) and option note from the table, checks that no wizard text matches the words I, me or my, and checks the answered line shows the bare value.
