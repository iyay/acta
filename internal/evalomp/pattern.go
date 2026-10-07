package evalomp

import (
	"fmt"
	"strings"
	"time"

	"github.com/dlclark/regexp2"
)

// matchTimeout stops a pattern that backtracks without end, so one bad
// grader cannot hang the whole run.
const matchTimeout = 2 * time.Second

// CompilePattern reads a grader pattern the way JavaScript does. The suite
// writes these patterns for JavaScript, and Go's own regexp has no lookaround.
func CompilePattern(pattern, flags string) (*regexp2.Regexp, error) {
	opts := regexp2.RegexOptions(regexp2.ECMAScript)
	if strings.Contains(flags, "i") {
		opts |= regexp2.IgnoreCase
	}
	re, err := regexp2.Compile(pattern, opts)
	if err != nil {
		return nil, err
	}
	re.MatchTimeout = matchTimeout
	return re, nil
}

// PatternProblems lists every grader pattern that does not compile.
func PatternProblems(cases []Case) []string {
	var out []string
	for _, c := range cases {
		for _, g := range c.Graders {
			pat, flags := "", ""
			switch g.Type {
			case "regex":
				pat, flags = g.Pattern, g.Flags
			case "tool_used":
				pat = g.InputMatch
			default:
				continue
			}
			if _, err := CompilePattern(pat, flags); err != nil {
				out = append(out, fmt.Sprintf("case %s, grader %s: %v", c.Name, g.Name, err))
			}
		}
	}
	return out
}
