---
id: DBT-0102
hash: mkal5uc
parent: plans/2026-10-08-curl-install-release-pipeline
---
# Windows CLAUDE_PLUGIN_ROOT path bug

- [ ] On Windows, the hook command path built from `CLAUDE_PLUGIN_ROOT` can come out mangled, so the hooks may not run at all. This is tracked upstream in anthropics/claude-code #18527 and #21878.
- [ ] Hooks stay broken on Windows until that is fixed upstream, or until the hooks call `acta` directly and no longer depend on the plugin root path.
- [ ] (medium) ci.yml and release.yml have never run on GitHub; the first push is the first proof that go test ./... passes on a clean ubuntu runner and that the windows-hooks job and the install.ps1 parse step work.
- [ ] (low) install.ps1 behaviour (the scope wrap, the v prefix, the hash check, the Git Bash warning) has no test beyond the CI parse step.
- [ ] (low) scripts/release: when git commit succeeds and git tag -a then fails (for example tag.gpgSign with no key), the release commit stays with no tag, and a second run bumps the version again.
- [ ] (low) No test covers the --tlsv1.2 floor in install.sh; removing that flag alone keeps every test green.
- [ ] (low) scripts/install_test.go passes https_proxy, HTTPS_PROXY and ALL_PROXY through to curl, so a machine with an https proxy and no no_proxy for 127.0.0.1 fails the install tests.
- [ ] (low) scripts/install_test.go reads r.hits without r.mu in failure messages and in one check; a locked copy helper like sawPath would close it.
- [ ] (low) The spec and plan still say ACTA_VERSION=vX.Y.Z; a bare X.Y.Z works too since the polish round.
