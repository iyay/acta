package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// RepoKeys are the user settings a repo may set for itself in .acta.yaml,
// in the order config show prints them. Everything else in User belongs to
// the person and follows them from repo to repo.
var RepoKeys = []string{"repo_language", "build_executor", "plan_depth", "commit_history", "coding_guide"}

// personalKeys may never sit in .acta.yaml: that file is committed, so a
// value there would change how the agent talks to everyone who clones it.
var personalKeys = []string{"chat_language", "style", "tone", "theme", "subagent_models"}

// MergeRepo lays the repo keys of repoRoot/.acta.yaml over v. The map holds
// each key the repo file set, so show can say where a value came from. On
// any error v comes back as it was, so a hook can still use it.
func MergeRepo(v User, repoRoot string) (User, map[string]bool, error) {
	path := filepath.Join(repoRoot, ".acta.yaml")
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return v, map[string]bool{}, nil
	}
	if err != nil {
		return v, nil, err
	}
	var keys map[string]any
	if err := yaml.Unmarshal(raw, &keys); err != nil {
		return v, nil, fmt.Errorf("%s: %w", path, err)
	}
	for _, k := range personalKeys {
		if _, ok := keys[k]; ok {
			return v, nil, fmt.Errorf("%s: %w: %s is a personal setting; set it with acta config set, not per repo", path, ErrBadUser, k)
		}
	}
	out, from := v, map[string]bool{}
	for _, k := range RepoKeys {
		s, ok := keys[k].(string)
		if !ok || strings.TrimSpace(s) == "" {
			continue
		}
		from[k] = true
		switch k {
		case "repo_language":
			out.RepoLanguage = strings.TrimSpace(s)
		case "build_executor":
			out.BuildExecutor = strings.TrimSpace(s)
		case "plan_depth":
			out.PlanDepth = strings.TrimSpace(s)
		case "commit_history":
			out.CommitHistory = strings.TrimSpace(s)
		case "coding_guide":
			out.CodingGuide = strings.TrimSpace(s)
		}
	}
	if err := out.Validate(); err != nil {
		return v, nil, fmt.Errorf("%s: %w", path, err)
	}
	return out, from, nil
}

// SaveRepoUser writes the given repo keys into repoRoot/.acta.yaml and
// returns its path. It edits the yaml tree instead of rewriting the file,
// so the user's comments and other keys (root, dirs, links) stay.
func SaveRepoUser(repoRoot string, set map[string]string) (string, error) {
	path := filepath.Join(repoRoot, ".acta.yaml")
	for k := range set {
		if !slices.Contains(RepoKeys, k) {
			return path, fmt.Errorf("%w: %s cannot be set per repo; only %s can", ErrBadUser, k, strings.Join(RepoKeys, ", "))
		}
	}
	raw, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return path, err
	}
	var doc yaml.Node
	if len(bytes.TrimSpace(raw)) > 0 {
		if err := yaml.Unmarshal(raw, &doc); err != nil {
			return path, fmt.Errorf("%s: %w", path, err)
		}
	}
	if doc.Kind == 0 {
		doc = yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode}}}
	}
	m := doc.Content[0]
	if m.Kind != yaml.MappingNode {
		return path, fmt.Errorf("%s: the top level is not a list of keys", path)
	}
	// Walk RepoKeys, not the map, so the file comes out the same every time.
	for _, k := range RepoKeys {
		val, ok := set[k]
		if !ok {
			continue
		}
		node := &yaml.Node{Kind: yaml.ScalarNode, Value: val}
		i := slices.IndexFunc(m.Content, func(n *yaml.Node) bool { return n.Value == k })
		// Keys sit at even places and values right after them.
		if i >= 0 && i%2 == 0 {
			m.Content[i+1] = node
			continue
		}
		m.Content = append(m.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: k}, node)
	}
	out, err := yaml.Marshal(&doc)
	if err != nil {
		return path, err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, out, 0o644); err != nil {
		return path, err
	}
	return path, os.Rename(tmp, path)
}

// Override is one repo key whose value in .acta.yaml differs from the one the
// user would get without that file.
type Override struct{ Key, Repo, Yours string }

// repoValue reads one repo key from u. A key nobody set has the default the
// hooks use, so "unset" and "set to the default" compare as equal.
func repoValue(u User, key string) string {
	var v, def string
	switch key {
	case "repo_language":
		v = u.RepoLanguage
	case "build_executor":
		v = u.BuildExecutor
	case "plan_depth":
		v, def = u.PlanDepth, "full"
	case "commit_history":
		v, def = u.CommitHistory, "tidy"
	case "coding_guide":
		v, def = u.CodingGuide, "lean"
	}
	if v == "" {
		return def
	}
	return v
}

// Overrides lists the repo keys that change what the user chose, in RepoKeys
// order. A key the repo repeats with the same value is left out: it changes
// nothing. It reads the file through MergeRepo, so it fails the same way.
func Overrides(user User, repoRoot string) ([]Override, error) {
	merged, from, err := MergeRepo(user, repoRoot)
	if err != nil {
		return nil, err
	}
	var out []Override
	for _, k := range RepoKeys {
		if !from[k] {
			continue
		}
		repo, yours := repoValue(merged, k), repoValue(user, k)
		if repo != yours {
			out = append(out, Override{Key: k, Repo: repo, Yours: yours})
		}
	}
	return out, nil
}

// UnsetRepoUser removes the given repo keys from repoRoot/.acta.yaml and
// returns its path. Like SaveRepoUser it edits the yaml tree, so other keys,
// comments and order stay. A key that is not in the file, or no file at all,
// is fine: the user's value already applies.
func UnsetRepoUser(repoRoot string, keys []string) (string, error) {
	path := filepath.Join(repoRoot, ".acta.yaml")
	for _, k := range keys {
		if !slices.Contains(RepoKeys, k) {
			return path, fmt.Errorf("%w: %s cannot be unset per repo; only %s can", ErrBadUser, k, strings.Join(RepoKeys, ", "))
		}
	}
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return path, nil
	}
	if err != nil {
		return path, err
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return path, fmt.Errorf("%s: %w", path, err)
	}
	if doc.Kind == 0 {
		return path, nil
	}
	m := doc.Content[0]
	if m.Kind != yaml.MappingNode {
		return path, fmt.Errorf("%s: the top level is not a list of keys", path)
	}
	// Keys sit at even places and values right after them, so drop in pairs.
	// yaml ties the comment above a key to that key, so a removed key would
	// take it along. The person wrote it for the file, so hand it on: it goes
	// above the next key that stays. The comment on the same line as a removed
	// value was about that value and goes with it.
	kept, changed, carried := make([]*yaml.Node, 0, len(m.Content)), false, ""
	for i := 0; i+1 < len(m.Content); i += 2 {
		k := m.Content[i]
		if slices.Contains(keys, k.Value) {
			changed = true
			carried = joinComments(joinComments(carried, k.HeadComment), k.FootComment)
			continue
		}
		if carried != "" {
			k.HeadComment = joinComments(carried, k.HeadComment)
			carried = ""
		}
		kept = append(kept, k, m.Content[i+1])
	}
	if !changed {
		return path, nil
	}
	m.Content = kept
	out := []byte{}
	switch {
	case len(kept) > 0:
		// No key is left below the comment, so it closes the file.
		if carried != "" {
			last := kept[len(kept)-2]
			last.FootComment = joinComments(last.FootComment, carried)
		}
		if out, err = yaml.Marshal(&doc); err != nil {
			return path, err
		}
	case carried != "":
		// An empty map would print as "{}", so write the comment alone.
		out = []byte(carried + "\n")
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, out, 0o644); err != nil {
		return path, err
	}
	return path, os.Rename(tmp, path)
}

// joinComments stacks two yaml comment blocks, the first above the second.
func joinComments(a, b string) string {
	switch {
	case a == "":
		return b
	case b == "":
		return a
	}
	return a + "\n" + b
}
