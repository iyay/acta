package board

import (
	"reflect"
	"testing"
)

func TestCheckBody(t *testing.T) {
	cases := []struct {
		name string
		kind Kind
		body string
		want []string
	}{
		{"scratch all sections", KindScratch, "# T\n\n## Words\n\nx\n\n## Context\n\n## Log\n\n## Open questions\n", nil},
		{"bug all sections", KindBug, "# T\n## Symptom\n## Root cause\n## Repro\n## Found in\n## Context\n", nil},
		{"debt all sections", KindDebt, "# T\n## Notes\n## Context\n", nil},
		{"spec all sections", KindStory, "# T\n## Why\n## Design\n## Testing\n## Context\n", nil},
		{"plan all sections", KindPlan, "# T\n## Global Constraints\n## File Map\n## Waves\n## Context\n", nil},

		{"scratch optional sections missing", KindScratch, "# T\n\n## Words\nx\n", nil},

		{"scratch missing words", KindScratch, "# T\n\n## Context\n", []string{"missing ## Words"}},
		{"bug missing symptom", KindBug, "# T\n## Repro\n", []string{"missing ## Symptom"}},
		{"debt missing notes", KindDebt, "# T\n## Context\n", []string{"missing ## Notes"}},
		{"spec missing testing", KindStory, "# T\n## Why\n## Design\n", []string{"missing ## Testing"}},
		{"plan missing global constraints", KindPlan, "# T\n## File Map\n## Waves\n", []string{"missing ## Global Constraints"}},

		{"out of order", KindScratch, "# T\n## Context\n## Words\n", []string{"## Words is out of order"}},
		{"extra at end", KindScratch, "# T\n## Words\n## Context\n## Log\n## Open questions\n\n## Extra\n", nil},
		{"extra between", KindScratch, "# T\n## Words\n## Extra\n## Context\n", []string{"## Extra is out of order"}},

		{"no title", KindScratch, "## Words\n", []string{"missing # title"}},
		{"crlf", KindScratch, "# T\r\n## Words\r\n", nil},
		{"trailing spaces", KindScratch, "# T\n## Words  \n", nil},
		{"h3 does not count", KindScratch, "# T\n### Words\n", []string{"missing ## Words"}},

		{"kind not in table", KindTask, "anything", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := CheckBody(c.kind, c.body); !reflect.DeepEqual(got, c.want) {
				t.Errorf("got %q want %q", got, c.want)
			}
		})
	}
}

func TestSchemaOnlyScratchIsOn(t *testing.T) {
	for _, k := range []Kind{KindStory, KindPlan, KindBug, KindDebt} {
		if SchemaOn(k) {
			t.Errorf("%s must stay off in plan 1", k)
		}
	}
	if !SchemaOn(KindScratch) {
		t.Error("scratch must be on")
	}
}

func TestHasSchema(t *testing.T) {
	for v, want := range map[any]bool{1: true, "1": true, 2: false, "": false, nil: false, true: false} {
		if got := HasSchema(map[string]any{"schema": v}); got != want {
			t.Errorf("schema %v: got %v", v, want)
		}
	}
	if HasSchema(nil) {
		t.Error("nil front must be false")
	}
}

func TestSectionNames(t *testing.T) {
	want := map[Kind][]string{
		KindScratch: {"Words", "Context", "Log", "Open questions"},
		KindBug:     {"Symptom", "Root cause", "Repro", "Found in", "Context"},
		KindDebt:    {"Notes", "Context"},
		KindStory:   {"Why", "Design", "Testing", "Context"},
		KindPlan:    {"Global Constraints", "File Map", "Waves", "Context"},
	}
	for k, names := range want {
		if got := SectionNames(k); !reflect.DeepEqual(got, names) {
			t.Errorf("%s: got %q want %q", k, got, names)
		}
	}
	if got := SectionNames(KindTask); got != nil {
		t.Errorf("kind not in table must give nil, got %q", got)
	}
}
