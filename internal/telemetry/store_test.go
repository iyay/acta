package telemetry

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// openTest gives a store in a fresh folder with an endpoint set and no
// environment switch, so every test starts from a machine that counts.
func openTest(t *testing.T) *Store {
	t.Helper()
	old := Endpoint
	Endpoint = "https://script.google.com/macros/s/test/exec"
	t.Cleanup(func() { Endpoint = old })
	t.Setenv("DO_NOT_TRACK", "")
	t.Setenv("ACTA_TELEMETRY", "")
	return Open(filepath.Join(t.TempDir(), "telemetry"))
}

func today(t *testing.T, s *Store) Day {
	t.Helper()
	days, err := s.Days()
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range days {
		if d.Day == Today() {
			return d
		}
	}
	t.Fatalf("no day file for today in %v", days)
	return Day{}
}

func TestStoreCountsNothingWhenSwitchedOff(t *testing.T) {
	cases := map[string]func(t *testing.T, s *Store){
		"no endpoint":      func(t *testing.T, s *Store) { Endpoint = "" },
		"DO_NOT_TRACK":     func(t *testing.T, s *Store) { t.Setenv("DO_NOT_TRACK", "1") },
		"ACTA_TELEMETRY=0": func(t *testing.T, s *Store) { t.Setenv("ACTA_TELEMETRY", "0") },
		"ACTA_TELEMETRY=false": func(t *testing.T, s *Store) {
			t.Setenv("ACTA_TELEMETRY", "false")
		},
		"ACTA_TELEMETRY=off": func(t *testing.T, s *Store) { t.Setenv("ACTA_TELEMETRY", "off") },
		"ACTA_TELEMETRY=no":  func(t *testing.T, s *Store) { t.Setenv("ACTA_TELEMETRY", "NO") },
		"setting off": func(t *testing.T, s *Store) {
			if err := s.Off(); err != nil {
				t.Fatal(err)
			}
		},
	}
	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			s := openTest(t)
			setup(t, s)
			if s.Enabled() {
				t.Fatal("Enabled() = true, want false")
			}
			if err := s.Add("commands", "show"); err != nil {
				t.Fatal(err)
			}
			if err := s.SetSettings(map[string]string{"plan_depth": "full"}); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(filepath.Join(s.dir, "days")); !os.IsNotExist(err) {
				t.Fatalf("days folder exists after a switched-off count: %v", err)
			}
			if s.InstallID() != "" {
				t.Fatalf("install ID made while switched off: %q", s.InstallID())
			}
		})
	}
}

func TestStoreSourceNamesWhatDecides(t *testing.T) {
	s := openTest(t)
	if got := s.Source(); got != "setting" {
		t.Fatalf("Source() = %q, want setting", got)
	}
	t.Setenv("DO_NOT_TRACK", "1")
	if got := s.Source(); got != "environment" {
		t.Fatalf("Source() = %q, want environment", got)
	}
	Endpoint = ""
	if got := s.Source(); got != "no endpoint in this build" {
		t.Fatalf("Source() = %q, want no endpoint in this build", got)
	}
}

func TestStoreMapsUnknownKeys(t *testing.T) {
	s := openTest(t)
	adds := [][2]string{
		{"commands", "show"}, {"commands", "frobnicate"}, {"commands", "../etc"},
		{"sessions", "codex"}, {"sessions", "vim"},
		{"skills", "shape"}, {"skills", "my-private-skill"},
		{"workflow", "plan_created"}, {"workflow", "weird"},
		{"blocks", "go_test"}, {"blocks", "rm"},
	}
	for _, a := range adds {
		if err := s.Add(a[0], a[1]); err != nil {
			t.Fatal(err)
		}
	}
	d := today(t, s)
	want := map[string]map[string]int{
		"commands": {"show": 1, "other": 2},
		"sessions": {"codex": 1, "other": 1},
		"skills":   {"shape": 1},
		"workflow": {"plan_created": 1, "other": 1},
		"blocks":   {"go_test": 1, "other": 1},
	}
	got := map[string]map[string]int{
		"commands": d.Commands, "sessions": d.Sessions, "skills": d.Skills,
		"workflow": d.Workflow, "blocks": d.Blocks,
	}
	for g, w := range want {
		gotJSON, _ := json.Marshal(got[g])
		wantJSON, _ := json.Marshal(w)
		if string(gotJSON) != string(wantJSON) {
			t.Errorf("%s = %s, want %s", g, gotJSON, wantJSON)
		}
	}
	if !regexp.MustCompile(`^[0-9a-f]{32}$`).MatchString(d.InstallID) {
		t.Fatalf("install ID %q is not 32 lowercase hex", d.InstallID)
	}
	if d.OS != "darwin" && d.OS != "linux" && d.OS != "other" {
		t.Fatalf("os %q is outside the fixed list", d.OS)
	}
	if d.Arch != "amd64" && d.Arch != "arm64" && d.Arch != "other" {
		t.Fatalf("arch %q is outside the fixed list", d.Arch)
	}
}

func TestStoreSettingsSnapshot(t *testing.T) {
	s := openTest(t)
	err := s.SetSettings(map[string]string{
		"build_executor": "inline",
		"plan_depth":     "sideways",
		"review_depth":   "",
		"chat_language":  " Indonesian ",
	})
	if err != nil {
		t.Fatal(err)
	}
	got := today(t, s).Settings
	want := map[string]string{
		"build_executor": "inline",
		"plan_depth":     "other",
		"review_depth":   "unset",
		"chat_language":  "indonesian",
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("settings[%s] = %q, want %q", k, got[k], v)
		}
	}
	if err := s.SetSettings(map[string]string{"chat_language": "Klingon"}); err != nil {
		t.Fatal(err)
	}
	if got := today(t, s).Settings["chat_language"]; got != "other" {
		t.Fatalf("chat_language = %q, want other", got)
	}
	if len(today(t, s).Settings) != 4 {
		t.Fatalf("settings hold keys outside the snapshot: %v", today(t, s).Settings)
	}
}

func TestStoreCapsCounts(t *testing.T) {
	s := openTest(t)
	if err := s.Add("commands", "show"); err != nil {
		t.Fatal(err)
	}
	d := today(t, s)
	d.Commands["show"] = MaxCount
	if err := s.writeDay(d); err != nil {
		t.Fatal(err)
	}
	if err := s.Add("commands", "show"); err != nil {
		t.Fatal(err)
	}
	if got := today(t, s).Commands["show"]; got != MaxCount {
		t.Fatalf("count = %d, want the cap %d", got, MaxCount)
	}
}

func TestStoreOffDeletesDataKeepsNotice(t *testing.T) {
	s := openTest(t)
	if err := s.Add("commands", "show"); err != nil {
		t.Fatal(err)
	}
	if first, err := s.MarkNoticed(); err != nil || !first {
		t.Fatalf("MarkNoticed() = %v, %v, want true", first, err)
	}
	oldID := s.InstallID()
	if err := s.Off(); err != nil {
		t.Fatal(err)
	}
	if s.InstallID() != "" {
		t.Fatal("install ID kept after off")
	}
	entries, _ := os.ReadDir(filepath.Join(s.dir, "days"))
	if len(entries) != 0 {
		t.Fatalf("unsent days kept after off: %v", entries)
	}
	if err := s.On(); err != nil {
		t.Fatal(err)
	}
	if first, _ := s.MarkNoticed(); first {
		t.Fatal("notice would show a second time after off and on")
	}
	if err := s.Add("commands", "show"); err != nil {
		t.Fatal(err)
	}
	if id := s.InstallID(); id == "" || id == oldID {
		t.Fatalf("install ID after on = %q, want a new one (old %q)", id, oldID)
	}
}

func TestStoreConcurrentAddsKeepFilesValid(t *testing.T) {
	s := openTest(t)
	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := Open(s.dir).Add("commands", "show"); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	d := today(t, s)
	if d.Commands["show"] < 1 || d.Commands["show"] > 40 {
		t.Fatalf("count = %d, want between 1 and 40", d.Commands["show"])
	}
}

// TestStoreHelperProcess is the child the two-process test starts. It only
// runs when that test asks for it.
func TestStoreHelperProcess(t *testing.T) {
	dir := os.Getenv("ACTA_TELEMETRY_HELPER_DIR")
	if dir == "" {
		t.Skip("helper only")
	}
	Endpoint = "https://script.google.com/macros/s/test/exec"
	n, _ := strconv.Atoi(os.Getenv("ACTA_TELEMETRY_HELPER_N"))
	for i := 0; i < n; i++ {
		if err := Open(dir).Add("commands", "list"); err != nil {
			t.Fatal(err)
		}
	}
}

func TestStoreTwoProcessesKeepFilesValid(t *testing.T) {
	s := openTest(t)
	var cmds []*exec.Cmd
	for i := 0; i < 2; i++ {
		c := exec.Command(os.Args[0], "-test.run=^TestStoreHelperProcess$")
		c.Env = append(os.Environ(), "ACTA_TELEMETRY_HELPER_DIR="+s.dir, "ACTA_TELEMETRY_HELPER_N=25", "DO_NOT_TRACK=", "ACTA_TELEMETRY=")
		if err := c.Start(); err != nil {
			t.Fatal(err)
		}
		cmds = append(cmds, c)
	}
	for _, c := range cmds {
		if err := c.Wait(); err != nil {
			t.Fatal(err)
		}
	}
	if got := today(t, s).Commands["list"]; got < 1 || got > 50 {
		t.Fatalf("count = %d, want between 1 and 50", got)
	}
}

func TestStoreDropsBrokenDayFiles(t *testing.T) {
	s := openTest(t)
	if err := s.Add("commands", "show"); err != nil {
		t.Fatal(err)
	}
	broken := filepath.Join(s.dir, "days", "2026-01-02.json")
	if err := os.WriteFile(broken, []byte(`{"day": "2026-01-02", "commands": {`), 0o600); err != nil {
		t.Fatal(err)
	}
	wrongDay := filepath.Join(s.dir, "days", "2026-01-03.json")
	if err := os.WriteFile(wrongDay, []byte(`{"day": "2026-01-09"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	days, err := s.Days()
	if err != nil {
		t.Fatal(err)
	}
	if len(days) != 1 || days[0].Day != Today() {
		t.Fatalf("Days() = %v, want only today", days)
	}
	for _, p := range []string{broken, wrongDay} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Fatalf("%s kept: %v", p, err)
		}
	}
}

func TestStoreKeepsSevenDays(t *testing.T) {
	s := openTest(t)
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.Local)
	for i := 0; i < 10; i++ {
		day := base.AddDate(0, 0, i)
		setNow(t, day)
		if err := s.Add("commands", "show"); err != nil {
			t.Fatal(err)
		}
	}
	days, err := s.Days()
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, d := range days {
		names = append(names, d.Day)
	}
	sort.Strings(names)
	if len(names) != MaxDays || names[0] != "2026-01-04" || names[len(names)-1] != "2026-01-10" {
		t.Fatalf("kept days %v, want the newest %d", names, MaxDays)
	}
}

func TestStoreFilesAreOwnerOnly(t *testing.T) {
	s := openTest(t)
	if err := s.Add("commands", "show"); err != nil {
		t.Fatal(err)
	}
	err := filepath.Walk(s.dir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().Perm()&0o077 != 0 {
			t.Errorf("%s has mode %v, want owner only", p, info.Mode().Perm())
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func setNow(t *testing.T, at time.Time) {
	t.Helper()
	old := now
	now = func() time.Time { return at }
	t.Cleanup(func() { now = old })
}

func TestSkillKeysMatchPluginSkills(t *testing.T) {
	entries, err := os.ReadDir(filepath.Join("..", "..", "plugin", "skills"))
	if err != nil {
		t.Fatal(err)
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	var keys []string
	for k := range skillKeys {
		keys = append(keys, k)
	}
	sort.Strings(dirs)
	sort.Strings(keys)
	if strings.Join(dirs, ",") != strings.Join(keys, ",") {
		t.Fatalf("skill keys %v, want the plugin skill folders %v", keys, dirs)
	}
}

func TestStoreBrokenStateReadsAsOff(t *testing.T) {
	s := openTest(t)
	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s.dir, "state.json"), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if s.Enabled() {
		t.Fatal("a broken state file turned telemetry on")
	}
}

func TestStoreSurvivesDayFilesWithMissingMaps(t *testing.T) {
	s := openTest(t)
	if err := s.Add("commands", "show"); err != nil {
		t.Fatal(err)
	}
	id := s.InstallID()
	raw := `{"day": "` + Today() + `", "install_id": "` + id + `", "commands": null}`
	if err := os.WriteFile(s.dayPath(Today()), []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, g := range []string{"commands", "sessions", "skills", "workflow", "blocks"} {
		if err := s.Add(g, "shape"); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.SetSettings(map[string]string{"plan_depth": "full"}); err != nil {
		t.Fatal(err)
	}
	if got := today(t, s).Settings["plan_depth"]; got != "full" {
		t.Fatalf("plan_depth = %q after a file with missing maps", got)
	}
}

func TestOffAndDaysStillWorkWithoutAnEndpoint(t *testing.T) {
	s := openTest(t)
	if err := s.Add("commands", "show"); err != nil {
		t.Fatal(err)
	}
	Endpoint = ""
	days, err := s.Days()
	if err != nil || len(days) != 1 {
		t.Fatalf("Days() = %v, %v; want the day a release build stored", days, err)
	}
	if err := s.Off(); err != nil {
		t.Fatal(err)
	}
	if s.InstallID() != "" {
		t.Fatal("off on a dev build kept the install ID")
	}
	if names, _ := s.dayNames(); len(names) != 0 {
		t.Fatalf("off on a dev build kept days %v", names)
	}
	Endpoint = "https://script.google.com/macros/s/test/exec"
	if s.Enabled() {
		t.Fatal("off on a dev build did not stick for the release build")
	}
}
