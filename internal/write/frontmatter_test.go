package write

import (
	"strings"
	"testing"
)

func TestSetField(t *testing.T) {
	cases := []struct {
		name, in, key, value, want string
	}{
		{"change a field, keep order and body",
			"---\nref: New-261\nstatus: open\nowner: rian\n---\n# T\n\nbody  \n",
			"status", "fixed",
			"---\nref: New-261\nstatus: fixed\nowner: rian\n---\n# T\n\nbody  \n"},
		{"add a missing field at the end",
			"---\nref: New-261\n---\n# T\n",
			"fixed_in", "abc123",
			"---\nref: New-261\nfixed_in: abc123\n---\n# T\n"},
		{"no frontmatter gets a new block",
			"# T\nbody\n",
			"status", "done",
			"---\nstatus: done\n---\n# T\nbody\n"},
		{"empty block",
			"---\n---\n# T\n",
			"status", "open",
			"---\nstatus: open\n---\n# T\n"},
		{"closing line at end of file",
			"---\nstatus: open\n---",
			"status", "fixed",
			"---\nstatus: fixed\n---\n"},
		{"a number-like value stays a string",
			"---\nstatus: open\n---\n",
			"ref", "261",
			"---\nstatus: open\nref: \"261\"\n---\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := SetField([]byte(c.in), c.key, c.value)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != c.want {
				t.Fatalf("got\n%q\nwant\n%q", got, c.want)
			}
		})
	}
}

func TestSetFieldLineEndings(t *testing.T) {
	cases := []struct {
		name, in, key, value, wantBody string
		crlf                           bool
	}{
		{"LF keeps body byte for byte",
			"---\nref: TICK-7\nstatus: draft\n---\n# Win spec\n",
			"status", "approved", "# Win spec\n", false},
		{"CRLF keeps ref and body, stays CRLF",
			"---\r\nref: TICK-7\r\nstatus: draft\r\n---\r\n# Win spec\r\n",
			"status", "approved", "# Win spec\r\n", true},
		{"no trailing newline keeps body",
			"---\nref: TICK-7\nstatus: draft\n---\n# Win spec",
			"status", "approved", "# Win spec", false},
		{"CRLF closing fence as last line",
			"---\r\nref: TICK-7\r\nstatus: draft\r\n---",
			"status", "approved", "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := SetField([]byte(c.in), c.key, c.value)
			if err != nil {
				t.Fatal(err)
			}
			s := string(got)
			norm := strings.ReplaceAll(s, "\r\n", "\n")
			if strings.Count(norm, "---\n") != 2 {
				t.Fatalf("want one frontmatter block, got %q", s)
			}
			if !strings.Contains(norm, "ref: TICK-7\n") {
				t.Fatalf("ref lost: %q", s)
			}
			if !strings.Contains(norm, "status: approved\n") {
				t.Fatalf("new value missing: %q", s)
			}
			close := strings.Index(norm, "\n---\n")
			body := ""
			if close >= 0 {
				body = norm[close+len("\n---\n"):]
			}
			if want := strings.ReplaceAll(c.wantBody, "\r\n", "\n"); body != want {
				t.Fatalf("body = %q, want %q (full %q)", body, want, s)
			}
			if c.crlf && strings.Contains(strings.ReplaceAll(s, "\r\n", ""), "\n") {
				t.Fatalf("bare LF in CRLF file: %q", s)
			}
			if !c.crlf && strings.Contains(s, "\r") {
				t.Fatalf("CR leaked into LF file: %q", s)
			}
		})
	}
}

func TestSetFieldMixedLineEndings(t *testing.T) {
	repro := "---\r\nref: TICK-7\r\nstatus: draft\n---\r\n# Win spec\r\n\r\nGoal text.\r\n\r\n---\r\n\r\n## Tasks\r\n"
	got, err := SetField([]byte(repro), "status", "approved")
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	for _, keep := range []string{"# Win spec\r\n", "Goal text.\r\n", "---\r\n\r\n## Tasks\r\n"} {
		if !strings.Contains(s, keep) {
			t.Errorf("body lost %q in %q", keep, s)
		}
	}
	norm := strings.ReplaceAll(s, "\r\n", "\n")
	if want := "---\nref: TICK-7\nstatus: approved\n---\n# Win spec\n\nGoal text.\n\n---\n\n## Tasks\n"; norm != want {
		t.Errorf("got\n%q\nwant\n%q", norm, want)
	}
}

func TestSetFieldClosingFenceLineEndings(t *testing.T) {
	cases := []struct{ name, in, wantBody string }{
		{"LF open CRLF close", "---\nref: TICK-7\nstatus: draft\r\n---\r\n# Win spec\n", "# Win spec\n"},
		{"CRLF file body fence", "---\r\nref: TICK-7\r\nstatus: draft\r\n---\r\n# Win spec\r\n\r\n---\r\n\r\ntail\r\n", "# Win spec\r\n\r\n---\r\n\r\ntail\r\n"},
		{"LF file body fence", "---\nref: TICK-7\nstatus: draft\n---\n# Win spec\n\n---\n\ntail\n", "# Win spec\n\n---\n\ntail\n"},
		{"closing fence last line LF", "---\nref: TICK-7\nstatus: draft\n---", ""},
		{"closing fence last line CRLF", "---\r\nref: TICK-7\r\nstatus: draft\r\n---", ""},
		{"closing fence last line LF newline", "---\nref: TICK-7\nstatus: draft\n---\n", ""},
		{"closing fence last line CRLF newline", "---\r\nref: TICK-7\r\nstatus: draft\r\n---\r\n", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := SetField([]byte(c.in), "status", "approved")
			if err != nil {
				t.Fatal(err)
			}
			s := string(got)
			norm := strings.ReplaceAll(s, "\r\n", "\n")
			head, body, ok := strings.Cut(norm, "\n---\n")
			if !ok {
				t.Fatalf("no closing fence in %q", s)
			}
			if head != "---\nref: TICK-7\nstatus: approved" {
				t.Fatalf("frontmatter changed: %q (full %q)", head, s)
			}
			if want := strings.ReplaceAll(c.wantBody, "\r\n", "\n"); body != want {
				t.Fatalf("body = %q, want %q (full %q)", body, want, s)
			}
		})
	}
}

func TestSetFieldMissingFenceWritesNothing(t *testing.T) {
	for name, in := range map[string]string{
		"LF file no fence":   "---\nref: TICK-7\nstatus: draft\n# Win spec\n",
		"CRLF file no fence": "---\r\nref: TICK-7\r\nstatus: draft\r\n# Win spec\r\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := SetField([]byte(in), "status", "approved"); err == nil {
				t.Fatal("want an error for a missing closing fence")
			}
		})
	}
}

func TestSetFieldErrors(t *testing.T) {
	for name, in := range map[string]string{
		"unclosed block": "---\nstatus: open\n# T\n",
		"not a mapping":  "---\n- a\n- b\n---\n",
		"broken yaml":    "---\nref: [x\n---\n",
	} {
		if _, err := SetField([]byte(in), "status", "done"); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}
