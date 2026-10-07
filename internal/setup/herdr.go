package setup

import "os"

// isHerdr says whether this session runs inside a herdr pane. Only then is
// dispatch offered as a build executor: herdr on PATH is not enough, since
// dispatch needs this session's own pane.
func isHerdr() bool {
	return os.Getenv("HERDR_ENV") == "1"
}
