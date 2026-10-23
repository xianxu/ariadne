package reviewexport

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

func TestParseBrief(t *testing.T) {
	b, err := ParseBrief("---\nname: a\n---\nBrief.\n", "a")
	if err != nil || b != (Brief{Name: "a"}) {
		t.Fatalf("got %+v, %v", b, err)
	}
	for _, c := range []struct{ name, text, want string }{
		{"no frontmatter", "name: a\n", "missing frontmatter"},
		{"unterminated", "---\nname: a\n", "unterminated"},
		{"name mismatch", "---\nname: b\n---\n", "basename"},
		// #575: the v3 mode flags are retired; a leftover one refuses the build.
		{"retired scope", "---\nname: a\nscope: whole-doc\n---\n", "scope"},
		{"retired frontier", "---\nname: a\nfrontier: on\n---\n", "frontier"},
		{"retired order", "---\nname: a\norder: 2\n---\n", "order"},
		{"retired deletions", "---\nname: a\ndeletions: propose-strike\n---\n", "deletions"},
		{"unknown key", "---\nname: a\ntone: dry\n---\n", "tone"},
		// Parity with parley's parser at the edges.
		{"crlf", "---\r\nname: a\r\n---\r\n", "missing frontmatter"},
		{"quoted value is literal", "---\nname: \"a\"\n---\n", "basename"},
		{"empty frontmatter", "---\n---\nBody\n", "basename"},
	} {
		if _, err := ParseBrief(c.text, "a"); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: err %v, want %q", c.name, err, c.want)
		}
	}
}

func fixture() fstest.MapFS {
	f := func(s string) *fstest.MapFile { return &fstest.MapFile{Data: []byte(s)} }
	return fstest.MapFS{
		"round.md":               f("round\n"),
		"review-convention.md":   f("grammar\n"),
		"collaboration-modes.md": f("collaboration modes\n"),
		"adapters/file-edit.md":  f("adapter\n"),
		"briefs/zeta.md":         f("---\nname: zeta\n---\nZ\n"),
		"briefs/beta.md":         f("---\nname: beta\n---\nB\n"),
		"briefs/alpha.md":        f("---\nname: alpha\n---\nA\n"),
	}
}

// The order is total: role, then name, then path. Briefs carry no order
// (#575), so they sort by name.
func TestBuildOrder(t *testing.T) {
	m, err := Build(fixture())
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, mem := range m.Members {
		got = append(got, mem.Role+":"+mem.Path)
	}
	want := "round:round.md grammar:review-convention.md collaboration-modes:collaboration-modes.md adapter:adapters/file-edit.md brief:briefs/alpha.md brief:briefs/beta.md brief:briefs/zeta.md"
	if strings.Join(got, " ") != want {
		t.Errorf("order\n got %s\nwant %s", strings.Join(got, " "), want)
	}
	if g := m.Members[1]; g.Source != GrammarSource || g.Bytes != len("grammar\n") {
		t.Errorf("grammar member %+v", g)
	}
	if z := m.Members[6]; z.Name != "zeta" {
		t.Errorf("a brief is named by its file: %+v", z)
	}
}

// The move cases' files (#424) and the named exportedCases (#575) are
// exported as fixtures; the other examples are not. Fixtures sort last, by
// case, then path.
func TestBuildFixtures(t *testing.T) {
	fs := fixture()
	f := func(s string) *fstest.MapFile { return &fstest.MapFile{Data: []byte(s)} }
	fs["examples/move-b/case.md"] = f("submitted: d.md\noutcome: ok\nmove: d.md e.md\n")
	fs["examples/move-b/before/d.md"] = f("D\n")
	fs["examples/move-b/after/e.md"] = f("D\n")
	fs["examples/move-a/case.md"] = f("move: x.md y.md\n")
	fs["examples/chat/case.md"] = f("submitted: s.md\noutcome: ok\n")
	fs["examples/chat/before/s.md"] = f("S\n")
	fs["examples/chat-tag/case.md"] = f("submitted: s.md\noutcome: ok\n")
	m, err := Build(fs)
	if err != nil {
		t.Fatal(err)
	}
	if m.Version != 4 {
		t.Errorf("version %d, want 4 (flagless briefs, #575)", m.Version)
	}
	var got []string
	for _, mem := range m.Members {
		if mem.Role == "fixture" {
			got = append(got, mem.Name+":"+mem.Path)
		}
	}
	want := "chat-tag:examples/chat-tag/case.md move-a:examples/move-a/case.md move-b:examples/move-b/after/e.md move-b:examples/move-b/before/d.md move-b:examples/move-b/case.md"
	if strings.Join(got, " ") != want {
		t.Errorf("fixtures\n got %s\nwant %s", strings.Join(got, " "), want)
	}
	if last := m.Members[len(m.Members)-1]; last.Role != "fixture" {
		t.Errorf("fixtures do not sort last: %+v", last)
	}
}

// #575: v4 has no mode role and no mode flags; v3 readers refuse it.
func TestV4ExportHasNoModeFlags(t *testing.T) {
	m, err := Build(fixture())
	if err != nil {
		t.Fatal(err)
	}
	data, err := Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	for _, retired := range []string{`"deletions"`, `"order"`, `"scope"`, `"frontier"`, `"role": "mode"`} {
		if bytes.Contains(data, []byte(retired)) {
			t.Errorf("v4 export still carries %s", retired)
		}
	}
	if !bytes.Contains(data, []byte(`"version": 4`)) {
		t.Errorf("not version 4: %s", data)
	}
}

func TestBuildRefusesMissingMembers(t *testing.T) {
	for _, drop := range []string{"round.md", "review-convention.md", "collaboration-modes.md", "adapters/file-edit.md", "briefs/alpha.md briefs/beta.md briefs/zeta.md"} {
		fs := fixture()
		for _, p := range strings.Fields(drop) {
			delete(fs, p)
		}
		if _, err := Build(fs); err == nil {
			t.Errorf("built without %s", drop)
		}
	}
}

func TestMarshalDeterministic(t *testing.T) {
	a, _ := Build(fixture())
	b, _ := Build(fixture())
	ja, err := Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	jb, _ := Marshal(b)
	if !bytes.Equal(ja, jb) || !bytes.HasSuffix(ja, []byte("}\n")) {
		t.Errorf("two builds differ or lack the trailing newline:\n%s\n%s", ja, jb)
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("go", "env", "GOMOD").Output()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Dir(strings.TrimSpace(string(out)))
}

// The committed export.json is the build of the committed skill directory.
// Regenerate it with UPDATE_REVIEW_EXPORT=1 go test ./pkg/reviewexport.
func TestExportCurrent(t *testing.T) {
	root := moduleRoot(t)
	dir := filepath.Join(root, "construct", "local", "universal-review")
	m, err := Build(os.DirFS(dir))
	if err != nil {
		t.Fatal(err)
	}
	want, err := Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "export.json")
	if os.Getenv("UPDATE_REVIEW_EXPORT") == "1" {
		if err := os.WriteFile(path, want, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (generate it with UPDATE_REVIEW_EXPORT=1 go test ./pkg/reviewexport)", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("export.json is stale; regenerate it with UPDATE_REVIEW_EXPORT=1 go test ./pkg/reviewexport")
	}
	// The grammar is the one authored copy, linked rather than duplicated.
	link := filepath.Join(dir, "review-convention.md")
	if target, err := os.Readlink(link); err != nil || target != "../fix/review-convention.md" {
		t.Errorf("review-convention.md must link ../fix/review-convention.md: %q, %v", target, err)
	}
	canon, err := os.ReadFile(filepath.Join(root, GrammarSource))
	if err != nil {
		t.Fatal(err)
	}
	if m.Members[1].Bytes != len(canon) {
		t.Errorf("grammar member covers %d bytes, the source has %d", m.Members[1].Bytes, len(canon))
	}
	var briefs []string
	collab := 0
	exported := map[string]bool{}
	for _, mem := range m.Members {
		switch mem.Role {
		case "brief":
			briefs = append(briefs, mem.Name)
		case "collaboration-modes":
			collab++
		case "fixture":
			exported[mem.Name] = true
		}
		// Only the move cases and exportedCases are exported, as fixtures (#424, #575).
		if strings.HasPrefix(mem.Path, "examples/") != (mem.Role == "fixture") || mem.Role == "fixture" && !strings.HasPrefix(mem.Name, "move-") && !slices.Contains(exportedCases, mem.Name) {
			t.Errorf("%s exported as %s/%s; only the move cases and exportedCases are, as fixtures", mem.Path, mem.Role, mem.Name)
		}
	}
	// Every named case exists: a renamed or missing one would silently drop a fixture pair and parley pin.
	for _, c := range exportedCases {
		if !exported[c] {
			t.Errorf("exportedCases names %q, which is not an example", c)
		}
	}
	// #412: parley's sync requires exactly one collaboration-modes member.
	if collab != 1 {
		t.Errorf("%d collaboration-modes members, want exactly one", collab)
	}
	if got := strings.Join(briefs, ","); got != "copy-editing,developmental,fact-check,line-editing,proofreading" {
		t.Errorf("stock briefs by name: %s", got)
	}
}

// These are the two agent-facing adapters for the same Pair edit engine.
// Guard the migration's refusal and recovery paths, not just manifest hashes.
func TestReviewSkillApplyContract(t *testing.T) {
	root := moduleRoot(t)
	read := func(relative string) string {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(root, relative))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	adapter := read("construct/local/universal-review/adapters/file-edit.md")
	workbench := read("construct/local/fix/SKILL.md")
	round := read("construct/local/universal-review/round.md")
	for _, tc := range []struct {
		name, body string
		want       []string
		obsolete   []string
	}{
		{"tracked adapter", adapter,
			[]string{"couch --session-sync apply EDITS.json FILE", "couch --session-sync round FILE --json", "edits-invalid", "propose-too-large", "propose-unencodable", "operator-owned", "Exit 3", "Exit 6", "next submit", "Ordinary prose edits still become proposals"},
			[]string{"couch --session-sync write --all BODYFILE FILE"}},
		{"Pair workbench", workbench,
			[]string{"pair review apply EDITS.json FILE", "docflow round --side agent", "propose-too-large", "operator-owned", "operator-lock", "stale", "Body"},
			[]string{"PAIR_REVIEW_HANDOFF_PATH", "PAIR_REVIEW_LANDED_PATH", "[reconcile —", "applied N edits"}},
		{"review round", round,
			[]string{"highlight", "Never type proposal marker syntax", "@@chat@@", "question", "command"},
			[]string{"deletions:", "propose-strike", "presentation-mode:", "processing-mode:", "frontier"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, phrase := range tc.want {
				if !strings.Contains(tc.body, phrase) {
					t.Errorf("missing %q", phrase)
				}
			}
			for _, phrase := range tc.obsolete {
				if strings.Contains(tc.body, phrase) {
					t.Errorf("retired instruction %q remains", phrase)
				}
			}
		})
	}
}

func TestReviewSkillCloseRecoveryContract(t *testing.T) {
	root := moduleRoot(t)
	for _, tc := range []struct {
		file string
		want []string
	}{
		{"construct/local/universal-review/adapters/file-edit.md", []string{"untracked", "presentation-mode: propose", "refuse", "observed before/after diff", "pending", "operator rejects"}},
		{"construct/local/universal-review/round.md", []string{"untracked", "observed before/after diff", "pending"}},
		{"construct/local/universal-review/collaboration-modes.md", []string{"pending", "operator rejects", "discussions:"}},
		{"construct/local/fix/SKILL.md", []string{"base_hash", "Current", "read back", "pending", "operator rejects", "write it yet fail to record a round"}},
		{"README.md", []string{"universal review", "presentation-mode: propose", "atlas/workflow/universal-review.md"}},
	} {
		t.Run(tc.file, func(t *testing.T) {
			body, err := os.ReadFile(filepath.Join(root, tc.file))
			if err != nil {
				t.Fatal(err)
			}
			for _, phrase := range tc.want {
				if !strings.Contains(string(body), phrase) {
					t.Errorf("missing %q", phrase)
				}
			}
		})
	}
}

// The adapters are agent instructions, so their routing and recovery tables
// are the executable contract. A keyword elsewhere cannot rescue an unsafe row.
func TestReviewSkillDecisionScenarios(t *testing.T) {
	root := moduleRoot(t)
	read := func(name string) string {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	rows := func(doc, heading string) map[string][]string {
		t.Helper()
		start := strings.Index(doc, heading)
		if start < 0 {
			t.Fatalf("missing decision table %q", heading)
		}
		section := doc[start+len(heading):]
		out := map[string][]string{}
		inside := false
		for _, line := range strings.Split(section, "\n") {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "|") {
				if inside {
					break
				}
				continue
			}
			inside = true
			cells := strings.Split(strings.Trim(line, "|"), "|")
			if len(cells) != 3 {
				continue
			}
			for i := range cells {
				cells[i] = strings.Trim(strings.TrimSpace(cells[i]), "`")
			}
			if cells[0] == "Case" || strings.HasPrefix(cells[0], "---") {
				continue
			}
			if _, duplicate := out[cells[0]]; duplicate {
				t.Fatalf("duplicate case %q", cells[0])
			}
			out[cells[0]] = cells[1:]
		}
		return out
	}
	routing := rows(read("construct/local/universal-review/adapters/file-edit.md"), "## Routing decisions")
	if len(routing) != 5 {
		t.Fatalf("routing has %d rows, want 5 unambiguous cases", len(routing))
	}
	for _, role := range []string{"submitted", "parent"} {
		for _, tc := range []struct{ state, action, evidence string }{
			{"tracked direct/highlight", "apply", "round spans"},
			{"tracked propose", "apply", "round spans"},
			{"untracked direct/highlight", "file tools", "observed diff"},
			{"untracked propose", "refuse", "track or Pair workbench"},
			{"unknown owner", "refuse", "lookup error"},
		} {
			key := tc.state + " (submitted or parent)"
			if got := routing[key]; !slices.Equal(got, []string{tc.action, tc.evidence}) {
				t.Errorf("%s routing %q = %v, want %q / %q", role, tc.state, got, tc.action, tc.evidence)
			}
		}
	}
	workbench := read("construct/local/fix/SKILL.md")
	start := strings.Index(workbench, "## Pair review workbench")
	if start < 0 {
		t.Fatal("missing Pair workbench section")
	}
	workbench = workbench[start:]
	for _, phrase := range []string{
		"Always call `pair review apply`",
		`{"records":[]}`,
		"explanations or questions in your session reply",
		"never commit an empty reply",
		"including after a lost apply reply",
	} {
		if !strings.Contains(workbench, phrase) {
			t.Errorf("empty-round producer contract missing %q", phrase)
		}
	}
	success := rows(workbench, "### Successful apply decisions")
	if len(success) != 2 {
		t.Errorf("success decisions have %d rows, want changed and empty", len(success))
	}
	for _, tc := range []struct{ scenario, action, evidence string }{
		{"changed reply", "commit Body verbatim", "exit 0; nonempty Body and Round"},
		{"empty reply", "no agent commit", "exit 0; no changed records; empty Body"},
	} {
		if got := success[tc.scenario]; !slices.Equal(got, []string{tc.action, tc.evidence}) {
			t.Errorf("success %q = %v, want %q / %q", tc.scenario, got, tc.action, tc.evidence)
		}
	}
	recovery := rows(read("construct/local/fix/SKILL.md"), "### Apply recovery decisions")
	if len(recovery) != 5 {
		t.Fatalf("recovery has %d rows, want 5 unambiguous cases", len(recovery))
	}
	for _, tc := range []struct{ scenario, action, evidence string }{
		{"matching landed", "commit Body", "base hash, Current, readback and spans match"},
		{"old current", "stop", "base hash differs"},
		{"null unchanged", "fresh read then retry", "file still equals attempted base"},
		{"null changed", "stop", "file changed without matching round"},
		{"stale or uncertain", "stop", "Current false or readback mismatch"},
	} {
		if got := recovery[tc.scenario]; !slices.Equal(got, []string{tc.action, tc.evidence}) {
			t.Errorf("recovery %q = %v, want %q / %q", tc.scenario, got, tc.action, tc.evidence)
		}
	}
}
