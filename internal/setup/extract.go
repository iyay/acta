package setup

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ExtractPlugin writes the plugin tree to <home>/.acta/plugin and returns
// that folder. It builds the new tree in a temp folder inside <home>/.acta
// first, then swaps it in, so an older tree is fully replaced and a failed
// extract leaves the old one alone.
func ExtractPlugin(src fs.FS, home string) (string, error) {
	if home == "" {
		return "", errors.New("no home folder to extract the plugin into")
	}
	base := filepath.Join(home, ".acta")
	final := filepath.Join(base, "plugin")
	if err := os.MkdirAll(base, 0o755); err != nil {
		return "", err
	}
	tmp, err := os.MkdirTemp(base, "plugin-new-")
	if err != nil {
		return "", err
	}
	if err := writeTree(src, tmp); err != nil {
		_ = os.RemoveAll(tmp)
		return "", err
	}
	// Move the old tree aside, so a failed swap can put it back.
	aside := ""
	if _, err := os.Lstat(final); err == nil {
		aside = tmp + "-old"
		if err := os.Rename(final, aside); err != nil {
			_ = os.RemoveAll(tmp)
			return "", err
		}
	}
	if err := os.Rename(tmp, final); err != nil {
		if aside != "" {
			_ = os.Rename(aside, final)
		}
		_ = os.RemoveAll(tmp)
		return "", err
	}
	if aside != "" {
		_ = os.RemoveAll(aside)
	}
	return final, nil
}

// writeTree copies every file of src under root. A name that could step
// outside root is refused.
func writeTree(src fs.FS, root string) error {
	return fs.WalkDir(src, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !fs.ValidPath(p) || strings.Contains(p, "..") {
			return fmt.Errorf("unsafe plugin path %q", p)
		}
		out := filepath.Join(root, filepath.FromSlash(p))
		if d.IsDir() {
			return os.MkdirAll(out, 0o755)
		}
		data, err := fs.ReadFile(src, p)
		if err != nil {
			return err
		}
		mode := fs.FileMode(0o644)
		if isScript(p) {
			mode = 0o755
		}
		return os.WriteFile(out, data, mode)
	})
}

// isScript says which files must stay executable. The embed drops mode
// bits, so the rule is by path: hook scripts have no extension, and shell
// helpers end in .sh.
func isScript(p string) bool {
	if strings.HasSuffix(p, ".sh") {
		return true
	}
	if path.Dir(p) != "hooks" {
		return false
	}
	return path.Ext(p) == ""
}
