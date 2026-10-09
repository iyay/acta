---
parent: specs/2026-10-09-acta-update-command-design
depth: minimal
closes: [SCR-0053]
id: PLN-0126
created: "2026-10-09 16:13:15"
hash: vbuiaa9
---
# acta update command

**Goal:** `acta update` replaces a release binary with the latest verified release and refreshes the plugin; `--check` only reports versions.

**Spec:** `.acta/specs/2026-10-09-acta-update-command-design.md`

**Tests:** fast `scripts/test ./internal/update ./internal/cli`; full `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Stdlib only (`net/http`, `archive/tar`, `compress/gzip`, `crypto/sha256`). No new dependency.
- Every download is https, also on redirects. Timeout 60 s per request.
- Base URL is `ACTA_DOWNLOAD_URL`, default `https://github.com/iyay/acta/releases`. Tests use `httptest.NewTLSServer`; no test reaches GitHub.
- Failure exit code is 1 (`exitBadInput` value), as the spec says. Messages copied verbatim from the spec.
- No version bump. Windows out of scope.
- Tests never run `rm -rf` or `rm -f`; use `t.TempDir()`.

## Waves

- Wave 1: Task 01, Task 02, Task 03
- Wave 2: Task 04

### Task 01: Release marker

**Files:**
- Create: `internal/update/release.go`
- Modify: `.goreleaser.yaml` (builds `ldflags`, and its top comment that says nothing is set at link time)
- Test: `internal/update/release_test.go`

**verify:** A binary counts as a release build only when goreleaser stamped it. List every way a binary is built in this repo (`go install`, `go build`, `go test`, goreleaser) and what `IsRelease` returns for each.

- [ ] **Step 1: Check claude commands first.** Run `claude plugin marketplace update --help` and `claude plugin update --help`. Write the exact working forms with `acta state set <plan id> next`. If either does not exist, stop and report; Task 04 depends on them.
- [ ] **Step 2: Failing test** `TestGoreleaserStampsRelease`: read `../../.goreleaser.yaml`, check it holds `-X github.com/iyay/acta/internal/update.release=true`. `TestIsReleaseUnstamped`: `IsRelease()` is false in a test binary.
- [ ] **Step 3: Watch it fail:** `scripts/test ./internal/update -run 'TestGoreleaserStampsRelease|TestIsReleaseUnstamped'`
- [ ] **Step 4: Minimal code:** `var release string` and `func IsRelease() bool { return release == "true" }`; add the `ldflags` line under `builds[0]` and fix the comment.
- [ ] **Step 5: Pass, then commit** with gofmt and `go vet ./internal/update` clean: `feat(update): stamp release builds`

### Task 02: Find and fetch the latest release

**Files:**
- Create: `internal/update/fetch.go`
- Test: `internal/update/fetch_test.go`

**verify:** No bytes that fail the checksum ever come back as a binary. List every failure path (bad status, http redirect, timeout, missing checksum line, mismatch, archive without `acta`) and what each returns.

**Interfaces:**
- Produces: `func Latest(c *http.Client, base string) (string, error)` returns the tag like `v0.1.46` from the `Location` of `<base>/latest` (client must not follow redirects for this call).
- Produces: `func Fetch(c *http.Client, base, tag, goos, goarch string) ([]byte, error)` downloads `<base>/download/<tag>/acta_<goos>_<goarch>.tar.gz` and `checksums.txt`, checks SHA-256 (accept `name` and `*name`), returns the `acta` file bytes from the archive.
- Produces: `func NewClient() *http.Client` with 60 s timeout and a `CheckRedirect` that refuses any non-https target.

- [ ] **Step 1: Failing tests**, table-driven over one `httptest.NewTLSServer`: good tag, good fetch, mismatch, missing line, `*name` line, archive without `acta`, redirect to `http://`. Use `srv.Client()` transport inside a client built like `NewClient` so TLS trusts the test cert.
- [ ] **Step 2: Watch them fail:** `scripts/test ./internal/update -run 'TestLatest|TestFetch'`
- [ ] **Step 3: Minimal code** in `fetch.go`.
- [ ] **Step 4: Pass:** `scripts/test ./internal/update`
- [ ] **Step 5: Commit** with gofmt and vet clean: `feat(update): fetch and verify the latest release`

### Task 03: Replace the binary in place

**Files:**
- Create: `internal/update/replace.go`
- Test: `internal/update/replace_test.go`

**verify:** The old binary is never left missing or half-written. List every failure path (dir not writable, write fails, chmod fails, rename fails) and the state of the old file and the temp file after each.

**Interfaces:**
- Produces: `func Replace(exe string, data []byte) error`. Resolves `exe` with `filepath.EvalSymlinks`, writes `.acta.<pid>` in its dir, chmod 755, renames over it, removes the temp file on any failure. A dir without write access returns an error whose text is `no write access to <dir>`.

- [ ] **Step 1: Failing tests:** good replace (content and mode 0755); symlinked exe replaces the target and keeps the link; read-only dir (chmod 0555 on a `t.TempDir()` subdir, skip when running as root) gives the exact message and leaves the old file and no `.acta.*`.
- [ ] **Step 2: Watch them fail:** `scripts/test ./internal/update -run TestReplace`
- [ ] **Step 3: Minimal code** in `replace.go`.
- [ ] **Step 4: Pass:** `scripts/test ./internal/update`
- [ ] **Step 5: Commit** with gofmt and vet clean: `feat(update): replace the binary atomically`

### Task 04: acta update command

**Files:**
- Create: `internal/cli/update_cmd.go`
- Modify: `internal/cli/cli.go` (dispatch `case "update"`, add `update` to the unknown-command list)
- Test: `internal/cli/update_cmd_test.go`

**verify:** No run of `acta update` exits 0 unless the binary is the latest and the plugin refresh finished. List every path (source build, latest already, `--check`, fetch error, replace error, refresh error, success) with its stdout, stderr and exit code.

**Interfaces:**
- Consumes: `update.IsRelease`, `update.NewClient`, `update.Latest`, `update.Fetch`, `update.Replace`, `setup.ExtractPlugin(plugin.Files, homeDir())`, `plugin.Version()`.
- Produces: `func cmdUpdate(args []string, stdout, stderr io.Writer) int`; flags `--check` and hidden `--refresh-plugin`. Package vars `isRelease = update.IsRelease` and `exePath = os.Executable` so tests can swap them.

- [ ] **Step 1: Failing tests** in `update_cmd_test.go`, one per verify path. TLS test server behind `ACTA_DOWNLOAD_URL`; client swap via a package var so the test cert is trusted. Refresh tests put a fake `claude` script on PATH (`t.Setenv`) that logs its args to a file, and set HOME to `t.TempDir()`. Success test: exe is a temp file; after run it holds the new bytes; the re-exec of `--refresh-plugin` is a package var `runRefresh func(exe string) error` swapped to call `cmdUpdate([]string{"--refresh-plugin"}, ...)` in process.
- [ ] **Step 2: Watch them fail:** `scripts/test ./internal/cli -run TestUpdate`
- [ ] **Step 3: Minimal code:** flow exactly as spec section "Flow of `acta update`", steps 1 to 5, messages verbatim. `--refresh-plugin` extracts, then when `exec.LookPath("claude")` finds it, runs the two commands recorded in Task 01's state note. Any refresh error: `plugin refresh failed: run acta setup`, exit 1.
- [ ] **Step 4: Pass:** `scripts/test ./internal/cli -run TestUpdate` then `scripts/test ./internal/update ./internal/cli`
- [ ] **Step 5: Commit** with gofmt and vet clean: `feat(cli): acta update command`
