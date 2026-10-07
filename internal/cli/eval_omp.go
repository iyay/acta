package cli

import (
	"flag"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"

	"github.com/iyay/acta/internal/evalomp"
)

// exitCaseFailed is what claude plugin eval also returns when a case fails,
// so scripts treat both runners the same.
const exitCaseFailed = 1

const evalOmpUsage = "usage: acta eval-omp [--eval-dir <dir>] [--case <glob>] [plugin-dir]"

// cmdEvalOmp runs the plugin's eval cases through omp. Each case costs model
// quota, so no test runs it with the real omp.
func cmdEvalOmp(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("eval-omp", flag.ContinueOnError)
	fs.SetOutput(stderr)
	only := fs.String("case", "", "run only the cases whose name matches this glob")
	evalDir := fs.String("eval-dir", "evals", "eval case folder below the plugin dir")
	pos, err := parseMixed(fs, args)
	if err != nil || len(pos) > 1 {
		fmt.Fprintln(stderr, evalOmpUsage)
		return exitBadInput
	}
	plugin := "plugin"
	if len(pos) == 1 {
		plugin = pos[0]
	}
	omp, err := exec.LookPath("omp")
	if err != nil {
		fmt.Fprintln(stderr, "acta eval-omp: omp is not on PATH; install omp first")
		return exitOther
	}
	abs, err := filepath.Abs(plugin)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	skills, err := evalomp.SkillNames(abs)
	if err != nil {
		fmt.Fprintf(stderr, "acta eval-omp: %s is not a plugin folder: %v\n", plugin, err)
		return exitOther
	}
	cases, err := evalomp.LoadCases(filepath.Join(abs, *evalDir))
	if err != nil {
		fmt.Fprintf(stderr, "acta eval-omp: %v\n", err)
		return exitOther
	}
	o := evalomp.Options{Omp: omp, PluginDir: abs, Skills: skills}
	if evalomp.RunAll(cases, o, *only, evalomp.OmpJudge(omp), stdout) {
		return exitCaseFailed
	}
	return exitOK
}
