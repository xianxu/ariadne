package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
)

// Admission is tool infrastructure, independent of the replay's product pin.
// The verified Metis tool_source supplies this binding; no replay fallback.
type benchSuiteHelper struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

func captureBenchSuiteHelper(source string) (benchSuiteHelper, error) {
	h := benchSuiteHelper{Path: filepath.Join(source, "scripts", "test-lock.py")}
	digest, err := hashBenchStepFile(h.Path, 1<<20)
	if err != nil {
		return h, fmt.Errorf("bench admission supplier: %w", err)
	}
	h.SHA256 = digest
	return h, nil
}

func (h benchSuiteHelper) prepare(ctx context.Context, root string) (string, error) {
	if !filepath.IsAbs(h.Path) || filepath.Clean(h.Path) != h.Path || !regexp.MustCompile(`^[a-f0-9]{64}$`).MatchString(h.SHA256) {
		return "", errors.New("bench admission supplier binding missing or invalid; use a fresh Metis repeat")
	}
	digest, err := hashBenchStepFile(h.Path, 1<<20)
	if err != nil || digest != h.SHA256 {
		return "", fmt.Errorf("bench admission supplier changed or unavailable: %s; use a fresh Metis repeat", h.Path)
	}
	if err := preflightSuiteHelper(ctx, root, root, h.Path); err != nil {
		return "", err
	}
	return h.Path, nil
}
