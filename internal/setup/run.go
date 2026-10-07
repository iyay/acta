package setup

import (
	"fmt"
	"io"
	"strings"

	"github.com/iyay/acta/internal/config"
)

// Runner runs one harness install command. Tests use a fake, so they never
// call the real claude or omp.
type Runner interface {
	Run(argv []string) error
}

// Apply carries out every action from Plan in order: it saves the user
// config after Validate, runs installs through r, writes the block, and
// prints hints. A failed install prints its command so the user can run it
// by hand, then the rest carries on. It returns the first hard error: a
// bad config value or an unknown action kind.
func Apply(actions []Action, r Runner, out io.Writer) error {
	for _, a := range actions {
		switch a.Kind {
		case ActionConfig:
			if err := a.User.Validate(); err != nil {
				return err
			}
			if err := config.SaveUser(a.User); err != nil {
				return err
			}
			fmt.Fprintln(out, "Saved your answers.")
		case ActionInstall:
			for _, argv := range a.Argv {
				if err := r.Run(argv); err != nil {
					fmt.Fprintf(out, "Could not run `%s`: run it by hand.\n", strings.Join(argv, " "))
				} else {
					fmt.Fprintf(out, "Ran `%s`.\n", strings.Join(argv, " "))
				}
			}
		case ActionPrint:
			fmt.Fprintln(out, a.Path)
		case ActionBlock:
			if err := WriteBlock(a.Path); err != nil {
				return err
			}
			fmt.Fprintf(out, "Wrote the acta block to %s.\n", a.Path)
		default:
			return fmt.Errorf("unknown setup action %q", a.Kind)
		}
	}
	return nil
}
