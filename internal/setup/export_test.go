package setup

// AnsweredLines gives tests the collapsed line each question would leave
// behind, built from its default answer.
func AnsweredLines(e Env) []string {
	var out []string
	for _, q := range buildFormState(e).questions {
		out = append(out, Collapsed(q.title, q.answer()))
	}
	return out
}
