package plugincheck

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// releaseFile reads one file by its path from the repo root. The root is the
// parent of plugin/.
func releaseFile(t *testing.T, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "..", filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// parseYAML turns one repo file into a plain map, so a test reads the same
// keys the tool reads.
func parseYAML(t *testing.T, rel string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := yaml.Unmarshal([]byte(releaseFile(t, rel)), &m); err != nil {
		t.Fatalf("%s: %v", rel, err)
	}
	return m
}

// at walks nested maps by key. It returns nil when a key is missing.
func at(v any, keys ...string) any {
	for _, k := range keys {
		m, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		v = m[k]
	}
	return v
}

// list reads a YAML list as a slice of values.
func list(v any) []any {
	l, _ := v.([]any)
	return l
}

// strs reads a YAML list of plain values as sorted strings.
func strs(v any) []string {
	var out []string
	for _, e := range list(v) {
		s, ok := e.(string)
		if !ok {
			continue
		}
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

func TestReleaseGoreleaserBuilds(t *testing.T) {
	cfg := parseYAML(t, ".goreleaser.yaml")
	if cfg["version"] != 2 {
		t.Errorf("version = %v, want 2", cfg["version"])
	}
	builds := list(cfg["builds"])
	if len(builds) != 1 {
		t.Fatalf("want one build, got %d", len(builds))
	}
	b := builds[0]
	if at(b, "main") != "./cmd/acta" {
		t.Errorf("main = %v", at(b, "main"))
	}
	// A static binary runs on any Linux, with no C library needed.
	if !reflect.DeepEqual(strs(at(b, "env")), []string{"CGO_ENABLED=0"}) {
		t.Errorf("env = %v, want only CGO_ENABLED=0", at(b, "env"))
	}
	if got := strs(at(b, "goos")); !reflect.DeepEqual(got, []string{"darwin", "linux", "windows"}) {
		t.Errorf("goos = %v", got)
	}
	if got := strs(at(b, "goarch")); !reflect.DeepEqual(got, []string{"amd64", "arm64"}) {
		t.Errorf("goarch = %v", got)
	}
	// Ignore rules would drop one of the six pairs.
	if at(b, "ignore") != nil {
		t.Errorf("builds[0].ignore = %v, all six os/arch pairs must build", at(b, "ignore"))
	}
}

func TestReleaseGoreleaserAssets(t *testing.T) {
	cfg := parseYAML(t, ".goreleaser.yaml")
	archives := list(cfg["archives"])
	if len(archives) != 1 {
		t.Fatalf("want one archive rule, got %d", len(archives))
	}
	a := archives[0]
	// No version in the name, so the latest/download URL never changes.
	if at(a, "name_template") != "acta_{{ .Os }}_{{ .Arch }}" {
		t.Errorf("archive name_template = %v", at(a, "name_template"))
	}
	if strings.Contains(releaseFile(t, ".goreleaser.yaml"), "Version") {
		t.Error(".goreleaser.yaml mentions Version, asset names must not carry it")
	}
	overrides := list(at(a, "format_overrides"))
	if len(overrides) != 1 || at(overrides[0], "goos") != "windows" ||
		!reflect.DeepEqual(strs(at(overrides[0], "formats")), []string{"zip"}) {
		t.Errorf("format_overrides = %v, want windows zip", at(a, "format_overrides"))
	}
	if at(cfg, "checksum", "name_template") != "checksums.txt" {
		t.Errorf("checksum name_template = %v", at(cfg, "checksum", "name_template"))
	}
	if at(cfg, "checksum", "algorithm") != "sha256" {
		t.Errorf("checksum algorithm = %v", at(cfg, "checksum", "algorithm"))
	}
	var globs []string
	for _, e := range list(at(cfg, "release", "extra_files")) {
		g, _ := at(e, "glob").(string)
		globs = append(globs, g)
	}
	sort.Strings(globs)
	want := []string{"scripts/install.ps1", "scripts/install.sh"}
	if !reflect.DeepEqual(globs, want) {
		t.Errorf("release.extra_files = %v, want %v", globs, want)
	}
	for _, g := range want {
		if !exists(filepath.Join(pluginRoot(t), "..", filepath.FromSlash(g))) {
			t.Errorf("%s is listed as a release asset but does not exist", g)
		}
	}
}

// steps reads the steps of one job as maps.
func steps(wf map[string]any, job string) []any {
	return list(at(wf, "jobs", job, "steps"))
}

// stepIndex is the place of the first step that matches, or -1.
func stepIndex(steps []any, match func(step any) bool) int {
	for i, s := range steps {
		if match(s) {
			return i
		}
	}
	return -1
}

func runOf(step any) string {
	s, _ := at(step, "run").(string)
	return s
}

func TestReleaseWorkflowTagCheck(t *testing.T) {
	wf := parseYAML(t, ".github/workflows/release.yml")
	if !reflect.DeepEqual(strs(at(wf, "on", "push", "tags")), []string{"v*"}) {
		t.Errorf("trigger tags = %v, want v*", at(wf, "on", "push", "tags"))
	}
	if at(wf, "jobs", "release", "permissions", "contents") != "write" {
		t.Error("release job needs permissions: contents: write")
	}
	ss := steps(wf, "release")
	check := stepIndex(ss, func(s any) bool {
		r := runOf(s)
		return strings.Contains(r, "GITHUB_REF_NAME") &&
			strings.Contains(r, "jq -r .version plugin/.claude-plugin/plugin.json")
	})
	goreleaser := stepIndex(ss, func(s any) bool {
		u, _ := at(s, "uses").(string)
		return strings.HasPrefix(u, "goreleaser/goreleaser-action@")
	})
	if check < 0 {
		t.Fatal("no step compares GITHUB_REF_NAME with the plugin.json version")
	}
	if goreleaser < 0 {
		t.Fatal("no goreleaser step")
	}
	// The check has to stop the job before anything is published.
	if check > goreleaser {
		t.Errorf("tag check is step %d, goreleaser is step %d, check must come first", check, goreleaser)
	}
	r := runOf(ss[check])
	if !strings.Contains(r, `"v$`) || !strings.Contains(r, "exit 1") {
		t.Errorf("tag check must compare against v plus the version and exit 1:\n%s", r)
	}
	// A skipped or forgiven check would let a wrong tag through.
	if at(ss[check], "if") != nil || at(ss[check], "continue-on-error") != nil {
		t.Error("tag check must always run and always fail the job")
	}
	if at(ss[goreleaser], "with", "args") != "release --clean" {
		t.Errorf("goreleaser args = %v", at(ss[goreleaser], "with", "args"))
	}
	if at(ss[goreleaser], "env", "GITHUB_TOKEN") != "${{ secrets.GITHUB_TOKEN }}" {
		t.Errorf("goreleaser GITHUB_TOKEN = %v", at(ss[goreleaser], "env", "GITHUB_TOKEN"))
	}
	// Goreleaser reads the tag history, so a shallow clone breaks it.
	if co := stepIndex(ss, func(s any) bool {
		u, _ := at(s, "uses").(string)
		return strings.HasPrefix(u, "actions/checkout@") && at(s, "with", "fetch-depth") == 0
	}); co < 0 {
		t.Error("checkout step needs fetch-depth: 0")
	}
	if tests := stepIndex(ss, func(s any) bool { return strings.Contains(runOf(s), "go test ./...") }); tests < 0 || tests > goreleaser {
		t.Errorf("go test step is %d, goreleaser is %d, tests must run first", tests, goreleaser)
	}
}

func TestReleaseCI(t *testing.T) {
	wf := parseYAML(t, ".github/workflows/ci.yml")
	// "push:" with no options reads as a key with an empty value.
	if on, _ := at(wf, "on").(map[string]any); on == nil {
		t.Errorf("ci must run on push, on = %v", at(wf, "on"))
	} else if _, ok := on["push"]; !ok {
		t.Errorf("ci must run on push, on = %v", on)
	}
	test := steps(wf, "test")
	if at(wf, "jobs", "test", "runs-on") != "ubuntu-latest" {
		t.Errorf("test job runs on %v", at(wf, "jobs", "test", "runs-on"))
	}
	for _, want := range []string{"go vet ./...", "gofmt -l .", "go test ./..."} {
		if stepIndex(test, func(s any) bool { return strings.Contains(runOf(s), want) }) < 0 {
			t.Errorf("test job has no step that runs %q", want)
		}
	}
	if at(wf, "jobs", "windows-hooks", "runs-on") != "windows-latest" {
		t.Errorf("windows-hooks runs on %v", at(wf, "jobs", "windows-hooks", "runs-on"))
	}
	if at(wf, "jobs", "windows-hooks", "defaults", "run", "shell") != "bash" {
		t.Error("windows-hooks needs defaults.run.shell: bash")
	}
	win := steps(wf, "windows-hooks")
	hooks := stepIndex(win, func(s any) bool {
		return strings.Contains(runOf(s), "plugin/hooks/") && strings.Contains(runOf(s), `"tool_name":"Bash"`)
	})
	if hooks < 0 {
		t.Error("windows-hooks has no step that feeds sample JSON to plugin/hooks/ scripts")
	} else if !strings.Contains(runOf(win[hooks]), "exit 1") {
		t.Error("the hooks step must fail the job when a hook exits non-zero")
	}
	parse := stepIndex(win, func(s any) bool {
		return strings.Contains(runOf(s), "scripts/install.ps1") && strings.Contains(runOf(s), "Parser]::ParseFile")
	})
	if parse < 0 {
		t.Error("windows-hooks has no step that parses scripts/install.ps1")
	}
}

func TestReleaseInstallPS1(t *testing.T) {
	ps := releaseFile(t, "scripts/install.ps1")
	for _, want := range []string{
		"ACTA_DOWNLOAD_URL", "ACTA_VERSION", "ACTA_INSTALL_DIR",
		"https://github.com/iyay/acta/releases", "latest/download", "download/",
		"LOCALAPPDATA", "PROCESSOR_ARCHITECTURE", "checksums.txt",
		"Get-FileHash", "SHA256", "Expand-Archive", "acta.exe", "setup",
	} {
		if !strings.Contains(ps, want) {
			t.Errorf("install.ps1 does not mention %q", want)
		}
	}
	// The script never edits the user PATH; it only tells the user.
	if strings.Contains(ps, "SetEnvironmentVariable") {
		t.Error("install.ps1 must not change the user PATH")
	}
}
