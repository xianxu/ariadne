// estimatesource.go — `sdlc estimate-source`: the PULL surface that names both
// halves of estimate derivation in one discoverable output (#134), the
// estimate-side counterpart to `arch-principles`. The SHARED METHOD is
// single-sourced in sdlc (the `## Estimate` grammar + `estimate.Models()`); the
// REPO-LOCAL CALIBRATION (the per-primitive hour ranges, drifting as closes
// accrue) lives in a brain artifact resolved here. start-plan/change-code PUSH a
// one-line variant at the gates; this command is the standalone pull.
//
// Fail-loud lives HERE: when the calibration doc is missing/inaccessible the pull
// exits non-zero. The gate pushes only warn-and-continue (a downstream repo with
// no brain must never break change-code), so this command is the one place that
// refuses — reconciling "fail loudly" with "don't break base-layer dependents."
//
// Rendering is the pure estimate.SourceGuidance; estimateSourceStatus is the thin
// IO seam (stat the doc + its sibling ledger for the stale compare).
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
)

// estimateSourceStatus is the IO seam: resolve the calibration doc, stat it, and
// (for the default brain layout) compare its mtime against the sibling
// calibration ledger — a ledger newer than the doc means closes have accrued
// since the last recalibration, so the per-primitive numbers may have drifted.
func estimateSourceStatus(brainDir, model, override string) estimate.SourceStatus {
	path := estimate.SourcePath(brainDir, model, override)
	st := estimate.SourceStatus{Path: path, Model: model}
	fi, err := os.Stat(path)
	if err != nil {
		return st // Exists stays false
	}
	st.Exists = true
	// Stale check only for the default brain layout (an override is a custom path
	// with no implied sibling ledger). The ledger sits next to the model docs.
	if override == "" {
		ledger := estimate.VelocityPath(brainDir, "calibration-ledger.tsv")
		if lfi, lerr := os.Stat(ledger); lerr == nil && lfi.ModTime().After(fi.ModTime()) {
			st.Stale = true
		}
	}
	return st
}

// NewEstimateSourceCmd returns the cobra command for `sdlc estimate-source`.
func NewEstimateSourceCmd() *cobra.Command {
	var model, brainDir string
	cmd := &cobra.Command{
		Use:           "estimate-source",
		Short:         "Name the shared estimate method + the repo-local calibration source (pull)",
		Long:          "Placeholder — replaced by helptext.MustGet(\"estimate-source\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !estimate.KnownModel(model) {
				return fmt.Errorf("unknown --model %q: want one of %v", model, estimate.Models())
			}
			brainAbs, _ := filepath.Abs(brainDir)
			override := os.Getenv("WF_ESTIMATOR_SRC")
			st := estimateSourceStatus(brainAbs, model, override)
			return reportEstimateSource(cmd.OutOrStdout(), cmd.ErrOrStderr(), st)
		},
	}
	cmd.Flags().StringVar(&model, "model", "estimate-logic-v2", "estimate model whose calibration doc to resolve")
	cmd.Flags().StringVar(&brainDir, "brain-dir", "../brain", "path to the brain repo (holds the calibration doc + ledger)")
	return cmd
}

// reportEstimateSource prints the guidance to stdout and, when the calibration
// source is missing, returns an error so the pull exits non-zero (fail loud).
func reportEstimateSource(stdout, stderr io.Writer, st estimate.SourceStatus) error {
	fmt.Fprint(stdout, estimate.SourceGuidance(st))
	if !st.Exists {
		return fmt.Errorf("estimator calibration source missing: %s", st.Path)
	}
	return nil
}
