// Package update holds the pieces of `acta update`: the release marker, the
// download and the in-place binary swap.
package update

// release is set to "true" at link time by goreleaser, and by nothing else.
// A binary without it was built from source.
var release string

// IsRelease says whether goreleaser built this binary. Only release builds
// can update themselves, because only they have a matching archive online.
func IsRelease() bool { return release == "true" }
