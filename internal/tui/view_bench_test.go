package tui

import (
	"testing"
)

// BenchmarkWheelFrame is one frame of trackpad scrolling in the detail box:
// ten notches, one tick, one draw.
func BenchmarkWheelFrame(b *testing.B) {
	t := &testing.T{}
	m := paneModel(t, paneDetail)
	box := scrollBox(m, paneDetail)
	up := true
	for b.Loop() {
		for range 10 {
			m = wheelOnly(m, box.x+1, box.y+2, up)
			up = !up
		}
		m = wheelTick(m)
		m.View()
	}
}
