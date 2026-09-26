package board

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"pm-board/internal/config"
)

// agentRec is one record of who last ticked a task, in the file a tick writes.
type agentRec struct {
	Agent string `json:"agent"`
	At    string `json:"at"`
}

// readAgents collects the agent of every task from the .agents.json of each
// root on disk, newest record first. A file that is missing or not JSON is
// skipped: a broken record must not hide the board.
func readAgents(roots []string) map[string]agentRec {
	out := map[string]agentRec{}
	for _, root := range roots {
		b, err := os.ReadFile(filepath.Join(root, ".agents.json"))
		if err != nil {
			continue
		}
		var recs map[string]agentRec
		if json.Unmarshal(b, &recs) != nil {
			continue
		}
		for id, rec := range recs {
			if old, ok := out[id]; ok && !newer(rec, old) {
				continue
			}
			out[id] = rec
		}
	}
	return out
}

// newer says whether a came after b. A record whose time cannot be read never
// beats one that can, so a garbled file cannot win over a real tick.
func newer(a, b agentRec) bool {
	at, err := time.Parse(time.RFC3339, a.At)
	if err != nil {
		return false
	}
	old, err := time.Parse(time.RFC3339, b.At)
	if err != nil {
		return true
	}
	return at.After(old)
}

// fillAgents tells the board who works on what. Every open task takes the agent
// that last ticked it; a plan, spec or bug takes the distinct agents of its
// open tasks. A tree read from git has no file to read, so it adds nothing.
func (b *Board) fillAgents(main config.Config, others []Tree) {
	roots := []string{main.Root}
	for _, t := range others {
		if t.Files == nil {
			roots = append(roots, t.Cfg.Root)
		}
	}
	recs := readAgents(roots)
	for _, it := range b.Items {
		if it.Kind == KindTask {
			if it.Status != "done" {
				it.Agent = recs[it.ID].Agent
			}
			continue
		}
		it.Agent = b.agentsOfTasks(it, recs)
	}
}

// agentsOfTasks joins the agents of the open tasks under an item, each one
// named once, in a fixed order so the row does not wobble between reads.
func (b *Board) agentsOfTasks(it *Item, recs map[string]agentRec) string {
	seen := map[string]bool{}
	for _, id := range it.Children {
		t := b.byID[id]
		if t == nil || t.Status == "done" {
			continue
		}
		if a := recs[id].Agent; a != "" {
			seen[a] = true
		}
	}
	names := make([]string, 0, len(seen))
	for a := range seen {
		names = append(names, a)
	}
	sort.Strings(names)
	return strings.Join(names, ",")
}
