package write

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMain points the lock folder at a temp folder for the whole package, so
// no test of this package ever writes a lock into the real cache folder. A
// test that wants the real cache path sets lockRoot to "" itself.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "acta-write-test-")
	if err != nil {
		panic(err)
	}
	lockRoot = filepath.Join(dir, "locks")
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}
