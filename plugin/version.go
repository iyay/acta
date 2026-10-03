// Package plugin reads facts about the acta plugin itself, so the Go
// binary and the plugin files it ships with never disagree.
package plugin

import (
	_ "embed"
	"encoding/json"
	"strings"
)

// shippedPlugin holds the plugin file the binary shipped with, so the
// version shown always matches what the user installed.
type shippedPlugin struct {
	Version string `json:"version"`
}

//go:embed .claude-plugin/plugin.json
var shipped []byte

// Version is the version in the shipped plugin.json, or dev when the
// file has no usable version, so the footer never shows a blank.
func Version() string {
	return parseVersion(shipped)
}

// parseVersion reads the version out of raw plugin file bytes. Anything
// it cannot read gives dev, the local-build word, and never a panic.
func parseVersion(raw []byte) string {
	var got shippedPlugin
	if err := json.Unmarshal(raw, &got); err != nil {
		return "dev"
	}
	if v := strings.TrimSpace(got.Version); v != "" {
		return v
	}
	return "dev"
}
