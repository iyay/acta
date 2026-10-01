package plugincheck

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// watchCall is the call every Go test package has to make, so a test run
// nobody waits for any more stops itself instead of burning the CPU.
const watchCall = "testguard.Watch()"

// selfPackage is the testguard package itself. A package cannot import its own
// name, so it calls Watch() directly and the walk skips it.
const selfPackage = "internal/testguard"

// TestEveryTestPackageWatches walks the whole repo and fails on any folder
// that holds a Go test but never calls testguard.Watch(). The walk is the
// point: a package added later has to fail here too, so no list of folders
// is written down.
func TestEveryTestPackageWatches(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	skip := map[string]bool{
		".git": true, "testdata": true, "node_modules": true, "plugin": true,
	}
	self := filepath.Join(root, "internal", "plugincheck", "testguard_test.go")
	seen := map[string]bool{}
	folders := 0
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path == root || skip[d.Name()] {
				return nil
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			if rel == selfPackage {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		dir := filepath.Dir(path)
		if seen[dir] {
			return nil
		}
		seen[dir] = true
		folders++
		rel, err := filepath.Rel(root, dir)
		if err != nil {
			return err
		}
		files, err := filepath.Glob(filepath.Join(dir, "*_test.go"))
		if err != nil {
			return err
		}
		names := make([]string, 0, len(files))
		for _, f := range files {
			// This checker is not a test package of its own; naming the
			// call here must not stand in for a real TestMain.
			if f == self {
				continue
			}
			names = append(names, filepath.Base(f))
			raw, err := os.ReadFile(f)
			if err != nil {
				return err
			}
			if strings.Contains(string(raw), watchCall) {
				return nil
			}
		}
		if len(names) == 0 {
			return nil
		}
		t.Errorf("%s has no %s in any of its test files: %s", rel, watchCall, strings.Join(names, ", "))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if folders == 0 {
		t.Fatal("the walk found no test folder at all")
	}
}
