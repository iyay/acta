package cli

import (
	"fmt"
	"io"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/iyay/acta/internal/tui"
)

// tuiTrace turns on the TUI trace when path names a file. It gives the
// tracer, the program option that notes each screen write, and a func that
// writes the summary and closes the file when the TUI is done. A file that
// cannot open only costs the trace: the TUI still runs.
func tuiTrace(path string, stdout *os.File, stderr io.Writer) (*tui.Tracer, []tea.ProgramOption, func()) {
	if path == "" {
		return nil, nil, func() {}
	}
	f, err := os.Create(path)
	if err != nil {
		fmt.Fprintf(stderr, "ACTA_TUI_TRACE: trace is off: %v\n", err)
		return nil, nil, func() {}
	}
	tr := tui.NewTracer(f, time.Now)
	// The TUI ends once, but a second call must not try to write a summary
	// into a file that is already shut.
	closed := false
	done := func() {
		if closed {
			return
		}
		closed = true
		if err := tr.Close(); err != nil {
			fmt.Fprintf(stderr, "ACTA_TUI_TRACE: summary not written: %v\n", err)
		}
		f.Close()
	}
	return tr, []tea.ProgramOption{tea.WithOutput(tr.Output(stdout))}, done
}
