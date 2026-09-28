package write

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
)

// NewScratch writes a scratch item from the body an agent sent and commits
// it. The body is kept byte for byte, so an idea reads back the way it was
// written. An empty body is refused before anything touches the disk.
func NewScratch(cfg config.Config, slug, title string, body []byte) (Outcome, error) {
	if strings.TrimSpace(string(body)) == "" {
		return Outcome{}, bad("scratch body is empty")
	}
	path, err := datedPath(cfg, cfg.Dirs.Scratch, slug)
	if err != nil {
		return Outcome{}, err
	}
	if title == "" {
		title = slug
	}
	b, err := board.Load(cfg)
	if err != nil {
		return Outcome{}, err
	}
	next, taken := scanIDs(b)
	shortID := fmt.Sprintf("SCRATCH-%d", next["SCRATCH"])
	content := body
	for _, f := range []struct{ key, value string }{
		{"id", shortID},
		{"hash", freeHash(taken)},
		{"title", title},
		{"status", "raw"},
		{"created", Now().Format("2006-01-02")},
	} {
		if content, err = SetField(content, f.key, f.value); err != nil {
			return Outcome{}, err
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return Outcome{}, err
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return Outcome{}, err
	}
	o := finish(cfg, path, "acta: new scratch "+stem(path), false)
	o.ShortID = shortID
	return o, nil
}

// AppendScratch puts text at the end of a scratch body, one blank line below
// what the body already holds. Any status takes more text: an idea that was
// specced or dropped can still collect an answer.
func AppendScratch(cfg config.Config, b *board.Board, id string, text []byte) (Outcome, error) {
	it := b.Get(id)
	switch {
	case it == nil:
		return Outcome{}, bad("unknown id %s", id)
	case it.Kind != board.KindScratch:
		return Outcome{}, bad("%s is not a scratch item", id)
	}
	if strings.TrimSpace(string(text)) == "" {
		return Outcome{}, bad("scratch body is empty")
	}
	dirty, err := dirtyBefore(cfg, it.Path)
	if err != nil {
		return Outcome{}, err
	}
	src, err := os.ReadFile(it.Path)
	if err != nil {
		return Outcome{}, err
	}
	out := string(src)
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	out += "\n" + string(text)
	if err := os.WriteFile(it.Path, []byte(out), 0o644); err != nil {
		return Outcome{}, err
	}
	o := finish(cfg, it.Path, "acta: add to scratch "+stem(it.Path), dirty)
	o.ShortID = it.ShortID
	return o, nil
}

// stem is the file name without ".md": the way a commit message names it.
func stem(path string) string {
	return strings.TrimSuffix(filepath.Base(path), ".md")
}
