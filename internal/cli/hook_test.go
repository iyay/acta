package cli

import (
	"os"
	"strings"
	"testing"
)

// The hook reads HERDR_ENV, so the session text can name a herdr tab as the
// extra way to run a second brainstorm. Without it, the text stays silent.
// The phrase belongs to the herdr block only; the skill index has its own line
// about the dispatch skill.
const herdrLine = "this session runs in a herdr tab"

func TestHookSessionStartNamesHerdrOnlyInsideHerdr(t *testing.T) {
	t.Chdir(hookRepo(t))
	t.Setenv("PM_VOICE_FILE", "")

	// Unset, not empty: an unset variable must not count as herdr either.
	saved, had := os.LookupEnv("HERDR_ENV")
	if err := os.Unsetenv("HERDR_ENV"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if had {
			_ = os.Setenv("HERDR_ENV", saved)
		}
	})

	if _, out, _ := runHook(t, "session-start", ""); strings.Contains(out, herdrLine) {
		t.Errorf("session start outside herdr offers a herdr tab:\n%s", out)
	}
	t.Setenv("HERDR_ENV", "1")
	if _, out, _ := runHook(t, "session-start", ""); !strings.Contains(out, herdrLine) {
		t.Errorf("herdr session start missing the herdr tab sentence:\n%s", out)
	}
	t.Setenv("HERDR_ENV", "0")
	if _, out, _ := runHook(t, "session-start", ""); strings.Contains(out, herdrLine) {
		t.Errorf("HERDR_ENV=0 counts as herdr:\n%s", out)
	}
}
