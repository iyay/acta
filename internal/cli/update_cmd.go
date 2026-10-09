package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"

	"github.com/iyay/acta/internal/setup"
	"github.com/iyay/acta/internal/update"
	"github.com/iyay/acta/plugin"
)

const (
	updateUsage         = "usage: acta update [--check]"
	defaultDownloadBase = "https://github.com/iyay/acta/releases"
	refreshFailedMsg    = "plugin refresh failed: run acta setup"
)

// These are variables so a test can swap them. Each test puts the old value
// back when it ends.
var (
	isRelease  = update.IsRelease
	exePath    = os.Executable
	newClient  = update.NewClient
	runRefresh = execRefresh
)

// cmdUpdate swaps the running binary for the latest release, then has the
// new binary refresh the plugin copies. It exits 0 only when both are done.
func cmdUpdate(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.SetOutput(stderr)
	check := fs.Bool("check", false, "show the current and latest version, change nothing")
	// Hidden: the old binary runs the new one with this flag, so the new
	// plugin text is what gets unpacked.
	refresh := fs.Bool("refresh-plugin", false, "")
	if err := fs.Parse(args); err != nil || fs.NArg() != 0 {
		fmt.Fprintln(stderr, updateUsage)
		return exitBadInput
	}
	if *refresh {
		return cmdRefreshPlugin(stderr)
	}
	if !isRelease() {
		fmt.Fprintln(stderr, "acta was built from source; rebuild it with go install")
		return exitBadInput
	}
	return updateBinary(*check, stdout, stderr)
}

// updateBinary is the download and swap part of cmdUpdate.
func updateBinary(check bool, stdout, stderr io.Writer) int {
	base := os.Getenv("ACTA_DOWNLOAD_URL")
	if base == "" {
		base = defaultDownloadBase
	}
	client := newClient()
	latest, err := update.Latest(client, base)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitBadInput
	}
	current := "v" + plugin.Version()
	if check {
		fmt.Fprintf(stdout, "current %s, latest %s\n", current, latest)
		return exitOK
	}
	if latest == current {
		fmt.Fprintf(stdout, "acta %s is the latest\n", current)
		return exitOK
	}
	data, err := update.Fetch(client, base, latest, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitBadInput
	}
	exe, err := exePath()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitBadInput
	}
	if err := update.Replace(exe, data); err != nil {
		fmt.Fprintln(stderr, err)
		return exitBadInput
	}
	if err := runRefresh(exe); err != nil {
		// A failed child already printed the message on the shared stderr.
		// Any other error means the child never ran, so say it here.
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			fmt.Fprintln(stderr, refreshFailedMsg)
		}
		return exitBadInput
	}
	fmt.Fprintf(stdout, "acta updated from %s to %s\n", current, latest)
	return exitOK
}

// execRefresh starts the new binary at exe to refresh the plugin. A new
// process is needed because the running code still holds the old plugin.
func execRefresh(exe string) error {
	cmd := exec.Command(exe, "update", "--refresh-plugin")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// cmdRefreshPlugin unpacks the plugin that ships inside this binary. When
// claude is installed it also asks claude to take the new copy, because
// claude keeps its own. omp reads the unpacked folder and needs nothing.
func cmdRefreshPlugin(stderr io.Writer) int {
	if err := refreshPlugin(); err != nil {
		fmt.Fprintln(stderr, refreshFailedMsg)
		return exitBadInput
	}
	return exitOK
}

func refreshPlugin() error {
	if _, err := setup.ExtractPlugin(plugin.Files, homeDir()); err != nil {
		return err
	}
	claude, err := exec.LookPath("claude")
	if err != nil {
		return nil
	}
	for _, argv := range [][]string{
		{"plugin", "marketplace", "update", "acta-local"},
		// Without a terminal claude cannot ask to confirm, so --yes is needed.
		{"plugin", "update", "acta@acta-local", "--yes"},
	} {
		if err := exec.Command(claude, argv...).Run(); err != nil {
			return err
		}
	}
	return nil
}
