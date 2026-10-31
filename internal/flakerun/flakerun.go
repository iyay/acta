// Package flakerun runs go test, then runs each failed test again on its
// own. A test that passes on a second try is a flake, not a real red, so it
// gets a FLAKY line instead of failing the whole run.
package flakerun

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// maxRuns is how many times a failed test runs in all: the first run plus
// two reruns.
const maxRuns = 3

// maxFailed is the most failed tests one run may rerun. More than that
// points at something broken for real, not at a few flaky tests.
const maxFailed = 5

// event is one line of go test -json output. Only the fields this package
// reads are kept.
type event struct {
	Action  string
	Package string
	Test    string
	Output  string
}

// failure is one failed top-level test in one package.
type failure struct {
	pkg  string
	test string
}

// Run runs go test with args through run, prints its output as plain text
// to out, and reruns each failed test on its own. run gets a go test
// argument list and returns the -json event stream; its error is non-nil
// when go test could not start or ended with a non-zero exit code. The
// result is the exit code: 0 only when every failed test passed on a rerun.
func Run(args []string, run func(args []string) (io.Reader, error), out io.Writer) int {
	events, err := runOnce(args, run, out)
	failed := failedTests(events)
	module := modulePath()
	if pkg := failedWithNoTest(events, failed); pkg != "" {
		fmt.Fprintf(out, "flakerun: %s failed with no failed test (build failure, panic or timeout), so nothing reruns\n",
			strings.TrimPrefix(pkg, module+"/"))
		return 1
	}
	if pkg := crashed(events); pkg != "" {
		fmt.Fprintf(out, "flakerun: %s crashed before all its tests ended (panic, timeout or os.Exit), so nothing reruns\n",
			strings.TrimPrefix(pkg, module+"/"))
		return 1
	}
	if len(failed) == 0 {
		if err != nil {
			// go test failed before any test or package could, such as on
			// a bad flag. Its own message is already printed above.
			fmt.Fprintf(out, "flakerun: go test failed: %v\n", err)
			return 1
		}
		return 0
	}
	if len(failed) > maxFailed {
		fmt.Fprintf(out, "flakerun: %d tests failed, more than %d, so nothing reruns\n", len(failed), maxFailed)
		return 1
	}
	code := 0
	for _, f := range failed {
		if !rerun(f, flagsOf(args), run, module, out) {
			code = 1
		}
	}
	return code
}

// rerun runs one failed test alone until it passes or has run maxRuns times
// in all, and prints how it went. It reports whether the test passed.
func rerun(f failure, flags []string, run func([]string) (io.Reader, error), module string, out io.Writer) bool {
	args := append(append([]string{}, flags...),
		"-run", "^("+regexp.QuoteMeta(f.test)+")$", "-count=1", f.pkg)
	name := strings.TrimPrefix(f.pkg, module+"/") + "." + f.test
	for n := 2; n <= maxRuns; n++ {
		events, err := runOnce(args, run, out)
		if err == nil && passed(events, f.test) {
			fmt.Fprintf(out, "FLAKY: %s (failed %d of %d)\n", name, n-1, n)
			return true
		}
	}
	fmt.Fprintf(out, "FAIL: %s (failed %d of %d)\n", name, maxRuns, maxRuns)
	return false
}

// runOnce calls run once and reads the whole event stream it returns.
func runOnce(args []string, run func([]string) (io.Reader, error), out io.Writer) ([]event, error) {
	r, err := run(args)
	if r == nil {
		return nil, err
	}
	return readEvents(r, out), err
}

// readEvents reads the stream line by line and prints each event's Output,
// so the log reads like plain go test output. A line that is not an event,
// such as an error from the go command itself, prints as it is, so no
// message is lost.
func readEvents(r io.Reader, out io.Writer) []event {
	var events []event
	br := bufio.NewReader(r)
	for {
		line, readErr := br.ReadString('\n')
		if line != "" {
			var e event
			if json.Unmarshal([]byte(line), &e) == nil {
				events = append(events, e)
				io.WriteString(out, e.Output)
			} else {
				io.WriteString(out, line)
			}
		}
		if readErr != nil {
			return events
		}
	}
}

// failedTests lists the failed top-level tests in the order they failed. A
// failed subtest counts as its top-level test, since -run cannot pick a
// subtest out without its parent running too.
func failedTests(events []event) []failure {
	var failed []failure
	seen := map[failure]bool{}
	for _, e := range events {
		if e.Action != "fail" || e.Test == "" {
			continue
		}
		top, _, _ := strings.Cut(e.Test, "/")
		f := failure{pkg: e.Package, test: top}
		if !seen[f] {
			seen[f] = true
			failed = append(failed, f)
		}
	}
	return failed
}

// failedWithNoTest returns the first package that failed while none of its
// tests failed, or "" when there is none. That is a build failure, or a
// panic or timeout that killed the test binary before it could name a
// test. A rerun cannot pick one test out of that, so nothing reruns.
func failedWithNoTest(events []event, failed []failure) string {
	hasTest := map[string]bool{}
	for _, f := range failed {
		hasTest[f.pkg] = true
	}
	for _, e := range events {
		if e.Action == "fail" && e.Test == "" && !hasTest[e.Package] {
			return e.Package
		}
	}
	return ""
}

// crashed returns the first package whose test binary died, or "" when
// none did. A binary that dies stops every test after the one that killed
// it, so a test that failed before it passing on a rerun proves nothing.
// go test sends no end event for a test that was still running when the
// binary died, and a panic prints a line that starts with "panic: ", even
// when the panicked test got a fail event of its own.
func crashed(events []event) string {
	type key struct{ pkg, test string }
	open := map[key]bool{}
	var started []key
	for _, e := range events {
		if strings.HasPrefix(e.Output, "panic: ") {
			return e.Package
		}
		k := key{e.Package, e.Test}
		switch {
		case e.Test == "":
			// A package event, not a test, so it never starts or ends one.
		case e.Action == "run":
			open[k] = true
			started = append(started, k)
		case e.Action == "pass" || e.Action == "fail" || e.Action == "skip":
			delete(open, k)
		}
	}
	// Walk in start order, so the package named is the first one that died.
	for _, k := range started {
		if open[k] {
			return k.pkg
		}
	}
	return ""
}

// passed reports whether the stream shows test passing.
func passed(events []event, test string) bool {
	for _, e := range events {
		if e.Action == "pass" && e.Test == test {
			return true
		}
	}
	return false
}

// flagsOf drops the package names from the first run's arguments, so a
// rerun keeps the same flags but names one package. scripts/test only ever
// names packages as ./ paths, so that is the rule used here too.
func flagsOf(args []string) []string {
	var flags []string
	for _, a := range args {
		if !strings.HasPrefix(a, "./") {
			flags = append(flags, a)
		}
	}
	return flags
}

// modulePath reads the module line of the nearest go.mod at or above the
// working folder. go test runs inside the module, so the go.mod is always
// there. With no go.mod it returns "", and names keep their full path.
func modulePath() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if rest, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
					return strings.TrimSpace(rest)
				}
			}
			return ""
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
