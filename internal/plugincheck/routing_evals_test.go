package plugincheck

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// routingEvalCases is every routing eval folder. The folder name starts with
// the route the prompt should get, so the report shows one pass rate per
// route.
var routingEvalCases = []string{
	"routing-scratch-01",
	"routing-scratch-02",
	"routing-scratch-03",
	"routing-scratch-04",
	"routing-arch-01",
	"routing-arch-02",
	"routing-arch-03",
	"routing-arch-04",
	"routing-light-01",
	"routing-light-02",
	"routing-light-03",
	"routing-light-04",
	"routing-chat-01",
	"routing-chat-02",
	"routing-chat-03",
	"routing-chat-04",
	"routing-second-01",
	"routing-second-02",
}

// routingBanned never appears in a prompt body. The prompt must read like a
// user talking normally, never naming its route, so even everyday words like
// later and idea are out.
var routingBanned = []string{
	"note this",
	"later",
	"idea",
	"follow the acta workflow",
	"design",
	"brainstorm",
	"light path",
	"architectural",
	"bounded",
	"scratch",
}

// routingCLIEndsEveryPrompt is the last line of each routing prompt. The
// scaffold puts the CLI there, and the prompt has to say so.
const routingCLIEnding = "The acta CLI for this repo is ./bin/acta. It is on PATH as well."

// routingLang is the main language of each prompt body. Mixed prompts blend
// in English sentences but still count under their main language.
var routingLang = map[string]string{
	"routing-scratch-01": "en",
	"routing-scratch-02": "id",
	"routing-scratch-03": "en",
	"routing-scratch-04": "id",
	"routing-arch-01":    "en",
	"routing-arch-02":    "id",
	"routing-arch-03":    "en",
	"routing-arch-04":    "id",
	"routing-light-01":   "en",
	"routing-light-02":   "id",
	"routing-light-03":   "en",
	"routing-light-04":   "id",
	"routing-chat-01":    "id",
	"routing-chat-02":    "en",
	"routing-chat-03":    "id",
	"routing-chat-04":    "en",
	"routing-second-01":  "en",
	"routing-second-02":  "id",
}

// routingEnWords and routingIdWords are common function words with no overlap.
// A pure English prompt holds none of the Indonesian ones and vice versa.
var routingEnWords = map[string]bool{
	"the": true, "and": true, "with": true, "this": true, "that": true,
	"these": true, "those": true, "from": true, "have": true, "has": true,
	"had": true, "will": true, "would": true, "should": true, "could": true,
	"there": true, "their": true, "them": true, "they": true, "your": true,
	"you": true, "about": true, "into": true, "over": true, "when": true,
	"where": true, "which": true, "while": true, "what": true, "then": true,
	"than": true, "are": true, "was": true, "were": true, "been": true,
	"does": true, "did": true, "not": true, "but": true, "for": true,
	"can": true, "all": true, "any": true, "out": true, "just": true,
	"more": true, "most": true, "other": true, "some": true, "such": true,
	"only": true, "also": true, "how": true, "why": true, "who": true,
	"because": true, "until": true, "between": true, "through": true,
	"during": true, "before": true, "after": true, "above": true,
	"below": true, "again": true, "once": true, "here": true, "now": true,
	"very": true, "every": true, "each": true,
}

var routingIdWords = map[string]bool{
	"saya": true, "aku": true, "gue": true, "gw": true, "kamu": true,
	"anda": true, "kalian": true, "kita": true, "kami": true, "mereka": true,
	"dia": true, "beliau": true, "ini": true, "itu": true, "apa": true,
	"siapa": true, "kapan": true, "mana": true, "berapa": true, "yang": true,
	"dan": true, "dengan": true, "untuk": true, "dari": true, "tidak": true,
	"nggak": true, "enggak": true, "sudah": true, "udah": true, "belum": true,
	"lagi": true, "bisa": true, "harus": true, "mesti": true, "mau": true,
	"ingin": true, "tolong": true, "dong": true, "nih": true, "tuh": true,
	"banget": true, "sangat": true, "gimana": true, "bagaimana": true,
	"kenapa": true, "mengapa": true, "dimana": true, "kemana": true,
	"kalau": true, "kalo": true, "karena": true, "tapi": true, "tetapi": true,
	"juga": true, "masih": true, "pernah": true, "selalu": true,
	"sering": true, "kadang": true, "jarang": true, "baru": true,
	"lama": true, "cepat": true, "lambat": true, "besar": true, "kecil": true,
	"bagus": true, "jelek": true, "rusak": true, "susah": true,
	"gampang": true, "mudah": true, "sulit": true, "paham": true,
	"mengerti": true, "ngerti": true, "kok": true, "sih": true, "deh": true,
	"yuk": true, "ayo": true, "oke": true, "terima": true, "kasih": true,
	"sama": true, "antara": true, "oleh": true, "pada": true, "dalam": true,
	"luar": true, "atas": true, "bawah": true, "depan": true,
	"belakang": true, "samping": true, "kiri": true, "kanan": true,
	"satu": true, "dua": true, "tiga": true, "banyak": true, "sedikit": true,
	"semua": true, "setiap": true, "tiap": true, "lain": true,
	"ubah": true, "salah": true, "benar": true, "betul": true,
	"tulisan": true, "kata": true, "kalimat": true, "muncul": true,
	"cerita": true, "dengar": true, "lihat": true, "baca": true,
	"tulis": true, "bikin": true, "buat": true, "pakai": true,
	"gunakan": true, "ambil": true, "taruh": true, "simpan": true,
	"hapus": true, "tambah": true, "kurang": true, "cek": true, "coba": true,
	"bro": true, "mas": true, "mba": true, "mbak": true, "pak": true,
	"bu": true, "kak": true, "bang": true, "lu": true, "elo": true,
	"sesi": true, "bahasan": true, "jangan": true, "sedang": true,
	"tengah": true, "saja": true, "cukup": true, "hanya": true,
	"tanpa": true, "atau": true, "agar": true, "supaya": true, "jika": true,
	"bila": true, "saat": true, "ketika": true, "sambil": true,
	"setelah": true, "sebelum": true, "sesudah": true, "sejak": true,
	"sampai": true, "hingga": true, "tentang": true, "soal": true,
	"hal": true, "orang": true, "waktu": true, "hari": true, "pagi": true,
	"siang": true, "sore": true, "malam": true, "kemarin": true,
	"tadi": true, "nanti": true, "sekarang": true, "besok": true,
}

var routingWordRe = regexp.MustCompile(`[a-z]+`)
var routingContextRe = regexp.MustCompile(`(?m)^[ \t]*context:`)

// routingSplit cuts a case file into its frontmatter and its body.
func routingSplit(t *testing.T, path string) (fm, body string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.SplitN(string(raw), "---", 3)
	if len(parts) != 3 || strings.TrimSpace(parts[0]) != "" {
		t.Fatalf("%s has no frontmatter between --- lines", path)
	}
	return parts[1], strings.TrimSpace(parts[2])
}

// routingLangBody is the prompt body without the English CLI closing line,
// so the closing line cannot pass as English content in an Indonesian prompt.
func routingLangBody(body string) string {
	if i := strings.Index(body, "The acta CLI"); i >= 0 {
		return body[:i]
	}
	return body
}

func routingCounts(body string) (en, id int) {
	for _, w := range routingWordRe.FindAllString(strings.ToLower(body), -1) {
		if routingEnWords[w] {
			en++
		}
		if routingIdWords[w] {
			id++
		}
	}
	return en, id
}

// TestRoutingEvalCases checks every routing eval folder: the set is exact,
// each prompt carries the routing frontmatter and ends on the CLI line, each
// folder holds a scaffold wired through case.yaml and at least one action
// grader, no grader calls the judge, no prompt names its route, and the
// language split holds.
func TestRoutingEvalCases(t *testing.T) {
	root := filepath.Join(pluginRoot(t), "evals")

	entries, err := filepath.Glob(filepath.Join(root, "routing-*"))
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, e := range entries {
		st, err := os.Stat(e)
		if err != nil {
			t.Fatal(err)
		}
		if st.IsDir() {
			got = append(got, filepath.Base(e))
		}
	}
	sort.Strings(got)
	want := append([]string(nil), routingEvalCases...)
	sort.Strings(want)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("routing folders = %v, want %v", got, want)
	}

	for _, folder := range routingEvalCases {
		t.Run(folder, func(t *testing.T) {
			dir := filepath.Join(root, folder)
			fm, body := routingSplit(t, filepath.Join(dir, "prompt.md"))

			var meta struct {
				Name    string   `yaml:"name"`
				Desc    string   `yaml:"description"`
				Tags    []string `yaml:"tags"`
				Runs    int      `yaml:"runs"`
				Max     int      `yaml:"max_turns"`
				Timeout int      `yaml:"timeout_seconds"`
			}
			if err := yaml.Unmarshal([]byte(fm), &meta); err != nil {
				t.Fatalf("prompt.md frontmatter does not parse: %v", err)
			}
			if meta.Name != folder {
				t.Errorf("name = %q, want %q", meta.Name, folder)
			}
			if meta.Desc == "" || strings.Contains(meta.Desc, "\n") {
				t.Errorf("description must be one non-empty line, got %q", meta.Desc)
			}
			if len(meta.Tags) != 1 || meta.Tags[0] != "routing" {
				t.Errorf("tags = %v, want [routing]", meta.Tags)
			}
			if meta.Runs != 3 {
				t.Errorf("runs = %d, want 3", meta.Runs)
			}
			if meta.Max != 6 {
				t.Errorf("max_turns = %d, want 6", meta.Max)
			}
			if meta.Timeout != 240 {
				t.Errorf("timeout_seconds = %d, want 240", meta.Timeout)
			}
			if !strings.Contains(fm, "allowed_tools:") {
				t.Error("prompt.md must set allowed_tools")
			}
			if routingContextRe.MatchString(fm) {
				t.Error("prompt.md must not set a context key; the scaffold lives in the case folder")
			}
			if !strings.HasSuffix(body, routingCLIEnding) {
				t.Error("prompt body must end with the acta CLI line")
			}

			lower := strings.ToLower(routingLangBody(body))
			for _, bad := range routingBanned {
				if strings.Contains(lower, bad) {
					t.Errorf("prompt body names its route with %q", bad)
				}
			}
			if st, err := os.Stat(filepath.Join(dir, "scaffold.sh")); err != nil || st.IsDir() {
				t.Error("scaffold.sh is missing")
			}
			raw, err := os.ReadFile(filepath.Join(dir, "case.yaml"))
			if err != nil {
				t.Error("case.yaml is missing; the scaffold runs only when case.yaml wires it")
			} else {
				var cy struct {
					Name    string `yaml:"name"`
					Context struct {
						Scaffold string `yaml:"scaffold_script"`
					} `yaml:"context"`
				}
				if err := yaml.Unmarshal(raw, &cy); err != nil {
					t.Errorf("case.yaml does not parse: %v", err)
				} else {
					if cy.Name != folder {
						t.Errorf("case.yaml name = %q, want %q", cy.Name, folder)
					}
					if cy.Context.Scaffold == "" {
						t.Error("case.yaml must set context.scaffold_script")
					} else if st, err := os.Stat(filepath.Join(dir, cy.Context.Scaffold)); err != nil || st.IsDir() {
						t.Errorf("scaffold_script %q is missing from the case folder", cy.Context.Scaffold)
					}
				}
			}

			graders, err := filepath.Glob(filepath.Join(dir, "graders", "*.md"))
			if err != nil {
				t.Fatal(err)
			}
			if len(graders) == 0 {
				t.Error("no grader in graders/")
			}
			for _, g := range graders {
				raw, err := os.ReadFile(g)
				if err != nil {
					t.Fatal(err)
				}
				if strings.Contains(string(raw), "type: llm") {
					t.Errorf("grader %s calls the judge; routing graders check actions, not words", filepath.Base(g))
				}
				if !strings.Contains(string(raw), "# guards: ") {
					t.Errorf("grader %s does not say which route it guards", filepath.Base(g))
				}
			}

			en, id := routingCounts(routingLangBody(body))
			switch routingLang[folder] {
			case "en":
				if id != 0 || en < 4 {
					t.Errorf("English prompt holds %d English and %d Indonesian marker words", en, id)
				}
			case "id":
				if id < 5 {
					t.Errorf("Indonesian prompt holds only %d Indonesian marker words", id)
				}
			default:
				t.Fatalf("no language listed for %s", folder)
			}
		})
	}

	enCount := map[string]int{}
	idCount := map[string]int{}
	mixed := 0
	for _, folder := range routingEvalCases {
		_, body := routingSplit(t, filepath.Join(root, folder, "prompt.md"))
		en, _ := routingCounts(routingLangBody(body))
		route := strings.Split(folder, "-")[1]
		switch routingLang[folder] {
		case "en":
			enCount[route]++
		case "id":
			idCount[route]++
			if en >= 4 {
				mixed++
			}
		}
	}
	for _, route := range []string{"scratch", "arch", "light", "chat"} {
		if enCount[route] != 2 || idCount[route] != 2 {
			t.Errorf("route %s has %d English and %d Indonesian prompts, want 2 and 2",
				route, enCount[route], idCount[route])
		}
	}
	if enCount["second"] != 1 || idCount["second"] != 1 {
		t.Errorf("route second has %d English and %d Indonesian prompts, want 1 and 1",
			enCount["second"], idCount["second"])
	}
	if mixed < 2 {
		t.Errorf("only %d prompts mix English and Indonesian, want at least 2", mixed)
	}
}
