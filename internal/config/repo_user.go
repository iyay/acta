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
var RepoKeys = []string{"repo_language", "build_executor", "plan_depth", "coding_guide"}

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
