//go:build !notui

package cli

import "testing"

func TestVariantFull(t *testing.T) {
	if Variant != "full" {
		t.Errorf("Variant = %q, want full", Variant)
	}
}

// What this build must call itself in acta --version and acta doctor.
const (
	wantVersionWord = "full"
	wantDoctorBuild = "full"
)
