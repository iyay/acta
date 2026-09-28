package board

import "strings"

// linkScratch ties each spec that names a scratch item as its parent to that
// item. A spec is where a raw idea ends up, so the item learns it was used.
// dir is the scratch folder this repo named, which is not always "scratch".
func (b *Board) linkScratch(dir string) {
	for _, it := range b.Items {
		if it.Kind != KindStory {
			continue
		}
		want := it.fmParent
		if !strings.HasPrefix(want, "scratch/") && !strings.HasPrefix(want, dir+"/") {
			continue
		}
		target := b.byID[want]
		if target == nil || target.Kind != KindScratch {
			it.Problems = append(it.Problems, "parent "+want+" not found")
			continue
		}
		target.Children = append(target.Children, it.ID)
	}
}

// scratchStatus picks the one status a scratch item shows. A written dropped
// is a choice the user already made, so it wins over a spec link, and a link
// turns a raw idea into specced.
func scratchStatus(written string, linked bool) (status, source string) {
	switch {
	case written == "dropped":
		return "dropped", "frontmatter"
	case linked:
		return "specced", "derived"
	case written != "":
		return written, "frontmatter"
	}
	return "raw", "derived"
}
