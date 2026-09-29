package write

import "testing"

func TestBugTemplate(t *testing.T) {
	t.Parallel()

	want := "# ack dup on resend\n\n## Symptom\n\n## Root cause\n\n## Repro\n\n## Found in\n"
	if got := string(BugTemplate("ack dup on resend", "")); got != want {
		t.Fatalf("got %q", got)
	}
	if got := string(BugTemplate("x", "New-261")); got != "---\nref: New-261\n---\n# x\n\n## Symptom\n\n## Root cause\n\n## Repro\n\n## Found in\n" {
		t.Fatalf("with ref got %q", got)
	}
}

func TestBugFile(t *testing.T) {
	t.Parallel()

	cases := []struct{ name, title, slug, ref, body, want string }{
		{"title flag", "Print shows one type", "print-type", "", "## Symptom\nWrong value.\n",
			"# Print shows one type\n\n## Symptom\nWrong value.\n"},
		{"title from stdin heading", "", "print-type", "", "# From stdin\n\n## Symptom\nX\n",
			"# From stdin\n\n## Symptom\nX\n"},
		{"flag beats stdin heading", "Flag", "s", "", "# From stdin\n## Symptom\nX\n",
			"# Flag\n\n## Symptom\nX\n"},
		{"title from slug", "", "ack-dup-on-resend", "", "## Symptom\nX\n",
			"# ack dup on resend\n\n## Symptom\nX\n"},
		{"ref", "T", "s", "B-1", "## Symptom\nX\n",
			"---\nref: B-1\n---\n# T\n\n## Symptom\nX\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := string(BugFile(c.title, c.slug, c.ref, []byte(c.body))); got != c.want {
				t.Fatalf("got %q\nwant %q", got, c.want)
			}
		})
	}
}

func TestHasSymptom(t *testing.T) {
	t.Parallel()

	if !HasSymptom([]byte("# T\n\n## Symptom\nx\n")) || !HasSymptom([]byte("## Symptom  \n")) {
		t.Error("missed a Symptom section")
	}
	if HasSymptom([]byte("# T\n### Symptom\n")) || HasSymptom([]byte("# T\nSymptom\n")) {
		t.Error("matched a line that is not a ## Symptom heading")
	}
}
