package tui

import (
	"testing"
)

// BenchmarkWheelFrame is one frame of trackpad scrolling in the detail box:
// ten notches, one tick, one draw. The wheel always turns down, so every frame
// really scrolls; at the end the box goes back to the top.
func BenchmarkWheelFrame(b *testing.B) {
	t := &testing.T{}
	m := paneModel(t, paneDetail)
	box := scrollBox(m, paneDetail)
	for b.Loop() {
		if m.off[paneDetail] >= m.lastOff(paneDetail) {
			m.off[paneDetail] = 0
		}
		for range 10 {
			m = wheelOnly(m, box.x+1, box.y+2, false)
		}
		m = wheelTick(m)
		m.View()
	}
}
