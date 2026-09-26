package board

import "strings"

// Prefix gives the short ID prefix of a kind. A plan file is always PLAN,
// even when the board shows it as a story.
func Prefix(k Kind, plan bool) string {
	switch {
	case plan:
		return "PLAN"
	case k == KindBug:
		return "BUG"
	default:
		return "SPEC"
	}
}

// IsHash says if s is a hash: 4 characters, a lower-case letter first, then
// lower-case letters or digits. The letter first keeps it apart from numbers.
func IsHash(s string) bool {
	if len(s) != 4 || s[0] < 'a' || s[0] > 'z' {
		return false
	}
	for _, c := range s[1:] {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}

// setIDs copies id and hash from frontmatter. A value in the wrong shape is
// a problem, not an ID, so a typo cannot steal another item's name. Calling
// it again replaces its own old verdict, so a plan first read as a story and
// then as a plan keeps only the plan reading.
func setIDs(it *Item, prefix string, doc Doc) {
	it.ShortID, it.Hash = "", ""
	kept := it.Problems[:0]
	for _, pr := range it.Problems {
		if !strings.HasPrefix(pr, "bad id ") && !strings.HasPrefix(pr, "bad hash ") {
			kept = append(kept, pr)
		}
	}
	it.Problems = kept
	if id := field(doc.Front, "id"); id != "" {
		n, ok := strings.CutPrefix(id, prefix+"-")
		if ok && n != "" && strings.Trim(n, "0123456789") == "" && n[0] != '0' {
			it.ShortID = id
		} else {
			it.Problems = append(it.Problems, "bad id "+id)
		}
	}
	if h := field(doc.Front, "hash"); h != "" {
		if IsHash(h) {
			it.Hash = prefix + "-" + h
		} else {
			it.Problems = append(it.Problems, "bad hash "+h)
		}
	}
}

// addAlias maps a short ID to an item. Two items with one short ID both get a
// problem line, and the first one keeps the name.
func (b *Board) addAlias(key string, it *Item, what string) {
	if key == "" {
		return
	}
	k := strings.ToLower(key)
	if old := b.alias[k]; old != nil && old != it {
		old.Problems = append(old.Problems, "duplicate "+what+" "+key)
		it.Problems = append(it.Problems, "duplicate "+what+" "+key)
		return
	}
	b.alias[k] = it
}

func (b *Board) aliasItem(it *Item) {
	b.addAlias(it.ShortID, it, "id")
	b.addAlias(it.Hash, it, "hash")
}

// aliasPlan names a plan whose tasks belong to a spec or bug. Its short IDs
// lead to that item, which is where the plan's tasks show.
func (b *Board) aliasPlan(p planFile, holder *Item) {
	tmp := &Item{}
	setIDs(tmp, "PLAN", p.doc)
	holder.Problems = append(holder.Problems, tmp.Problems...)
	b.addAlias(tmp.ShortID, holder, "id")
	b.addAlias(tmp.Hash, holder, "hash")
}

// aliasTask gives a task its plan's IDs plus its own number.
func (b *Board) aliasTask(t *Item, planID, planHash string) {
	if planID != "" {
		t.ShortID = planID + "." + t.TaskNum
	}
	if planHash != "" {
		t.Hash = planHash + "." + t.TaskNum
	}
	b.aliasItem(t)
}
