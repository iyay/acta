package evalomp

import (
	"fmt"
	"time"

	"github.com/dlclark/regexp2"
)

// matchTimeout stops a pattern that backtracks without end, so one bad
// grader cannot hang the whole run.
var matchTimeout = 2 * time.Second

// CompilePattern reads a grader pattern the way JavaScript does. The suite
// writes these patterns for JavaScript, and Go's own regexp has no lookaround.
func CompilePattern(pattern, flags string) (*regexp2.Regexp, error) {
	opts := regexp2.RegexOptions(regexp2.ECMAScript)
	// A flag we ignore would change what a grader means without a word, so
	// every flag is applied or refused.
	for _, f := range flags {
		switch f {
		case 'i':
			opts |= regexp2.IgnoreCase
		case 'm':
			opts |= regexp2.Multiline
		default:
			return nil, fmt.Errorf("unsupported regex flag %q: only i and m are supported", string(f))
		}
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
