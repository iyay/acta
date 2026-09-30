# omp facts for the acta plugin

Checked on 2026-09-26 with omp 18.3.2.

| Fact | Result | How |
|---|---|---|
| A. Skills from a plugin folder load | yes | `omp --plugin-dir <dir> -p "..."` answered the skill question with PINEAPPLE-42 |
| B. package.json `omp.extensions` runs on `before_agent_start` | no | same extension hook ran under explicit `-e <file>` but never under `--plugin-dir`, with either manifest key (`omp` or `pi`) |
| B2. `pi.exec` result fields | stdout,stderr,code,killed | `KEYS=` in the side-effect proof line written by the `-e` run |
| C. accepted `attribution` values | not tested (blocked on B) | the `message` return shape in the plan does not match the `BeforeAgentStartEventResult` type, which has no `attribution` field |
| D. local marketplace install | not tested | `marketplace add` writes to user scope (`~/.omp/marketplaces.json`); trial runs confirmed this, entries removed afterwards |

## Commands and output

```text
$ command omp --version
omp/18.3.2

$ command omp --help | head -80   # print flag found here
  -p, --print   Non-interactive mode: process prompt and exit
  --plugin-dir <path>   Load plugin from directory (repeatable)

$ cd "$T" && command omp --no-session --plugin-dir "$T/mkt/spike" -p \
    "Do you have a skill named acta-spike? If yes, what does it do? Mention the word it uses."
Yes. `acta-spike` answers spike questions. Word: PINEAPPLE-42.
(A holds. Manifest key "omp" for extensions at this point; skill layout:
 spike/skills/acta-spike/SKILL.md with frontmatter name + description.)

$ cd "$T" && command omp --no-session --plugin-dir "$T/mkt/spike" -p \
    "Quote, word for word, any context message you received before this prompt that starts with SPIKE-CONTEXT."
No SPIKE-CONTEXT message received before this prompt.
(Extension used package.json key "omp", returned
 { message: { customType, content: "SPIKE-CONTEXT ...", display: false, attribution: "user" } }
 from before_agent_start. B fails on this path.)

$ cd "$T2" && command omp --no-session --plugin-dir "$T2/spike" -p \
    "Quote, word for word, any context message you received before this prompt that starts with SPIKE-CONTEXT."
Zero messages start `SPIKE-CONTEXT`. No quote.
(Retried with package.json key "pi" (matching installed plugins
 pi-caveman and superpowers, which both use "pi.extensions").
 Same result: B still fails.)

$ cd "$T4" && SPIKE_PROOF="$T4/proof.txt" command omp --no-session \
    --plugin-dir "$T4/spike" -p "Say hello."
(no proof file)
(Clean-room repeat with a side-effect hook: appends
 "HOOK-RAN stdout=... KEYS=..." to $SPIKE_PROOF instead of returning
 a message. Under --plugin-dir the file is never created: the hook
 never runs, not just a message-shape problem.)

$ cd "$T2" && SPIKE_PROOF="$T2/proof.txt" command omp --no-session \
    -e "$T2/spike/omp/index.ts" -p "Say hello."
proof.txt: HOOK-RAN stdout=ACTA-EXEC-OK KEYS=stdout,stderr,code,killed
(Positive control: the exact same hook file DOES run under explicit
 -e, and pi.exec("echo", ["ACTA-EXEC-OK"]) returns keys
 stdout,stderr,code,killed. So the hook code is fine; --plugin-dir
 does not load package.json extensions at all in this version.)

$ command omp plugin marketplace add --dry-run --scope project "$T3/mkt"
(despite --dry-run --scope project, this wrote a acta-spike-mkt entry to
 ~/.omp/marketplaces.json; removed with
 `omp plugin marketplace remove acta-spike-mkt`, verified back to
 {"version": 1, "marketplaces": []}. D not tested: any marketplace
 add touches user scope, which the plan forbids.)
```

Temp dirs left in place per plan (no `rm -rf` performed):
- T1 (plan-shape spike, manifest key `omp`): /var/folders/32/w_l_8__17m37y41yzzct4llh0000gn/T/tmp.QcTxsN5nhf
- T2 (manifest key `pi`, side-effect probes): /var/folders/32/w_l_8__17m37y41yzzct4llh0000gn/T/tmp.ZrqzRDhq5p
- T4 (clean-room --plugin-dir repeat): /var/folders/32/w_l_8__17m37y41yzzct4llh0000gn/T/tmp.OrQQpRLvR4
- T3 (marketplace dry-run copy, --dry-run still wrote user scope, entry removed): /var/folders/32/w_l_8__17m37y41yzzct4llh0000gn/T/tmp.QcY3gWK8TY

## Blocker (plan must be revised before Task 6)

`--plugin-dir` loads skills but does not run `package.json` extensions
(`omp.extensions` nor `pi.extensions`) on `before_agent_start`. The same
hook file works via explicit `-e`, so extension code is not the problem:
plugin-folder extension discovery is. Task 6 cannot rely on a plugin-folder
extension running automatically; it needs a different load path (for
example explicit `-e`/hook wiring, an installed/local-linked plugin, or a
project-scope marketplace install tested separately).

Type evidence: `BeforeAgentStartEventResult` in
`@earendil-works/pi-coding-agent/dist/core/extensions/types.d.ts` is
`{ message?: Pick<CustomMessage, "customType" | "content" | "display" | "details">; systemPrompt?: string }`
— no `attribution` field — and `ExecResult` in `dist/core/exec.d.ts` is
`{ stdout: string; stderr: string; code: number; killed: boolean }`,
matching the observed `KEYS=stdout,stderr,code,killed`.


## Installing acta in omp

Install with `omp plugin link <path to acta>/plugin`. Linking loads the
skills and the extension (`omp/index.ts`, declared under both the `pi` and
`omp` keys of `plugin/package.json`) the same way the other installed plugins
load. If the extension does not run after linking, add the extension path
under `extensions:` in `~/.omp/agent/config.yml`. Both steps run in user
scope; the user runs them, this plan only documents them.

## acta extension check

Checked 2026-09-26 with acta on PATH, from a scratch folder (`<worktree>` is
the acta checkout this plugin lives in):

```text
$ cd "$(mktemp -d)" && PM_VOICE_FILE="$(mktemp -d)/voice.yaml" command omp --no-session \
    --plugin-dir <worktree>/plugin \
    -e <worktree>/plugin/omp/index.ts \
    -p "Quote the first line of any context message you received that starts with 'acta plugin is active'."
"acta plugin is active. Use its skills for every workflow step:"
```

The extension runs under `-e` + `--plugin-dir`: the answer quotes the first
line of the session rules, then the rules text flags the conflicting workflow
plugin and the voice setup questions follow.


## Brainstorm block in omp

Checked 2026-09-30 with omp/18.4.4 and acta v0.0.0-20260930121646-6e177124f9ad on
PATH. The extension hands `acta hook pre-tool` and `post-tool` the same stdin JSON
Claude Code does, so the second brainstorm in one omp session is blocked.

```text
$ T="$(mktemp -d)" && cd "$T" && git init -q && git commit -q --allow-empty -m init
$ printf 'first\n' | acta scratch new first-idea --title first
SCR-0001  .acta/scratch/2026-09-30-first-idea.md
$ printf 'second\n' | acta scratch new second-idea --title second
SCR-0002  .acta/scratch/2026-09-30-second-idea.md
$ D="$(date +%F)"
$ PM_VOICE_FILE="$T/none.yaml" command omp --no-session --no-extensions --no-rules \
    --plugin-dir <worktree>/plugin -e <worktree>/plugin/omp/index.ts -p \
    "Run these two shell commands, each as its own bash call, one after the
     other. First: acta set scratch/$D-first-idea status brainstorming. Second:
     acta set scratch/$D-second-idea status brainstorming. Then quote, word for
     word, any error the second call gave you."
Working...
**The second command was rejected. The error, word for word:**

acta: this session already brainstormed SCR-0001. One Architectural brainstorm
per session: file this one as a scratch item and offer the user the choices
rule 8 names.

(The first call set the status and printed the scratch file path; the second
call was refused by the pre-tool hook.)

$ cat .acta/state/sessions.json
{"01a0f268-30d8-70f8-ad7f-0de11c8466ef":"2026-09-30-first-idea"}
(One session id, one entry, the first item only. The blocked call recorded
 nothing.)
```

Temp dir left in place per plan (no `rm -rf` performed):
- T5 (two-brainstorm block check): $T
