package flakerun

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
)

// tui is a full import path, so the tests also prove the FLAKY line cuts
// the module path from go.mod off the front.
const tui = "github.com/iyay/acta/internal/tui"

// failA is a first run where TestA fails.
const failA = `{"Action":"run","Package":"github.com/iyay/acta/internal/tui","Test":"TestA"}
{"Action":"output","Package":"github.com/iyay/acta/internal/tui","Test":"TestA","Output":"--- FAIL: TestA (0.00s)\n"}
{"Action":"fail","Package":"github.com/iyay/acta/internal/tui","Test":"TestA"}
{"Action":"output","Package":"github.com/iyay/acta/internal/tui","Output":"FAIL\n"}
{"Action":"fail","Package":"github.com/iyay/acta/internal/tui"}
`

// passA is a rerun where TestA passes.
const passA = `{"Action":"run","Package":"github.com/iyay/acta/internal/tui","Test":"TestA"}
{"Action":"output","Package":"github.com/iyay/acta/internal/tui","Test":"TestA","Output":"--- PASS: TestA (0.00s)\n"}
{"Action":"pass","Package":"github.com/iyay/acta/internal/tui","Test":"TestA"}
{"Action":"output","Package":"github.com/iyay/acta/internal/tui","Output":"ok  \tgithub.com/iyay/acta/internal/tui\t0.1s\n"}
{"Action":"pass","Package":"github.com/iyay/acta/internal/tui"}
`

// failSub is a first run where the subtest TestA/b fails, which also fails
// its parent TestA.
const failSub = `{"Action":"run","Package":"github.com/iyay/acta/internal/tui","Test":"TestA"}
{"Action":"run","Package":"github.com/iyay/acta/internal/tui","Test":"TestA/b"}
{"Action":"output","Package":"github.com/iyay/acta/internal/tui","Test":"TestA/b","Output":"    --- FAIL: TestA/b (0.00s)\n"}
{"Action":"fail","Package":"github.com/iyay/acta/internal/tui","Test":"TestA/b"}
{"Action":"output","Package":"github.com/iyay/acta/internal/tui","Test":"TestA","Output":"--- FAIL: TestA (0.00s)\n"}
{"Action":"fail","Package":"github.com/iyay/acta/internal/tui","Test":"TestA"}
{"Action":"output","Package":"github.com/iyay/acta/internal/tui","Output":"FAIL\n"}
{"Action":"fail","Package":"github.com/iyay/acta/internal/tui"}
`

// buildFail is a first run where internal/cli does not compile. The tests
// put it after failA, so a failed test in another package must not make
// the run rerun anything either.
const buildFail = `{"ImportPath":"github.com/iyay/acta/internal/cli","Action":"build-output","Output":"internal/cli/a.go:3:1: syntax error\n"}
{"ImportPath":"github.com/iyay/acta/internal/cli","Action":"build-fail"}
{"Action":"output","Package":"github.com/iyay/acta/internal/cli","Output":"FAIL\tgithub.com/iyay/acta/internal/cli [build failed]\n"}
{"Action":"fail","Package":"github.com/iyay/acta/internal/cli","FailedBuild":"github.com/iyay/acta/internal/cli"}
`

// timeout is a first run where the test binary panics on a timeout and
// only the package is marked failed, with no test name.
const timeout = `{"Action":"run","Package":"github.com/iyay/acta/internal/tui","Test":"TestA"}
{"Action":"output","Package":"github.com/iyay/acta/internal/tui","Test":"TestA","Output":"panic: test timed out after 1s\n"}
{"Action":"output","Package":"github.com/iyay/acta/internal/tui","Output":"FAIL\tgithub.com/iyay/acta/internal/tui\t1.0s\n"}
{"Action":"fail","Package":"github.com/iyay/acta/internal/tui"}
`

// failMany is a first run where six tests fail, one more than a run may
// rerun.
var failMany = func() string {
	var b strings.Builder
	for _, name := range []string{"TestA", "TestB", "TestC", "TestD", "TestE", "TestF"} {
		fmt.Fprintf(&b, `{"Action":"fail","Package":"github.com/iyay/acta/internal/tui","Test":%q}`+"\n", name)
	}
	b.WriteString(`{"Action":"fail","Package":"github.com/iyay/acta/internal/tui"}` + "\n")
	return b.String()
}()

// call is one canned answer of the fake run.
type call struct {
	stream string
	err    error
}

// fakeRun hands out the canned answers in order and keeps every argument
// list it got, so a test can check what each rerun asked go test for.
func fakeRun(calls []call) (func([]string) (io.Reader, error), *[][]string) {
	var got [][]string
	run := func(args []string) (io.Reader, error) {
		got = append(got, args)
		if len(got) > len(calls) {
			return strings.NewReader(""), io.ErrUnexpectedEOF
		}
		c := calls[len(got)-1]
		return strings.NewReader(c.stream), c.err
	}
	return run, &got
}

func TestRun(t *testing.T) {
	first := []string{"-tags", "notui", "-short", "./..."}
	rerunA := []string{"-tags", "notui", "-short", "-run", "^(TestA)$", "-count=1", tui}
	cases := []struct {
		name      string
		calls     []call
		wantCode  int
		wantFlaky []string
		wantArgs  [][]string
		// wantText is a line the output must hold, or "" for none.
		wantText string
	}{
		{
			name:      "fails then passes",
			calls:     []call{{failA, errExit}, {passA, nil}},
			wantCode:  0,
			wantFlaky: []string{"FLAKY: internal/tui.TestA (failed 1 of 2)"},
			wantArgs:  [][]string{first, rerunA},
		},
		{
			name:     "fails 3 of 3",
			calls:    []call{{failA, errExit}, {failA, errExit}, {failA, errExit}},
			wantCode: 1,
			wantArgs: [][]string{first, rerunA, rerunA},
			wantText: "FAIL: internal/tui.TestA (failed 3 of 3)",
		},
		{
			name:      "failed subtest reruns its top-level test",
			calls:     []call{{failSub, errExit}, {passA, nil}},
			wantCode:  0,
			wantFlaky: []string{"FLAKY: internal/tui.TestA (failed 1 of 2)"},
			wantArgs:  [][]string{first, rerunA},
		},
		{
			name:     "build failure reruns nothing",
			calls:    []call{{failA + buildFail, errExit}, {passA, nil}},
			wantCode: 1,
			wantArgs: [][]string{first},
		},
		{
			name:     "panic or timeout with no test name reruns nothing",
			calls:    []call{{timeout, errExit}},
			wantCode: 1,
			wantArgs: [][]string{first},
		},
		{
			name:     "more than 5 failed tests rerun nothing",
			calls:    []call{{failMany, errExit}},
			wantCode: 1,
			wantArgs: [][]string{first},
		},
		{
			name:     "go test fails with no failed test or package",
			calls:    []call{{"go: unknown flag -x\n", errExit}},
			wantCode: 1,
			wantArgs: [][]string{first},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			run, got := fakeRun(tc.calls)
			var out strings.Builder
			code := Run(first, run, &out)
			if code != tc.wantCode {
				t.Errorf("exit code = %d, want %d\noutput:\n%s", code, tc.wantCode, out.String())
			}
			if flaky := flakyLines(out.String()); !reflect.DeepEqual(flaky, tc.wantFlaky) {
				t.Errorf("FLAKY lines = %q, want %q", flaky, tc.wantFlaky)
			}
			if !reflect.DeepEqual(*got, tc.wantArgs) {
				t.Errorf("run args = %q, want %q", *got, tc.wantArgs)
			}
			if tc.wantText != "" && !strings.Contains(out.String(), tc.wantText+"\n") {
				t.Errorf("output lacks %q\noutput:\n%s", tc.wantText, out.String())
			}
		})
	}
}

// TestRunPrintsPlainText checks the log reads like plain go test output, from
// the first run and the rerun, and never shows the JSON events.
func TestRunPrintsPlainText(t *testing.T) {
	run, _ := fakeRun([]call{{"go: warning: plain line\n" + failA, errExit}, {passA, nil}})
	var out strings.Builder
	Run([]string{"./..."}, run, &out)
	for _, want := range []string{"go: warning: plain line\n", "--- FAIL: TestA (0.00s)\n", "--- PASS: TestA (0.00s)\n"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output lacks %q\noutput:\n%s", want, out.String())
		}
	}
	if strings.Contains(out.String(), `"Action"`) {
		t.Errorf("output shows JSON events:\n%s", out.String())
	}
}

// errExit stands for go test ending with a non-zero exit code.
var errExit = errors.New("exit status 1")

// flakyLines picks the FLAKY lines out of the printed output.
func flakyLines(out string) []string {
	var lines []string
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "FLAKY:") {
			lines = append(lines, line)
		}
	}
	return lines
}
