package setup

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// renameFile is os.Rename. A test swaps it to force a failed swap.
var renameFile = os.Rename

// ExtractPlugin writes the plugin tree to <home>/.acta/plugin and returns
// that folder. It builds the new tree in a temp folder inside <home>/.acta
// first, then swaps it in, so an older tree is fully replaced and a failed
// extract leaves the old one alone.
func ExtractPlugin(src fs.FS, home string) (string, error) {
	// A relative home would unpack under the current folder and wipe an old
	// tree there. Stop before any file call.
	if !filepath.IsAbs(home) {
		return "", fmt.Errorf("home folder %q is not an absolute path", home)
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
		if err := renameFile(final, aside); err != nil {
			_ = os.RemoveAll(tmp)
			return "", err
		}
	}
	if err := renameFile(tmp, final); err != nil {
		if aside != "" {
			if rerr := renameFile(aside, final); rerr != nil {
				// Keep the old tree: it now sits at aside, so say where.
				return "", fmt.Errorf("%w; could not restore the old plugin, it is at %s: %v", err, aside, rerr)
			}
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
