package plugincheck

import "testing"

// The cleanup skill must keep every step of the spec, in order, so no write
// happens before the user has seen the proposals and said yes.
func TestSkillCleanup(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "cleanup",
		MaxLines: 60,
		Must: []string{
			"acta cleanup --fix", "acta cleanup --json", "read-only subagent", "in parallel",
			"keep, done, dropped, wontfix or fixed", "file:line", "commit sha", "spec id", "one-line reason",
			"never write", "Weak evidence becomes keep", "one message per kind", "current channel",
			"10 numbered items", "ok", "numbers to skip", "under the plan", "acta cleanup apply",
			"one batch", "Keep writes nothing",
		},
		MustNot: []string{"superpowers:", "rm -rf", "git push"},
	})
}
