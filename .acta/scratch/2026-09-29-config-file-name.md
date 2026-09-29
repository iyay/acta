---
id: SCR-0024
hash: uehvwoo
title: Rename ~/.acta/voice.yaml to config.yaml
status: raw
created: "2026-09-29"
---
User: "kenapa file config kita voice.yml ya. kok aneh. gabisa pake config.yml aja?" then "catet" (2026-09-29)

Agent lines:
- voice.yaml started as voice only (chat_language, style, tone); it now also holds build_executor and subagent_models, so the name no longer fits.
- Proposal: ~/.acta/config.yaml. Read voice.yaml when config.yaml is missing; on first write, rename the old file so no setting is lost.
- About 21 files name voice.yaml (Go code, skills, tests). `acta voice show` can stay, or gain an `acta config show` alias.
- Start after PLAN-32 (short ids) and PLAN-33 (rule 8) land, to avoid clashes.
