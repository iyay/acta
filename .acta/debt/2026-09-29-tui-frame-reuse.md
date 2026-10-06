---
id: DBT-0031
hash: u1m36g7
parent: plans/2026-09-29-tui-frame-reuse
---
# Review NOTEs: TUI Frame Reuse Implementation Plan

- [ ] (low) TestViewDrawsAgainAfterEveryChange does not cover the editor returning, a tab click, a status-line link click, or watchFailedMsg and a failed reload; wrongly setting same on any of them would stay green.
- [ ] (low) TestViewDrawsAgainAfterEveryChange copies base per case in a map-ordered loop, so the cases share the sel, idx and off slices and can leak cursor state into each other.
- [ ] (low) The modal-guard same=true in mouse is not pinned by any test; removing it leaves the suite green.
- [ ] (low) Nothing tests that WithTheme makes a fresh frame cache.
- [ ] (low) The status-line link click always draws although it changes nothing on screen; allowed, only a missed saving.
- [ ] (low) The flush-to-another-pane test fakes the focus change with m.focus = paneList instead of a real key or click.
- [ ] (low) drewNew in view_test.go assumes m.frame is not nil, so it panics on a model built without New.
- [ ] (low) Model copies share the off, sel, idx and newest slices and now the frame pointer, which is only safe while one model is live; keep in mind if an old copy is ever drawn.
