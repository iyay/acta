package gitc

import (
	"fmt"
	"strings"
	"testing"
)

func TestFoldSubject(t *testing.T) {
	long := make([]string, 0, 12)
	for _, n := range []string{"alpha-one", "bravo-two", "charlie-three", "delta-four", "echo-five", "foxtrot-six", "golf-seven"} {
		long = append(long, "chore(scratch): new "+n)
	}
	cases := []struct {
		name string
		in   []string
		want string
	}{
		{"one subject comes back as it is", []string{"chore(scratch): new a"}, "chore(scratch): new a"},
		{"same verb joins the names", []string{"chore(scratch): new a", "chore(scratch): new b"}, "chore(scratch): new a, b"},
		{"names already joined stay split", []string{"chore(debt): tick DBT-1, DBT-2", "chore(debt): tick DBT-3"}, "chore(debt): tick DBT-1, DBT-2, DBT-3"},
		{"mixed verbs group per verb", []string{"chore(bug): new x", "chore(bug): fixed_in y", "chore(bug): new z"}, "chore(bug): new x, z; fixed_in y"},
		{"a name seen twice is said once", []string{"chore(plan): assign short ids", "chore(plan): assign short ids"}, "chore(plan): assign short ids"},
		{"a subject with one word is a verb with no names", []string{"chore(plan): tidy", "chore(plan): new p"}, "chore(plan): tidy; new p"},
		{"no subjects give nothing", nil, ""},
		{"different kinds do not fold", []string{"chore(bug): new x", "chore(scratch): new y"}, ""},
		{"a subject that is not a chore does not fold", []string{"feat: x", "chore(bug): new y"}, ""},
		{"too long ends with a count", long, "chore(scratch): new alpha-one, bravo-two, charlie-three, +4 more"},
	}
	for _, c := range cases {
		got := FoldSubject(c.in)
		if got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
		if len(got) > maxSubject {
			t.Errorf("%s: %d characters, the cap is %d", c.name, len(got), maxSubject)
		}
	}
	// Every name either shows or is counted, at any length.
	for n := 1; n <= len(long); n++ {
		got := FoldSubject(long[:n])
		shown := strings.Count(got, "-")
		more := 0
		if i := strings.LastIndex(got, "+"); i >= 0 {
			if _, err := fmt.Sscan(got[i+1:], &more); err != nil {
				t.Fatalf("%q: %v", got, err)
			}
		}
		if shown+more != n || len(got) > maxSubject {
			t.Errorf("%d subjects gave %q", n, got)
		}
	}
}
