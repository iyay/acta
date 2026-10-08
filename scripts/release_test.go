package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The three version files, with the same line shape as the real ones: the
// same indent and the trailing comma.
var releaseFiles = map[string]string{
	"plugin/.claude-plugin/plugin.json": "{\n  \"name\": \"acta\",\n  \"description\": \"x\",\n  \"version\": \"0.1.43\",\n  \"author\": \"x\"\n}\n",
	"plugin/.claude-plugin/marketplace.json": "{\n  \"name\": \"acta\",\n  \"plugins\": [\n    {\n      \"name\": \"acta\",\n      \"source\": \"./\",\n" +
		"      \"description\": \"x\",\n      \"category\": \"dev\",\n      \"version\": \"0.1.43\",\n      \"tags\": []\n    }\n  ]\n}\n",
	"plugin/package.json": "{\n  \"name\": \"acta\",\n  \"version\": \"0.1.43\",\n  \"private\": true\n}\n",
}

// releaseGates says what the fake gates do: the exit code of scripts/test and
// of go vet, and what gofmt -l prints.
type releaseGates struct {
	test   int
	vet    int
	fmtOut string
}

// releaseRepo is a temp repo on main, with a copy of scripts/release, a fake
// scripts/test, the version files at 0.1.43, and a bare "origin" that must
// stay empty because the script never pushes.
type releaseRepo struct {
	t      *testing.T
	dir    string
	bin    string
	origin string
}

func (r *releaseRepo) env() []string {
	env := make([]string, 0, len(os.Environ())+4)
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "GIT_CONFIG_") && !strings.HasPrefix(e, "PATH=") {
			env = append(env, e)
		}
	}
	// No global git config, so a signing key or hook path on this machine
	// cannot reach the temp repo.
	return append(env, "PATH="+r.bin+":/usr/bin:/bin", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1", "HOME="+r.dir)
}

func (r *releaseRepo) git(args ...string) string {
	r.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = r.dir
	cmd.Env = r.env()
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.t.Fatalf("git %v: %v %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func (r *releaseRepo) write(rel, body string, mode os.FileMode) {
	r.t.Helper()
	path := filepath.Join(r.dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		r.t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), mode); err != nil {
		r.t.Fatal(err)
	}
}

func newReleaseRepo(t *testing.T, g releaseGates) *releaseRepo {
	t.Helper()
	root := t.TempDir()
	r := &releaseRepo{t: t, dir: filepath.Join(root, "repo"), bin: filepath.Join(root, "bin"), origin: filepath.Join(root, "origin.git")}
	for _, d := range []string{r.dir, r.bin, r.origin} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	script, err := os.ReadFile("release")
	if err != nil {
		t.Fatal(err)
	}
	r.write("scripts/release", string(script), 0o755)
	r.write("scripts/test", "#!/bin/sh\nexit "+itoa(g.test)+"\n", 0o755)
	// Fake go and gofmt: the gates run for real only in the real repo.
	r.write("../bin/go", "#!/bin/sh\nexit "+itoa(g.vet)+"\n", 0o755)
	r.write("../bin/gofmt", "#!/bin/sh\nprintf '%s' '"+g.fmtOut+"'\n", 0o755)
	for rel, body := range releaseFiles {
		r.write(rel, body, 0o644)
	}
	if out, err := exec.Command("git", "init", "-q", "--bare", r.origin).CombinedOutput(); err != nil {
		t.Fatalf("git init bare: %v %s", err, out)
	}
	r.git("init", "-q", "-b", "main")
	r.git("config", "user.name", "Test")
	r.git("config", "user.email", "test@example.com")
	r.git("remote", "add", "origin", r.origin)
	r.git("add", ".")
	r.git("commit", "-q", "-m", "init")
	return r
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	return "1"
}

func (r *releaseRepo) run(args ...string) (string, error) {
	cmd := exec.Command(filepath.Join(r.dir, "scripts", "release"), args...)
	cmd.Dir = r.dir
	cmd.Env = r.env()
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (r *releaseRepo) versions() map[string]string {
	got := map[string]string{}
	for rel := range releaseFiles {
		b, err := os.ReadFile(filepath.Join(r.dir, rel))
		if err != nil {
			r.t.Fatal(err)
		}
		got[rel] = string(b)
	}
	return got
}

// assertNoPush fails when anything reached origin: no branch and no tag.
func (r *releaseRepo) assertNoPush() {
	r.t.Helper()
	out, err := exec.Command("git", "--git-dir", r.origin, "for-each-ref").CombinedOutput()
	if err != nil {
		r.t.Fatal(err)
	}
	if strings.TrimSpace(string(out)) != "" {
		r.t.Errorf("origin got refs, the script must never push: %s", out)
	}
}

func TestReleaseBumpsTagsAndPrintsPush(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct{ arg, want string }{
		{"patch", "0.1.44"},
		{"minor", "0.2.0"},
		{"1.2.3", "1.2.3"},
	} {
		t.Run(tc.arg, func(t *testing.T) {
			t.Parallel()
			r := newReleaseRepo(t, releaseGates{})
			before := r.git("rev-list", "--count", "HEAD")

			out, err := r.run(tc.arg)
			if err != nil {
				t.Fatalf("release %s: %v %s", tc.arg, err, out)
			}
			for rel, body := range r.versions() {
				want := strings.Replace(releaseFiles[rel], "0.1.43", tc.want, 1)
				if body != want {
					t.Errorf("%s:\n got %q\nwant %q", rel, body, want)
				}
			}
			if after := r.git("rev-list", "--count", "HEAD"); after == before {
				t.Error("no new commit")
			}
			if got := r.git("rev-list", "--count", "HEAD"); got != "2" {
				t.Errorf("commit count: got %s, want 2", got)
			}
			if got := r.git("log", "-1", "--format=%s"); got != "chore(release): v"+tc.want {
				t.Errorf("subject: got %q", got)
			}
			if got := r.git("show", "--name-only", "--format=", "HEAD"); len(strings.Split(got, "\n")) != 3 {
				t.Errorf("commit should hold exactly the three version files, got %q", got)
			}
			if got := r.git("cat-file", "-t", "refs/tags/v"+tc.want); got != "tag" {
				t.Errorf("tag type: got %q, want an annotated tag", got)
			}
			if got := r.git("tag", "-l"); got != "v"+tc.want {
				t.Errorf("tags: got %q", got)
			}
			if got := r.git("status", "--porcelain"); got != "" {
				t.Errorf("tree not clean after release: %q", got)
			}
			if want := "git push origin main v" + tc.want; !strings.Contains(out, want) {
				t.Errorf("output should name %q, got %q", want, out)
			}
			r.assertNoPush()
		})
	}
}

// Every path that must stop leaves no new commit, no tag and the version
// files as they were.
func TestReleaseStopsAndLeavesRepoUntouched(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		gates releaseGates
		arg   string
		setup func(r *releaseRepo)
	}{
		{name: "no argument", arg: ""},
		{name: "bad argument", arg: "banana"},
		{name: "two part version", arg: "1.2"},
		{name: "version with v", arg: "v1.2.3"},
		{name: "other branch", arg: "patch", setup: func(r *releaseRepo) { r.git("checkout", "-q", "-b", "feature") }},
		{name: "dirty tree", arg: "patch", setup: func(r *releaseRepo) { r.write("stray.txt", "x", 0o644) }},
		{name: "dirty version file", arg: "patch", setup: func(r *releaseRepo) {
			r.write("plugin/package.json", strings.Replace(releaseFiles["plugin/package.json"], "private", "other", 1), 0o644)
		}},
		{name: "tag exists", arg: "patch", setup: func(r *releaseRepo) { r.git("tag", "v0.1.44") }},
		{name: "explicit tag exists", arg: "1.2.3", setup: func(r *releaseRepo) { r.git("tag", "v1.2.3") }},
		{name: "red tests", arg: "patch", gates: releaseGates{test: 1}},
		{name: "red vet", arg: "patch", gates: releaseGates{vet: 1}},
		{name: "gofmt lists a file", arg: "patch", gates: releaseGates{fmtOut: "internal/x.go\n"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := newReleaseRepo(t, tc.gates)
			if tc.setup != nil {
				tc.setup(r)
			}
			head := r.git("rev-parse", "HEAD")
			tags := r.git("tag", "-l")
			status := r.git("status", "--porcelain")
			files := r.versions()

			args := []string{}
			if tc.arg != "" {
				args = append(args, tc.arg)
			}
			out, err := r.run(args...)
			if err == nil {
				t.Fatalf("release should fail, got %s", out)
			}
			if got := r.git("rev-parse", "HEAD"); got != head {
				t.Errorf("HEAD moved: %s -> %s", head, got)
			}
			if got := r.git("tag", "-l"); got != tags {
				t.Errorf("tags changed: %q -> %q", tags, got)
			}
			if got := r.git("status", "--porcelain"); got != status {
				t.Errorf("tree changed: %q -> %q", status, got)
			}
			for rel, body := range r.versions() {
				if body != files[rel] {
					t.Errorf("%s changed:\n got %q\nwant %q", rel, body, files[rel])
				}
			}
			if strings.Contains(out, "git push") {
				t.Errorf("a failed release must not print the push line: %q", out)
			}
			r.assertNoPush()
		})
	}
}
