// Package telemetry counts how acta is used, per local day, and keeps the
// counts on this machine until a flush sends them. It never records text,
// paths or names: only counts under fixed keys.
package telemetry

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/iyay/acta/plugin"
)

// Endpoint is where finished days go. Release builds set it with -X. Every
// other build leaves it empty, and then acta counts and sends nothing.
var Endpoint = ""

const (
	// MaxCount stops one count from growing without end inside a loop.
	MaxCount = 100000
	// MaxDays is how many unsent days stay on disk. Older ones go unsent.
	MaxDays = 7
)

// now is swapped by tests that need another day.
var now = time.Now

// Today is the local calendar day that counts go to now.
func Today() string { return now().Format("2006-01-02") }

// Day is one local day of counts, in the shape the collector gets.
type Day struct {
	Kind      string            `json:"kind"`
	InstallID string            `json:"install_id"`
	Day       string            `json:"day"`
	Version   string            `json:"acta_version"`
	OS        string            `json:"os"`
	Arch      string            `json:"arch"`
	Sessions  map[string]int    `json:"sessions"`
	Commands  map[string]int    `json:"commands"`
	Skills    map[string]int    `json:"skills"`
	Workflow  map[string]int    `json:"workflow"`
	Blocks    map[string]int    `json:"blocks"`
	Settings  map[string]string `json:"settings"`
}

// state is the machine's own record. The notice mark sits here and not under
// the install ID, so turning telemetry off and on again shows no second notice.
type state struct {
	InstallID string    `json:"install_id,omitempty"`
	Off       bool      `json:"off,omitempty"`
	Noticed   bool      `json:"noticed,omitempty"`
	LastFlush time.Time `json:"last_flush,omitempty"`
}

// Store reads and writes one telemetry folder.
type Store struct {
	dir string
}

// Open gives the store for dir. It touches nothing on disk.
func Open(dir string) *Store { return &Store{dir: dir} }

// Dir is ~/.acta/telemetry. It is kept apart from config.yaml because the
// setup wizard rewrites that file and must not lose the user's off choice.
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".acta", "telemetry"), nil
}

// Default opens the store in Dir. A missing home gives a store that is never
// enabled, so callers need no error path.
func Default() *Store {
	dir, err := Dir()
	if err != nil {
		return Open("")
	}
	return Open(dir)
}

// envOff says whether the environment turns telemetry off. It beats the
// stored setting in both directions of a process: counting and sending.
func envOff() bool {
	if os.Getenv("DO_NOT_TRACK") != "" && os.Getenv("DO_NOT_TRACK") != "0" {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(os.Getenv("ACTA_TELEMETRY"))) {
	case "0", "false", "off", "no":
		return true
	}
	return false
}

// Enabled says whether this process may count and send.
func (s *Store) Enabled() bool {
	if Endpoint == "" || s.dir == "" || envOff() {
		return false
	}
	return !s.readState().Off
}

// Source names what decides the current state, for acta telemetry status.
func (s *Store) Source() string {
	switch {
	case Endpoint == "":
		return "no endpoint in this build"
	case envOff():
		return "environment"
	}
	return "setting"
}

// InstallID is the current random ID, or empty when there is none.
func (s *Store) InstallID() string { return s.readState().InstallID }

// Noticed says whether the notice was shown on this machine.
func (s *Store) Noticed() bool { return s.readState().Noticed }

// LastFlush is when a flush last tried to send.
func (s *Store) LastFlush() time.Time { return s.readState().LastFlush }

var (
	commandKeys = set("list", "show", "set", "state", "plan", "bug", "debt", "scratch", "flake",
		"config", "hook", "tick", "commit", "commits", "cleanup", "worktrees", "usage", "id",
		"migrate-root", "doctor", "setup", "wiki", "dispatch", "reply-back", "run-one",
		"eval-omp", "tidy", "ready", "spec-dates", "update", "telemetry", "tui", "help", "version")
	sessionKeys  = set("claude", "codex", "omp")
	skillKeys    = set("bug", "build", "cleanup", "debug", "frame", "land", "lean", "migrate", "review", "scratch", "setup", "shape", "slice", "tdd")
	workflowKeys = set("spec_created", "plan_created", "bug_created", "plan_landed")
	blockKeys    = set("go_test", "destructive")
	settingKeys  = map[string]map[string]bool{
		"build_executor": set("subagent", "dispatch", "inline"),
		"plan_depth":     set("minimal", "full"),
		"review_depth":   set("quick", "standard", "deep"),
		"chat_language": set("english", "indonesian", "sundanese", "javanese", "malay", "spanish",
			"portuguese", "french", "german", "italian", "dutch", "russian", "turkish", "arabic",
			"hindi", "japanese", "korean", "chinese", "vietnamese", "thai"),
	}
)

// KnownCommand says whether a top-level command name has its own count.
func KnownCommand(name string) bool { return commandKeys[name] }

func set(keys ...string) map[string]bool {
	m := make(map[string]bool, len(keys))
	for _, k := range keys {
		m[k] = true
	}
	return m
}

// Add counts one event. A key outside the fixed list becomes other, except
// for skills: a user's own skill name can be private, so it is not kept at all.
func (s *Store) Add(group, key string) error {
	if !s.Enabled() {
		return nil
	}
	return s.update(func(d *Day) {
		var m map[string]int
		var known map[string]bool
		switch group {
		case "commands":
			m, known = d.Commands, commandKeys
		case "sessions":
			m, known = d.Sessions, sessionKeys
		case "skills":
			if !skillKeys[key] {
				return
			}
			m, known = d.Skills, skillKeys
		case "workflow":
			m, known = d.Workflow, workflowKeys
		case "blocks":
			m, known = d.Blocks, blockKeys
		default:
			return
		}
		if !known[key] {
			key = "other"
		}
		if m[key] < MaxCount {
			m[key]++
		}
	})
}

// SetSettings stores today's snapshot of the four settings. Each value outside
// its fixed list is other, and an empty one is unset.
func (s *Store) SetSettings(v map[string]string) error {
	if !s.Enabled() {
		return nil
	}
	return s.update(func(d *Day) {
		for k, known := range settingKeys {
			val := strings.ToLower(strings.TrimSpace(v[k]))
			switch {
			case val == "":
				val = "unset"
			case !known[val]:
				val = "other"
			}
			d.Settings[k] = val
		}
	})
}

// On turns counting back on. The next count makes a new install ID.
func (s *Store) On() error {
	return s.withLock(func() error {
		st := s.readState()
		st.Off = false
		return s.writeState(st)
	})
}

// Off stops counting and deletes the install ID and every unsent day, so
// rows from before and after cannot be linked. The notice mark stays.
func (s *Store) Off() error {
	return s.withLock(func() error {
		st := s.readState()
		st.Off = true
		st.InstallID = ""
		if err := s.writeState(st); err != nil {
			return err
		}
		names, _ := s.dayNames()
		for _, n := range names {
			if err := os.Remove(s.dayPath(n)); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
		return nil
	})
}

// MarkNoticed records that the notice was shown. It says true only the first
// time, so the caller shows the notice once per machine.
func (s *Store) MarkNoticed() (bool, error) {
	first := false
	err := s.withLock(func() error {
		st := s.readState()
		if st.Noticed {
			return nil
		}
		first = true
		st.Noticed = true
		return s.writeState(st)
	})
	return first, err
}

// SetLastFlush records the time of a flush attempt, so the next one waits an hour.
func (s *Store) SetLastFlush(at time.Time) error {
	return s.withLock(func() error {
		st := s.readState()
		st.LastFlush = at
		return s.writeState(st)
	})
}

// Days lists the stored days that parse, oldest first. A file that does not
// parse, or names another day than its file name, is deleted and never sent.
func (s *Store) Days() ([]Day, error) {
	var days []Day
	err := s.withLock(func() error {
		names, err := s.dayNames()
		if err != nil {
			return err
		}
		for _, n := range names {
			d, ok := s.readDay(n)
			if !ok {
				os.Remove(s.dayPath(n))
				continue
			}
			days = append(days, d)
		}
		return nil
	})
	return days, err
}

// Delete removes one day, after the collector took it or refused it for good.
func (s *Store) Delete(day string) error {
	return s.withLock(func() error {
		err := os.Remove(s.dayPath(day))
		if os.IsNotExist(err) {
			return nil
		}
		return err
	})
}

// update changes today's file under the lock and keeps at most MaxDays files.
func (s *Store) update(change func(*Day)) error {
	return s.withLock(func() error {
		st := s.readState()
		if st.InstallID == "" {
			id, err := newID()
			if err != nil {
				return err
			}
			st.InstallID = id
			if err := s.writeState(st); err != nil {
				return err
			}
		}
		d, ok := s.readDay(Today())
		if !ok || d.InstallID != st.InstallID {
			d = newDay(Today(), st.InstallID)
		}
		change(&d)
		if err := s.writeDay(d); err != nil {
			return err
		}
		return s.prune()
	})
}

func newDay(day, id string) Day {
	return Day{
		Kind: "acta_telemetry", InstallID: id, Day: day, Version: plugin.Version(),
		OS: pick(runtime.GOOS, "darwin", "linux"), Arch: pick(runtime.GOARCH, "amd64", "arm64"),
		Sessions: map[string]int{}, Commands: map[string]int{}, Skills: map[string]int{},
		Workflow: map[string]int{}, Blocks: map[string]int{}, Settings: map[string]string{},
	}
}

func pick(v string, known ...string) string {
	for _, k := range known {
		if v == k {
			return v
		}
	}
	return "other"
}

func newID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// prune deletes the oldest days past MaxDays. Names are dates, so they sort in time order.
func (s *Store) prune() error {
	names, err := s.dayNames()
	if err != nil {
		return err
	}
	for len(names) > MaxDays {
		if err := os.Remove(s.dayPath(names[0])); err != nil && !os.IsNotExist(err) {
			return err
		}
		names = names[1:]
	}
	return nil
}

var dayName = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func (s *Store) dayNames() ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(s.dir, "days"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		n := strings.TrimSuffix(e.Name(), ".json")
		if e.Type().IsRegular() && n != e.Name() && dayName.MatchString(n) {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	return names, nil
}

func (s *Store) dayPath(day string) string {
	return filepath.Join(s.dir, "days", day+".json")
}

func (s *Store) readDay(day string) (Day, bool) {
	raw, err := os.ReadFile(s.dayPath(day))
	if err != nil {
		return Day{}, false
	}
	var d Day
	if err := json.Unmarshal(raw, &d); err != nil || d.Day != day || d.InstallID == "" {
		return Day{}, false
	}
	return d, true
}

func (s *Store) writeDay(d Day) error {
	return writeJSON(s.dayPath(d.Day), d)
}

// readState gives the zero state when the file is missing or broken. A broken
// file must not turn a user's off choice into on, so it reads as off.
func (s *Store) readState() state {
	raw, err := os.ReadFile(filepath.Join(s.dir, "state.json"))
	if err != nil {
		return state{}
	}
	var st state
	if err := json.Unmarshal(raw, &st); err != nil {
		return state{Off: true}
	}
	return st
}

func (s *Store) writeState(st state) error {
	return writeJSON(filepath.Join(s.dir, "state.json"), st)
}

// writeJSON writes a temp file and renames it, so a reader never sees half a file.
func writeJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// errBusy means another acta process held the lock too long. The count is
// lost, which the spec allows; a half-written file is not.
var errBusy = errors.New("telemetry folder is busy")

// withLock runs fn while holding the folder lock, so two acta processes never
// write the same file at once.
func (s *Store) withLock(fn func() error) error {
	if s.dir == "" {
		return fmt.Errorf("no telemetry folder")
	}
	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(s.dir, "lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	deadline := time.Now().Add(2 * time.Second)
	for {
		err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			break
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) {
			return err
		}
		if time.Now().After(deadline) {
			return errBusy
		}
		time.Sleep(5 * time.Millisecond)
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	return fn()
}
