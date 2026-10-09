package tui

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
)

func TestWatchDirs(t *testing.T) {
	t.Parallel()

	cfg := config.Default("/r")
	want := []string{"/r/.acta", "/r/.acta/specs", "/r/.acta/plans", "/r/.acta/bugs",
		"/r/docs/superpowers", "/r/docs/superpowers/specs", "/r/docs/superpowers/plans", "/r/docs/superpowers/bugs"}
	if got := WatchDirs(cfg); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v", got)
	}
}

func TestWatchGathersEventsIntoOneReload(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	var mu sync.Mutex
	loads := 0
	msgs := make(chan tea.Msg, 10)
	reload := func() tea.Msg {
		mu.Lock()
		loads++
		mu.Unlock()
		return reloadMsg{b: &board.Board{}}
	}
	stop, err := Watch(func() []string { return []string{dir, filepath.Join(dir, "missing")} }, reload, func(m tea.Msg) { msgs <- m })
	if err != nil {
		t.Fatal(err)
	}
	defer stop()

	for i := range 5 {
		if err := os.WriteFile(filepath.Join(dir, "f.md"), []byte{byte('a' + i)}, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	select {
	case m := <-msgs:
		if _, ok := m.(reloadMsg); !ok {
			t.Fatalf("got %T, want reloadMsg", m)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("no reload after files changed")
	}
	if loads != 1 {
		t.Fatalf("loads = %d, want 1 for a burst of writes", loads)
	}
}

func TestWatchPicksUpFoldersCreatedLater(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	later := filepath.Join(dir, "later")
	msgs := make(chan tea.Msg, 10)
	stop, err := Watch(func() []string { return []string{dir, later} },
		func() tea.Msg { return reloadMsg{b: &board.Board{}} }, func(m tea.Msg) { msgs <- m })
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	if err := os.Mkdir(later, 0o755); err != nil { // an event in dir: first reload, which adds later
		t.Fatal(err)
	}
	<-msgs
	if err := os.WriteFile(filepath.Join(later, "f.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	select {
	case <-msgs:
	case <-time.After(3 * time.Second):
		t.Fatal("a folder created after start was never watched")
	}
}

func TestReloadKeyStillLoadsInManualMode(t *testing.T) {
	t.Parallel()

	m := newModel(t)
	called := false
	m.load = func() (*board.Board, error) { called = true; return m.board, nil }
	_, cmd := m.Update(key("r"))
	if cmd == nil {
		t.Fatal("r should return a reload command")
	}
	if msg := cmd(); msg == nil {
		t.Fatal("reload command returned nil")
	} else if _, ok := msg.(reloadMsg); !ok {
		t.Fatalf("got %T, want reloadMsg", msg)
	}
	if !called {
		t.Fatal("reload command did not call load")
	}
}

func TestStartLoadsOnceAfterTheWatcherIsReady(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	var mu sync.Mutex
	dirsRan := false
	dirsFirst := false
	loads := 0
	msgs := make(chan tea.Msg, 10)
	dirs := func() []string {
		mu.Lock()
		dirsRan = true
		mu.Unlock()
		return []string{dir}
	}
	reload := func() tea.Msg {
		mu.Lock()
		loads++
		dirsFirst = dirsRan
		mu.Unlock()
		return reloadMsg{b: &board.Board{}}
	}
	stop := Start(dirs, reload, func(m tea.Msg) { msgs <- m })
	defer stop()

	select {
	case m := <-msgs:
		if _, ok := m.(reloadMsg); !ok {
			t.Fatalf("got %T, want reloadMsg", m)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("no first load after the watcher started")
	}
	select {
	case m := <-msgs:
		t.Fatalf("got a second message %T, want only one load", m)
	case <-time.After(2 * Debounce):
	}
	mu.Lock()
	defer mu.Unlock()
	if loads != 1 {
		t.Fatalf("loads = %d, want 1", loads)
	}
	if !dirsFirst {
		t.Fatal("the load ran before the watcher asked for its folders")
	}
}

func TestStartStopWaitsForTheSetupToEnd(t *testing.T) {
	t.Parallel()

	release := make(chan struct{})
	msgs := make(chan tea.Msg, 10)
	dirs := func() []string {
		<-release
		return nil
	}
	reload := func() tea.Msg { return reloadMsg{b: &board.Board{}} }
	stop := Start(dirs, reload, func(m tea.Msg) { msgs <- m })

	stopped := make(chan struct{})
	go func() { stop(); close(stopped) }()
	select {
	case <-stopped:
		t.Fatal("stop returned while the setup was still running")
	case <-time.After(100 * time.Millisecond):
	}
	close(release)
	select {
	case <-stopped:
	case <-time.After(3 * time.Second):
		t.Fatal("stop never returned after the setup ended")
	}
}

// It swaps a package variable, so it must not run in parallel. Go holds every
// parallel test until the serial ones are done, so the others never see the swap.
func TestStartStillLoadsWhenTheWatcherFails(t *testing.T) {
	boom := errors.New("no watch slots")
	old := newWatcher
	newWatcher = func() (*fsnotify.Watcher, error) { return nil, boom }
	t.Cleanup(func() { newWatcher = old })

	var mu sync.Mutex
	loads := 0
	msgs := make(chan tea.Msg, 10)
	reload := func() tea.Msg {
		mu.Lock()
		loads++
		mu.Unlock()
		return reloadMsg{b: &board.Board{}}
	}
	stop := Start(func() []string { return nil }, reload, func(m tea.Msg) { msgs <- m })

	select {
	case m := <-msgs:
		failed, ok := m.(watchFailedMsg)
		if !ok {
			t.Fatalf("first message is %T, want watchFailedMsg", m)
		}
		if !errors.Is(failed.err, boom) {
			t.Fatalf("watchFailedMsg err = %v, want %v", failed.err, boom)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("no watchFailedMsg after the watcher failed to start")
	}
	select {
	case m := <-msgs:
		if _, ok := m.(reloadMsg); !ok {
			t.Fatalf("second message is %T, want reloadMsg", m)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("no load after the watcher failed to start")
	}

	stopped := make(chan struct{})
	go func() { stop(); close(stopped) }()
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("stop did not return after the watcher failed to start")
	}
	mu.Lock()
	defer mu.Unlock()
	if loads != 1 {
		t.Fatalf("loads = %d, want 1", loads)
	}
}

func TestInitStartsNoLoad(t *testing.T) {
	t.Parallel()

	m := newModel(t)
	var mu sync.Mutex
	loads := 0
	m.load = func() (*board.Board, error) {
		mu.Lock()
		loads++
		mu.Unlock()
		return m.board, nil
	}
	runNow(m.Init(), 200*time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if loads != 0 {
		t.Fatalf("Init loaded the board %d times, want 0", loads)
	}
}
