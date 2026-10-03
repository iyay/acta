package tui

import (
	"strings"
	"testing"
)

// The bar reads Scratchpad, Debt and Activity, so no screen the TUI draws
// still shows the old plurals.
func TestTopTabsReadScratchpadDebtActivity(t *testing.T) {
	t.Parallel()

	frames := []Model{
		sized(newModel(t), 120, 40),
		press(sized(newModel(t), 120, 40), "?"),
	}
	for _, m := range frames {
		v := plain(m.View())
		for _, want := range []string{"1 Scratchpad", "3 Debt", "6 Activity"} {
			if !strings.Contains(v, want) {
				t.Errorf("frame is missing %q:\n%s", want, v)
			}
		}
		for _, gone := range []string{"Scratches", "Debts", "Activities"} {
			if strings.Contains(v, gone) {
				t.Errorf("frame still shows %q:\n%s", gone, v)
			}
		}
	}
}
