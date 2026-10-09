package cli

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/update"
	"github.com/iyay/acta/plugin"
)

var (
	oldBytes = []byte("old acta binary\n")
	newBytes = []byte("new acta binary\n")
)

const newTag = "v99.0.0"

// updateTarGz packs one file named acta, like the release archive does.
func updateTarGz(t *testing.T, body []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	hdr := &tar.Header{Name: "acta", Mode: 0o755, Size: int64(len(body)), Typeflag: tar.TypeReg}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// releaseServer plays GitHub: /latest redirects to tag, and the archive and
// checksums.txt sit under /download/<tag>/. A bad sum makes the check fail.
func releaseServer(t *testing.T, tag string, badSum bool) string {
	t.Helper()
	archive := updateTarGz(t, newBytes)
	sum := sha256.Sum256(archive)
	hexSum := hex.EncodeToString(sum[:])
	if badSum {
		hexSum = strings.Repeat("0", 64)
	}
	name := "acta_" + runtime.GOOS + "_" + runtime.GOARCH + ".tar.gz"
	mux := http.NewServeMux()
	mux.HandleFunc("/rel/latest", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/rel/tag/"+tag, http.StatusFound)
	})
	mux.HandleFunc("/rel/download/"+tag+"/"+name, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(archive)
	})
	mux.HandleFunc("/rel/download/"+tag+"/checksums.txt", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(hexSum + "  " + name + "\n"))
	})
	srv := httptest.NewTLSServer(mux)
	t.Cleanup(srv.Close)
	t.Setenv("ACTA_DOWNLOAD_URL", srv.URL+"/rel")
	newClient = func() *http.Client {
		c := update.NewClient()
		c.Transport = srv.Client().Transport
		return c
	}
	return srv.URL
}

// updateEnv swaps the package vars for one test and puts them back after.
// It returns the path of a fake binary holding the old bytes.
func updateEnv(t *testing.T, release bool) (exe string, refreshed *[]string) {
	t.Helper()
	exe = filepath.Join(t.TempDir(), "acta")
	if err := os.WriteFile(exe, oldBytes, 0o755); err != nil {
		t.Fatal(err)
	}
	oldRelease, oldExe, oldClient, oldRefresh := isRelease, exePath, newClient, runRefresh
	t.Cleanup(func() {
		isRelease, exePath, newClient, runRefresh = oldRelease, oldExe, oldClient, oldRefresh
	})
	isRelease = func() bool { return release }
	exePath = func() (string, error) { return exe, nil }
	calls := []string{}
	runRefresh = func(path string) error {
		calls = append(calls, path)
		return nil
	}
	return exe, &calls
}

func runUpdate(args ...string) (code int, out, errOut string) {
	var o, e strings.Builder
	code = cmdUpdate(args, &o, &e)
	return code, o.String(), e.String()
}

func assertExe(t *testing.T, exe string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("exe holds %q, want %q", got, want)
	}
}

func TestUpdateSourceBuildRefused(t *testing.T) {
	exe, refreshed := updateEnv(t, false)
	t.Setenv("ACTA_DOWNLOAD_URL", "https://127.0.0.1:1/never")
	code, out, errOut := runUpdate()
	if code != 1 || out != "" {
		t.Errorf("code=%d out=%q, want 1 and empty", code, out)
	}
	if want := "acta was built from source; rebuild it with go install"; !strings.Contains(errOut, want) {
		t.Errorf("stderr = %q, want %q", errOut, want)
	}
	assertExe(t, exe, oldBytes)
	if len(*refreshed) != 0 {
		t.Error("refresh ran on a source build")
	}
}

func TestUpdateSourceBuildRefusedWithCheck(t *testing.T) {
	updateEnv(t, false)
	code, _, errOut := runUpdate("--check")
	if code != 1 || !strings.Contains(errOut, "built from source") {
		t.Errorf("code=%d stderr=%q, want 1 and the source message", code, errOut)
	}
}

func TestUpdateAlreadyLatest(t *testing.T) {
	exe, refreshed := updateEnv(t, true)
	releaseServer(t, "v"+plugin.Version(), false)
	code, out, errOut := runUpdate()
	if code != 0 || errOut != "" {
		t.Errorf("code=%d stderr=%q, want 0 and empty", code, errOut)
	}
	if want := "acta v" + plugin.Version() + " is the latest"; strings.TrimSpace(out) != want {
		t.Errorf("stdout = %q, want %q", out, want)
	}
	assertExe(t, exe, oldBytes)
	if len(*refreshed) != 0 {
		t.Error("refresh ran when already latest")
	}
}

func TestUpdateCheckOnly(t *testing.T) {
	exe, refreshed := updateEnv(t, true)
	releaseServer(t, newTag, false)
	code, out, errOut := runUpdate("--check")
	if code != 0 || errOut != "" {
		t.Errorf("code=%d stderr=%q, want 0 and empty", code, errOut)
	}
	if !strings.Contains(out, "v"+plugin.Version()) || !strings.Contains(out, newTag) {
		t.Errorf("stdout = %q, want current and latest version", out)
	}
	assertExe(t, exe, oldBytes)
	if len(*refreshed) != 0 {
		t.Error("refresh ran on --check")
	}
}

func TestUpdateFetchErrorLeavesExe(t *testing.T) {
	exe, refreshed := updateEnv(t, true)
	releaseServer(t, newTag, true)
	code, out, errOut := runUpdate()
	if code != 1 || out != "" {
		t.Errorf("code=%d stdout=%q, want 1 and empty", code, out)
	}
	if !strings.Contains(errOut, "checksum mismatch") {
		t.Errorf("stderr = %q, want checksum mismatch", errOut)
	}
	assertExe(t, exe, oldBytes)
	if len(*refreshed) != 0 {
		t.Error("refresh ran after a fetch error")
	}
}

func TestUpdateNetworkErrorLeavesExe(t *testing.T) {
	exe, _ := updateEnv(t, true)
	t.Setenv("ACTA_DOWNLOAD_URL", "https://127.0.0.1:1/rel")
	code, _, errOut := runUpdate()
	if code != 1 || errOut == "" {
		t.Errorf("code=%d stderr=%q, want 1 and a message", code, errOut)
	}
	assertExe(t, exe, oldBytes)
}

func TestUpdateReplaceErrorLeavesExe(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can write into a read-only dir")
	}
	exe, refreshed := updateEnv(t, true)
	releaseServer(t, newTag, false)
	dir := filepath.Dir(exe)
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
	code, out, errOut := runUpdate()
	if code != 1 || out != "" {
		t.Errorf("code=%d stdout=%q, want 1 and empty", code, out)
	}
	// Replace names the real dir; on macOS the temp dir is a link.
	real, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	if want := "no write access to " + real; !strings.Contains(errOut, want) {
		t.Errorf("stderr = %q, want %q", errOut, want)
	}
	assertExe(t, exe, oldBytes)
	if len(*refreshed) != 0 {
		t.Error("refresh ran after a replace error")
	}
}

func TestUpdateRefreshErrorStillExit1(t *testing.T) {
	exe, _ := updateEnv(t, true)
	releaseServer(t, newTag, false)
	runRefresh = func(string) error { return os.ErrInvalid }
	code, out, errOut := runUpdate()
	if code != 1 {
		t.Errorf("code = %d, want 1", code)
	}
	if want := "plugin refresh failed: run acta setup"; !strings.Contains(errOut, want) {
		t.Errorf("stderr = %q, want %q", errOut, want)
	}
	if strings.Contains(out, "updated") {
		t.Errorf("stdout = %q, must not claim success", out)
	}
	assertExe(t, exe, newBytes)
}

func TestUpdateSuccess(t *testing.T) {
	exe, refreshed := updateEnv(t, true)
	releaseServer(t, newTag, false)
	code, out, errOut := runUpdate()
	if code != 0 || errOut != "" {
		t.Errorf("code=%d stderr=%q, want 0 and empty", code, errOut)
	}
	if want := "acta updated from v" + plugin.Version() + " to " + newTag; !strings.Contains(out, want) {
		t.Errorf("stdout = %q, want %q", out, want)
	}
	assertExe(t, exe, newBytes)
	if len(*refreshed) != 1 || (*refreshed)[0] != exe {
		t.Errorf("refresh calls = %v, want one call with %s", *refreshed, exe)
	}
}

// fakeClaude puts a claude script first on PATH. It logs its args and
// exits with the given code.
func fakeClaude(t *testing.T, exit int) (log string) {
	t.Helper()
	bin := t.TempDir()
	log = filepath.Join(t.TempDir(), "claude.log")
	script := "#!/bin/sh\necho \"$@\" >> '" + log + "'\nexit " + string(rune('0'+exit)) + "\n"
	if err := os.WriteFile(filepath.Join(bin, "claude"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	return log
}

func TestUpdateRefreshPluginRunsClaude(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	log := fakeClaude(t, 0)
	code, _, errOut := runUpdate("--refresh-plugin")
	if code != 0 || errOut != "" {
		t.Fatalf("code=%d stderr=%q, want 0 and empty", code, errOut)
	}
	got, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	want := "plugin marketplace update acta-local\nplugin update acta@acta-local --yes\n"
	if string(got) != want {
		t.Errorf("claude log = %q, want %q", got, want)
	}
	if _, err := os.Stat(filepath.Join(home, ".acta", "plugin")); err != nil {
		t.Errorf("plugin was not extracted: %v", err)
	}
}

func TestUpdateRefreshPluginClaudeFails(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	fakeClaude(t, 1)
	code, _, errOut := runUpdate("--refresh-plugin")
	if code != 1 {
		t.Errorf("code = %d, want 1", code)
	}
	if want := "plugin refresh failed: run acta setup"; !strings.Contains(errOut, want) {
		t.Errorf("stderr = %q, want %q", errOut, want)
	}
}

func TestUpdateRefreshPluginWithoutClaude(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", t.TempDir())
	code, _, errOut := runUpdate("--refresh-plugin")
	if code != 0 || errOut != "" {
		t.Errorf("code=%d stderr=%q, want 0 and empty", code, errOut)
	}
	if _, err := os.Stat(filepath.Join(home, ".acta", "plugin")); err != nil {
		t.Errorf("plugin was not extracted: %v", err)
	}
}

func TestUpdateRefreshPluginExtractFails(t *testing.T) {
	t.Setenv("HOME", "relative-home")
	t.Setenv("PATH", t.TempDir())
	code, _, errOut := runUpdate("--refresh-plugin")
	if code != 1 || !strings.Contains(errOut, "plugin refresh failed: run acta setup") {
		t.Errorf("code=%d stderr=%q, want 1 and the refresh message", code, errOut)
	}
}

func TestUpdateBadFlag(t *testing.T) {
	updateEnv(t, true)
	code, _, errOut := runUpdate("--nope")
	if code != 1 || !strings.Contains(errOut, "usage: acta update") {
		t.Errorf("code=%d stderr=%q, want 1 and usage", code, errOut)
	}
}
