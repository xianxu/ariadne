package main

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/repoenum"
	"github.com/xianxu/ariadne/pkg/vocab"
)

const (
	gateGood   = "---\nid: 1\nstatus: open\n---\n## Spec\nsome spec\n## Plan\n- [ ] do the thing\n## Done when\n- it works\n"
	gateNoPlan = "---\nid: 1\nstatus: working\n---\n## Spec\nsome spec\n## Done when\n- it works\n"
)

// stubGate swaps the three seams and returns a restore func. fmOK/fmRunErr drive the
// frontmatter validator; files drives readIssueFileFn.
func stubGate(t *testing.T, changes []gitx.FileChange, fmOK bool, fmRunErr error, files map[string]string) func() {
	t.Helper()
	od, of, or := diffNameStatusFn, validateFrontmatterFn, readIssueFileFn
	diffNameStatusFn = func(_, _ string) ([]gitx.FileChange, error) { return changes, nil }
	validateFrontmatterFn = func(_, _ string) (string, bool, error) {
		return "  - status: \"in-progress\" is not valid", fmOK, fmRunErr
	}
	readIssueFileFn = func(p string) ([]byte, error) { return []byte(files[p]), nil }
	stubRepoEnums(t, "", nil)
	return func() { diffNameStatusFn, validateFrontmatterFn, readIssueFileFn = od, of, or }
}

// stubRepoEnums answers the gate's declaration read (#348) with decl at any
// base, or with err; it counts reads.
func stubRepoEnums(t *testing.T, decl string, err error) *int {
	t.Helper()
	prev := repoEnumsAtBaseFn
	t.Cleanup(func() { repoEnumsAtBaseFn = prev })
	reads := new(int)
	repoEnumsAtBaseFn = func(string) (repoenum.Declaration, error) {
		*reads++
		if err != nil {
			return repoenum.Declaration{}, err
		}
		return parseRepoEnums([]byte(decl))
	}
	return reads
}

func runGate() error {
	var out, errw bytes.Buffer
	return validateChangedInstances("BASE", "", nounGates("workshop/issues"), &out, &errw)
}

func TestValidateChangedIssues(t *testing.T) {
	t.Run("modified file with bad frontmatter is rejected (universal)", func(t *testing.T) {
		defer stubGate(t, []gitx.FileChange{{Status: "M", Path: "workshop/issues/000052-x.md"}}, false, nil, nil)()
		if err := runGate(); err == nil {
			t.Error("expected gate failure on a modified file with bad frontmatter")
		}
	})

	t.Run("newly-added file missing ## Plan is rejected (section, added-only)", func(t *testing.T) {
		defer stubGate(t, []gitx.FileChange{{Status: "A", Path: "workshop/issues/000901-noplan.md"}}, true, nil,
			map[string]string{"workshop/issues/000901-noplan.md": gateNoPlan})()
		err := runGate()
		if err == nil {
			t.Fatal("expected gate failure on an added file missing ## Plan")
		}
	})

	t.Run("MODIFIED legacy ticket lacking sections passes (grandfathered)", func(t *testing.T) {
		defer stubGate(t, []gitx.FileChange{{Status: "M", Path: "workshop/issues/000052-legacy.md"}}, true, nil,
			map[string]string{"workshop/issues/000052-legacy.md": gateNoPlan})()
		if err := runGate(); err != nil {
			t.Errorf("a modified legacy ticket (frontmatter OK) must be grandfathered, got: %v", err)
		}
	})

	t.Run("RENAMED file is not section-validated (R != A)", func(t *testing.T) {
		defer stubGate(t, []gitx.FileChange{{Status: "R", Path: "workshop/issues/000052-renamed.md"}}, true, nil,
			map[string]string{"workshop/issues/000052-renamed.md": gateNoPlan})()
		if err := runGate(); err != nil {
			t.Errorf("a rename (R) must not be section-validated, got: %v", err)
		}
	})

	t.Run("added file with all sections passes", func(t *testing.T) {
		defer stubGate(t, []gitx.FileChange{{Status: "A", Path: "workshop/issues/000900-good.md"}}, true, nil,
			map[string]string{"workshop/issues/000900-good.md": gateGood})()
		if err := runGate(); err != nil {
			t.Errorf("a well-formed added file must pass, got: %v", err)
		}
	})

	t.Run("validator that cannot RUN fails the gate loudly (not a silent pass)", func(t *testing.T) {
		defer stubGate(t, []gitx.FileChange{{Status: "M", Path: "workshop/issues/000052-x.md"}}, false,
			errors.New("executable file not found in $PATH"), nil)()
		err := runGate()
		if err == nil || !strings.Contains(err.Error(), "could not run") {
			t.Errorf("a non-runnable validator must fail the gate with a setup error, got: %v", err)
		}
	})

	t.Run("non-issue files are ignored", func(t *testing.T) {
		defer stubGate(t, []gitx.FileChange{{Status: "M", Path: "cmd/sdlc/foo.go"}}, false, nil, nil)()
		if err := runGate(); err != nil {
			t.Errorf("non-issue files must be ignored, got: %v", err)
		}
	})

	t.Run("deleted issue file is ignored", func(t *testing.T) {
		defer stubGate(t, []gitx.FileChange{{Status: "D", Path: "workshop/issues/000052-x.md"}}, false, nil, nil)()
		if err := runGate(); err != nil {
			t.Errorf("a deleted issue file must be ignored, got: %v", err)
		}
	})
}

func TestIsInstanceFile(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"workshop/issues/000124-x.md", true},
		{"workshop/issues/sub/x.md", true}, // prefix match is fine (issues are flat in practice)
		{"workshop/history/000124-x.md", false},
		{"workshop/issues/x.txt", false},
		{"cmd/sdlc/x.go", false},
	}
	for _, c := range cases {
		if got := isInstanceFile(c.path, "workshop/issues"); got != c.want {
			t.Errorf("isInstanceFile(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestNounGatesDefaultIssueHomeComesFromVocabulary(t *testing.T) {
	t.Setenv("WF_ISSUES_DIR", "")
	gates := nounGates("")
	if got, want := gates[0].dir, vocab.Issue().Discovery().Home; got != want {
		t.Fatalf("issue gate dir = %q, want vocabulary discovery home %q", got, want)
	}
}

func TestValidateChangedInstancesDispatchesByNoun(t *testing.T) {
	od, of, or := diffNameStatusFn, validateFrontmatterFn, readIssueFileFn
	t.Cleanup(func() { diffNameStatusFn, validateFrontmatterFn, readIssueFileFn = od, of, or })
	stubRepoEnums(t, "", nil)

	diffNameStatusFn = func(_, _ string) ([]gitx.FileChange, error) {
		return []gitx.FileChange{
			{Status: "A", Path: "custom/issues/000900-good.md"},
			{Status: "A", Path: "workshop/projects/demo.md"},
			{Status: "M", Path: "workshop/issues/000901-default-dir.md"},
		}, nil
	}
	var validated []string
	validateFrontmatterFn = func(noun, path string) (string, bool, error) {
		validated = append(validated, noun+":"+path)
		return "", true, nil
	}
	reads := 0
	readIssueFileFn = func(path string) ([]byte, error) {
		reads++
		return []byte(gateGood), nil
	}

	var out, errw bytes.Buffer
	err := validateChangedInstances("BASE", "", nounGates("custom/issues"), &out, &errw)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"issue:custom/issues/000900-good.md",
		"project:workshop/projects/demo.md",
	}
	if strings.Join(validated, "\n") != strings.Join(want, "\n") {
		t.Fatalf("validated = %v, want %v", validated, want)
	}
	if reads != 1 {
		t.Fatalf("issue section reads = %d, want 1 (projects have no issue-section gate)", reads)
	}
}

// #348: the gate checks every changed issue's classification against the
// declaration at its base, reading the declaration once and only when an issue
// changed; an unreadable declaration fails closed.
func TestValidateChangedIssuesClassification(t *testing.T) {
	issueFile := func(fm string) string {
		return "---\nid: 000001\nstatus: open\n" + fm + "---\n\n# x\n"
	}
	changes := []gitx.FileChange{{Status: "M", Path: "workshop/issues/000001-x.md"}, {Status: "M", Path: "workshop/issues/000002-y.md"}}
	for _, tc := range []struct {
		name, fm1, fm2 string
		want           string // "" = passes
	}{
		{"declared and ancestor pass", "area: [couch/broadcast/web]\n", "area:\n  - couch\nseverity: high\n", ""},
		{"undeclared area", "area: [couch/session]\n", "", "couch/session"},
		{"invalid severity", "", "severity: urgent\n", "urgent"},
		{"bad area shape", "area: Couch\n", "", "classification"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			restore := stubGate(t, changes, true, nil, map[string]string{
				"workshop/issues/000001-x.md": issueFile(tc.fm1), "workshop/issues/000002-y.md": issueFile(tc.fm2)})
			defer restore()
			reads := stubRepoEnums(t, `{"version":1,"enums":{"area":["couch/broadcast/web"]}}`, nil)
			var out bytes.Buffer
			err := validateChangedInstances("BASE", "HEAD", nounGates("workshop/issues"), &out, io.Discard)
			if tc.want == "" && err != nil {
				t.Fatalf("%v\n%s", err, out.String())
			}
			if tc.want != "" && (err == nil || !strings.Contains(out.String(), tc.want)) {
				t.Fatalf("err %v, out %q, want %q", err, out.String(), tc.want)
			}
			if *reads != 1 {
				t.Fatalf("declaration read %d times, want once", *reads)
			}
		})
	}
	t.Run("no issue changed: declaration never read", func(t *testing.T) {
		restore := stubGate(t, []gitx.FileChange{{Status: "M", Path: "README.md"}}, true, nil, nil)
		defer restore()
		reads := stubRepoEnums(t, "", nil)
		if err := validateChangedInstances("BASE", "HEAD", nounGates("workshop/issues"), io.Discard, io.Discard); err != nil || *reads != 0 {
			t.Fatalf("err %v reads %d", err, *reads)
		}
	})
	t.Run("unreadable declaration fails closed", func(t *testing.T) {
		restore := stubGate(t, changes, true, nil, map[string]string{"workshop/issues/000001-x.md": issueFile(""), "workshop/issues/000002-y.md": issueFile("")})
		defer restore()
		stubRepoEnums(t, "", errors.New(".sdlc/vocabulary.json: version 2, want 1"))
		if err := validateChangedInstances("BASE", "HEAD", nounGates("workshop/issues"), io.Discard, io.Discard); err == nil || !strings.Contains(err.Error(), ".sdlc/vocabulary.json") {
			t.Fatalf("err %v", err)
		}
	})
}
