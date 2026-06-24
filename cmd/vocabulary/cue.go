package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

// CueRunner is the IO seam over the `cue` CLI (ARCH-PURE — callers stay testable
// with a fake). CUE is a BUILD-TIME tool here: we shell out to export JSON and
// vet, so consumers (sdlc) carry no cuelang Go dependency.
type CueRunner interface {
	Vet(path string) error
	Export(path string) ([]byte, error)
}

// osCue is the production runner backed by the `cue` binary on PATH.
type osCue struct{}

func (osCue) Vet(path string) error {
	out, err := exec.Command("cue", "vet", path).CombinedOutput()
	if err != nil {
		return fmt.Errorf("cue vet %s failed: %v\n%s", path, err, out)
	}
	return nil
}

func (osCue) Export(path string) ([]byte, error) {
	cmd := exec.Command("cue", "export", path, "--out", "json")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr // stdout stays clean for the JSON; CUE's diagnostic lands here
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("cue export %s failed: %v\n%s", path, err, stderr.Bytes())
	}
	return out, nil
}
