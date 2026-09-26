package editor

import (
	"reflect"
	"testing"
)

func TestCmd(t *testing.T) {
	cases := []struct {
		editor string
		line   int
		want   []string
	}{
		{"nvim", 12, []string{"nvim", "+12", "/f.md"}},
		{"", 3, []string{"vi", "+3", "/f.md"}},
		{"code -w", 1, []string{"code", "-w", "/f.md"}},
		{"hx", 0, []string{"hx", "/f.md"}},
	}
	for _, c := range cases {
		t.Setenv("EDITOR", c.editor)
		if got := Cmd("/f.md", c.line).Args; !reflect.DeepEqual(got, c.want) {
			t.Errorf("EDITOR=%q line %d: args %v, want %v", c.editor, c.line, got, c.want)
		}
	}
}
