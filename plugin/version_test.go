package plugin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/testguard"
)

// TestMain stops the binary when nobody waits for it, so a stray run
// does not burn CPU.
func TestMain(m *testing.M) {
	testguard.Watch()
	os.Exit(m.Run())
}

// pluginFileVersion reads the version field straight from the plugin file
// on disk, so this test tracks the real file instead of a copy.
func pluginFileVersion(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(".claude-plugin", "plugin.json"))
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	return got.Version
}

func TestVersionMatchesPluginFile(t *testing.T) {
	if got, want := Version(), pluginFileVersion(t); got != want {
		t.Fatalf("Version() = %q, want %q from plugin.json", got, want)
	}
}

func TestParseVersionBadInput(t *testing.T) {
	cases := map[string]string{
		"empty field":      `{"version": ""}`,
		"whitespace field": `{"version": "   "}`,
		"missing key":      `{"name": "acta"}`,
		"malformed json":   `{"version": `,
		"not an object":    `["0.1.6"]`,
		"empty file":       ``,
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			if got := parseVersion([]byte(in)); got != "dev" {
				t.Fatalf("parseVersion(%q) = %q, want %q", in, got, "dev")
			}
		})
	}
}

func TestParseVersionGoodInput(t *testing.T) {
	if got := parseVersion([]byte(`{"version": "0.1.6"}`)); got != "0.1.6" {
		t.Fatalf("parseVersion good input = %q, want %q", got, "0.1.6")
	}
	if got := parseVersion([]byte(strings.TrimSpace(`{"version": " 0.1.6 "}`))); got != "0.1.6" {
		t.Fatalf("parseVersion padded input = %q, want %q", got, "0.1.6")
	}
}
