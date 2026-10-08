package scripts

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// The fake acta writes its args to a file, so a test can see the setup call.
const fakeActa = "#!/bin/sh\necho \"$@\" > \"$ACTA_ARGS_FILE\"\n"

// installRig holds one fake release server and the folders one run of
// install.sh uses.
type installRig struct {
	srv   *httptest.Server
	dir   string // ACTA_INSTALL_DIR
	home  string
	bin   string // fake uname lives here
	args  string // file the fake acta writes to
	mu    sync.Mutex
	hits  []string
	uname [2]string
}

func tarGz(t *testing.T, name, body string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(body))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte(body)); err != nil {
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

func sum(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// newRig serves files under both the latest and the pinned v9.9.9 path.
// A file missing from the map answers 404.
func newRig(t *testing.T, files map[string][]byte, osName, arch string) *installRig {
	t.Helper()
	r := &installRig{
		dir:   filepath.Join(t.TempDir(), "bin"),
		home:  t.TempDir(),
		bin:   t.TempDir(),
		args:  filepath.Join(t.TempDir(), "args"),
		uname: [2]string{osName, arch},
	}
	r.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.mu.Lock()
		r.hits = append(r.hits, req.URL.Path)
		r.mu.Unlock()
		for _, prefix := range []string{"/latest/download/", "/download/v9.9.9/"} {
			if name, ok := strings.CutPrefix(req.URL.Path, prefix); ok {
				if body, found := files[name]; found {
					_, _ = w.Write(body)
					return
				}
			}
		}
		http.NotFound(w, req)
	}))
	t.Cleanup(r.srv.Close)
	fake := "#!/bin/sh\ncase \"$1\" in\n-s) echo " + osName + " ;;\n-m) echo " + arch + " ;;\n*) echo " + osName + " ;;\nesac\n"
	if err := os.WriteFile(filepath.Join(r.bin, "uname"), []byte(fake), 0o755); err != nil {
		t.Fatal(err)
	}
	return r
}

// goodRig serves one valid archive and its checksums line for the asset name.
func goodRig(t *testing.T, asset, osName, arch string) *installRig {
	t.Helper()
	archive := tarGz(t, "acta", fakeActa)
	sums := sum(archive) + "  " + asset + "\n" + strings.Repeat("0", 64) + "  other.tar.gz\n"
	return newRig(t, map[string][]byte{asset: archive, "checksums.txt": []byte(sums)}, osName, arch)
}

// run executes install.sh and returns its output and error. PATH holds only
// the fake uname plus the system dirs, as in runScript.
func (r *installRig) run(t *testing.T, pathExtra string, env ...string) (string, error) {
	t.Helper()
	script, err := filepath.Abs("install.sh")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("/bin/sh", script)
	cmd.Dir = t.TempDir()
	keep := make([]string, 0, len(os.Environ())+8)
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "GIT_CONFIG_") && !strings.HasPrefix(e, "PATH=") &&
			!strings.HasPrefix(e, "HOME=") && !strings.HasPrefix(e, "ACTA_") {
			keep = append(keep, e)
		}
	}
	path := r.bin + ":/usr/bin:/bin"
	if pathExtra != "" {
		path = r.bin + ":" + pathExtra + ":/usr/bin:/bin"
	}
	keep = append(keep,
		"HOME="+r.home,
		"ACTA_DOWNLOAD_URL="+r.srv.URL,
		"ACTA_INSTALL_DIR="+r.dir,
		"ACTA_ARGS_FILE="+r.args,
		"PATH="+path,
	)
	cmd.Env = append(keep, env...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (r *installRig) installed() bool {
	_, err := os.Stat(filepath.Join(r.dir, "acta"))
	return err == nil
}

func (r *installRig) sawPath(p string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, h := range r.hits {
		if h == p {
			return true
		}
	}
	return false
}

// setupSeen is true when the fake acta got `setup`, or the script told the
// user to run it. A terminal decides which one happens, so both count.
func (r *installRig) setupSeen(out string) bool {
	b, err := os.ReadFile(r.args)
	if err == nil && strings.TrimSpace(string(b)) == "setup" {
		return true
	}
	return strings.Contains(out, "run: acta setup")
}

func TestInstallPutsActaInDirWhenHashMatches(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct{ osName, arch, asset string }{
		{"Darwin", "arm64", "acta_darwin_arm64.tar.gz"},
		{"Darwin", "x86_64", "acta_darwin_amd64.tar.gz"},
		{"Linux", "x86_64", "acta_linux_amd64.tar.gz"},
		{"Linux", "amd64", "acta_linux_amd64.tar.gz"},
		{"Linux", "aarch64", "acta_linux_arm64.tar.gz"},
		{"Linux", "arm64", "acta_linux_arm64.tar.gz"},
	} {
		r := goodRig(t, tc.asset, tc.osName, tc.arch)
		out, err := r.run(t, "")
		if err != nil {
			t.Fatalf("%s %s: %v\n%s", tc.osName, tc.arch, err, out)
		}
		st, statErr := os.Stat(filepath.Join(r.dir, "acta"))
		if statErr != nil {
			t.Fatalf("%s %s: acta not installed: %s", tc.osName, tc.arch, out)
		}
		if st.Mode().Perm() != 0o755 {
			t.Errorf("%s %s: mode %v, want 0755", tc.osName, tc.arch, st.Mode().Perm())
		}
		got, _ := os.ReadFile(filepath.Join(r.dir, "acta"))
		if string(got) != fakeActa {
			t.Errorf("%s %s: installed file is not the archive content", tc.osName, tc.arch)
		}
		if !r.sawPath("/latest/download/" + tc.asset) {
			t.Errorf("%s %s: archive not fetched from the latest path: %v", tc.osName, tc.arch, r.hits)
		}
		if !r.setupSeen(out) {
			t.Errorf("%s %s: no setup call and no hint: %s", tc.osName, tc.arch, out)
		}
	}
}

func TestInstallPinnedVersionUsesPinnedPath(t *testing.T) {
	t.Parallel()

	r := goodRig(t, "acta_linux_amd64.tar.gz", "Linux", "x86_64")
	out, err := r.run(t, "", "ACTA_VERSION=v9.9.9")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if !r.installed() {
		t.Fatalf("acta not installed: %s", out)
	}
	for _, want := range []string{"/download/v9.9.9/acta_linux_amd64.tar.gz", "/download/v9.9.9/checksums.txt"} {
		if !r.sawPath(want) {
			t.Errorf("missing request %s, got %v", want, r.hits)
		}
	}
	if r.sawPath("/latest/download/acta_linux_amd64.tar.gz") {
		t.Errorf("pinned run still asked for latest: %v", r.hits)
	}
}

func TestInstallBadHashLeavesNothing(t *testing.T) {
	t.Parallel()

	asset := "acta_linux_amd64.tar.gz"
	archive := tarGz(t, "acta", fakeActa)
	sums := strings.Repeat("a", 64) + "  " + asset + "\n"
	r := newRig(t, map[string][]byte{asset: archive, "checksums.txt": []byte(sums)}, "Linux", "x86_64")
	out, err := r.run(t, "")
	if err == nil {
		t.Fatalf("want failure on a bad hash, got success: %s", out)
	}
	if r.installed() {
		t.Errorf("acta landed despite a bad hash")
	}
}

func TestInstallMissingChecksumLineLeavesNothing(t *testing.T) {
	t.Parallel()

	asset := "acta_linux_amd64.tar.gz"
	archive := tarGz(t, "acta", fakeActa)
	sums := sum(archive) + "  acta_darwin_arm64.tar.gz\n"
	r := newRig(t, map[string][]byte{asset: archive, "checksums.txt": []byte(sums)}, "Linux", "x86_64")
	out, err := r.run(t, "")
	if err == nil {
		t.Fatalf("want failure on a missing line, got success: %s", out)
	}
	if r.installed() {
		t.Errorf("acta landed without a checksum line")
	}
}

func TestInstallFailedDownloadLeavesNothing(t *testing.T) {
	t.Parallel()

	asset := "acta_linux_amd64.tar.gz"
	archive := tarGz(t, "acta", fakeActa)
	sums := sum(archive) + "  " + asset + "\n"
	for name, files := range map[string]map[string][]byte{
		"archive 404":   {"checksums.txt": []byte(sums)},
		"checksums 404": {asset: archive},
	} {
		r := newRig(t, files, "Linux", "x86_64")
		out, err := r.run(t, "")
		if err == nil {
			t.Errorf("%s: want failure, got success: %s", name, out)
		}
		if r.installed() {
			t.Errorf("%s: acta landed after a failed download", name)
		}
	}
}

func TestInstallUnknownPlatformFails(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct{ osName, arch, want string }{
		{"Plan9", "x86_64", "Plan9"},
		{"Linux", "riscv64", "riscv64"},
	} {
		r := goodRig(t, "acta_linux_amd64.tar.gz", tc.osName, tc.arch)
		out, err := r.run(t, "")
		if err == nil {
			t.Errorf("%s %s: want failure, got success: %s", tc.osName, tc.arch, out)
		}
		if !strings.Contains(out, tc.want) {
			t.Errorf("%s %s: message does not name the platform: %s", tc.osName, tc.arch, out)
		}
		if r.installed() {
			t.Errorf("%s %s: acta landed on an unknown platform", tc.osName, tc.arch)
		}
		if len(r.hits) != 0 {
			t.Errorf("%s %s: downloaded before checking the platform: %v", tc.osName, tc.arch, r.hits)
		}
	}
}

func TestInstallPathHintOnlyWhenDirNotOnPath(t *testing.T) {
	t.Parallel()

	off := goodRig(t, "acta_linux_amd64.tar.gz", "Linux", "x86_64")
	out, err := off.run(t, "")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if !strings.Contains(out, "export PATH=") || !strings.Contains(out, off.dir) {
		t.Errorf("no PATH line naming %s: %s", off.dir, out)
	}

	on := goodRig(t, "acta_linux_amd64.tar.gz", "Linux", "x86_64")
	out, err = on.run(t, on.dir)
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if strings.Contains(out, "export PATH=") {
		t.Errorf("PATH line printed although the dir is on PATH: %s", out)
	}
}

func TestInstallNeverEditsRcFiles(t *testing.T) {
	t.Parallel()

	r := goodRig(t, "acta_linux_amd64.tar.gz", "Linux", "x86_64")
	rc := map[string]string{".bashrc": "# bash\n", ".zshrc": "# zsh\n", ".profile": "# profile\n"}
	for name, body := range rc {
		if err := os.WriteFile(filepath.Join(r.home, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if out, err := r.run(t, ""); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	for name, body := range rc {
		got, _ := os.ReadFile(filepath.Join(r.home, name))
		if string(got) != body {
			t.Errorf("%s changed: %q", name, got)
		}
	}
	entries, _ := os.ReadDir(r.home)
	if len(entries) != len(rc) {
		t.Errorf("home gained files: %v", entries)
	}
}
