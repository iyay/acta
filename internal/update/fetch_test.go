package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	testOS      = "darwin"
	testArch    = "arm64"
	testTag     = "v0.1.46"
	testArchive = "acta_darwin_arm64.tar.gz"
)

var testBinary = []byte("#!fake acta binary\n")

// tarGz packs the given files into a gzip tar, like the release archive.
func tarGz(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, body := range files {
		hdr := &tar.Header{Name: name, Mode: 0o755, Size: int64(len(body)), Typeflag: tar.TypeReg}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(body); err != nil {
			t.Fatal(err)
		}
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
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

// testClient is NewClient with the TLS server's certificate trusted.
func testClient(srv *httptest.Server) *http.Client {
	c := NewClient()
	c.Transport = srv.Client().Transport
	return c
}

func TestNewClient(t *testing.T) {
	c := NewClient()
	if c.Timeout != 60*time.Second {
		t.Errorf("timeout = %v, want 60s", c.Timeout)
	}
	if c.CheckRedirect == nil {
		t.Fatal("CheckRedirect is nil")
	}
	httpReq, _ := http.NewRequest("GET", "http://example.com/x", nil)
	if err := c.CheckRedirect(httpReq, nil); err == nil {
		t.Error("redirect to http was allowed")
	}
	httpsReq, _ := http.NewRequest("GET", "https://example.com/x", nil)
	if err := c.CheckRedirect(httpsReq, nil); err != nil {
		t.Errorf("redirect to https refused: %v", err)
	}
}

func TestLatest(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		want    string
		wantErr string
	}{
		{
			name: "good tag",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/rel/tag/v0.1.46", http.StatusFound)
			},
			want: "v0.1.46",
		},
		{
			name: "absolute location",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Location", "https://github.com/iyay/acta/releases/tag/v0.2.0")
				w.WriteHeader(http.StatusFound)
			},
			want: "v0.2.0",
		},
		{
			name: "no redirect",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			wantErr: "no redirect",
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "boom", http.StatusInternalServerError)
			},
			wantErr: "500",
		},
		{
			name: "redirect without location",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusFound)
			},
			wantErr: "Location",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/rel/latest" {
					t.Errorf("path = %q, want /rel/latest", r.URL.Path)
				}
				tc.handler(w, r)
			}))
			defer srv.Close()
			c := testClient(srv)
			got, err := Latest(c, srv.URL+"/rel")
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %v, want containing %q", err, tc.wantErr)
				}
				if got != "" {
					t.Errorf("tag = %q on error, want empty", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("tag = %q, want %q", got, tc.want)
			}
		})
	}
}

// Latest must not change the client it was given: the caller reuses it for Fetch.
func TestLatestLeavesClientAlone(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/rel/tag/v1.0.0", http.StatusFound)
	}))
	defer srv.Close()
	c := testClient(srv)
	if _, err := Latest(c, srv.URL+"/rel"); err != nil {
		t.Fatal(err)
	}
	httpReq, _ := http.NewRequest("GET", "http://example.com/x", nil)
	if err := c.CheckRedirect(httpReq, nil); err == nil {
		t.Error("Latest changed the client's CheckRedirect")
	}
}

func TestFetch(t *testing.T) {
	good := tarGz(t, map[string][]byte{"acta": testBinary, "README.md": []byte("hi")})
	noActa := tarGz(t, map[string][]byte{"README.md": []byte("hi")})

	tests := []struct {
		name      string
		archive   []byte
		checksums string
		status    map[string]int // path suffix -> status override
		redirect  bool           // archive redirects to http://
		want      []byte
		wantErr   string
	}{
		{
			name:      "good fetch",
			archive:   good,
			checksums: sum(good) + "  " + testArchive + "\n",
			want:      testBinary,
		},
		{
			name:      "star name line",
			archive:   good,
			checksums: sum(good) + " *" + testArchive + "\n",
			want:      testBinary,
		},
		{
			name:    "picks the right line among many",
			archive: good,
			checksums: sum([]byte("other")) + "  acta_linux_amd64.tar.gz\n" +
				strings.ToUpper(sum(good)) + "  " + testArchive + "\n",
			want: testBinary,
		},
		{
			name:      "mismatch",
			archive:   good,
			checksums: sum([]byte("tampered")) + "  " + testArchive + "\n",
			wantErr:   "checksum mismatch for " + testArchive,
		},
		{
			name:      "missing line",
			archive:   good,
			checksums: sum(good) + "  acta_linux_amd64.tar.gz\n",
			wantErr:   "no checksum line for " + testArchive,
		},
		{
			name:      "archive without acta",
			archive:   noActa,
			checksums: sum(noActa) + "  " + testArchive + "\n",
			wantErr:   "no acta file in " + testArchive,
		},
		{
			name:      "archive not gzip",
			archive:   []byte("not a gzip"),
			checksums: sum([]byte("not a gzip")) + "  " + testArchive + "\n",
			wantErr:   "read " + testArchive,
		},
		{
			name:      "archive bad status",
			archive:   good,
			checksums: sum(good) + "  " + testArchive + "\n",
			status:    map[string]int{testArchive: http.StatusNotFound},
			wantErr:   "404",
		},
		{
			name:      "checksums bad status",
			archive:   good,
			checksums: sum(good) + "  " + testArchive + "\n",
			status:    map[string]int{"checksums.txt": http.StatusInternalServerError},
			wantErr:   "500",
		},
		{
			name:      "redirect to http refused",
			archive:   good,
			checksums: sum(good) + "  " + testArchive + "\n",
			redirect:  true,
			wantErr:   "https",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			base := "/rel/download/" + testTag + "/"
			mux := http.NewServeMux()
			serve := func(file string, body []byte) {
				mux.HandleFunc(base+file, func(w http.ResponseWriter, r *http.Request) {
					if code, ok := tc.status[file]; ok {
						http.Error(w, "nope", code)
						return
					}
					if tc.redirect && file == testArchive {
						http.Redirect(w, r, "http://127.0.0.1:1/evil", http.StatusFound)
						return
					}
					w.Write(body)
				})
			}
			serve(testArchive, tc.archive)
			serve("checksums.txt", []byte(tc.checksums))
			srv := httptest.NewTLSServer(mux)
			defer srv.Close()

			got, err := Fetch(testClient(srv), srv.URL+"/rel", testTag, testOS, testArch)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %v, want containing %q", err, tc.wantErr)
				}
				if got != nil {
					t.Errorf("returned %d bytes on error, want none", len(got))
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, tc.want) {
				t.Errorf("binary = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFetchTimeout(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
	}))
	defer srv.Close()
	c := testClient(srv)
	c.Timeout = 50 * time.Millisecond
	got, err := Fetch(c, srv.URL, testTag, testOS, testArch)
	if err == nil {
		t.Fatal("want a timeout error")
	}
	if got != nil {
		t.Errorf("returned %d bytes on error, want none", len(got))
	}
}
