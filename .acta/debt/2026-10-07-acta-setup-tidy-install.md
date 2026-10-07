---
id: DBT-0095
hash: g89hdi0
parent: plans/2026-10-07-acta-setup-tidy-install
---
# acta setup tidy install review debt

- [ ] (medium) config.UserPath has the same relative-HOME gap the extractor had: with HOME=. the config file would be written under the current folder (internal/config/user.go UserPath).
- [ ] (low) Two acta setup runs at once: the second swap fails with "could not unpack", and a crashed run leaves plugin-new-* folders in ~/.acta that nothing cleans up (internal/setup/extract.go).
- [ ] (low) No end-to-end test runs acta setup with a relative HOME to check the problem line and that setup carries on.
- [ ] (low) Tests that swap renameFile, os.Stdin or the working folder break if internal/setup or internal/cli tests ever use t.Parallel.
