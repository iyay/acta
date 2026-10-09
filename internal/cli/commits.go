package cli

import (
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/commits"
)

// commitRow is one commit of a plan, with every task of that plan it names.
type commitRow struct {
	sha     string
	date    time.Time
	tasks   []int
	subject string
	branch  string
	chore   bool
}

// commitJSON is the --json shape of one commit.
type commitJSON struct {
	Sha     string `json:"sha"`
	Date    string `json:"date"`
	Tasks   []int  `json:"tasks"`
	Subject string `json:"subject"`
	Branch  string `json:"branch"`
	Chore   bool   `json:"chore"`
}

// cmdCommits lists the commits that name a plan or one of its tasks in a
// Task: trailer, oldest first. Chore commits stay hidden unless --all.
func cmdCommits(args []string, stdout, stderr io.Writer) int {
	fs, root := flags("commits", stderr)
	all := fs.Bool("all", false, "include chore commits")
	asJSON := fs.Bool("json", false, "print JSON")
	usage := func() int {
		fmt.Fprintln(stderr, "usage: acta commits <plan> [task] [--all] [--json]")
		return exitBadInput
	}
	pos, err := parseMixed(fs, args)
	if err != nil || len(pos) < 1 || len(pos) > 2 {
		return usage()
	}
	only := 0
	if len(pos) == 2 {
		n, err := strconv.Atoi(pos[1])
		if err != nil || n < 1 {
			fmt.Fprintf(stderr, "bad task number %s\n", pos[1])
			return exitBadInput
		}
		only = n
	}
	cfg, b, code := loadAllTrees(*root, stderr)
	if code != exitOK {
		return code
	}
	plan := b.Get(pos[0])
	if plan == nil || plan.Kind != board.KindPlan {
		fmt.Fprintf(stderr, "%s is not a plan\n", pos[0])
		return exitBadInput
	}
	tasks := planTaskNumbers(b, plan)
	if only != 0 && !slices.Contains(tasks, only) {
		fmt.Fprintf(stderr, "%s has no task %d\n", pos[0], only)
		return exitBadInput
	}
	if only != 0 {
		tasks = []int{only}
	}
	var found map[string][]commits.Commit
	// A plan with no hash cannot be named in a trailer, so nothing links to it.
	if plan.RawHash != "" {
		found, err = commits.Find(cfg.RepoRoot, commits.RefsFor(cfg))
		if err != nil {
			fmt.Fprintln(stderr, err)
			return exitOther
		}
	}
	rows := gatherCommitRows(found, plan.RawHash, tasks, *all)
	if *asJSON {
		out := make([]commitJSON, 0, len(rows))
		for _, r := range rows {
			out = append(out, commitJSON{Sha: r.sha, Date: r.date.Format(time.RFC3339),
				Tasks: r.tasks, Subject: r.subject, Branch: r.branch, Chore: r.chore})
		}
		return printJSON(stdout, stderr, out)
	}
	if len(rows) == 0 {
		fmt.Fprintln(stdout, "no linked commits")
		return exitOK
	}
	for _, r := range rows {
		marks := make([]string, len(r.tasks))
		for i, n := range r.tasks {
			marks[i] = "#" + strconv.Itoa(n)
		}
		fmt.Fprintf(stdout, "%s  %s  %s  %s  (%s)\n", r.sha[:min(7, len(r.sha))],
			r.date.Format("2006-01-02"), strings.Join(marks, ","), r.subject, r.branch)
	}
	return exitOK
}

// planTaskNumbers gives the task numbers of a plan, lowest first. A task whose
// number is not a plain integer cannot be named in a trailer, so it is left out.
func planTaskNumbers(b *board.Board, plan *board.Item) []int {
	var nums []int
	for _, id := range plan.Children {
		t := b.Get(id)
		if t == nil || t.Kind != board.KindTask {
			continue
		}
		if n, err := strconv.Atoi(t.TaskNum); err == nil && !slices.Contains(nums, n) {
			nums = append(nums, n)
		}
	}
	slices.Sort(nums)
	return nums
}

// gatherCommitRows keeps the commits of the given tasks, folds one commit that
// names two tasks into one row, and sorts the rows oldest first.
func gatherCommitRows(found map[string][]commits.Commit, hash string, tasks []int, all bool) []commitRow {
	var rows []commitRow
	index := map[string]int{}
	for _, n := range tasks {
		for _, c := range found[hash+"#"+strconv.Itoa(n)] {
			if c.Chore && !all {
				continue
			}
			if i, ok := index[c.Sha]; ok {
				rows[i].tasks = append(rows[i].tasks, n)
				continue
			}
			index[c.Sha] = len(rows)
			rows = append(rows, commitRow{sha: c.Sha, date: c.Date, tasks: []int{n},
				subject: c.Subject, branch: c.Branch, chore: c.Chore})
		}
	}
	slices.SortStableFunc(rows, func(a, b commitRow) int { return a.date.Compare(b.date) })
	return rows
}
