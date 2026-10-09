package update

import (
	"os"
	"strings"
	"testing"
)

// The release workflow is the only thing that stamps a binary. If this line
// leaves the config, acta update would refuse every release build.
func TestGoreleaserStampsRelease(t *testing.T) {
	b, err := os.ReadFile("../../.goreleaser.yaml")
	if err != nil {
		t.Fatal(err)
	}
	want := "-X github.com/iyay/acta/internal/update.release=true"
	if !strings.Contains(string(b), want) {
		t.Fatalf(".goreleaser.yaml does not set %q", want)
	}
}

// A test binary is never built by goreleaser, so it must count as a source build.
func TestIsReleaseUnstamped(t *testing.T) {
	if IsRelease() {
		t.Fatal("IsRelease() is true in a binary goreleaser did not stamp")
	}
}
