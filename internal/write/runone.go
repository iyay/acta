package write

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
)

// runOneName is the one lock file every acta run-one of this user shares.
// Runs fight over the machine's CPU and memory, not over one repo, so the
// name holds no repo path.
const runOneName = "run-one.lock"

// HoldRunOne takes the machine-wide run-one lock and waits as long as it
// takes. waiting runs once, the first time the lock is busy, so the caller
// can say why nothing is happening yet. The kernel drops the lock when the
// process ends, even in a crash, so a dead holder never blocks anyone.
func HoldRunOne(waiting func()) (func(), error) {
	dir, err := lockDir()
	if err != nil {
		return nil, err
	}
	// The cache folder and the acta folder under it are not there yet on a
	// first run, so make them closed to everyone else.
	if err := os.MkdirAll(filepath.Dir(dir), 0o700); err != nil {
		return nil, err
	}
	if err := safeDir(dir); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filepath.Join(dir, runOneName), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if errors.Is(err, syscall.EWOULDBLOCK) {
		waiting()
		err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
	}
	if err != nil {
		f.Close()
		return nil, err
	}
	return func() {
		syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
	}, nil
}
