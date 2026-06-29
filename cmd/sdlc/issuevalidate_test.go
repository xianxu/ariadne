package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRunIssueValidate(t *testing.T) {
	of := validateFrontmatterFn
	defer func() { validateFrontmatterFn = of }()

	dir := t.TempDir()
	good := filepath.Join(dir, "000900-good.md")
	noPlan := filepath.Join(dir, "000901-noplan.md")
	if err := os.WriteFile(good, []byte(gateGood), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(noPlan, []byte(gateNoPlan), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Run("clean file passes", func(t *testing.T) {
		validateFrontmatterFn = func(string) (string, bool, error) { return "", true, nil }
		var out, errw bytes.Buffer
		if err := runIssueValidate(&out, &errw, &issueValidateFlags{}, []string{good}); err != nil {
			t.Errorf("clean file should pass, got: %v", err)
		}
	})

	t.Run("file missing ## Plan fails (section, full check on demand)", func(t *testing.T) {
		validateFrontmatterFn = func(string) (string, bool, error) { return "", true, nil }
		var out, errw bytes.Buffer
		if err := runIssueValidate(&out, &errw, &issueValidateFlags{}, []string{noPlan}); err == nil {
			t.Error("a file missing ## Plan should fail on-demand validation")
		}
	})

	t.Run("bad frontmatter fails", func(t *testing.T) {
		validateFrontmatterFn = func(string) (string, bool, error) { return "  - status: nope", false, nil }
		var out, errw bytes.Buffer
		if err := runIssueValidate(&out, &errw, &issueValidateFlags{}, []string{good}); err == nil {
			t.Error("bad frontmatter should fail validation")
		}
	})

	t.Run("no target specified errors", func(t *testing.T) {
		var out, errw bytes.Buffer
		if err := runIssueValidate(&out, &errw, &issueValidateFlags{}, nil); err == nil {
			t.Error("no <file>/--issue/--all should error")
		}
	})

	t.Run("multiple files all conforming passes", func(t *testing.T) {
		validateFrontmatterFn = func(string) (string, bool, error) { return "", true, nil }
		var out, errw bytes.Buffer
		if err := runIssueValidate(&out, &errw, &issueValidateFlags{}, []string{good, good}); err != nil {
			t.Errorf("an all-conforming batch should pass, got: %v", err)
		}
	})

	t.Run("multiple files, one nonconforming exits non-zero", func(t *testing.T) {
		validateFrontmatterFn = func(string) (string, bool, error) { return "", true, nil }
		var out, errw bytes.Buffer
		if err := runIssueValidate(&out, &errw, &issueValidateFlags{}, []string{good, noPlan}); err == nil {
			t.Error("a batch with one nonconforming file should exit non-zero")
		}
	})
}

// TestResolveValidateTargets exercises the multi-target resolution contract
// (#133): comma-separated --issue IDs, multiple positional files, --all, and the
// two mutual-exclusion rejections (mix --issue+files, --all+explicit targets).
func TestResolveValidateTargets(t *testing.T) {
	dir := t.TempDir()
	mk := func(name, body string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	f1 := mk("000001-one.md", gateGood)
	f2 := mk("000002-two.md", gateGood)
	mk("000003-three.md", gateNoPlan)

	t.Run("comma-separated IDs resolve to their files in order", func(t *testing.T) {
		got, err := resolveValidateTargets(&issueValidateFlags{Issues: []int{1, 2}, IssuesDir: dir}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if want := []string{f1, f2}; !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("unknown ID errors", func(t *testing.T) {
		if _, err := resolveValidateTargets(&issueValidateFlags{Issues: []int{99}, IssuesDir: dir}, nil); err == nil {
			t.Error("an unknown --issue ID should error")
		}
	})

	t.Run("multiple positional files pass through unchanged", func(t *testing.T) {
		args := []string{"a.md", "b.md", "c.md"}
		got, err := resolveValidateTargets(&issueValidateFlags{IssuesDir: dir}, args)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, args) {
			t.Errorf("got %v, want %v", got, args)
		}
	})

	t.Run("--all globs every issue file in the dir", func(t *testing.T) {
		got, err := resolveValidateTargets(&issueValidateFlags{All: true, IssuesDir: dir}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 3 {
			t.Errorf("--all should glob all 3 issue files, got %d: %v", len(got), got)
		}
	})

	t.Run("mixing --issue and positional files is rejected", func(t *testing.T) {
		if _, err := resolveValidateTargets(&issueValidateFlags{Issues: []int{1}, IssuesDir: dir}, []string{"a.md"}); err == nil {
			t.Error("mixing --issue with positional files should error")
		}
	})

	t.Run("--all combined with explicit targets is rejected", func(t *testing.T) {
		if _, err := resolveValidateTargets(&issueValidateFlags{All: true, IssuesDir: dir}, []string{"a.md"}); err == nil {
			t.Error("--all + positional file should error")
		}
		if _, err := resolveValidateTargets(&issueValidateFlags{All: true, Issues: []int{1}, IssuesDir: dir}, nil); err == nil {
			t.Error("--all + --issue should error")
		}
	})

	t.Run("no target errors", func(t *testing.T) {
		if _, err := resolveValidateTargets(&issueValidateFlags{IssuesDir: dir}, nil); err == nil {
			t.Error("no target should error")
		}
	})
}

// TestIssueValidateCmdCommaIDs pins the end-to-end cobra wiring: the IntSliceVar
// flag actually parses `--issue 1,2` into both IDs and validates each (#133).
func TestIssueValidateCmdCommaIDs(t *testing.T) {
	of := validateFrontmatterFn
	defer func() { validateFrontmatterFn = of }()
	validateFrontmatterFn = func(string) (string, bool, error) { return "", true, nil }

	dir := t.TempDir()
	for _, name := range []string{"000001-one.md", "000002-two.md"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(gateGood), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	cmd := newIssueValidateCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--issues-dir", dir, "--issue", "1,2"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("`validate --issue 1,2` should parse + validate both, got: %v\n%s", err, out.String())
	}
}
