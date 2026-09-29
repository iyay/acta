package write

import "testing"

func TestDatesFor(t *testing.T) {
	fixNow(t)
	src := "---\nid: SCRATCH-1\nstatus: raw\n---\n# T\n"
	started := "---\nid: SCRATCH-1\nstatus: raw\nstarted: \"2026-09-26\"\n---\n# T\n"
	finished := "---\nid: SCRATCH-1\nstatus: raw\nfinished: \"2026-09-26\"\n---\n# T\n"
	cases := []struct{ status, want string }{
		{"brainstorming", started},
		{"in-progress", started},
		{"fixing", started},
		{"done", finished},
		{"wontfix", finished},
		{"dropped", finished},
		{"fixed", finished},
		{"specced", finished},
	}
	for _, c := range cases {
		out, err := DatesFor([]byte(src), c.status)
		if err != nil {
			t.Errorf("%s: %v", c.status, err)
			continue
		}
		// The whole file is compared, so a body or another field that moved
		// or changed is a failure too.
		if string(out) != c.want {
			t.Errorf("%s:\n got %q\nwant %q", c.status, out, c.want)
		}
	}
}

func TestStartedIsWrittenOnce(t *testing.T) {
	fixNow(t)
	in := "---\nstarted: 2026-01-01\n---\n# T\n"
	out, err := DatesFor([]byte(in), "in-progress")
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != in {
		t.Errorf("started was rewritten:\n got %q\nwant %q", out, in)
	}
}

func TestReCloseGetsTheNewDay(t *testing.T) {
	fixNow(t)
	in := "---\nfinished: 2026-02-01\n---\n# T\n"
	out, err := DatesFor([]byte(in), "done")
	if err != nil {
		t.Fatal(err)
	}
	if want := "---\nfinished: \"2026-09-26\"\n---\n# T\n"; string(out) != want {
		t.Errorf("\n got %q\nwant %q", out, want)
	}
}

func TestMarkFinishedOnce(t *testing.T) {
	fixNow(t)
	// The close day a person already read stays the day they read.
	old := "---\nfinished: \"2026-01-01\"\n---\n# T\n"
	out, err := MarkFinishedOnce([]byte(old))
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != old {
		t.Errorf("the close day moved:\n got %q\nwant %q", out, old)
	}
	// A file with no close day gets today.
	open := "---\nstatus: raw\n---\n# T\n"
	out, err = MarkFinishedOnce([]byte(open))
	if err != nil {
		t.Fatal(err)
	}
	if want := "---\nstatus: raw\nfinished: \"2026-09-26\"\n---\n# T\n"; string(out) != want {
		t.Errorf("\n got %q\nwant %q", out, want)
	}
}

func TestReopenClearsFinished(t *testing.T) {
	fixNow(t)
	in := "---\nstarted: 2026-01-01\nfinished: 2026-02-01\n---\n# T\n"
	out, err := DatesFor([]byte(in), "fixing")
	if err != nil {
		t.Fatal(err)
	}
	// The first started stays, so the work keeps the day it really began.
	if want := "---\nstarted: 2026-01-01\n---\n# T\n"; string(out) != want {
		t.Errorf("\n got %q\nwant %q", out, want)
	}
}

func TestOtherStatusChangesNoDate(t *testing.T) {
	fixNow(t)
	for _, s := range []string{"raw", "draft", "approved", "open"} {
		in := "---\nstatus: x\n---\n# T\n"
		out, err := DatesFor([]byte(in), s)
		if err != nil {
			t.Errorf("%s: %v", s, err)
			continue
		}
		if string(out) != in {
			t.Errorf("%s changed the file: %q", s, out)
		}
	}
}
