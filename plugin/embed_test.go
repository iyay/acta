package plugin

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// shippedOnDisk says whether a plugin file belongs in the binary: evals
// and Go files stay out, everything else goes in.
func shippedOnDisk(rel string) bool {
	top := strings.SplitN(rel, "/", 2)[0]
	if top == "evals" || top == "evals-routing" {
		return false
	}
	return !strings.HasSuffix(rel, ".go")
}

func TestEmbedMatchesDisk(t *testing.T) {
	onDisk := map[string]bool{}
	err := filepath.WalkDir(".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel := filepath.ToSlash(p)
		if shippedOnDisk(rel) {
			onDisk[rel] = true
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	embedded := map[string]bool{}
	err = fs.WalkDir(Files, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		embedded[p] = true
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for p := range onDisk {
		if !embedded[p] {
			t.Errorf("%s is on disk but not embedded", p)
		}
	}
	for p := range embedded {
		if !onDisk[p] {
			t.Errorf("%s is embedded but not shipped", p)
		}
		if strings.HasPrefix(p, "evals") || strings.HasSuffix(p, ".go") {
			t.Errorf("%s must not be embedded", p)
		}
	}
	for _, must := range []string{".claude-plugin/plugin.json", ".claude-plugin/marketplace.json", "hooks/hooks.json", "package.json", "README.md", "NOTICE"} {
		if !embedded[must] {
			t.Errorf("%s missing from the embed", must)
		}
	}
}

func TestEmbedBytesMatchDisk(t *testing.T) {
	got, err := Files.ReadFile("package.json")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("package.json")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatal("embedded package.json differs from disk")
	}
}
