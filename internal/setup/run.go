package setup

import (
	"fmt"
	"io"
	"path/filepath"

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
// by hand, then the rest carries on. A repo action edits .acta.yaml and
// never commits it, so the run ends with one line when the file changed. It
// returns the first hard error: a bad config value or an unknown action kind.
func Apply(actions []Action, r Runner, out io.Writer) error {
	repoChanged := false
	for _, a := range actions {
		switch a.Kind {
		case ActionConfig:
			if err := a.User.Validate(); err != nil {
				return err
			}
			if err := config.SaveUser(a.User); err != nil {
				return err
			}
			fmt.Fprint(out, RailLine("✓ config saved"))
		case ActionInstall:
			for _, argv := range a.Argv {
				fmt.Fprint(out, InstallLine(a.Harness, argv, r.Run(argv)))
			}
		case ActionInstalled:
			fmt.Fprint(out, RailLine("✓ "+a.Harness+"  already installed"))
		case ActionPrint:
			fmt.Fprint(out, PrintLines(a))
		case ActionBlock:
			if err := WriteBlock(a.Path); err != nil {
				return err
			}
			fmt.Fprint(out, RailLine("✓ acta block → "+a.Path))
		case ActionRepoSet:
			if _, err := config.SaveRepoUser(filepath.Dir(a.Path), map[string]string{a.Key: a.Value}); err != nil {
				return err
			}
			repoChanged = true
			fmt.Fprint(out, RailLine("✓ .acta.yaml  "+a.Key+" set to "+a.Value))
		case ActionRepoUnset:
			if _, err := config.UnsetRepoUser(filepath.Dir(a.Path), []string{a.Key}); err != nil {
				return err
			}
			repoChanged = true
			fmt.Fprint(out, RailLine("✓ .acta.yaml  "+a.Key+" removed"))
		default:
			return fmt.Errorf("unknown setup action %q", a.Kind)
		}
	}
	if repoChanged {
		fmt.Fprint(out, RailLine(".acta.yaml changed; commit it when the team should get it."))
	}
	return nil
}
