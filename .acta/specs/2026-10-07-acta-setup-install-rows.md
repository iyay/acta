---
parent: specs/2026-10-07-acta-setup-question-text
id: SPC-0103
created: "2026-10-07 13:07:52"
hash: sfyxand
started: "2026-10-07 13:08:12"
finished: "2026-10-07 13:09:11"
---
Status: Bounded, approved by the user in chat on 2026-10-07.
Why: on the plugin install screen of `acta setup`, each tool name runs straight into its answer (`claude● Yes`), and the rows do not line up because the names differ in length. Its title and description also say the same thing, which SPC-0102 removed everywhere else.

# acta setup: aligned install rows

## Changes

- Each install row reads `<tool name>` padded to the longest tool name found, then two spaces, then `● Yes  ○ No`, so the Yes/No marks start in the same column on every row. Example with claude and omp:
  ```
  claude  ● Yes  ○ No
  omp     ● Yes  ○ No
  ```
- The title becomes `Plugin install`, with the description `One row per tool found.` (no "I", in line with SPC-0102).
- The answered line stays `◇ Plugin install` / `│ claude: yes, omp: no`.

- The closing summary drops its `installed:` and `block:` lines (the install rows and the `acta block →` line already show both) and its `next:` line. It keeps `config: <path>`, and the closing `└` line reads `└  Run acta in a repo to browse specs, plans and bugs.` (user ruling 2026-10-07).

Nothing else changes: which tools are offered, the default yes, the install commands, `Plan`, `Apply`, layout and colors.

## Testing

A render test with tool names of different lengths (one, two and three tools) checks that the first mark sits in the same column on every row and that a space separates the name from the mark.
