package main

import (
	"bytes"
	"os"
	"path/filepath"
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
}
