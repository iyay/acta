package tui

import (
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
)

// Debounce is how long file events are gathered before one reload.
var Debounce = 200 * time.Millisecond

// WatchDirs lists every folder whose files the board reads.
func WatchDirs(cfg config.Config) []string {
	dirs := []string{cfg.Root,
		filepath.Join(cfg.Root, cfg.Dirs.Specs),
		filepath.Join(cfg.Root, cfg.Dirs.Plans),
		filepath.Join(cfg.Root, cfg.Dirs.Bugs)}
	for _, l := range cfg.Legacy {
		dirs = append(dirs, l, filepath.Join(l, "specs"), filepath.Join(l, "plans"), filepath.Join(l, "bugs"))
	}
	return dirs
}

// Watch reloads the board after files change and hands the result to send.
// dirs is asked again after every reload, so a folder or worktree that appears
// later is watched from then on. Missing folders are skipped.
func Watch(dirs func() []string, load func() (*board.Board, error), send func(tea.Msg)) (func(), error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	watched := map[string]bool{}
	addAll := func() {
		for _, d := range dirs() {
			if watched[d] {
				continue
			}
			if st, err := os.Stat(d); err != nil || !st.IsDir() {
				continue
			}
			if w.Add(d) == nil {
				watched[d] = true
			}
		}
	}
	addAll()
	done := make(chan struct{})
	fire := make(chan struct{}, 1)
	go func() {
		var timer *time.Timer
		for {
			select {
			case <-done:
				return
			case _, ok := <-w.Events:
				if !ok {
					return
				}
				if timer != nil {
					timer.Stop()
				}
				timer = time.AfterFunc(Debounce, func() {
					select {
					case fire <- struct{}{}:
					default:
					}
				})
			case err, ok := <-w.Errors:
				if !ok {
					return
				}
				send(watchFailedMsg{err: err})
			case <-fire:
				b, err := load()
				addAll()
				send(reloadMsg{b: b, err: err})
			}
		}
	}()
	return func() { close(done); w.Close() }, nil
}
