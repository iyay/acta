# omp facts for the pm plugin

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
    "Do you have a skill named pm-spike? If yes, what does it do? Mention the word it uses."
Yes. `pm-spike` answers spike questions. Word: PINEAPPLE-42.
(A holds. Manifest key "omp" for extensions at this point; skill layout:
 spike/skills/pm-spike/SKILL.md with frontmatter name + description.)

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
proof.txt: HOOK-RAN stdout=PMB-EXEC-OK KEYS=stdout,stderr,code,killed
(Positive control: the exact same hook file DOES run under explicit
 -e, and pi.exec("echo", ["PMB-EXEC-OK"]) returns keys
 stdout,stderr,code,killed. So the hook code is fine; --plugin-dir
 does not load package.json extensions at all in this version.)

$ command omp plugin marketplace add --dry-run --scope project "$T3/mkt"
(despite --dry-run --scope project, this wrote a pm-spike-mkt entry to
 ~/.omp/marketplaces.json; removed with
 `omp plugin marketplace remove pm-spike-mkt`, verified back to
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


## Installing pm in omp

Install with `omp plugin link <path to pm-board>/plugin`. Linking loads the
skills and the extension (`omp/index.ts`, declared under both the `pi` and
`omp` keys of `plugin/package.json`) the same way the other installed plugins
load. If the extension does not run after linking, add the extension path
under `extensions:` in `~/.omp/agent/config.yml`. Both steps run in user
scope; the user runs them, this plan only documents them.

## pm extension check

Checked 2026-09-26 with pmb on PATH, from a scratch folder (`<worktree>` is
the pm-board checkout this plugin lives in):

```text
$ cd "$(mktemp -d)" && PM_VOICE_FILE="$(mktemp -d)/voice.yaml" command omp --no-session \
    --plugin-dir <worktree>/plugin \
    -e <worktree>/plugin/omp/index.ts \
    -p "Quote the first line of any context message you received that starts with 'pm plugin is active'."
"pm plugin is active. Use its skills for every workflow step:"
```

The extension runs under `-e` + `--plugin-dir`: the answer quotes the first
line of the session rules, then the rules text flags the conflicting workflow
plugin and the voice setup questions follow.
