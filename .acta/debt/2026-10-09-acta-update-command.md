---
id: DBT-0103
hash: q0n1hcw
parent: plans/2026-10-09-acta-update-command
---
# acta update hardening

- [ ] (medium) ACTA_DOWNLOAD_URL accepts an http:// base; https-only covers redirects only (internal/cli/update_cmd.go:57, internal/update/fetch.go:51,84). Reject a non-https base.
- [ ] (low) The tag read from Location is not validated before it goes into the download URL (internal/update/fetch.go:71). Check it against ^v\d+\.\d+\.\d+$.
- [ ] (low) Any latest tag that differs from the current version is installed, so a /latest pointing at an older tag downgrades. Compare versions.
