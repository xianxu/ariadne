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
	// VetInstance unifies the YAML data at dataPath against the `def` definition
	// (e.g. "#Issue") in schemaPath, via `cue vet -d <def> <data> <schema>` (#124
	// instance-conformance). Returns cue's combined output (empty on a clean vet,
	// the verbose diagnostics on a conformance failure) — the caller renders it via
	// parseCueDiagnostics. A non-nil error means cue couldn't be INVOKED (binary
	// missing); a conformance failure is NOT a Go error, it's diagnostic output.
	VetInstance(dataPath, schemaPath, def string) (output string, err error)
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

func (osCue) VetInstance(dataPath, schemaPath, def string) (string, error) {
	out, err := exec.Command("cue", "vet", "-d", def, dataPath, schemaPath).CombinedOutput()
	if err == nil {
		return "", nil // clean
	}
	// A non-zero exit from `cue vet` (an *exec.ExitError) means cue RAN and found
	// violations — that's the expected failure path; hand the diagnostics back as
	// output, not a Go error. Anything else (cue not on PATH) is a real run error.
	if _, ok := err.(*exec.ExitError); ok {
		return string(out), nil
	}
	return "", fmt.Errorf("cue vet -d %s failed to run: %v\n%s", def, err, out)
}
