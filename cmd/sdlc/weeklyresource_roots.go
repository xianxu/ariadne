package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/pkg/workspace"
)

const resourceInventoryLimit = 1 << 20

func collectWeeklyResources(ctx context.Context, id workspace.Identity, paths []string, flags weeklySnapshotFlags) (ResourceCollection, ResourceMachine) {
	brainIndependent := ResourceCollection{Families: []ResourceObservation{}, LargestLedgers: []ResourceFile{}}
	inventoryPath := flags.ResourceInventory
	if inventoryPath == "" {
		inventoryPath = filepath.Join(id.FleetRoot, "pair", "cmd", "internal", "historyarchive", "sources.json")
	}
	info, err := os.Stat(inventoryPath)
	if err != nil || info.Size() > resourceInventoryLimit {
		brainIndependent.CoverageUnknown = append(brainIndependent.CoverageUnknown, fmt.Sprintf("resource inventory unavailable or exceeds 1 MiB: %s: %v", inventoryPath, err))
	} else {
		data, readErr := os.ReadFile(inventoryPath)
		if readErr != nil {
			brainIndependent.CoverageUnknown = append(brainIndependent.CoverageUnknown, "resource inventory: "+readErr.Error())
		} else {
			roots, warnings := resolveResourceRoots(ctx, id, paths, flags)
			brainIndependent, err = collectResourceDetails(data, roots)
			if err != nil {
				brainIndependent = ResourceCollection{CoverageUnknown: []string{"resource inventory: " + err.Error()}}
			}
			brainIndependent.CoverageUnknown = append(brainIndependent.CoverageUnknown, warnings...)
		}
	}
	scratch := flags.ResourceScratch
	if scratch == "" {
		scratch = "/Volumes/AgentScratch"
	}
	disk := flags.ResourceDisk
	if disk == "" {
		disk = "disk0"
	}
	machine := collectResourceMachine(ctx, resourceMachineDeps{Runner: resourceExecRunner{}, Capacity: resourceStatfsCapacity}, scratch, disk)
	return brainIndependent, machine
}

// resolveResourceRoots binds Pair's declared tokens to the same fleet's
// primary stores. PAIR_DATA_DIR is intentionally not used: a Couch slot sets
// it to one repository's subtree, not the global repos/* namespace.
func resolveResourceRoots(ctx context.Context, id workspace.Identity, paths []string, flags weeklySnapshotFlags) ([]ResourceRoot, []string) {
	home, _ := os.UserHomeDir()
	globalPair := flags.ResourcePairData
	if globalPair == "" {
		base := os.Getenv("XDG_DATA_HOME")
		if base == "" {
			base = filepath.Join(home, ".local", "share")
		}
		globalPair = filepath.Join(base, "pair")
	}
	couchStore := os.Getenv("COUCH_STORE_DIR")
	if couchStore == "" {
		couchStore = filepath.Join(globalPair, "couch")
	}
	fromEnv := func(key, fallback string) string {
		if value := os.Getenv(key); value != "" {
			return value
		}
		return fallback
	}
	roots := []ResourceRoot{
		{Token: "pair-data", Path: globalPair},
		{Token: "couch-store", Path: couchStore},
		{Token: "claude-home", Path: fromEnv("CLAUDE_CONFIG_DIR", filepath.Join(home, ".claude"))},
		{Token: "codex-home", Path: fromEnv("CODEX_HOME", filepath.Join(home, ".codex"))},
		{Token: "qoder-home", Path: fromEnv("QODER_HOME", filepath.Join(home, ".qoder"))},
	}
	var warnings []string
	addGlob := func(token, pattern string) {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			warnings = append(warnings, "resolve "+token+": "+err.Error())
			return
		}
		for _, match := range matches {
			roots = append(roots, ResourceRoot{Token: token, Path: match})
		}
	}
	addGlob("slot-store", filepath.Join(id.FleetRoot, "worktree", "*-slot*", ".couch"))
	addGlob("ops-worktree", filepath.Join(id.FleetRoot, "ops", "teams", "*"))
	seenCommon := map[string]bool{}
	for _, repo := range paths {
		out, err := weeklyGit(ctx, repo, "rev-parse", "--path-format=absolute", "--git-common-dir")
		if err != nil {
			warnings = append(warnings, "git common directory for "+repo+": "+err.Error())
			continue
		}
		common := strings.TrimSpace(string(out))
		if !seenCommon[common] {
			seenCommon[common] = true
			roots = append(roots, ResourceRoot{Token: "git-common", Path: common})
		}
	}
	sort.Slice(roots, func(i, j int) bool {
		if roots[i].Token != roots[j].Token {
			return roots[i].Token < roots[j].Token
		}
		return roots[i].Path < roots[j].Path
	})
	return roots, warnings
}
