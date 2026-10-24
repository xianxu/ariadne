package main

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/pkg/workspace"
)

type ownedActualLimits struct {
	Dirs         []string
	Since, Until time.Time
}

func withinManualDirs(path string, dirs []string) bool {
	for _, dir := range dirs {
		rel, err := filepath.Rel(dir, path)
		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

func limitOwnedSourceWindows(original map[string][]activetime.ActivityWindow, limits ownedActualLimits) (map[string][]activetime.ActivityWindow, error) {
	out := map[string][]activetime.ActivityWindow{}
	for file, windows := range original {
		path, err := workspace.CanonicalPath(file)
		if err != nil {
			return nil, err
		}
		if !withinManualDirs(path, limits.Dirs) {
			continue
		}
		for _, w := range windows {
			if !limits.Since.IsZero() && w.Start.Before(limits.Since) {
				w.Start = limits.Since
			}
			if !limits.Until.IsZero() && w.End.After(limits.Until) {
				w.End = limits.Until
			}
			if w.Start.Before(w.End) {
				out[file] = append(out[file], w)
			}
		}
	}
	return out, nil
}
