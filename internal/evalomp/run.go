package evalomp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// Options says which omp to run and which plugin to load into it.
type Options struct {
	Omp       string
	PluginDir string
	Skills    []string
}

// ErrTimeout marks a case that ran past its timeout_seconds.
var ErrTimeout = errors.New("timeout")

// judgeTimeout caps one judge call. A judge only reads one reply.
const judgeTimeout = 120 * time.Second

// SkillNames lists the plugin's skill folders. omp's --skills filter takes
// these names, which keeps acta's skills in and every other skill out.
func SkillNames(pluginDir string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(pluginDir, "skills"))
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// RunCase runs one case in a new throwaway folder and returns what it left.
// The caller removes the folder's parent when done with it.
func RunCase(c Case, o Options) (Workspace, error) {
	base, err := os.MkdirTemp("", "acta-eval-omp-")
	if err != nil {
		return Workspace{}, err
	}
	home, work := filepath.Join(base, "home"), filepath.Join(base, "work")
	for _, d := range []string{home, work} {
		if err := os.Mkdir(d, 0o755); err != nil {
			return Workspace{Dir: work}, err
		}
	}
	if c.Scaffold != "" {
		// The scaffold writes the voice file into HOME. A throwaway home keeps
		// it off the user's real config, which omp itself still needs.
		cmd := exec.Command("bash", c.Scaffold)
		cmd.Dir = work
		cmd.Env = withEnv(os.Environ(), "HOME", home)
		if out, err := cmd.CombinedOutput(); err != nil {
			return Workspace{Dir: work}, fmt.Errorf("scaffold: %v: %s", err, strings.TrimSpace(string(out)))
		}
	}
	before, err := ListFiles(work)
	if err != nil {
		return Workspace{Dir: work}, err
	}
	w := Workspace{Dir: work, Before: before}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.TimeoutSeconds)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, o.Omp, ompArgs(c.Prompt, o)...)
	cmd.Dir = work
	cmd.Env = withEnv(os.Environ(), "PM_VOICE_FILE", filepath.Join(home, ".acta", "config.yaml"))
	// A killed omp can leave a child holding the pipes open. Stop waiting for
	// them soon after, so one stuck case cannot hang the whole run.
	cmd.WaitDelay = 5 * time.Second
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	runErr := cmd.Run()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return w, ErrTimeout
	}
	if runErr != nil {
		return w, fmt.Errorf("omp: %v: %s", runErr, strings.TrimSpace(stderr.String()))
	}
	w.Result, err = ParseStream(&stdout)
	return w, err
}

func ompArgs(prompt string, o Options) []string {
	return []string{
		"-p", "--mode", "json", "--no-session", "--no-extensions", "--no-rules",
		"--skills=" + strings.Join(o.Skills, ","),
		"--plugin-dir", o.PluginDir,
		"-e", filepath.Join(o.PluginDir, "omp", "index.ts"),
		prompt,
	}
}

// withEnv sets one variable, dropping any copy already in env so the new
// value is the one the child sees.
func withEnv(env []string, key, value string) []string {
	out := make([]string, 0, len(env)+1)
	for _, kv := range env {
		if !strings.HasPrefix(kv, key+"=") {
			out = append(out, kv)
		}
	}
	return append(out, key+"="+value)
}

// OmpJudge grades with omp's own default model. No plugin, skill or rule is
// loaded: the judge only reads the rubric and the reply.
func OmpJudge(omp string) Judge {
	return func(prompt string) (string, error) {
		ctx, cancel := context.WithTimeout(context.Background(), judgeTimeout)
		defer cancel()
		cmd := exec.CommandContext(ctx, omp, "-p", "--no-session", "--no-extensions", "--no-rules", "--no-skills", prompt)
		cmd.Dir = os.TempDir()
		out, err := cmd.Output()
		return string(out), err
	}
}

// RunAll runs every case whose name matches the only glob (all when only is
// empty), prints one line per case and a total, and says whether any failed.
func RunAll(cases []Case, o Options, only string, judge Judge, out io.Writer) (failed bool) {
	passed, failedN, skipped := 0, 0, 0
	for _, c := range cases {
		if only != "" {
			if ok, _ := path.Match(only, c.Name); !ok {
				continue
			}
		}
		if c.ClaudeOnly() {
			fmt.Fprintf(out, "SKIP %s (claude-only)\n", c.Name)
			skipped++
			continue
		}
		if why := runOne(c, o, judge); why != "" {
			fmt.Fprintf(out, "FAIL %s: %s\n", c.Name, why)
			failedN++
			continue
		}
		fmt.Fprintf(out, "PASS %s\n", c.Name)
		passed++
	}
	fmt.Fprintf(out, "%d passed, %d failed, %d skipped\n", passed, failedN, skipped)
	return failedN > 0
}

// runOne gives "" when every grader passed, else what went wrong.
func runOne(c Case, o Options, judge Judge) string {
	w, err := RunCase(c, o)
	if w.Dir != "" {
		defer os.RemoveAll(filepath.Dir(w.Dir))
	}
	if err != nil {
		return err.Error()
	}
	var fails []string
	for _, g := range c.Graders {
		if r := Grade(g, w, judge); !r.Pass {
			fails = append(fails, fmt.Sprintf("%s (%s)", r.Grader, r.Why))
		}
	}
	return strings.Join(fails, "; ")
}
