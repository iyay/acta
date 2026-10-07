package setup

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const beginMarker = "<!-- acta:begin -->"
const endMarker = "<!-- acta:end -->"

// Block is the acta block, the single source since the setup skill no
// longer holds it. It matches byte for byte the block that used to live in
// plugin/skills/setup/SKILL.md.
const Block = "<!-- acta:begin -->\n" +
	"## acta\n" +
	"This repo uses the acta plugin. Before each workflow step, load the matching acta skill and follow it.\n" +
	"Specs, plans, bugs, debt, scratch items and the wiki live in `.acta/`.\n" +
	"Project knowledge (gotchas, runbooks, decisions with their why) goes to `.acta/wiki/`, never to agent memory. Write a page only when a fresh agent would lose time or repeat a mistake without it. When a fact changes, rewrite its page.\n" +
	"Before changing a file, run `acta wiki match <file>` and read each page it names.\n" +
	"Work in flight goes to `acta state set <plan id> next` with the text on stdin (read it back with `acta state <plan id>`), not to agent memory. Agent memory keeps only the user's own setup.\n" +
	"Raw ideas go to Scratchpad with `acta scratch new`. A finished scratch item is specced, never dropped; dropped means not done or not valid.\n" +
	"<!-- acta:end -->\n"

// WriteBlock writes Block to path. A missing file ends up holding only the
// block. A file without the markers keeps its bytes and gains the block at
// the end. A file with one full marker pair keeps everything outside it:
// only the bytes between the two markers are replaced, so a re-run changes
// nothing else. A file with half a pair (a begin without an end, an end
// without a begin, the end before the begin, or a repeated marker) is left
// alone and an error comes back, so a broken file never loses user text.
// The block follows the file's own line endings, so a CRLF file never
// grows a stray carriage return however often the wizard rewrites it.
// Every write lands through a temp file in the same folder plus a rename,
// so a failed write leaves the old file whole.
// A link never turns into a plain file: the path is resolved first, so the
// swap lands on the file the link points to and the link itself is left
// alone, with the temp file made in that file's own folder.
// Only the mode bits carry over to the new file: owner, group and xattrs
// stay as the fresh temp file leaves them.
func WriteBlock(path string) error {
	// A path with no file behind it keeps its own name.
	real := path
	if dst, err := filepath.EvalSymlinks(path); err == nil {
		real = dst
	}
	raw, err := os.ReadFile(real)
	perm := os.FileMode(0o644)
	if err == nil {
		// An existing file keeps its own mode: WriteFile only used its
		// perm for new files, so the rename path must do the same.
		if fi, statErr := os.Stat(real); statErr == nil {
			perm = fi.Mode().Perm()
		}
	} else {
		if errors.Is(err, os.ErrNotExist) {
			return swapFile(real, []byte(Block), perm)
		}
		return err
	}
	old := string(raw)
	eol := "\n"
	if strings.Contains(old, "\r\n") {
		eol = "\r\n"
	}
	block := blockText(eol)
	begin := strings.Index(old, beginMarker)
	end := strings.Index(old, endMarker)
	switch {
	case begin < 0 && end < 0:
		sep := ""
		if old != "" && !strings.HasSuffix(old, "\n") {
			sep = eol
		}
		return swapFile(real, []byte(old+sep+block), perm)
	case begin < 0:
		return fmt.Errorf("setup: %s has an end marker without a begin marker", path)
	case end < 0:
		return fmt.Errorf("setup: %s has a begin marker without an end marker", path)
	case end < begin:
		return fmt.Errorf("setup: %s has its end marker before its begin marker", path)
	case strings.Contains(old[begin+len(beginMarker):], beginMarker):
		return fmt.Errorf("setup: %s has more than one begin marker", path)
	case strings.Contains(old[end+len(endMarker):], endMarker):
		return fmt.Errorf("setup: %s has more than one end marker", path)
	}
	tail := old[end+len(endMarker):]
	tail = strings.TrimPrefix(tail, "\r\n")
	tail = strings.TrimPrefix(tail, "\n")
	return swapFile(real, []byte(old[:begin]+block+tail), perm)
}

// blockText returns Block with the file's own line endings. LF files keep
// the block as is; CRLF files get a CRLF copy so no lone carriage return
// ever lands, however often the wizard rewrites.
func blockText(eol string) string {
	if eol == "\r\n" {
		return strings.ReplaceAll(Block, "\n", "\r\n")
	}
	return Block
}

// swapFile lands bytes through a temp file in the same folder plus a
// rename. A failed write leaves the old file whole because the new bytes
// never touch it; the rename is the only move, and the temp file is
// cleaned up on the way out.
func swapFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".acta-block-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
