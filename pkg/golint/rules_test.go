package golint

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// header imports what the snippets call, as real files must for the rules to
// resolve a package by the name the file gives it.
const header = "package p\n\nimport (\n\t\"os\"\n\t\"os/exec\"\n\t\"path/filepath\"\n\t\"time\"\n\t\"strings\"\n)\n\n"

// findings checks one function body against the catalog, as a test file or not.
func findings(t *testing.T, body string, test bool) []Finding {
	t.Helper()
	name := "x.go"
	if test {
		name = "x_test.go"
	}
	got, err := Check(name, []byte(header+"func f() {\n"+body+"\n}\n"), Catalog)
	if err != nil {
		t.Fatalf("parse %q: %v", body, err)
	}
	return got
}

func ruleOf(t *testing.T, name string) Rule {
	t.Helper()
	for _, r := range Catalog {
		if r.Name == name {
			return r
		}
	}
	t.Fatalf("no rule %s", name)
	return Rule{}
}

func TestRules(t *testing.T) {
	for _, tc := range []struct {
		rule string
		body string
		test bool
		want bool
	}{
		// test-hardcoded-arch-count (#454)
		{"test-hardcoded-arch-count", `strings.Contains(prompt, "each of the 8 entries")`, true, true},
		{"test-hardcoded-arch-count", `strings.Contains(prompt, "each of the 9 entries below")`, true, true},
		{"test-hardcoded-arch-count", `strings.Contains(prompt, fmt.Sprintf("each of the %d entries", len(markers)))`, true, false},
		{"test-hardcoded-arch-count", `strings.Contains(prompt, "each of the 8 entries")`, false, false},
		// test-batch-shell-runner (#448): the known aggregate entrypoint.
		{"test-batch-shell-runner", `exec.CommandContext(ctx, "bash", "construct-scripts.test.sh")`, true, true},
		{"test-batch-shell-runner", `exec.Command("bash", "scripts/test/construct-scripts.test.sh")`, true, true},
		{"test-batch-shell-runner", `exec.CommandContext(ctx, "bash", "construct-scripts.test.sh", "--list")`, true, false},
		{"test-batch-shell-runner", `exec.Command("bash", script)`, true, false},
		{"test-batch-shell-runner", `exec.Command("bash", "leaf.test.sh")`, true, false},
		{"test-batch-shell-runner", `exec.Command("bash", "construct-scripts.test.sh")`, false, false},
		{"test-batch-shell-runner", "//lint:allow test-batch-shell-runner — intentionally measures the batch entrypoint\nexec.Command(\"bash\", \"construct-scripts.test.sh\")", true, false},
		// test-shrunk-deadline (#434)
		{"test-shrunk-deadline", "fooTimeout = 50 * time.Millisecond", true, true},
		{"test-shrunk-deadline", "reviewShutdownGrace = time.Millisecond * 50", true, true},
		{"test-shrunk-deadline", "pushDeadline = time.Millisecond", true, true},
		{"test-shrunk-deadline", "fooTimeout = 2 * time.Second", true, false},
		{"test-shrunk-deadline", "fooTimeout = prev", true, false},
		{"test-shrunk-deadline", "fooTimeout := 50 * time.Millisecond", true, false},
		{"test-shrunk-deadline", "limit = 50 * time.Millisecond", true, false},
		{"test-shrunk-deadline", "fooTimeout = 50 * time.Millisecond", false, false},
		{"test-shrunk-deadline", "cfg.ShutdownGrace = 10 * time.Millisecond", true, true},
		// tmp-literal (#342)
		{"tmp-literal", `p := filepath.Join("/tmp", "x")`, false, true},
		{"tmp-literal", `os.MkdirAll("/tmp/work", 0o755)`, true, true},
		{"tmp-literal", `f, _ := os.Create("/tmp/x.log")`, false, true},
		{"tmp-literal", `os.MkdirTemp("", "x-*")`, false, false},
		{"tmp-literal", `strings.HasPrefix(cwd, "/tmp/")`, false, false},
		{"tmp-literal", `args := []string{"/tmp/probe"}`, true, false},
		{"tmp-literal", `os.Getenv("/tmpdir")`, false, false},
		{"tmp-literal", `want := filepath.FromSlash("/tmp/hx/issues")`, true, false},
		// test-go-build (pair#602)
		{"test-go-build", `exec.Command("go", "build", "-o", bin, ".")`, true, true},
		{"test-go-build", `run(t, time.Minute, "", true, "go", "build", "-o", bin, ".")`, true, true},
		{"test-go-build", `exec.Command("go", "vet", "./...")`, true, false},
		{"test-go-build", `exec.Command("go", "build", ".")`, false, false},
		// test-tight-wallclock (10-15 flakes)
		{"test-tight-wallclock", "if time.Since(start) > 700*time.Millisecond { panic(0) }", true, true},
		{"test-tight-wallclock", "elapsed := time.Since(start)\nif elapsed >= 500*time.Millisecond { panic(0) }", true, true},
		{"test-tight-wallclock", "if 200*time.Millisecond < time.Since(start) { panic(0) }", true, true},
		{"test-tight-wallclock", "if time.Since(start) > 3*time.Second { panic(0) }", true, false},
		{"test-tight-wallclock", "if elapsed := time.Since(start); elapsed < budget { panic(0) }", true, false},
		{"test-tight-wallclock", "if time.Since(start) < 500*time.Millisecond { panic(0) }", true, false},
		{"test-tight-wallclock", "if time.Since(start) > 700*time.Millisecond { panic(0) }", false, false},
		// test-fixed-sleep (10-15 flakes)
		{"test-fixed-sleep", "time.Sleep(200 * time.Millisecond)", true, true},
		{"test-fixed-sleep", "time.Sleep(time.Second)", true, true},
		{"test-fixed-sleep", "time.Sleep(5 * time.Millisecond)", true, false},
		{"test-fixed-sleep", "time.Sleep(time.Hour)", true, false},
		{"test-fixed-sleep", "time.Sleep(d)", true, false},
		{"test-fixed-sleep", "time.Sleep(time.Second)", false, false},
		// test-reads-workshop (#469): plans and issues move to history on landing.
		{"test-reads-workshop", `os.ReadFile("../../../../workshop/plans/000449-landing-train-plan.md")`, true, true},
		{"test-reads-workshop", `os.ReadFile(filepath.Join("..", "..", "../workshop/issues/000001-x.md"))`, true, true},
		{"test-reads-workshop", `migrate(Opts{file: "../dst/workshop/issues/000005-y.md"})`, true, false},
		{"test-reads-workshop", `procedureWrite(t, root, "workshop/issues/000001-x.md", body)`, true, false},
		{"test-reads-workshop", `os.ReadFile("../../../../atlas/workflow/sdlc-binary.md")`, true, false},
		{"test-reads-workshop", `os.ReadFile("../../workshop/history/plans/x.md")`, true, false},
		{"test-reads-workshop", `os.ReadFile("../../workshop/plans/x.md")`, false, false},
		{"test-reads-workshop", `os.ReadFile(filepath.Join("..", "..", "workshop", "plans", "x.md"))`, true, true},
		{"test-reads-workshop", `os.ReadFile(filepath.Join(root, "workshop", "issues", "x.md"))`, true, false},
	} {
		var hit *Finding
		got := findings(t, tc.body, tc.test)
		for i := range got {
			if got[i].Rule == tc.rule {
				hit = &got[i]
			}
		}
		if (hit != nil) != tc.want {
			t.Errorf("%s on %q (test file %v): flagged=%v, want %v (all: %v)", tc.rule, tc.body, tc.test, hit != nil, tc.want, got)
			continue
		}
		if hit != nil && !strings.Contains(hit.Message, ruleOf(t, tc.rule).Incident) {
			t.Errorf("%s: message does not cite its incident: %s", tc.rule, hit.Message)
		}
	}
}

func TestCatalogHygiene(t *testing.T) {
	seen := map[string]bool{}
	for _, r := range Catalog {
		if r.Name == "" || r.Incident == "" || r.Message == "" || r.Match == nil {
			t.Errorf("rule %+v lacks a name, incident, message or matcher", r.Name)
		}
		if seen[r.Name] || r.Name == suppressionRule {
			t.Errorf("rule name %s is reused", r.Name)
		}
		seen[r.Name] = true
	}
}

func TestSuppressions(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
		want []string // rules of the findings, in order
	}{
		{"line above, with reason", "//lint:allow test-fixed-sleep — the fake clock needs real time to pass\ntime.Sleep(time.Second)", nil},
		{"same line", "time.Sleep(time.Second) //lint:allow test-fixed-sleep -- bounds a known race", nil},
		{"no reason", "//lint:allow test-fixed-sleep\ntime.Sleep(time.Second)", []string{suppressionRule, "test-fixed-sleep"}},
		{"dash only", "//lint:allow test-fixed-sleep —\ntime.Sleep(time.Second)", []string{suppressionRule, "test-fixed-sleep"}},
		{"unknown rule", "//lint:allow no-such-rule — why\ntime.Sleep(time.Second)", []string{suppressionRule, "test-fixed-sleep"}},
		{"unused", "//lint:allow test-fixed-sleep — stale\nx := 1\n_ = x", []string{suppressionRule}},
		{"two lines above does not cover", "//lint:allow test-fixed-sleep — too far\n\ntime.Sleep(time.Second)", []string{"test-fixed-sleep", suppressionRule}},
		{"other rule does not cover", "//lint:allow test-go-build — wrong rule\ntime.Sleep(time.Second)", []string{"test-fixed-sleep", suppressionRule}},
	} {
		var rules []string
		for _, f := range findings(t, tc.body, true) {
			rules = append(rules, f.Rule)
		}
		if strings.Join(rules, ",") != strings.Join(tc.want, ",") {
			t.Errorf("%s: findings %v, want %v", tc.name, rules, tc.want)
		}
	}
}

// Rules resolve packages by the name the file imports them under.
func TestRulesFollowImportNames(t *testing.T) {
	src := "package p\n\nimport (\n\tfp \"path/filepath\"\n\tclock \"time\"\n)\n\nfunc f() {\n\t_ = fp.Join(\"/tmp\", \"x\")\n\tclock.Sleep(clock.Second)\n}\n"
	got, err := Check("x_test.go", []byte(src), Catalog)
	if err != nil {
		t.Fatal(err)
	}
	var rules []string
	for _, f := range got {
		rules = append(rules, f.Rule)
	}
	if strings.Join(rules, ",") != "tmp-literal,test-fixed-sleep" {
		t.Errorf("aliased imports: findings %v", rules)
	}
	// Without the import, a same-named identifier is not the package.
	local := "package p\n\nfunc f(time clock) {\n\ttime.Sleep(time.Second)\n}\n"
	if got, _ := Check("x_test.go", []byte(local), Catalog); len(got) != 0 {
		t.Errorf("an unimported name matched: %v", got)
	}
}

func TestGeneratedFilesAreSkipped(t *testing.T) {
	src := "// Code generated by x. DO NOT EDIT.\n\n" + header + "func f() {\n\ttime.Sleep(time.Second)\n}\n"
	if got, err := Check("x_test.go", []byte(src), Catalog); err != nil || len(got) != 0 {
		t.Errorf("generated file: %v %v", got, err)
	}
}

func TestBatchRunnerFollowsExecImportAlias(t *testing.T) {
	for _, call := range []string{`command.Command("bash", "construct-scripts.test.sh")`, `command.CommandContext(ctx, "bash", "./construct-scripts.test.sh")`} {
		src := "package p\nimport command \"os/exec\"\nfunc f(){ " + call + " }\n"
		got, err := Check("alias_test.go", []byte(src), []Rule{ruleOf(t, "test-batch-shell-runner")})
		if err != nil || len(got) != 1 {
			t.Fatalf("%s: findings=%v, error=%v", call, got, err)
		}
	}
}

// The atlas's catalog paragraph names every rule (#469 close review): a rule
// added without its atlas line fails here.
func TestAtlasNamesEveryRule(t *testing.T) {
	atlas, err := os.ReadFile("../../atlas/workflow/sdlc-binary.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range Catalog {
		if !strings.Contains(string(atlas), "`"+r.Name+"`") {
			t.Errorf("atlas/workflow/sdlc-binary.md's Go lint catalog does not name %s", r.Name)
		}
	}
}

// livePlanRE is a numbered plan named at its live path: it moves to
// workshop/history/plans when its issue lands, so an atlas line naming it
// there breaks then (#469 close review). Cite the issue, or the history path.
var livePlanRE = regexp.MustCompile("workshop/plans/[0-9]{6}-")

func TestAtlasCitesNoLivePlanPath(t *testing.T) {
	files, err := filepath.Glob("../../atlas/*/*.md")
	if err != nil {
		t.Fatal(err)
	}
	more, _ := filepath.Glob("../../atlas/*.md")
	files = append(files, more...)
	if len(files) < 10 {
		t.Fatalf("found only %d atlas files; the glob is stale", len(files))
	}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for i, line := range strings.Split(string(data), "\n") {
			if livePlanRE.MatchString(line) {
				t.Errorf("%s:%d names a live plan path (it moves to workshop/history/plans on landing): %s", f, i+1, strings.TrimSpace(line))
			}
		}
	}
}
