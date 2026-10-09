package tui

import (
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/iyay/acta/internal/commits"
)

// diffFake stands in for git show: it answers from a map and writes down which
// shas were asked for.
type diffFake struct {
	mu    sync.Mutex
	texts map[string]string
	errs  map[string]error
	asked []string
}

func (f *diffFake) show(_, sha string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.asked = append(f.asked, sha)
	if err := f.errs[sha]; err != nil {
		return "", err
	}
	if t, ok := f.texts[sha]; ok {
		return t, nil
	}
	return "diff --git a/f b/f\n@@ -1 +1 @@\n-old " + sha[:7] + "\n+new " + sha[:7] + "\n", nil
}

func (f *diffFake) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.asked)
}

func swapShowDiff(t *testing.T, f *diffFake) {
	t.Helper()
	old := showDiff
	showDiff = f.show
	t.Cleanup(func() { showDiff = old })
}

// pagerFake notes what the o key asked the pager seam for.
type pagerFake struct{ repo, sha string }

func swapPager(t *testing.T, f *pagerFake) {
	t.Helper()
	old := pagerCmd
	pagerCmd = func(repo, sha string) *exec.Cmd {
		f.repo, f.sha = repo, sha
		return exec.Command("true")
	}
	t.Cleanup(func() { pagerCmd = old })
}

// mk makes a commit with a 7 letter sha, made h hours after the first of October.
func mk(sha string, h int, subject string) commits.Commit {
	return commits.Commit{
		Sha:     sha + strings.Repeat("0", 40-len(sha)),
		Date:    time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(h) * time.Hour),
		Subject: subject,
		Chore:   strings.HasPrefix(subject, "chore("),
	}
}

// atItem puts the cursor of the board on an item, on whichever tab lists it.
func atItem(t *testing.T, m Model, id string) Model {
	t.Helper()
	it := m.board.Get(id)
	if it == nil {
		t.Fatalf("the board holds no %s", id)
	}
	for i := range topTabs {
		m.openTab(i)
		for d := range max(1, len(topTabs[i].done)) {
			m.done = d
			for _, p := range m.panes() {
				rows, _, _ := m.slotOf(p)
				for n, r := range rows {
					if r.id == it.ID {
						m.focus, m.last = p, p
						m.sel[p], m.idx[p] = it.ID, n
						return m
					}
				}
			}
		}
	}
	t.Fatalf("%s is on no list of the board", id)
	return m
}

// settle runs the command a key gave back and hands the diffs it loaded to the
// model, the way the program would.
func settle(m Model, cmd tea.Cmd) Model {
	for _, msg := range runNow(cmd, time.Second) {
		if d, ok := msg.(diffLoadedMsg); ok {
			next, _ := m.Update(d)
			m = next.(Model)
		}
	}
	return m
}

// tap presses keys one by one and settles the diff loads after each.
func tap(m Model, keys ...string) Model {
	for _, k := range keys {
		next, cmd := m.Update(key(k))
		m = settle(next.(Model), cmd)
	}
	return m
}

func screenLines(m Model) []string { return strings.Split(plain(m.View()), "\n") }

func cscreen(m Model) string { return plain(m.View()) }

// longDiff is a diff with n added lines, so there is something to scroll.
func longDiff(n int) string {
	var b strings.Builder
	b.WriteString("diff --git a/f b/f\n@@ -1 +1 @@\n")
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&b, "+line %03d\n", i)
	}
	return b.String()
}

// openOn builds the model, puts the cursor on an item and presses d.
func openOn(t *testing.T, found map[string][]commits.Commit, id string, w, h int) Model {
	t.Helper()
	m := loadedModel(t, found)
	m = sized(atItem(t, m, id), w, h)
	return tap(m, "d")
}

func TestCommitScreenOpensOnATaskWithListAndColoredDiff(t *testing.T) {
	swapShowDiff(t, &diffFake{})
	found := map[string][]commits.Commit{"abcdefg#1": {mk("aaaaaaa", 1, "feat: one"), mk("bbbbbbb", 2, "feat: two")}}
	m := loadedModel(t, found)
	m = atItem(t, m, "PLN-0001.01")
	boardView := cscreen(m)
	m = tap(m, "d")
	if m.commitScreen == nil {
		t.Fatal("d on a task did not open the screen")
	}
	lines := screenLines(m)
	// 35% of 160 columns is 56, so the list sits left of it and the diff right.
	var listAt, diffAt = -1, -1
	for _, ln := range lines {
		if i := strings.Index(ln, "bbbbbbb  feat: two"); i >= 0 {
			listAt = i
		}
		if i := strings.Index(ln, "+new bbbbbbb"); i >= 0 {
			diffAt = i
		}
	}
	if listAt < 0 || listAt >= 56 {
		t.Errorf("want the newest commit in the list on the left, got column %d:\n%s", listAt, cscreen(m))
	}
	if diffAt < 56 {
		t.Errorf("want the diff of the pick on the right, got column %d:\n%s", diffAt, cscreen(m))
	}
	if strings.Index(cscreen(m), "aaaaaaa") < strings.Index(cscreen(m), "bbbbbbb  feat") {
		t.Errorf("want newest first")
	}
	if got := cscreen(m); strings.Contains(got, "Detail") || strings.Contains(got, "Activity") && strings.Contains(got, "Done") {
		t.Errorf("the screen must replace every pane:\n%s", got)
	}

	m = tap(m, "esc")
	if m.commitScreen != nil {
		t.Fatal("esc did not close the screen")
	}
	if got := cscreen(m); got != boardView {
		t.Errorf("the board after esc is not the board before d:\nbefore:\n%s\nafter:\n%s", boardView, got)
	}
}

func TestCommitScreenColorsTheDiff(t *testing.T) {
	swapShowDiff(t, &diffFake{})
	withColors(func() {
		m := openOn(t, map[string][]commits.Commit{"abcdefg#1": {mk("aaaaaaa", 1, "feat: one")}}, "PLN-0001.01", 160, 40)
		if got := m.View(); !strings.Contains(got, m.styles.done.Render("+new aaaaaaa")) {
			t.Errorf("the added line is not painted with the done brush:\n%q", got)
		}
	})
}

func TestCommitScreenOpensOnAPlanWithTaskLabels(t *testing.T) {
	swapShowDiff(t, &diffFake{})
	both := mk("sssssss", 5, "feat: shared")
	found := map[string][]commits.Commit{
		"abcdefg#1": {mk("aaaaaaa", 1, "feat: one"), both},
		"abcdefg#2": {mk("bbbbbbb", 2, "feat: two"), both},
		"abcdefg#3": {mk("ccccccc", 3, "feat: three")},
	}
	m := openOn(t, found, "PLN-0001", 160, 40)
	if m.commitScreen == nil {
		t.Fatal("d on a plan did not open the screen")
	}
	got := cscreen(m)
	for _, want := range []string{"aaaaaaa  #1  feat: one", "bbbbbbb  #2  feat: two", "ccccccc  #3  feat: three", "sssssss  #1,#2  feat: shared"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing row %q:\n%s", want, got)
		}
	}
	if n := strings.Count(got, "feat: shared"); n != 1 {
		t.Errorf("a shared commit must be one row, got %d", n)
	}
	// Newest first: shared (5h) above three (3h) above two above one.
	if !(strings.Index(got, "sssssss") < strings.Index(got, "ccccccc") && strings.Index(got, "ccccccc") < strings.Index(got, "bbbbbbb") && strings.Index(got, "bbbbbbb") < strings.Index(got, "aaaaaaa")) {
		t.Errorf("rows are not newest first:\n%s", got)
	}
}

func TestCommitScreenDOnOtherRowsDoesNothing(t *testing.T) {
	swapShowDiff(t, &diffFake{})
	all := map[string][]commits.Commit{"abcdefg#1": {mk("aaaaaaa", 1, "feat: one")}}
	for _, id := range []string{"SPEC-1", "BUG-1", "DEBT-1.1"} {
		m := atItem(t, loadedModel(t, all), id)
		next, cmd := m.Update(key("d"))
		got := next.(Model)
		if got.commitScreen != nil || cmd != nil {
			t.Errorf("d on %s opened the screen or ran a command", id)
		}
	}
}

func TestCommitScreenOpensFromTheDetailPane(t *testing.T) {
	swapShowDiff(t, &diffFake{})
	m := loadedModel(t, map[string][]commits.Commit{"abcdefg#1": {mk("aaaaaaa", 1, "feat: one")}})
	m = atItem(t, m, "PLN-0001.01")
	m.focus = paneDetail
	m = tap(m, "d")
	if m.commitScreen == nil {
		t.Fatal("d with the detail pane focused did not open the screen")
	}
	m = tap(m, "esc")
	if m.commitScreen != nil || m.focus != paneDetail {
		t.Errorf("esc must give back the detail focus, got focus %v", m.focus)
	}
}

func TestCommitScreenEmptyItemSaysSo(t *testing.T) {
	f := &diffFake{}
	swapShowDiff(t, f)
	m := openOn(t, nil, "PLN-0001.04", 160, 40)
	if m.commitScreen == nil {
		t.Fatal("an item with no commits must still open the screen")
	}
	if !strings.Contains(cscreen(m), "no linked commits") {
		t.Errorf("want the empty note:\n%s", cscreen(m))
	}
	for _, k := range []string{"j", "k", "tab", "j", "ctrl+d", "g", "G", "c", "o"} {
		next, cmd := m.Update(key(k))
		m = next.(Model)
		if cmd != nil {
			t.Errorf("%s on an empty screen ran a command", k)
		}
	}
	if f.count() != 0 {
		t.Errorf("no commit was picked, yet git show ran %d times", f.count())
	}
	if m = tap(m, "esc"); m.commitScreen != nil {
		t.Error("esc did not close the empty screen")
	}
}

func TestCommitScreenPickMovesAndDiffFollows(t *testing.T) {
	f := &diffFake{}
	swapShowDiff(t, f)
	found := map[string][]commits.Commit{"abcdefg#1": {mk("aaaaaaa", 1, "feat: one"), mk("bbbbbbb", 2, "feat: two"), mk("ccccccc", 3, "feat: three")}}
	m := openOn(t, found, "PLN-0001.01", 160, 40)
	if !strings.Contains(cscreen(m), "+new ccccccc") {
		t.Fatalf("the diff must start on the newest commit:\n%s", cscreen(m))
	}
	m = tap(m, "j")
	if !strings.Contains(cscreen(m), "+new bbbbbbb") || m.commitScreen.pick != "bbbbbbb"+strings.Repeat("0", 33) {
		t.Errorf("j did not move the pick:\n%s", cscreen(m))
	}
	m = tap(m, "down")
	if !strings.Contains(cscreen(m), "+new aaaaaaa") {
		t.Errorf("down did not move the pick")
	}
	m = tap(m, "j")
	if !strings.Contains(cscreen(m), "+new aaaaaaa") {
		t.Errorf("j at the last row must stay")
	}
	m = tap(m, "k", "up")
	if !strings.Contains(cscreen(m), "+new ccccccc") {
		t.Errorf("k and up did not move back to the top")
	}
	// Three commits, three loads: a pick that was loaded once is cached.
	if f.count() != 3 {
		t.Errorf("git show ran %d times, want 3", f.count())
	}
	m = tap(m, "G")
	if !strings.Contains(cscreen(m), "+new aaaaaaa") {
		t.Errorf("G did not go to the last row")
	}
	m = tap(m, "g")
	if !strings.Contains(cscreen(m), "+new ccccccc") {
		t.Errorf("g did not go to the first row")
	}
	if f.count() != 3 {
		t.Errorf("a cached diff was read again: %d runs", f.count())
	}
}

func TestCommitScreenListPageKeys(t *testing.T) {
	swapShowDiff(t, &diffFake{})
	many := fakeCommits("a", 30)
	m := openOn(t, map[string][]commits.Commit{"abcdefg#1": many}, "PLN-0001.01", 160, 20)
	// Newest first: row 0 is change 30.
	m = tap(m, "ctrl+d")
	if m.commitScreen.at != pageLines {
		t.Errorf("ctrl+d moved the pick to %d, want %d", m.commitScreen.at, pageLines)
	}
	if !strings.Contains(cscreen(m), "feat: a change 20") {
		t.Errorf("the pick is not on screen after ctrl+d:\n%s", cscreen(m))
	}
	m = tap(m, "ctrl+u", "ctrl+u")
	if m.commitScreen.at != 0 {
		t.Errorf("ctrl+u went to %d, want 0", m.commitScreen.at)
	}
	m = tap(m, "G")
	if m.commitScreen.at != 29 || !strings.Contains(cscreen(m), "feat: a change 1\n") && !strings.Contains(cscreen(m), "feat: a change 1 ") {
		t.Errorf("G did not show the last row (at %d):\n%s", m.commitScreen.at, cscreen(m))
	}
}

func TestCommitScreenTabMovesFocusAndKeysScrollTheDiff(t *testing.T) {
	swapShowDiff(t, &diffFake{texts: map[string]string{"aaaaaaa" + strings.Repeat("0", 33): longDiff(120)}})
	m := openOn(t, map[string][]commits.Commit{"abcdefg#1": {mk("aaaaaaa", 1, "feat: one"), mk("bbbbbbb", 0, "feat: zero")}}, "PLN-0001.01", 160, 20)
	if m.commitScreen.diffFocus {
		t.Fatal("the focus must start on the list")
	}
	m = tap(m, "tab")
	if !m.commitScreen.diffFocus {
		t.Fatal("tab did not move the focus to the diff")
	}
	pick := m.commitScreen.pick
	m = tap(m, "j", "j", "j")
	if m.commitScreen.diffOff != 3 || m.commitScreen.pick != pick {
		t.Errorf("j on the diff: off %d pick %q, want 3 and the same pick", m.commitScreen.diffOff, m.commitScreen.pick)
	}
	if !strings.Contains(cscreen(m), "+line 002") || strings.Contains(cscreen(m), "+line 001") && !strings.Contains(cscreen(m), "+line 004") {
		t.Errorf("the diff did not scroll:\n%s", cscreen(m))
	}
	m = tap(m, "k")
	if m.commitScreen.diffOff != 2 {
		t.Errorf("k: off %d, want 2", m.commitScreen.diffOff)
	}
	m = tap(m, "ctrl+d")
	if m.commitScreen.diffOff != 2+pageLines {
		t.Errorf("ctrl+d: off %d, want %d", m.commitScreen.diffOff, 2+pageLines)
	}
	m = tap(m, "ctrl+u", "ctrl+u")
	if m.commitScreen.diffOff != 0 {
		t.Errorf("ctrl+u must stop at the top, got %d", m.commitScreen.diffOff)
	}
	m = tap(m, "G")
	_, d, _ := m.commitBoxes()
	want := 122 - d.inner // 120 added lines plus the two header lines
	if m.commitScreen.diffOff != want {
		t.Errorf("G: off %d, want %d", m.commitScreen.diffOff, want)
	}
	if !strings.Contains(cscreen(m), "+line 120") {
		t.Errorf("G did not show the last diff line:\n%s", cscreen(m))
	}
	m = tap(m, "j")
	if m.commitScreen.diffOff != want {
		t.Errorf("j below the end moved the diff to %d", m.commitScreen.diffOff)
	}
	m = tap(m, "g")
	if m.commitScreen.diffOff != 0 {
		t.Errorf("g: off %d, want 0", m.commitScreen.diffOff)
	}
	m = tap(m, "tab")
	if m.commitScreen.diffFocus {
		t.Error("a second tab did not give the focus back to the list")
	}
	// A new pick starts its diff at the top.
	m = tap(m, "tab", "j", "j", "tab", "j")
	if m.commitScreen.diffOff != 0 {
		t.Errorf("a new pick must show its diff from the top, got %d", m.commitScreen.diffOff)
	}
}

func TestCommitScreenWheelScrollsTheBoxUnderThePointer(t *testing.T) {
	swapShowDiff(t, &diffFake{texts: map[string]string{"a60" + strings.Repeat("0", 40): longDiff(120)}})
	m := openOn(t, map[string][]commits.Commit{"abcdefg#1": fakeCommits("a", 60)}, "PLN-0001.01", 160, 20)
	boardOff := slices.Clone(m.off)
	before := m.commitScreen.pick

	m = wheel(m, 100, 6, false)
	if m.commitScreen.diffOff != wheelStep {
		t.Errorf("wheel down on the diff: off %d, want %d", m.commitScreen.diffOff, wheelStep)
	}
	m = wheel(m, 100, 6, true)
	m = wheel(m, 100, 6, true)
	if m.commitScreen.diffOff != 0 {
		t.Errorf("wheel up must stop at the top, got %d", m.commitScreen.diffOff)
	}
	if m.commitScreen.listOff != 0 {
		t.Errorf("a notch on the diff moved the list")
	}

	m = wheel(m, 10, 6, false)
	if m.commitScreen.listOff != wheelStep {
		t.Errorf("wheel down on the list: off %d, want %d", m.commitScreen.listOff, wheelStep)
	}
	if m.commitScreen.diffOff != 0 || m.commitScreen.pick != before {
		t.Errorf("a notch on the list moved the pick or the diff")
	}
	for range 100 {
		m = wheel(m, 10, 6, false)
	}
	l, _, _ := m.commitBoxes()
	if want := 60 - l.inner; m.commitScreen.listOff != want {
		t.Errorf("the list scrolled to %d, want the end at %d", m.commitScreen.listOff, want)
	}

	// A notch on the tab bar or the status line is on neither box.
	off := m.commitScreen.listOff
	m = wheel(m, 10, 0, true)
	m = wheel(m, 10, m.height-1, true)
	if m.commitScreen.listOff != off {
		t.Errorf("a notch outside the boxes scrolled the list")
	}
	if !slices.Equal(m.off, boardOff) {
		t.Errorf("the wheel moved the board behind the screen: %v -> %v", boardOff, m.off)
	}
	if m.commitScreen == nil {
		t.Error("the wheel closed the screen")
	}
}

func TestCommitScreenClicksDoNothing(t *testing.T) {
	swapShowDiff(t, &diffFake{})
	m := openOn(t, map[string][]commits.Commit{"abcdefg#1": fakeCommits("a", 5)}, "PLN-0001.01", 160, 40)
	pick, top := m.commitScreen.pick, m.top
	// A click on a tab name, on a list row and on the diff.
	for _, at := range [][2]int{{3, 1}, {5, 5}, {100, 6}, {10, m.height - 1}} {
		m = click(m, at[0], at[1])
	}
	if m.commitScreen == nil || m.commitScreen.pick != pick || m.top != top {
		t.Errorf("a click acted on the screen or the board behind it")
	}
}

func TestCommitScreenChoreToggleKeepsThePick(t *testing.T) {
	swapShowDiff(t, &diffFake{})
	found := map[string][]commits.Commit{"abcdefg#1": {
		mk("aaaaaaa", 1, "feat: one"),
		mk("cccccc1", 2, "chore(plan): tick"),
		mk("bbbbbbb", 3, "feat: two"),
		mk("cccccc2", 4, "chore(plan): tick again"),
	}}
	m := openOn(t, found, "PLN-0001.01", 160, 40)
	got := cscreen(m)
	if strings.Contains(got, "chore(plan)") {
		t.Fatalf("chore commits must start hidden:\n%s", got)
	}
	// Pick the older feat commit, then show the chores: the pick stays on it.
	m = tap(m, "j")
	pick := m.commitScreen.pick
	if !strings.HasPrefix(pick, "aaaaaaa") {
		t.Fatalf("pick = %q, want aaaaaaa", pick)
	}
	m = tap(m, "c")
	got = cscreen(m)
	if !strings.Contains(got, "cccccc1  chore(plan): tick") || !strings.Contains(got, "cccccc2  chore(plan): tick again") {
		t.Errorf("c did not show the chore commits:\n%s", got)
	}
	if m.commitScreen.pick != pick || m.commitScreen.at != 3 {
		t.Errorf("the pick moved off its sha: %q at %d", m.commitScreen.pick, m.commitScreen.at)
	}
	m = tap(m, "c")
	if strings.Contains(cscreen(m), "chore(plan)") || m.commitScreen.pick != pick || m.commitScreen.at != 1 {
		t.Errorf("hiding the chores moved the pick: %q at %d", m.commitScreen.pick, m.commitScreen.at)
	}
	// When the picked chore commit is hidden again, the pick lands on a row that exists.
	m = tap(m, "c", "k")
	if !strings.HasPrefix(m.commitScreen.pick, "cccccc1") {
		t.Fatalf("pick = %q, want the chore commit", m.commitScreen.pick)
	}
	m = tap(m, "c")
	if m.commitScreen.cursor(m.commitScreen.visible()) < 0 || m.commitScreen.pickSha() == "" {
		t.Errorf("hiding the picked chore commit left no pick")
	}
	if !strings.Contains(cscreen(m), "+new "+m.commitScreen.pickSha()[:7]) {
		t.Errorf("the diff does not follow the new pick:\n%s", cscreen(m))
	}
}

func TestCommitScreenOnlyChoreCommitsSaysSo(t *testing.T) {
	swapShowDiff(t, &diffFake{})
	m := openOn(t, map[string][]commits.Commit{"abcdefg#3": {mk("cccccc1", 1, "chore(plan): tick")}}, "PLN-0001.03", 160, 40)
	if !strings.Contains(cscreen(m), "c to show") {
		t.Errorf("want a note that chores are hidden:\n%s", cscreen(m))
	}
	m = tap(m, "c")
	if !strings.Contains(cscreen(m), "cccccc1  chore(plan): tick") {
		t.Errorf("c did not show the chore commit:\n%s", cscreen(m))
	}
}

func TestCommitScreenDiffStates(t *testing.T) {
	f := &diffFake{errs: map[string]error{"aaaaaaa" + strings.Repeat("0", 33): errors.New("git exploded")}}
	swapShowDiff(t, f)
	m := loadedModel(t, map[string][]commits.Commit{"abcdefg#1": {mk("aaaaaaa", 1, "feat: one"), mk("bbbbbbb", 2, "feat: two")}})
	m = sized(atItem(t, m, "PLN-0001.01"), 160, 40)
	// The open key gives a load, and until it comes back the box says so.
	next, cmd := m.Update(key("d"))
	m = next.(Model)
	if cmd == nil || !strings.Contains(cscreen(m), "loading…") {
		t.Fatalf("want a load command and loading… while the diff is on its way:\n%s", cscreen(m))
	}
	m = settle(m, cmd)
	if strings.Contains(cscreen(m), "loading…") || !strings.Contains(cscreen(m), "+new bbbbbbb") {
		t.Errorf("the loaded diff is not shown:\n%s", cscreen(m))
	}
	// A failed load shows its error in the diff box.
	m2 := loadedModel(t, map[string][]commits.Commit{"abcdefg#1": {mk("aaaaaaa", 1, "feat: one")}})
	m2 = sized(atItem(t, m2, "PLN-0001.01"), 160, 40)
	m2 = tap(m2, "d")
	if got := cscreen(m2); !strings.Contains(got, "git exploded") || strings.Contains(got, "loading…") {
		t.Errorf("want the error text in the diff box:\n%s", got)
	}
}

func TestCommitScreenOpensTheBoxWithoutBlockingOnTheLoad(t *testing.T) {
	// The diff loads in a command, not inside the key press.
	f := &diffFake{}
	swapShowDiff(t, f)
	m := loadedModel(t, map[string][]commits.Commit{"abcdefg#1": {mk("aaaaaaa", 1, "feat: one")}})
	m = atItem(t, m, "PLN-0001.01")
	if _, cmd := m.Update(key("d")); cmd == nil || f.count() != 0 {
		t.Errorf("git show ran inside the key press (%d runs) or no load was asked", f.count())
	}
}

func TestCommitScreenPagerRunsGitShowAndTurnsTheMouseBackOn(t *testing.T) {
	swapShowDiff(t, &diffFake{})
	p := &pagerFake{}
	swapPager(t, p)
	m := openOn(t, map[string][]commits.Commit{"abcdefg#1": {mk("aaaaaaa", 1, "feat: one"), mk("bbbbbbb", 2, "feat: two")}}, "PLN-0001.01", 160, 40)
	m = tap(m, "j")
	next, cmd := m.Update(key("o"))
	m = next.(Model)
	if cmd == nil {
		t.Fatal("o gave no command")
	}
	if p.sha != "aaaaaaa"+strings.Repeat("0", 33) || p.repo != m.cfg.RepoRoot {
		t.Errorf("the pager seam got repo %q sha %q", p.repo, p.sha)
	}
	if msgs := runNow(cmd, time.Second); len(msgs) != 1 || fmt.Sprintf("%T", msgs[0]) != "tea.execMsg" {
		t.Errorf("o must hand the terminal to the command with ExecProcess, got %v", msgs)
	}
	if m.commitScreen == nil {
		t.Error("o closed the screen")
	}
	// The pager ran with the program suspended, so the mouse is off.
	want := tea.EnableMouseCellMotion()
	for _, done := range []pagerDoneMsg{{}, {err: errors.New("boom")}} {
		next, cmd := m.Update(done)
		if !slices.Contains(runNow(cmd, time.Second), want) {
			t.Errorf("the mouse stayed off after the pager (err %v)", done.err)
		}
		if done.err != nil && !strings.Contains(next.(Model).status, "boom") {
			t.Errorf("the pager error is not shown")
		}
		if next.(Model).commitScreen == nil {
			t.Errorf("the screen closed when the pager did")
		}
	}
}

func TestCommitScreenPagerSkipsAShaThatStartsWithADash(t *testing.T) {
	swapShowDiff(t, &diffFake{})
	p := &pagerFake{}
	swapPager(t, p)
	m := openOn(t, map[string][]commits.Commit{"abcdefg#1": {mk("-aaaaaa", 1, "feat: one")}}, "PLN-0001.01", 160, 40)
	next, cmd := m.Update(key("o"))
	if cmd != nil || p.sha != "" || next.(Model).commitScreen == nil {
		t.Errorf("o ran a pager (cmd %v, sha %q) or closed the screen on a dash sha", cmd, p.sha)
	}
}

func TestCommitScreenPagerGitShowCommandShape(t *testing.T) {
	cmd := pagerCmd("/some/repo", "abc1234")
	if cmd.Dir != "/some/repo" || !slices.Contains(cmd.Args, "show") || !slices.Contains(cmd.Args, "abc1234") {
		t.Errorf("the default pager command is %v in %q", cmd.Args, cmd.Dir)
	}
}

func TestCommitScreenEscAndQCloseAndKeepTheBoard(t *testing.T) {
	swapShowDiff(t, &diffFake{})
	for _, k := range []string{"esc", "q"} {
		for _, focusDiff := range []bool{false, true} {
			m := loadedModel(t, map[string][]commits.Commit{"abcdefg#2": fakeCommits("a", 3)})
			m = atItem(t, m, "PLN-0001.02")
			m.off[paneDetail] = 2
			wasTop, wasFocus, wasLast, wasSel, wasIdx, wasOff := m.top, m.focus, m.last, slices.Clone(m.sel), slices.Clone(m.idx), slices.Clone(m.off)
			m = tap(m, "d", "j")
			if focusDiff {
				m = tap(m, "tab")
			}
			next, cmd := m.Update(key(k))
			m = next.(Model)
			if m.commitScreen != nil {
				t.Errorf("%s (diff focus %v) did not close the screen", k, focusDiff)
			}
			if cmd != nil {
				t.Errorf("%s ran a command: q on the screen must not quit", k)
			}
			if m.top != wasTop || m.focus != wasFocus || m.last != wasLast || !slices.Equal(m.sel, wasSel) || !slices.Equal(m.idx, wasIdx) || !slices.Equal(m.off, wasOff) {
				t.Errorf("%s changed the board: tab %d/%d focus %v/%v sel %v/%v off %v/%v", k, wasTop, m.top, wasFocus, m.focus, wasSel, m.sel, wasOff, m.off)
			}
		}
	}
	// On the board, q still quits.
	m := loadedModel(t, nil)
	if _, cmd := m.Update(key("q")); cmd == nil {
		t.Error("q on the board must still quit")
	}
}

// Every key the screen has no use for must leave the board and the screen as
// they were: no tab switch, no popup, no editor, no search, no help.
func TestCommitScreenKeysItDoesNotUseDoNothing(t *testing.T) {
	swapShowDiff(t, &diffFake{})
	m := openOn(t, map[string][]commits.Commit{"abcdefg#1": fakeCommits("a", 3)}, "PLN-0001.01", 160, 40)
	pick := m.commitScreen.pick
	top, focus, status := m.top, m.focus, m.status
	for _, k := range []string{"1", "2", "3", "4", "5", "6", "left", "right", "shift+tab", "z", "]", "[", "/", "?", "r", "t", "s", "p", "+", "-", "n", "e", "y", " ", "h", "l", "enter", "d", "x", "backspace"} {
		next, cmd := m.Update(key(k))
		got := next.(Model)
		if cmd != nil {
			t.Errorf("%q ran a command from the screen", k)
		}
		if got.commitScreen == nil || got.commitScreen.pick != pick || got.commitScreen.diffFocus || got.commitScreen.chore {
			t.Errorf("%q changed the screen", k)
		}
		if got.top != top || got.focus != focus || got.status != status || got.popup != nil || got.slug != nil || got.searching || got.help || got.query != "" {
			t.Errorf("%q reached the board behind the screen", k)
		}
	}
}

func TestCommitScreenCtrlCStillQuits(t *testing.T) {
	swapShowDiff(t, &diffFake{})
	m := openOn(t, map[string][]commits.Commit{"abcdefg#1": fakeCommits("a", 1)}, "PLN-0001.01", 160, 40)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("ctrl+c gave no command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("ctrl+c did not quit")
	}
}

func TestCommitScreenLayoutAroundEightyColumns(t *testing.T) {
	swapShowDiff(t, &diffFake{})
	found := map[string][]commits.Commit{"abcdefg#1": {mk("aaaaaaa", 1, "feat: one")}}
	for _, w := range []int{80, 120} {
		m := openOn(t, found, "PLN-0001.01", w, 30)
		listX, diffX := -1, -1
		for _, ln := range screenLines(m) {
			if i := strings.Index(ln, "aaaaaaa  feat: one"); i >= 0 {
				listX = i
			}
			if i := strings.Index(ln, "+new aaaaaaa"); i >= 0 {
				diffX = i
			}
		}
		if listX < 0 || listX >= w*35/100 || diffX < w*35/100 {
			t.Errorf("width %d: want the list left of column %d and the diff right of it, got list %d diff %d", w, w*35/100, listX, diffX)
		}
	}
	for _, w := range []int{79, 60, 40} {
		m := openOn(t, found, "PLN-0001.01", w, 30)
		var listY, diffY = -1, -1
		for y, ln := range screenLines(m) {
			if strings.Contains(ln, "aaaaaaa  feat: one") {
				listY = y
			}
			if strings.Contains(ln, "+new aaaaaaa") {
				diffY = y
			}
		}
		if listY < 0 || diffY <= listY {
			t.Errorf("width %d: want the diff below the list, list row %d diff row %d:\n%s", w, listY, diffY, cscreen(m))
		}
		// The list is about a third of the body: 30 rows less the bars is 26.
		l, d, wide := m.commitBoxes()
		if wide || l.h != 26/3 || d.y != l.y+l.h || l.h+d.h != 26 {
			t.Errorf("width %d: list %+v diff %+v", w, l, d)
		}
	}
}

func TestCommitScreenNeverDrawsPastTheScreen(t *testing.T) {
	swapShowDiff(t, &diffFake{})
	long := mk("aaaaaaa", 1, strings.Repeat("a very long subject ", 12))
	for _, size := range [][2]int{{200, 50}, {120, 40}, {80, 24}, {79, 24}, {60, 20}, {40, 12}, {20, 8}, {10, 5}, {5, 3}, {80, 4}} {
		m := openOn(t, map[string][]commits.Commit{"abcdefg#1": {long, mk("bbbbbbb", 2, "feat: \t tab and wide 日本語 text")}}, "PLN-0001.01", size[0], size[1])
		if m.commitScreen == nil {
			t.Fatalf("%v: the screen did not open", size)
		}
		lines := screenLines(m)
		if len(lines) > size[1] {
			t.Errorf("%v: %d lines, want at most %d", size, len(lines), size[1])
		}
		for _, ln := range lines {
			if w := lipgloss.Width(ln); w > size[0] {
				t.Errorf("%v: a line is %d wide: %q", size, w, ln)
			}
		}
		// Every key and wheel notch must stay inside the screen at any size.
		m = tap(m, "j", "tab", "j", "G", "g", "ctrl+d", "ctrl+u", "c", "tab", "k")
		m = wheel(wheel(m, 1, 4, false), 1, 6, true)
		_ = m.View()
	}
}

func TestCommitScreenReloadKeepsTheScreenAndThePick(t *testing.T) {
	f := &diffFake{}
	swapShowDiff(t, f)
	found := map[string][]commits.Commit{"abcdefg#1": {mk("aaaaaaa", 1, "feat: one"), mk("bbbbbbb", 2, "feat: two")}}
	m := openOn(t, found, "PLN-0001.01", 160, 40)
	m = tap(m, "j")
	pick := m.commitScreen.pick
	loads := f.count()

	// A newer commit lands: it sits on top and the pick stays on its sha.
	more := map[string][]commits.Commit{"abcdefg#1": {mk("aaaaaaa", 1, "feat: one"), mk("bbbbbbb", 2, "feat: two"), mk("ddddddd", 3, "feat: three")}}
	next, cmd := m.Update(reloadMsg{b: m.board, commits: more})
	m = next.(Model)
	if m.commitScreen == nil || m.commitScreen.pick != pick {
		t.Fatalf("the reload closed the screen or moved the pick: %+v", m.commitScreen)
	}
	if !strings.Contains(cscreen(m), "ddddddd  feat: three") {
		t.Errorf("the reload did not bring the new commit:\n%s", cscreen(m))
	}
	// The cache is empty after a reload, so the pick's diff is read again.
	if m.diffCache != nil {
		t.Errorf("the diff cache survived the reload")
	}
	m = settle(m, cmd)
	if f.count() != loads+1 || !strings.Contains(cscreen(m), "+new aaaaaaa") {
		t.Errorf("want the pick's diff read again after a reload (%d -> %d runs)", loads, f.count())
	}

	// The picked commit is gone: the pick keeps its row number.
	less := map[string][]commits.Commit{"abcdefg#1": {mk("bbbbbbb", 2, "feat: two"), mk("ddddddd", 3, "feat: three")}}
	next, cmd = m.Update(reloadMsg{b: m.board, commits: less})
	m = settle(next.(Model), cmd)
	if m.commitScreen == nil || !strings.HasPrefix(m.commitScreen.pick, "bbbbbbb") {
		t.Errorf("want the pick clamped to the last row, got %q", m.commitScreen.pick)
	}

	// Every commit is gone: the screen stays, with its empty note.
	next, cmd = m.Update(reloadMsg{b: m.board})
	m = settle(next.(Model), cmd)
	if m.commitScreen == nil || !strings.Contains(cscreen(m), "no linked commits") {
		t.Errorf("want the empty screen after every commit went:\n%s", cscreen(m))
	}
}

func TestCommitScreenDropsADiffThatComesAfterClose(t *testing.T) {
	swapShowDiff(t, &diffFake{})
	m := loadedModel(t, map[string][]commits.Commit{"abcdefg#1": {mk("aaaaaaa", 1, "feat: one")}})
	m = atItem(t, m, "PLN-0001.01")
	next, cmd := m.Update(key("d"))
	m = next.(Model)
	m = tap(m, "esc")
	m = settle(m, cmd)
	if m.diffCache != nil {
		t.Errorf("a diff that arrived after esc was kept")
	}
}

func TestCommitScreenGatheredBoardNotchesDoNotScrollTheBoard(t *testing.T) {
	swapShowDiff(t, &diffFake{})
	m := loadedModel(t, map[string][]commits.Commit{"abcdefg#1": {mk("aaaaaaa", 1, "feat: one")}})
	m = sized(atItem(t, m, "PLN-0001.01"), 160, 12)
	// Two notches in one frame leave a delta waiting for its tick.
	m = wheelOnly(wheelOnly(m, 100, 6, false), 100, 6, false)
	off := slices.Clone(m.off)
	m = tap(m, "d")
	m = wheelTick(m)
	if !slices.Equal(m.off, off) {
		t.Errorf("a board notch landed on the board behind the screen: %v -> %v", off, m.off)
	}
}

func TestCommitScreenHelpAndHints(t *testing.T) {
	swapShowDiff(t, &diffFake{})
	if !strings.Contains(helpLines, "\nd                open commits\n") {
		t.Errorf("the help has no d line in the key column layout")
	}
	m := loadedModel(t, nil)
	task := atItem(t, m, "PLN-0001.01")
	if !slices.Contains(task.hints(), "Commits: d") {
		t.Errorf("the board hints of a task miss d commits: %v", task.hints())
	}
	if !slices.Contains(atItem(t, m, "PLN-0001").hints(), "Commits: d") {
		t.Errorf("the board hints of a plan miss d commits")
	}
	if slices.Contains(atItem(t, m, "BUG-1").hints(), "Commits: d") {
		t.Errorf("a bug has no commits to open")
	}
	detail := task
	detail.focus = paneDetail
	if !slices.Contains(detail.hints(), "Commits: d") {
		t.Errorf("the detail hints of a task miss d commits: %v", detail.hints())
	}
	// The screen has its own hint bar, with no Help hint.
	open := tap(sized(task, 160, 40), "d")
	lines := screenLines(open)
	bar := lines[len(lines)-1]
	if !strings.Contains(bar, "j/k pick · tab focus · c chore · o pager · esc back") || strings.Contains(bar, "Help: ?") {
		t.Errorf("the hint bar of the screen is %q", bar)
	}
	// A search typed on the board before must not show on the screen.
	open.query = "zzz"
	if bar := screenLines(open); strings.Contains(bar[len(bar)-1], "/zzz") {
		t.Errorf("the board search shows on the screen: %q", bar[len(bar)-1])
	}
}
