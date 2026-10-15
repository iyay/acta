package scripts

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
)

// The fake acta writes its args to a file, so a test can see the setup call.
const fakeActa = "#!/bin/sh\necho \"$@\" > \"$ACTA_ARGS_FILE\"\n"

// installRig holds one fake release server and the folders one run of
// install.sh uses.
type installRig struct {
	srv   *httptest.Server
	base  string // ACTA_DOWNLOAD_URL; the TLS server unless a test swaps it
	ca    string // PEM file with the fake server certificate
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

// serveFiles answers the latest and pinned paths from the map, else 404.
func serveFiles(files map[string][]byte, w http.ResponseWriter, req *http.Request) {
	for _, prefix := range []string{"/latest/download/", "/download/v9.9.9/"} {
		if name, ok := strings.CutPrefix(req.URL.Path, prefix); ok {
			if body, found := files[name]; found {
				_, _ = w.Write(body)
				return
			}
		}
	}
	http.NotFound(w, req)
}

// newRig serves files under both the latest and the pinned v9.9.9 path.
// A file missing from the map answers 404.
func newRig(t *testing.T, files map[string][]byte, osName, arch string) *installRig {
	t.Helper()
	return newRigWith(t, files, osName, arch, nil)
}

// newRigWith is newRig with an optional handler that replaces the default one.
// The handler must be given here: the server reads it as soon as it starts.
func newRigWith(t *testing.T, files map[string][]byte, osName, arch string, override http.Handler) *installRig {
	t.Helper()
	r := &installRig{
		dir:   filepath.Join(t.TempDir(), "bin"),
		home:  t.TempDir(),
		bin:   t.TempDir(),
		args:  filepath.Join(t.TempDir(), "args"),
		uname: [2]string{osName, arch},
	}
	handler := override
	if handler == nil {
		handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			r.mu.Lock()
			r.hits = append(r.hits, req.URL.Path)
			r.mu.Unlock()
			serveFiles(files, w, req)
		})
	}
	r.srv = httptest.NewTLSServer(handler)
	t.Cleanup(r.srv.Close)
	r.base = r.srv.URL
	// install.sh only trusts https, so curl gets the fake certificate this way.
	r.ca = filepath.Join(t.TempDir(), "ca.pem")
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: r.srv.Certificate().Raw})
	if err := os.WriteFile(r.ca, certPEM, 0o600); err != nil {
		t.Fatal(err)
	}
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
// the fake uname plus the system dirs, as in runScript. The script runs in a
// new session with no terminal, so it never stops to ask a question.
func (r *installRig) run(t *testing.T, pathExtra string, env ...string) (string, error) {
	t.Helper()
	script, err := filepath.Abs("install.sh")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("/bin/sh", script)
	cmd.Env = r.env(pathExtra, env...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Dir = t.TempDir()
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// env is the environment one run of install.sh gets.
func (r *installRig) env(pathExtra string, env ...string) []string {
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
		"ACTA_DOWNLOAD_URL="+r.base,
		"CURL_CA_BUNDLE="+r.ca,
		"ACTA_INSTALL_DIR="+r.dir,
		"ACTA_ARGS_FILE="+r.args,
		"PATH="+path,
	)
	return append(keep, env...)
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

// A plain http base URL must fail before any byte is fetched.
func TestInstallRefusesPlainHTTPBaseURL(t *testing.T) {
	t.Parallel()

	asset := "acta_linux_amd64.tar.gz"
	archive := tarGz(t, "acta", fakeActa)
	files := map[string][]byte{asset: archive, "checksums.txt": []byte(sum(archive) + "  " + asset + "\n")}
	r := newRig(t, files, "Linux", "x86_64")
	var plainHits atomic.Int32
	plain := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		plainHits.Add(1)
		serveFiles(files, w, req)
	}))
	t.Cleanup(plain.Close)
	r.base = plain.URL
	out, err := r.run(t, "")
	if err == nil {
		t.Fatalf("want failure on an http base URL, got success: %s", out)
	}
	if r.installed() {
		t.Errorf("acta landed from an http base URL")
	}
	if n := plainHits.Load(); n != 0 {
		t.Errorf("http server got %d requests, want 0", n)
	}
}

// A redirect from https down to http must stop too, for each download.
func TestInstallRefusesRedirectToPlainHTTP(t *testing.T) {
	t.Parallel()

	asset := "acta_linux_amd64.tar.gz"
	archive := tarGz(t, "acta", fakeActa)
	files := map[string][]byte{asset: archive, "checksums.txt": []byte(sum(archive) + "  " + asset + "\n")}
	for _, name := range []string{asset, "checksums.txt"} {
		var plainHits atomic.Int32
		plain := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			plainHits.Add(1)
			serveFiles(files, w, req)
		}))
		t.Cleanup(plain.Close)
		r := newRigWith(t, files, "Linux", "x86_64", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if strings.HasSuffix(req.URL.Path, "/"+name) {
				http.Redirect(w, req, plain.URL+req.URL.Path, http.StatusFound)
				return
			}
			serveFiles(files, w, req)
		}))
		out, err := r.run(t, "")
		if err == nil {
			t.Errorf("%s: want failure on a redirect to http, got success: %s", name, out)
		}
		if r.installed() {
			t.Errorf("%s: acta landed after a redirect to http", name)
		}
		if n := plainHits.Load(); n != 0 {
			t.Errorf("%s: http server got %d requests, want 0", name, n)
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

// A version typed without the leading v still has to hit the v tag path.
func TestInstallPinnedVersionWithoutVGetsV(t *testing.T) {
	t.Parallel()

	r := goodRig(t, "acta_linux_amd64.tar.gz", "Linux", "x86_64")
	out, err := r.run(t, "", "ACTA_VERSION=9.9.9")
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

// bothRig serves the full and the acta-cli archive for linux/amd64, so a
// test can see which one install.sh asked for. The fake uname must win over
// the real one; a Linux rig makes that fail loudly on a Mac too.
func bothRig(t *testing.T) *installRig {
	t.Helper()
	archive := tarGz(t, "acta", fakeActa)
	sums := sum(archive) + "  acta_linux_amd64.tar.gz\n" + sum(archive) + "  acta-cli_linux_amd64.tar.gz\n"
	return newRig(t, map[string][]byte{
		"acta_linux_amd64.tar.gz":     archive,
		"acta-cli_linux_amd64.tar.gz": archive,
		"checksums.txt":               []byte(sums),
	}, "Linux", "x86_64")
}

const (
	fullAsset = "/latest/download/acta_linux_amd64.tar.gz"
	cliAsset  = "/latest/download/acta-cli_linux_amd64.tar.gz"
)

// assetIs fails unless install.sh fetched want and not the other archive.
func (r *installRig) assetIs(t *testing.T, name, want string) {
	t.Helper()
	other := fullAsset
	if want == fullAsset {
		other = cliAsset
	}
	if !r.sawPath(want) || r.sawPath(other) {
		r.mu.Lock()
		defer r.mu.Unlock()
		t.Errorf("%s: fetched %v, want %s only", name, r.hits, want)
	}
}

func TestInstallActaTUIPicksTheBuild(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct{ env, want string }{
		{"ACTA_TUI=1", fullAsset},
		{"ACTA_TUI=0", cliAsset},
	} {
		r := bothRig(t)
		out, err := r.run(t, "", tc.env)
		if err != nil || !r.installed() {
			t.Fatalf("%s: err %v, installed %v\n%s", tc.env, err, r.installed(), out)
		}
		r.assetIs(t, tc.env, tc.want)
	}
}

// A bad ACTA_TUI must stop the script before it asks the server for anything.
func TestInstallBadActaTUIFailsBeforeDownload(t *testing.T) {
	t.Parallel()

	for _, v := range []string{"2", "yes", " 1", "unset", ""} {
		r := bothRig(t)
		out, err := r.run(t, "", "ACTA_TUI="+v)
		if err == nil || !strings.Contains(out, "ACTA_TUI must be 0 or 1") {
			t.Errorf("ACTA_TUI=%q: err %v, out %q", v, err, out)
		}
		r.mu.Lock()
		hits := len(r.hits)
		r.mu.Unlock()
		if hits != 0 || r.installed() {
			t.Errorf("ACTA_TUI=%q: %d requests made, installed %v", v, hits, r.installed())
		}
	}
}

// With no terminal and no ACTA_TUI, nobody can answer, so the full build
// comes down the same as before this choice existed.
func TestInstallNoTerminalGetsFullBuild(t *testing.T) {
	t.Parallel()

	r := bothRig(t)
	out, err := r.run(t, "")
	if err != nil || !r.installed() {
		t.Fatalf("err %v\n%s", err, out)
	}
	if strings.Contains(out, "[Y/n]") {
		t.Errorf("asked with no terminal: %q", out)
	}
	r.assetIs(t, "no terminal", fullAsset)
}

// ptyRun runs install.sh on a fresh terminal, types each answer when the
// question shows, and gives back what the terminal showed.
func (r *installRig) ptyRun(t *testing.T, answers ...string) string {
	t.Helper()
	script, err := filepath.Abs("install.sh")
	if err != nil {
		t.Fatal(err)
	}
	// Python's pty.fork gives the child a real controlling terminal, which
	// Go cannot do without a new dependency.
	const driver = `
import os, pty, select, signal, sys, time
answers = sys.argv[2:]
pid, fd = pty.fork()
if pid == 0:
    os.execv("/bin/sh", ["/bin/sh", sys.argv[1]])
seen, asked = b"", 0
end = time.time() + 30
while True:
    left = end - time.time()
    # A script stuck on a question it never gets an answer to must not hang
    # the test, so give up and stop it when time runs out.
    if left <= 0:
        os.kill(pid, signal.SIGKILL)
        sys.stdout.write(seen.decode(errors="replace") + "\nTIMEOUT\n")
        sys.exit(1)
    ready, _, _ = select.select([fd], [], [], left)
    if not ready:
        continue
    try:
        chunk = os.read(fd, 4096)
    except OSError:
        break
    if not chunk:
        break
    seen += chunk
    if asked < len(answers) and seen.count(b"[Y/n]") > asked:
        os.write(fd, answers[asked].encode() + b"\n")
        asked += 1
os.waitpid(pid, 0)
sys.stdout.write(seen.decode(errors="replace"))
`
	py, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 is not installed, so there is no way to open a terminal")
	}
	// Run python by its full path. The script keeps the rig PATH, so its
	// fake uname still wins over the real one.
	cmd := exec.Command(py, append([]string{"-I", "-c", driver, script}, answers...)...)
	cmd.Env = r.env("")
	cmd.Dir = t.TempDir()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("pty run: %v\n%s", err, out)
	}
	return string(out)
}

func TestInstallAsksOnTerminal(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		answers []string
		asks    int
		want    string
	}{
		{"enter", []string{""}, 1, fullAsset},
		{"y", []string{"y"}, 1, fullAsset},
		{"n", []string{"n"}, 1, cliAsset},
		{"bad then n", []string{"maybe", "n"}, 2, cliAsset},
	} {
		r := bothRig(t)
		out := r.ptyRun(t, tc.answers...)
		if got := strings.Count(out, "Install the TUI board too? It adds about 7M. [Y/n]"); got != tc.asks {
			t.Errorf("%s: asked %d times, want %d\n%s", tc.name, got, tc.asks, out)
		}
		if !r.installed() {
			t.Fatalf("%s: not installed\n%s", tc.name, out)
		}
		r.assetIs(t, tc.name, tc.want)
	}
}
