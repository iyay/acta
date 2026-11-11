package cli

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/iyay/acta/internal/telemetry"
)

// cmdTelemetry turns the daily counts on or off, shows what would be sent, or
// sends finished days now. flush prints nothing: it runs in the background
// and a failure must never reach the user.
func cmdTelemetry(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, usageLine("telemetry"))
		return exitBadInput
	}
	s := telemetry.Default()
	switch args[0] {
	case "on":
		return reportErr(s.On(), stderr)
	case "off":
		return reportErr(s.Off(), stderr)
	case "status":
		return telemetryStatus(s, stdout, stderr)
	case "flush":
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		telemetry.Flush(ctx, s, telemetry.NewClient())
		return exitOK
	}
	fmt.Fprintln(stderr, usageLine("telemetry"))
	return exitBadInput
}

func reportErr(err error, stderr io.Writer) int {
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	return exitOK
}

// telemetryStatus prints the state, what decides it, and the exact body each
// stored day would be sent with, so the user can see every byte.
func telemetryStatus(s *telemetry.Store, stdout, stderr io.Writer) int {
	state := "off"
	if s.Enabled() {
		state = "on"
	}
	fmt.Fprintf(stdout, "Telemetry: %s\nSource: %s\n", state, s.Source())
	days, err := s.Days()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	if len(days) == 0 {
		fmt.Fprintln(stdout, "Nothing stored.")
	}
	for _, d := range days {
		body, err := telemetry.Payload(d)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return exitOther
		}
		label := "unsent"
		if d.Day == telemetry.Today() {
			label = "today, not finished"
		}
		fmt.Fprintf(stdout, "%s (%s): %s\n", d.Day, label, body)
	}
	return exitOK
}
