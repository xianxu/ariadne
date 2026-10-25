package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/fleet"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// Match Couch's existing inventory transport bound. Encode before publishing,
// so overflow is unavailable evidence rather than a truncated success.
const maxFleetInventoryJSON = 1 << 20

type fleetIntentObserver func(context.Context, string, string) ([]mergeIntentReport, error)

type fleetIntentRead struct {
	Common  string
	Root    string
	Reports []mergeIntentReport
	Err     error
}

type fleetIntentPlacementCounts struct{ Rows, Records, RootLookups int }

func observeFleetMergeIntents(ctx context.Context, root, pause string) ([]mergeIntentReport, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	reports, err := observeMergeIntentsWithPause(execGitRunner{}, root, func(mergeIntentReport) bool { return true }, func() string { return pause })
	for n := range reports {
		if reports[n].Root == "" {
			continue
		}
		canonical, canonicalErr := workspace.CanonicalPath(reports[n].Root)
		if canonicalErr != nil {
			reports[n].State, reports[n].Root, reports[n].Reentry = "unreadable", "", ""
			reports[n].Note = "cannot canonicalize intent root: " + canonicalErr.Error()
			continue
		}
		reports[n].Root = canonical
	}
	return reports, errors.Join(err, ctx.Err())
}

func collectFleetMergeIntents(ctx context.Context, inventory fleet.Inventory, observe fleetIntentObserver, pause func() string) (fleet.Inventory, error) {
	groups := map[string]string{}
	for _, row := range inventory.Rows {
		if old, ok := groups[row.RepoIdentity]; !ok || row.TreePath < old {
			groups[row.RepoIdentity] = row.TreePath
		}
	}
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	note := ""
	if len(keys) > 0 {
		note = pause()
	}
	reads := make([]fleetIntentRead, 0, len(keys))
	for _, common := range keys {
		if err := ctx.Err(); err != nil {
			return fleet.Inventory{}, err
		}
		reports, err := observe(ctx, groups[common], note)
		reads = append(reads, fleetIntentRead{common, groups[common], reports, err})
	}
	if err := ctx.Err(); err != nil {
		return fleet.Inventory{}, err
	}
	result, _ := placeFleetMergeIntents(inventory, reads)
	return result, nil
}

// placeFleetMergeIntents consumes canonical observations only. No filesystem,
// process, queue or Git calls occur here. All input slices remain caller-owned.
func placeFleetMergeIntents(inventory fleet.Inventory, reads []fleetIntentRead) (fleet.Inventory, fleetIntentPlacementCounts) {
	counts := fleetIntentPlacementCounts{}
	inventory.Rows = append([]fleet.TreeRow{}, inventory.Rows...)
	inventory.Diagnostics = append([]fleet.RepoDiagnostic{}, inventory.Diagnostics...)
	inventory.Capabilities.MergeIntents = 1
	index := map[string]int{}
	groupRows := map[string][]int{}
	groupErrors := map[string][]string{}
	unreadable := map[string]bool{}
	for n := range inventory.Rows {
		counts.Rows++
		row := &inventory.Rows[n]
		row.MergeIntents, row.MergeIntentsState, row.MergeIntentsError = []mergeIntentReport{}, "complete", ""
		key := row.RepoIdentity + "\x00" + row.TreePath
		if _, exists := index[key]; exists {
			index[key] = -1
		} else {
			index[key] = n
		}
		groupRows[row.RepoIdentity] = append(groupRows[row.RepoIdentity], n)
	}
	seen := map[string]mergeIntentReport{}
	conflicts := map[string]bool{}
	for _, read := range reads {
		for _, report := range read.Reports {
			if old, ok := seen[report.Identity]; ok && old != report {
				conflicts[report.Identity] = true
			}
			seen[report.Identity] = report
		}
	}
	diagnosed := map[string]bool{}
	placed := map[string]bool{}
	for _, read := range reads {
		if read.Err != nil {
			groupErrors[read.Common] = append(groupErrors[read.Common], read.Err.Error())
			unreadable[read.Common] = len(read.Reports) == 0
		}
		for _, report := range read.Reports {
			counts.Records++
			counts.RootLookups++
			rowIndex, matched := index[read.Common+"\x00"+report.Root]
			reason := ""
			switch {
			case conflicts[report.Identity]:
				reason = "conflicting observations of physical merge intent"
			case report.Identity == "" || report.Identity != report.Path:
				reason = "missing or inconsistent physical merge intent identity"
			case report.Root == "":
				reason = "merge intent has no trustworthy root: " + report.Note
			case !matched || rowIndex < 0:
				reason = "merge intent root has no unique row in its Git common-dir store"
			}
			if reason != "" {
				groupErrors[read.Common] = append(groupErrors[read.Common], reason+": "+report.Path)
				key := report.Identity
				if key == "" {
					key = read.Common + "\x00" + report.Path
				}
				if !diagnosed[key] {
					diagnosed[key] = true
					inventory.Diagnostics = append(inventory.Diagnostics, fleet.RepoDiagnostic{RepoIdentity: read.Common, RepoPath: read.Root, Stage: "merge_intents", Message: reason + ": " + report.Path})
				}
				continue
			}
			if placed[report.Identity] {
				continue
			}
			placed[report.Identity] = true
			row := &inventory.Rows[rowIndex]
			row.MergeIntents = append(row.MergeIntents, report)
			if report.State == "unreadable" {
				groupErrors[read.Common] = append(groupErrors[read.Common], report.Path+": "+report.Note)
			}
		}
	}
	for common, rows := range groupRows {
		message := strings.Join(groupErrors[common], "; ")
		for _, n := range rows {
			row := &inventory.Rows[n]
			if message != "" {
				row.MergeIntentsState, row.MergeIntentsError = "partial", message
				if unreadable[common] {
					row.MergeIntentsState = "unreadable"
				}
			}
			slices.SortFunc(row.MergeIntents, func(a, b mergeIntentReport) int { return strings.Compare(a.Identity, b.Identity) })
		}
	}
	slices.SortFunc(inventory.Diagnostics, func(a, b fleet.RepoDiagnostic) int {
		if c := strings.Compare(a.RepoIdentity, b.RepoIdentity); c != 0 {
			return c
		}
		if c := strings.Compare(a.Stage, b.Stage); c != 0 {
			return c
		}
		return strings.Compare(a.Message, b.Message)
	})
	return inventory, counts
}

func writeFleetInventoryJSON(ctx context.Context, out io.Writer, inventory fleet.Inventory) error {
	raw, err := json.Marshal(inventory)
	if err != nil {
		return err
	}
	if len(raw)+1 > maxFleetInventoryJSON {
		return fmt.Errorf("fleet inventory JSON exceeds %d bytes; no report published", maxFleetInventoryJSON)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	raw = append(raw, '\n')
	n, err := out.Write(raw)
	if err == nil && n != len(raw) {
		return io.ErrShortWrite
	}
	return err
}
