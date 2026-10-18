package board

// markWaiting marks each plan that names a plan in its after list that is not
// done yet. An id the board does not know also counts as not done, so a typo
// never makes a plan look free to start. acta ready gives the full reason.
func (b *Board) markWaiting() {
	for _, it := range b.Items {
		if it.Kind != KindPlan {
			continue
		}
		for _, id := range it.After {
			if dep := b.Get(id); dep == nil || dep.Status != "done" {
				it.Waiting = true
				break
			}
		}
	}
}
