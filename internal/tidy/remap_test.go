package tidy

import (
	"strings"
	"testing"
)

func TestRemapShortFullUnknownAndLastCommit(t *testing.T) {
	r := start(t)
	// Rides into the first new commit, so the new hash of a differs from the old one.
	r.commit("chore(plan): early", d(1), "ann", map[string]string{".acta/early.md": "e\n"})
	a := r.commit("feat: a", d(1), "ann", map[string]string{"a.go": "a\n"})
	b := r.commit("feat: b", d(2), "ann", map[string]string{"b.go": "b\n"})
	unknown := "0123456789abcdef0123456789abcdef01234567"
	notes := "short " + a[:7] + " mid " + a[:12] + " full " + a + "\nunknown " + unknown + " and " + unknown[:9] + "\nbig " + strings.Repeat("ab", 32) + "\n"
	r.commit("chore(plan): notes", d(3), "ann", map[string]string{".acta/notes.md": notes})
	// This one lands in the LAST new commit, so its hash cannot be remapped.
	r.commit("chore(plan): last ref", d(4), "ann", map[string]string{".acta/last.md": "b is " + b + "\n"})
	// An empty commit folds into the last one and changes nothing.
	r.commit("fix: oops", d(5), "ann", nil)

	res, err := Run(r.dir, opt("main", "feat"))
	if err != nil {
		t.Fatal(err)
	}
	newA := r.rev(res.Tip + "~1")
	if res.NewCount != 2 {
		t.Fatalf("counts %+v", res)
	}
	got := r.git(nil, "show", res.Tip+":.acta/notes.md")
	want := "short " + newA[:7] + " mid " + newA[:12] + " full " + newA + "\nunknown " + unknown + " and " + unknown[:9] + "\nbig " + strings.Repeat("ab", 32)
	if got != want {
		t.Fatalf("notes:\n%s\nwant:\n%s", got, want)
	}
	// Hashes of the last new commit are left as they were: the commit hash
	// depends on the tree, so the edit cannot name it.
	if r.git(nil, "show", res.Tip+":.acta/last.md") != "b is "+b {
		t.Fatal("hash of last commit changed")
	}
	// Only the planning file differs from the branch tree.
	diff := r.git(nil, "diff", "--name-only", "feat", res.Tip)
	if diff != ".acta/notes.md" {
		t.Fatalf("diff %q", diff)
	}
}

func TestRemapTextUnit(t *testing.T) {
	old := []string{"aaaaaaa1111111111111111111111111111111111", "bbbbbbb2222222222222222222222222222222222"}
	old[0] = old[0][:40]
	old[1] = old[1][:40]
	m := map[string]string{old[0]: strings.Repeat("c", 40)}
	in := "x " + old[0][:7] + " y " + old[1][:7] + " z" + old[0][:8] + "q"
	got := remapText(in, old, m)
	want := "x ccccccc y " + old[1][:7] + " z" + old[0][:8] + "q"
	if got != want {
		t.Fatalf("%q want %q", got, want)
	}
}
