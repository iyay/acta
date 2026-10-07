---
ref: SCR-0031
id: BUG-0021
hash: rgja94n
fixed_in: a852ada
finished: "2026-10-01 18:46:59"
---
# Setup never lets a user change the chat language once one is saved

## Symptom
A user whose config already has a chat language runs setup to change it. Setup reports "voice: English, adhd — already set" and asks only the unset parts (here the build executor). The user gets no way to change the language inside setup. The only way left is the `acta config set` CLI, which a new user does not know exists. First-run onboarding does not walk the user through every choice.

## Root cause
plugin/skills/setup/SKILL.md:13 tells the first run to ask "only for what `acta config show` says is not set yet". plugin/skills/setup/SKILL.md:56 says "A later run asks which part to change", but nothing in the skill says how to tell a first run from a later run. A config that is set only in part (voice set, executor missing) takes the First run path, so voice is never offered and the Change later path never runs.

## Repro
1. Make ~/.acta/config.yaml hold only `chat_language: English`, `style: adhd`, `repo_language: English` (no build_executor).
2. In omp (or Claude Code), run setup and say you want a different language.
3. Setup asks only for the build executor and reports voice as already set.

Seen in omp on 2026-09-30 after a review mutant wiped the user's config to one line.

## Found in
main, by the user while restoring their config after the PLN-0056 review; raw idea filed first as SCR-0031.
