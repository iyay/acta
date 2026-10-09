package update

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Replace swaps the binary at exe for data. It never edits the old binary.
// The new bytes go to a temp file next to it, then one rename puts them in
// place. Rename is all or nothing, so a crash leaves the old or the new
// binary, never a half-written one. Any failure before the rename leaves the
// old binary as it was and removes the temp file.
func Replace(exe string, data []byte) error {
	// If exe is a link, the real file is what must change. The link stays.
	target, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return err
	}
	dir := filepath.Dir(target)

	// The temp file must live in the same dir. A rename across disks is a
	// copy, and a copy can stop half way.
	tmp := filepath.Join(dir, fmt.Sprintf(".acta.%d", os.Getpid()))
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o755)
	if err != nil {
		if errors.Is(err, fs.ErrPermission) {
			return fmt.Errorf("no write access to %s", dir)
		}
		return err
	}

	// Only the rename moves the temp file away, so until then it must go.
	done := false
	defer func() {
		if !done {
			_ = os.Remove(tmp)
		}
	}()

	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	// The umask can strip bits at create time, so set the mode again.
	if err := os.Chmod(tmp, 0o755); err != nil {
		return err
	}
	if err := os.Rename(tmp, target); err != nil {
		return err
	}
	done = true
	return nil
}
