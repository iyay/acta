//go:build !notui

package cli

import (
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/trees"
	"github.com/iyay/acta/internal/tui"
)

// Variant names this build. The full build holds the TUI board.
const Variant = "full"

// runTUI opens the TUI with live reload. The board on the first frame holds
// only the main tree. The watcher's start brings the full one, with the other
// worktrees, branches and authors. When the watcher cannot start, the TUI
// still opens in manual mode.
var runTUI = func(cfg config.Config, stdout, stderr io.Writer) int {
	b, loader, err := tuiBoards(cfg)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	// Ask the terminal for its background now; asking inside the program
	// fights Bubble Tea for stdin.
	dark := lipgloss.HasDarkBackground()
	m := tui.New(cfg, b, dark).WithTheme(voiceTheme(), dark).WithVersion(tuiVersion()).WithUpdate(newerRelease()).WithLoad(loader)
	// 120fps halves how long a new frame waits to reach the screen.
	opts := []tea.ProgramOption{tea.WithAltScreen(), tea.WithMouseCellMotion(), tea.WithFPS(120)}
	tr, traceOpts, traceDone := tuiTrace(os.Getenv("ACTA_TUI_TRACE"), os.Stdout, stderr)
	defer traceDone()
	m = m.WithTrace(tr)
	p := tea.NewProgram(m, append(opts, traceOpts...)...)
	load := m.Reload()
	dirs := func() []string { return append(tui.WatchDirs(cfg), trees.WatchDirs(cfg, tui.WatchDirs)...) }
	stop := tui.Start(dirs, load, p.Send)
	defer stop()
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	return exitOK
}
