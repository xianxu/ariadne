package workspace

import (
	"os"
	"path/filepath"
)

// WorkflowFile resolves the workflow supplier, preserving a disposable tree anchor.
func WorkflowFile(root, anchor string) (string, error) {
	p := filepath.Join(root, "Makefile.workflow")
	mk, err := filepath.EvalSymlinks(p)
	if err == nil || anchor == "" {
		return mk, err
	}
	if link, lerr := os.Readlink(p); lerr == nil && !filepath.IsAbs(link) {
		if mk, lerr := filepath.EvalSymlinks(filepath.Join(anchor, link)); lerr == nil {
			return mk, nil
		}
	}
	if mk, aerr := filepath.EvalSymlinks(filepath.Join(anchor, "Makefile.workflow")); aerr == nil {
		return mk, nil
	}
	return "", err
}

// SuiteHelper resolves the real helper as well as the workflow symlink.
// A helper symlink may select a different checkout from Makefile.workflow.
func SuiteHelper(root, anchor string) (string, error) {
	mk, err := WorkflowFile(root, anchor)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(filepath.Join(filepath.Dir(mk), "scripts", "test-lock.py"))
}
