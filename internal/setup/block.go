package setup

import (
	"errors"
	"os"
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
// the end. A file with the markers keeps everything outside them: only the
// bytes between the two markers are replaced, so a re-run changes nothing
// else.
func WriteBlock(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return os.WriteFile(path, []byte(Block), 0o644)
		}
		return err
	}
	old := string(raw)
	begin := strings.Index(old, beginMarker)
	end := -1
	if begin >= 0 {
		if i := strings.Index(old[begin:], endMarker); i >= 0 {
			end = begin + i
		}
	}
	if begin < 0 || end < 0 {
		if old != "" && !strings.HasSuffix(old, "\n") {
			old += "\n"
		}
		return os.WriteFile(path, []byte(old+Block), 0o644)
	}
	tail := old[end+len(endMarker):]
	tail = strings.TrimPrefix(tail, "\n")
	return os.WriteFile(path, []byte(old[:begin]+Block+tail), 0o644)
}
