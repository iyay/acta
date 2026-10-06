---
id: DBT-0045
hash: zbwdlcc
parent: plans/2026-09-30-setup-subagent-models-default
---
# Review NOTEs: Setup Remembers a No to Split Subagent Models Implementation Plan

- [ ] (low) The User type comment in internal/config/user.go still says "how the agent should talk to the user" and stops at repo language; the type now also holds executor, subagent models and theme.
- [ ] (low) ErrBadUser keeps the text "bad voice setting" and doc comments on UserDefault, LoadUser and SaveUserFile still say "voice file".
- [ ] (low) Moved tests in internal/config/user_test.go keep names TestDefault, TestPath, TestLoadMissing, TestSaveThenLoad, TestValidate; in package config they read like tests of the planning-file config.Load.
- [ ] (low) Callers keep old voice naming: cli.voiceTheme, cli.loadVoice, hook.Input.Voice, doctor.Env.Voice, while the type is now config.User.
- [ ] (low) Setup skill "ask only while subagent_models is empty" is prose only; no test covers the ask rule itself.
