package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

type ResourceRoot struct {
	Token string
	Path  string
}

type resourceInventory struct {
	Version int `json:"version"`
	Sources []struct {
		Kind      string   `json:"kind"`
		RootToken string   `json:"root_token"`
		Patterns  []string `json:"patterns"`
	} `json:"sources"`
}

type resourceAssignment struct {
	kind string
	size int64
}

type ResourceFile struct {
	Path  string `json:"path"`
	Bytes int64  `json:"bytes"`
}

type ResourceCollection struct {
	Families        []ResourceObservation
	LargestLedgers  []ResourceFile
	Visited         int
	CoverageUnknown []string
}

const resourceFamilyEntryLimit = 2_000_000

func collectResourceFamilies(inventoryJSON []byte, roots []ResourceRoot) ([]ResourceObservation, int, error) {
	result, err := collectResourceDetails(inventoryJSON, roots)
	return result.Families, result.Visited, err
}

// collectResourceDetails walks each resolved root once. Selectors are Pair's
// slash-separated path.Match patterns; a selected directory owns descendants.
func collectResourceDetails(inventoryJSON []byte, roots []ResourceRoot) (ResourceCollection, error) {
	return collectResourceDetailsWithLimit(inventoryJSON, roots, resourceFamilyEntryLimit)
}

func collectResourceDetailsWithLimit(inventoryJSON []byte, roots []ResourceRoot, limit int) (ResourceCollection, error) {
	var inventory resourceInventory
	if err := json.Unmarshal(inventoryJSON, &inventory); err != nil || inventory.Version != 1 || len(inventory.Sources) == 0 {
		return ResourceCollection{}, fmt.Errorf("invalid resource inventory: %v", err)
	}
	byKind := map[string]*ResourceObservation{}
	byToken := map[string][]int{}
	for i, source := range inventory.Sources {
		if source.Kind == "" || source.RootToken == "" || len(source.Patterns) == 0 || byKind[source.Kind] != nil {
			return ResourceCollection{}, fmt.Errorf("invalid or duplicate resource source %q", source.Kind)
		}
		for _, pattern := range source.Patterns {
			if pattern == "" || strings.HasPrefix(pattern, "/") || strings.Contains(pattern, "..") {
				return ResourceCollection{}, fmt.Errorf("invalid resource pattern %q", pattern)
			}
			if _, err := path.Match(pattern, ""); err != nil {
				return ResourceCollection{}, fmt.Errorf("invalid resource pattern %q: %w", pattern, err)
			}
		}
		byKind[source.Kind] = &ResourceObservation{Kind: source.Kind, Complete: true}
		byToken[source.RootToken] = append(byToken[source.RootToken], i)
	}
	rootsByToken := map[string][]string{}
	for _, root := range roots {
		if root.Token != "" && root.Path != "" {
			rootsByToken[root.Token] = append(rootsByToken[root.Token], filepath.Clean(root.Path))
		}
	}
	for _, source := range inventory.Sources {
		paths := append([]string{}, rootsByToken[source.RootToken]...)
		sort.Strings(paths)
		definition, _ := json.Marshal(struct {
			Rule  any
			Roots []string
		}{source, paths})
		hash := sha256.Sum256(definition)
		byKind[source.Kind].Definition = hex.EncodeToString(hash[:])
		if len(paths) == 0 {
			markResourceUnknown(byKind[source.Kind], "no root for "+source.RootToken)
		}
	}
	assignments := map[string]resourceAssignment{}
	blocked := map[string]bool{}
	visited := 0
	for _, root := range roots {
		indices := byToken[root.Token]
		if len(indices) == 0 || root.Path == "" {
			continue
		}
		if info, err := os.Lstat(root.Path); err != nil || !info.IsDir() {
			for _, i := range indices {
				markResourceUnknown(byKind[inventory.Sources[i].Kind], "unreadable root "+root.Path)
			}
			continue
		}
		err := filepath.WalkDir(root.Path, func(name string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			visited++
			if visited > limit {
				return fmt.Errorf("resource metadata walk limit of %d entries reached", limit)
			}
			if entry.Type()&os.ModeSymlink != 0 || entry.IsDir() {
				return nil
			}
			if !entry.Type().IsRegular() {
				return nil
			}
			rel, err := filepath.Rel(root.Path, name)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)
			matched := map[string]bool{}
			for _, i := range indices {
				source := inventory.Sources[i]
				for _, pattern := range source.Patterns {
					if resourcePatternSelects(pattern, rel) {
						matched[source.Kind] = true
						break
					}
				}
			}
			if len(matched) == 0 {
				return nil
			}
			if len(matched) > 1 {
				for kind := range matched {
					markResourceUnknown(byKind[kind], "overlapping source assignment")
				}
				blocked[name] = true
				return nil
			}
			if blocked[name] {
				return nil
			}
			var kind string
			for kind = range matched {
			}
			if prior, exists := assignments[name]; exists {
				if prior.kind != kind {
					previous := byKind[prior.kind]
					previous.Bytes -= prior.size
					previous.Files--
					markResourceUnknown(previous, "overlapping source assignment")
					markResourceUnknown(byKind[kind], "overlapping source assignment")
					blocked[name] = true
				}
				return nil
			}
			info, err := entry.Info()
			if err != nil || info.Size() < 0 {
				markResourceUnknown(byKind[kind], "unreadable file metadata")
				return nil
			}
			byKind[kind].Bytes += info.Size()
			byKind[kind].Files++
			assignments[name] = resourceAssignment{kind: kind, size: info.Size()}
			return nil
		})
		if err != nil {
			for _, i := range indices {
				markResourceUnknown(byKind[inventory.Sources[i].Kind], "walk root "+root.Path+": "+err.Error())
			}
		}
	}
	rows := make([]ResourceObservation, 0, len(byKind))
	for _, row := range byKind {
		rows = append(rows, *row)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Kind < rows[j].Kind })
	var ledgers []ResourceFile
	for name, assignment := range assignments {
		if assignment.kind == "couch-ledgers" && !blocked[name] {
			ledgers = append(ledgers, ResourceFile{Path: name, Bytes: assignment.size})
		}
	}
	sort.Slice(ledgers, func(i, j int) bool {
		if ledgers[i].Bytes != ledgers[j].Bytes {
			return ledgers[i].Bytes > ledgers[j].Bytes
		}
		return ledgers[i].Path < ledgers[j].Path
	})
	if len(ledgers) > 10 {
		ledgers = ledgers[:10]
	}
	return ResourceCollection{Families: rows, LargestLedgers: ledgers, Visited: visited}, nil
}

func resourcePatternSelects(pattern, rel string) bool {
	for candidate := rel; candidate != "." && candidate != "/"; candidate = path.Dir(candidate) {
		if matched, _ := path.Match(pattern, candidate); matched {
			return true
		}
	}
	return false
}

func markResourceUnknown(row *ResourceObservation, reason string) {
	row.Complete = false
	if row.Reason == "" {
		row.Reason = reason
	} else if !strings.Contains(row.Reason, reason) {
		row.Reason += "; " + reason
	}
}
