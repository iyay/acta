package tui

import (
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
)

func TestWatchDirs(t *testing.T) {
	cfg := config.Default("/r")
	want := []string{"/r/.acta", "/r/.acta/specs", "/r/.acta/plans", "/r/.acta/bugs",
		"/r/docs/superpowers", "/r/docs/superpowers/specs", "/r/docs/superpowers/plans", "/r/docs/superpowers/bugs"}
	if got := WatchDirs(cfg); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v", got)
	}
}

func TestWatchGathersEventsIntoOneReload(t *testing.T) {
	dir := t.TempDir()
	var mu sync.Mutex
	loads := 0
	msgs := make(chan tea.Msg, 10)
	load := func() (*board.Board, error) {
		mu.Lock()
		loads++
		mu.Unlock()
		return &board.Board{}, nil
	}
	stop, err := Watch(func() []string { return []string{dir, filepath.Join(dir, "missing")} }, load, func(m tea.Msg) { msgs <- m })
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
	dir := t.TempDir()
	later := filepath.Join(dir, "later")
	msgs := make(chan tea.Msg, 10)
	stop, err := Watch(func() []string { return []string{dir, later} },
		func() (*board.Board, error) { return &board.Board{}, nil }, func(m tea.Msg) { msgs <- m })
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
