package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Old files can still hold repo_language and subagent_models. They must load
// as before, and every write must drop them while keeping the rest.

func TestLegacyKeysDroppedFromGlobalFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	body := "chat_language: Korean\nstyle: plain\nrepo_language: German\nsubagent_models: split\nbuild_executor: inline\nplan_depth: minimal\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	v, exists, err := LoadUser(path)
	if err != nil || !exists {
		t.Fatalf("LoadUser: exists=%v err=%v", exists, err)
	}
	if err := SaveUserFile(path, v); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := string(raw)
	for _, k := range []string{"repo_language", "subagent_models"} {
		if strings.Contains(got, k) {
			t.Errorf("saved file still holds %s:\n%s", k, got)
		}
	}
	for _, kv := range []string{"chat_language: Korean", "style: plain", "build_executor: inline", "plan_depth: minimal"} {
		if !strings.Contains(got, kv) {
			t.Errorf("saved file lost %q:\n%s", kv, got)
		}
	}
}

func TestLegacyKeysInRepoFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".acta.yaml")
	body := "# shared settings\nroot: .acta\nrepo_language: German\nsubagent_models: split\nplan_depth: full\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := MergeRepo(UserDefault(), dir); err != nil {
		t.Fatalf("MergeRepo on an old file: %v", err)
	}
	if _, err := SaveRepoUser(dir, map[string]string{"review_depth": "deep"}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := string(raw)
	for _, k := range []string{"repo_language", "subagent_models"} {
		if strings.Contains(got, k) {
			t.Errorf("saved file still holds %s:\n%s", k, got)
		}
	}
	for _, kv := range []string{"# shared settings", "root: .acta", "plan_depth: full", "review_depth: deep"} {
		if !strings.Contains(got, kv) {
			t.Errorf("saved file lost %q:\n%s", kv, got)
		}
	}
}

func TestRepoLanguageIsNotARepoKey(t *testing.T) {
	dir := t.TempDir()
	if _, err := SaveRepoUser(dir, map[string]string{"repo_language": "German"}); err == nil {
		t.Error("SaveRepoUser took repo_language")
	}
	if _, err := UnsetRepoUser(dir, []string{"repo_language"}); err == nil {
		t.Error("UnsetRepoUser took repo_language")
	}
}
