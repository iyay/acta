package tidy

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// hashWord finds short and full hashes standing alone in text.
var hashWord = regexp.MustCompile(`\b[0-9a-f]{7,40}\b`)

// remapText swaps old commit hashes for new ones at the same length.
// olds lists every old full hash. newOf maps an old full hash to its new one.
// A word that matches no old hash, or more than one, stays as it is.
func remapText(text string, olds []string, newOf map[string]string) string {
	return hashWord.ReplaceAllStringFunc(text, func(word string) string {
		var hit string
		for _, o := range olds {
			if strings.HasPrefix(o, word) {
				if hit != "" {
					return word
				}
				hit = o
			}
		}
		n, ok := newOf[hit]
		if hit == "" || !ok {
			return word
		}
		return n[:len(word)]
	})
}

// treeEntry is one file of a tree.
type treeEntry struct{ mode, sha, path string }

// planningFiles lists the plain files under root in a tree.
func planningFiles(repo, tree, root string) ([]treeEntry, error) {
	out, err := git(repo, nil, "", "ls-tree", "-r", "-z", tree, "--", root)
	if err != nil {
		return nil, err
	}
	var files []treeEntry
	for _, rec := range strings.Split(out, "\x00") {
		meta, path, ok := strings.Cut(rec, "\t")
		if !ok {
			continue
		}
		f := strings.Fields(meta)
		if len(f) != 3 || f[1] != "blob" || (f[0] != "100644" && f[0] != "100755") {
			continue
		}
		files = append(files, treeEntry{mode: f[0], sha: f[2], path: path})
	}
	return files, nil
}

// readBlobs reads many blobs with one git process.
func readBlobs(repo string, shas []string) (map[string][]byte, error) {
	out, err := git(repo, nil, strings.Join(shas, "\n")+"\n", "cat-file", "--batch")
	if err != nil {
		return nil, err
	}
	rd := bufio.NewReader(strings.NewReader(out))
	blobs := make(map[string][]byte, len(shas))
	for range shas {
		head, err := rd.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("cat-file --batch: %w", err)
		}
		f := strings.Fields(head)
		if len(f) != 3 {
			return nil, fmt.Errorf("cat-file --batch: odd line %q", head)
		}
		size, err := strconv.Atoi(f[2])
		if err != nil {
			return nil, err
		}
		body := make([]byte, size+1)
		if _, err := readFull(rd, body); err != nil {
			return nil, err
		}
		blobs[f[0]] = body[:size]
	}
	return blobs, nil
}

func readFull(rd *bufio.Reader, buf []byte) (int, error) {
	n := 0
	for n < len(buf) {
		m, err := rd.Read(buf[n:])
		n += m
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

// remapBlob returns the new text of a file, or false when it stays the same
// or is not text.
func remapBlob(body []byte, olds []string, newOf map[string]string) ([]byte, bool) {
	if bytes.IndexByte(body, 0) >= 0 {
		return nil, false
	}
	next := remapText(string(body), olds, newOf)
	return []byte(next), next != string(body)
}

// remapTree returns a tree like baseTree with old hashes swapped inside the
// planning root. It also returns the edited paths.
func remapTree(repo, baseTree, root string, olds []string, newOf map[string]string) (string, []string, error) {
	files, err := planningFiles(repo, baseTree, root)
	if err != nil || len(files) == 0 {
		return baseTree, nil, err
	}
	shas := make([]string, len(files))
	for i, f := range files {
		shas[i] = f.sha
	}
	blobs, err := readBlobs(repo, shas)
	if err != nil {
		return "", nil, err
	}
	var info strings.Builder
	var edited []string
	for _, f := range files {
		next, changed := remapBlob(blobs[f.sha], olds, newOf)
		if !changed {
			continue
		}
		sha, err := git(repo, nil, string(next), "hash-object", "-w", "--stdin")
		if err != nil {
			return "", nil, err
		}
		fmt.Fprintf(&info, "%s %s\t%s\n", f.mode, strings.TrimSpace(sha), f.path)
		edited = append(edited, f.path)
	}
	if len(edited) == 0 {
		return baseTree, nil, nil
	}
	tree, err := treeWithEdits(repo, baseTree, info.String())
	return tree, edited, err
}

// treeWithEdits writes a tree from baseTree plus index-info lines, using a
// throwaway index so the real one is never touched.
func treeWithEdits(repo, baseTree, info string) (string, error) {
	idx, err := tempIndex()
	if err != nil {
		return "", err
	}
	defer removeIndex(idx)
	env := []string{"GIT_INDEX_FILE=" + idx}
	if _, err := git(repo, env, "", "read-tree", baseTree); err != nil {
		return "", err
	}
	if _, err := git(repo, env, info, "update-index", "--index-info"); err != nil {
		return "", err
	}
	out, err := git(repo, env, "", "write-tree")
	return strings.TrimSpace(out), err
}
