package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// Stop a bad or hostile server from filling memory. The real archive is a few MB.
const (
	maxArchiveBytes   = 256 << 20
	maxChecksumsBytes = 1 << 20
)

// NewClient builds the client every download uses. It gives up after 60
// seconds and refuses to follow a redirect that leaves https, so a hop to
// plain http cannot swap the file on the way.
func NewClient() *http.Client {
	return &http.Client{
		Timeout: 60 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if req.URL.Scheme != "https" {
				return fmt.Errorf("refusing redirect to %s: only https is allowed", req.URL.Redacted())
			}
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			return nil
		},
	}
}

// Latest asks <base>/latest where it redirects and returns the tag at the end
// of that address, like v0.1.46. It must read the redirect itself, so it uses a
// copy of the client that does not follow it. The caller's client stays as is.
func Latest(c *http.Client, base string) (string, error) {
	noFollow := *c
	noFollow.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	u := strings.TrimRight(base, "/") + "/latest"
	resp, err := noFollow.Get(u)
	if err != nil {
		return "", fmt.Errorf("find latest release at %s: %w", u, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 300 || resp.StatusCode > 399 {
		if resp.StatusCode == http.StatusOK {
			return "", fmt.Errorf("find latest release at %s: no redirect, status 200", u)
		}
		return "", fmt.Errorf("find latest release at %s: status %d", u, resp.StatusCode)
	}
	loc := resp.Header.Get("Location")
	if loc == "" {
		return "", fmt.Errorf("find latest release at %s: redirect has no Location header", u)
	}
	parsed, err := url.Parse(loc)
	if err != nil {
		return "", fmt.Errorf("find latest release at %s: bad Location %q: %w", u, loc, err)
	}
	tag := path.Base(strings.TrimRight(parsed.Path, "/"))
	if tag == "" || tag == "." || tag == "/" {
		return "", fmt.Errorf("find latest release at %s: no tag in Location %q", u, loc)
	}
	return tag, nil
}

// Fetch downloads the release archive for this system and its checksums.txt,
// checks the archive against the checksum, and only then opens it. It returns
// the bytes of the acta file inside. On any failure it returns no bytes, so
// nothing that failed the check can ever be installed.
func Fetch(c *http.Client, base, tag, goos, goarch string) ([]byte, error) {
	name := fmt.Sprintf("acta_%s_%s.tar.gz", goos, goarch)
	dir := strings.TrimRight(base, "/") + "/download/" + tag + "/"

	sums, err := get(c, dir+"checksums.txt", maxChecksumsBytes)
	if err != nil {
		return nil, fmt.Errorf("download checksums.txt: %w", err)
	}
	want, ok := checksumFor(string(sums), name)
	if !ok {
		return nil, fmt.Errorf("no checksum line for %s", name)
	}

	archive, err := get(c, dir+name, maxArchiveBytes)
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", name, err)
	}
	have := sha256.Sum256(archive)
	if !strings.EqualFold(hex.EncodeToString(have[:]), want) {
		return nil, fmt.Errorf("checksum mismatch for %s", name)
	}

	bin, err := extractActa(archive, name)
	if err != nil {
		return nil, err
	}
	return bin, nil
}

// get reads one https address and insists on status 200. A body past limit is
// an error, not a silent cut, because a cut file would fail the check anyway
// with a confusing message.
func get(c *http.Client, u string, limit int64) ([]byte, error) {
	resp, err := c.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d from %s", resp.StatusCode, u)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", u, err)
	}
	if int64(len(body)) > limit {
		return nil, fmt.Errorf("%s is larger than %d bytes", u, limit)
	}
	return body, nil
}

// checksumFor finds the line for name in a checksums.txt. Each line is a hex
// hash, then the file name, which sha256sum may write as *name in binary mode.
func checksumFor(sums, name string) (string, bool) {
	for _, line := range strings.Split(sums, "\n") {
		f := strings.Fields(line)
		if len(f) < 2 {
			continue
		}
		if f[1] == name || f[1] == "*"+name {
			return f[0], true
		}
	}
	return "", false
}

// extractActa returns the bytes of the plain file called acta from a gzip tar.
func extractActa(archive []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", name, err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("no acta file in %s", name)
		}
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", name, err)
		}
		if hdr.Typeflag != tar.TypeReg || path.Clean(hdr.Name) != "acta" {
			continue
		}
		bin, err := io.ReadAll(io.LimitReader(tr, maxArchiveBytes+1))
		if err != nil {
			return nil, fmt.Errorf("read acta from %s: %w", name, err)
		}
		if int64(len(bin)) > maxArchiveBytes {
			return nil, fmt.Errorf("acta in %s is larger than %d bytes", name, maxArchiveBytes)
		}
		return bin, nil
	}
}
