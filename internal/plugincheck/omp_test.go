package plugincheck

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestOmpPackageJSON(t *testing.T) {
	var p struct {
		Name    string
		Private bool
		Pi      struct {
			Extensions []string
			Skills     []string
		}
		Omp  struct{ Extensions []string }
		Deps map[string]string `json:"dependencies"`
	}
	if err := json.Unmarshal([]byte(readFile(t, "package.json")), &p); err != nil {
		t.Fatal(err)
	}
	if p.Name != "acta" || !p.Private || len(p.Deps) != 0 {
		t.Fatalf("package.json = %+v", p)
	}
	if !reflect.DeepEqual(p.Pi.Extensions, []string{"./omp/index.ts"}) {
		t.Fatalf("pi.extensions = %v", p.Pi.Extensions)
	}
	if !reflect.DeepEqual(p.Pi.Skills, []string{"./skills"}) {
		t.Fatalf("pi.skills = %v", p.Pi.Skills)
	}
	if !reflect.DeepEqual(p.Omp.Extensions, []string{"./omp/index.ts"}) {
		t.Fatalf("omp.extensions = %v", p.Omp.Extensions)
	}
}
