package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/update"
	"github.com/iyay/acta/plugin"
)

// These are variables so a test can swap them.
var (
	spawnCheck       = execCheck
	stderrIsTerminal = isTerminal
)

// updateCachePath sits next to the user config, so a test that moves the
// config with PM_VOICE_FILE moves the cache too.
func updateCachePath() (string, error) {
	p, err := config.UserPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(p), "update.json"), nil
}

// updateCheckOn says whether this run may look at releases at all. Source
// builds have no release to compare with, and users can turn it off.
func updateCheckOn() bool {
	if !isRelease() || os.Getenv("ACTA_NO_UPDATE_CHECK") == "1" {
		return false
	}
	v, _, err := config.ResolveUser()
	return err != nil || v.UpdateCheck != "off"
}

// startUpdateCheck starts a background check when the cache is old. It never
// waits for the network: the child process does that and writes the cache.
func startUpdateCheck(args []string) {
	if len(args) > 0 && args[0] == "update" {
		return
	}
	if !updateCheckOn() {
		return
	}
	path, err := updateCachePath()
	if err != nil || !update.Stale(path, time.Now()) {
		return
	}
	// Touch first, so runs that start right after this one see a fresh file
	// and do not start a second check.
	if err := touch(path); err != nil {
		return
	}
	_ = spawnCheck()
}

func touch(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	now := time.Now()
	if err := os.Chtimes(path, now, now); err == nil {
		return nil
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	return f.Close()
}

// execCheck starts `acta update --check --cache` on its own, with no
// terminal, in a new session so it outlives this run.
func execCheck() error {
	exe, err := exePath()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, "update", "--check", "--cache")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

// printUpdateNotice adds one line on stderr when the cache knows a newer
// release. Hooks, --json and pipes stay clean, because a script or an agent
// reads them. The TUI shows its own badge, so it gets no line either.
func printUpdateNotice(args []string, stderr io.Writer) {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") || args[0] == "hook" || args[0] == "update" || hasJSONFlag(args) {
		return
	}
	if !stderrIsTerminal(stderr) {
		return
	}
	if latest := newerRelease(); latest != "" {
		fmt.Fprintf(stderr, "acta %s available (you have v%s). Run: acta update\n", latest, plugin.Version())
	}
}

// newerRelease is the cached version when it beats this binary, else "".
func newerRelease() string {
	if !updateCheckOn() {
		return ""
	}
	path, err := updateCachePath()
	if err != nil {
		return ""
	}
	latest := update.ReadCache(path)
	if !update.Newer(latest, "v"+plugin.Version()) {
		return ""
	}
	return latest
}

func hasJSONFlag(args []string) bool {
	for _, a := range args {
		if a == "--json" || a == "-json" || strings.HasPrefix(a, "--json=") || strings.HasPrefix(a, "-json=") {
			return true
		}
	}
	return false
}

func isTerminal(w any) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	st, err := f.Stat()
	return err == nil && st.Mode()&os.ModeCharDevice != 0
}
