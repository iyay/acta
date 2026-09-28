---
id: SCRATCH-6
hash: gc5a
title: harness-spec
status: raw
created: "2026-09-28"
---
# Spec 3: harness so agents do not skip the rules

Third of three specs, after spec 2 (skill rules).

User, 2026-09-27: skills and the main agent must be hardened so they do not misjudge or skip the scratch and brainstorm rules. Proposed:
1. Hook state: acta records each brainstorm start per session (session id), and the prompt-submit hook reminds "this session already brainstormed SCRATCH-n" when true.
2. plugincheck text guards per skill.
3. A behaviour eval suite (claude plugin eval) with scenarios: side idea mid-build goes to scratch, not memory; "catet aja" goes to scratch; second brainstorm asks the three choices; a one-file obvious fix gets no brainstorm; brainstorm start files a scratch item first; answers get appended.
4. Run the evals greenfield: a clean config and a scratch repo, because a global CLAUDE.md or AGENTS.md hides plugin gaps.
