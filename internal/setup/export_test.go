package setup

import (
	"io"

	"github.com/iyay/acta/internal/config"
)

// AnsweredLines gives tests the collapsed line each question would leave
// behind, built from its default answer.
func AnsweredLines(e Env) []string {
	var out []string
	for _, q := range buildFormState(e).questions {
		out = append(out, Collapsed(q.title, q.answer()))
	}
	return out
}

// AskOverridesPicking runs the override step without a terminal: each
// question takes the answer in picks, or keeps its default when the key is
// not there. It returns the same things Ask puts into Answers.
func AskOverridesPicking(u config.User, repoRoot string, out io.Writer, picks map[string]string) ([]config.Override, map[string]string, error) {
	return askOverrides(u, repoRoot, out, func(q *overrideQuestion) error {
		if p, ok := picks[q.override.Key]; ok {
			*q.choice = p
		}
		return nil
	})
}
