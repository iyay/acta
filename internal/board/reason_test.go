package board

import "testing"

// reason says why an item was closed, so the board must carry it as written.
func TestReasonIsRead(t *testing.T) {
	t.Parallel()

	const id = "bugs/2026-10-01-crash"
	b := boardWith(t, map[string]string{
		id + ".md": "---\nstatus: wontfix\nreason: the old runner is gone\n---\n# App crashes on start\n",
	})
	it := b.Get(id)
	if it == nil {
		t.Fatalf("%s missing from the board", id)
	}
	if it.Reason != "the old runner is gone" {
		t.Errorf("reason = %q", it.Reason)
	}
}
