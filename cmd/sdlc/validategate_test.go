package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

const (
	gateGood   = "---\nid: 1\nstatus: open\n---\n## Spec\nsome spec\n## Plan\n- [ ] do the thing\n## Done when\n- it works\n"
	gateNoPlan = "---\nid: 1\nstatus: open\n---\n## Spec\nsome spec\n## Done when\n- it works\n"
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
	return func() { diffNameStatusFn, validateFrontmatterFn, readIssueFileFn = od, of, or }
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

func TestValidateChangedInstancesDispatchesByNoun(t *testing.T) {
	od, of, or := diffNameStatusFn, validateFrontmatterFn, readIssueFileFn
	t.Cleanup(func() { diffNameStatusFn, validateFrontmatterFn, readIssueFileFn = od, of, or })

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
