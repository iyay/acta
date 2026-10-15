//go:build !notui

package cli

import "testing"

func TestVariantFull(t *testing.T) {
	if Variant != "full" {
		t.Errorf("Variant = %q, want full", Variant)
	}
}
