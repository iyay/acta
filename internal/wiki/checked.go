package wiki

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// CheckedFile sits in the wiki folder. It holds one line per page, the page id
// and the time it was last checked. Freshness reads the commits that added a
// line, never the time: the time only makes the line change on each check.
const CheckedFile = ".checked"

// RewriteChecked gives the new text of the checked file. Every touched page
// gets now. A page named twice keeps its newest time, which is what a union
// merge of two branches leaves behind. A page that is not known any more is
// dropped, and so is a line that is not "<id> <time>". Lines come out sorted,
// so the same pages always give the same file.
func RewriteChecked(old string, known, touched []string, now string) string {
	dates := map[string]string{}
	for _, line := range strings.Split(old, "\n") {
		f := strings.Fields(line)
		if len(f) != 2 || !slices.Contains(known, f[0]) {
			continue
		}
		if f[1] > dates[f[0]] {
			dates[f[0]] = f[1]
		}
	}
	for _, id := range touched {
		if slices.Contains(known, id) {
			dates[id] = now
		}
	}
	ids := make([]string, 0, len(dates))
	for id := range dates {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	var b strings.Builder
	for _, id := range ids {
		b.WriteString(id + " " + dates[id] + "\n")
	}
	return b.String()
}

// MarkChecked rewrites the checked file in dir, the wiki folder, and gives its
// path. A missing file counts as empty. A link is refused: writing through it
// would wipe a file that is not the wiki's.
func MarkChecked(dir string, known, touched []string, now string) (string, error) {
	p := filepath.Join(dir, CheckedFile)
	if fi, err := os.Lstat(p); err == nil && !fi.Mode().IsRegular() {
		return "", fmt.Errorf("%s is not a plain file", p)
	}
	old, err := os.ReadFile(p)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	return p, os.WriteFile(p, []byte(RewriteChecked(string(old), known, touched, now)), 0o644)
}
