---
id: DBT-0053
hash: rh5pqso
parent: plans/2026-09-30-config-set-broken-file-path
---
# Review NOTEs: Config Set Broken File Path Implementation Plan

- [ ] (low) TestConfigSetNamesBrokenConfigFile and TestConfigSetNamesBrokenPMVoiceFile stay green when the fix is reverted; they guard paths that already printed the right file.
- [ ] (low) ResolveUserFile returns UserPath, not the failing file, when voicePath or oldPath fails; only possible if UserHomeDir fails after UserPath succeeded.
- [x] Plan named one test; the branch added three more plus two helpers to cover the whole verify line. (stale)
- [ ] (low) setWithBrokenFile returns home, used by one of three callers; TestConfigSetNamesBrokenOldFile repeats its setup and the brokenVoice literal.
