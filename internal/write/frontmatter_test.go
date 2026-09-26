package write

import "testing"

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
