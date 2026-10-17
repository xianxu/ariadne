package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issueassets"
)

// prepareAssetArchive validates the complete body-indexed bundle before moving
// a legacy issue body. Returned paths retain the caller's Git path convention.
func prepareAssetArchive(ctx context.Context, root, detail, archived string, body []byte, untracked func(string) bool) ([]preparedArchiveMove, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	lexicalRoot := absRoot
	absRoot, err = filepath.EvalSymlinks(absRoot)
	if err != nil {
		return nil, err
	}
	detailLocal := detail
	if filepath.IsAbs(detailLocal) {
		detailLocal, err = filepath.Rel(lexicalRoot, detailLocal)
		if err != nil {
			return nil, err
		}
	}
	detailFull := filepath.Join(absRoot, detailLocal)
	id := issueIDPrefix(filepath.Base(detail))
	body = issueAssetBody(body)
	refs, err := issueassets.ParseRefs(id, body)
	if err != nil {
		return nil, err
	}
	archivedLocal := archived
	if filepath.IsAbs(archivedLocal) {
		archivedLocal, err = filepath.Rel(lexicalRoot, archivedLocal)
		if err != nil {
			return nil, err
		}
	}
	archiveDir := filepath.Join(filepath.Dir(archivedLocal), "assets", id)
	// Legacy callers without either asset directory need no repository-root IO.
	if len(refs) == 0 {
		_, sourceErr := os.Lstat(filepath.Join(filepath.Dir(detailFull), "assets", id))
		_, archiveErr := os.Lstat(filepath.Join(absRoot, archiveDir))
		if os.IsNotExist(sourceErr) && os.IsNotExist(archiveErr) {
			return nil, nil
		}
	}
	// The destination must be empty as a whole: a differently named old member
	// would become an unindexed asset after the body moves, including when the
	// new body carries no assets. Check directories before opening any member.
	if err := assetDirectories(absRoot, filepath.ToSlash(archiveDir), false); err == nil {
		dir, err := os.Open(filepath.Join(absRoot, archiveDir))
		if err != nil {
			return nil, fmt.Errorf("read archive assets %s: %w; repair that directory before retrying", archiveDir, err)
		}
		names, readErr := dir.Readdirnames(1)
		dir.Close()
		if len(names) != 0 {
			return nil, fmt.Errorf("archive asset already exists: %s; remove that archived asset before retrying", filepath.Join(archiveDir, names[0]))
		}
		if readErr != nil && readErr != io.EOF {
			return nil, fmt.Errorf("read archive assets %s: %w; repair that directory before retrying", archiveDir, readErr)
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	detailRel, err := gitx.InsideRoot(absRoot, detailFull)
	if err != nil {
		return nil, err
	}
	files, err := readAssetFiles(absRoot, detailRel, id)
	if err != nil {
		return nil, err
	}
	bundle, err := issueassets.BuildBundle(id, body, files)
	if err != nil {
		return nil, err
	}
	var moves []preparedArchiveMove
	for name := range bundle.Files {
		src := filepath.Join(filepath.Dir(detail), filepath.FromSlash(name))
		dst := filepath.Join(filepath.Dir(archived), filepath.FromSlash(name))
		local := dst
		if filepath.IsAbs(local) {
			local, err = filepath.Rel(lexicalRoot, local)
			if err != nil {
				return nil, err
			}
		}
		full := filepath.Join(absRoot, local)
		rel, err := gitx.InsideRoot(absRoot, local)
		if err != nil {
			return nil, err
		}
		if err := assetDirectories(absRoot, filepath.ToSlash(filepath.Dir(rel)), false); err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		if _, err := os.Lstat(full); err == nil {
			return nil, fmt.Errorf("archive destination occupied: %s; reconcile or remove that destination before retrying", dst)
		} else if !os.IsNotExist(err) {
			return nil, err
		}
		if err := verifyAssetAttributes(ctx, absRoot, rel); err != nil {
			return nil, err
		}
		moves = append(moves, preparedArchiveMove{IssuePath: src, HistoryPath: dst, SourceUntracked: untracked(src)})
	}
	sort.Slice(moves, func(i, j int) bool { return moves[i].IssuePath < moves[j].IssuePath })
	return moves, nil
}

// finishArchiveMemberMoves moves the prepared members of a folder that travels
// with its issue: its assets, and its documents' side folders (#424). It
// links then removes, never replacing a destination, and skips a member whose
// source is already gone, which makes a resumed archive idempotent.
func finishArchiveMemberMoves(root string, moves []preparedArchiveMove) error {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	lexicalRoot := absRoot
	absRoot, err = filepath.EvalSymlinks(absRoot)
	if err != nil {
		return err
	}
	for _, m := range moves {
		src, dst := m.IssuePath, m.HistoryPath
		if !filepath.IsAbs(src) {
			src = filepath.Join(absRoot, src)
		}
		if !filepath.IsAbs(dst) {
			dst = filepath.Join(absRoot, dst)
		}
		// Already prepared moves need no filesystem effect on recovery.
		if _, err := os.Lstat(src); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return err
		}
		local := m.HistoryPath
		if filepath.IsAbs(local) {
			local, err = filepath.Rel(lexicalRoot, local)
			if err != nil {
				return err
			}
		}
		rel, err := gitx.InsideRoot(absRoot, local)
		if err != nil {
			return err
		}
		if err := assetDirectories(absRoot, filepath.ToSlash(filepath.Dir(rel)), true); err != nil {
			return err
		}
		// No replacement: an independently created destination must never be lost.
		if err := os.Link(src, dst); err != nil {
			return fmt.Errorf("archive retained %s: %w; reconcile the two paths then rerun sdlc push --yes", m.IssuePath, err)
		}
		if err := os.Remove(src); err != nil {
			return fmt.Errorf("archive retained both %s and %s: %w; remove the duplicate active file then rerun sdlc push --yes", m.IssuePath, m.HistoryPath, err)
		}
	}
	return nil
}

// recoverAssetArchiveMoves uses moved bodies as authority, checking active and
// archived members together so a stopped legacy move can resume safely.
func recoverAssetArchiveMoves(ctx context.Context, moves []preparedArchiveMove, entries []gitx.StatusEntry, issuesDir string) ([]preparedArchiveMove, []string, error) {
	expected := map[string]bool{}
	var assets []preparedArchiveMove
	for _, m := range moves {
		if !isIssuePath(m.IssuePath, issuesDir) {
			continue
		}
		body, err := os.ReadFile(m.HistoryPath)
		if err != nil {
			return nil, nil, err
		}
		id := issueIDPrefix(filepath.Base(m.IssuePath))
		root, err := os.Getwd()
		if err != nil {
			return nil, nil, err
		}
		active, err := readAssetFiles(root, m.IssuePath, id)
		if err != nil {
			return nil, nil, err
		}
		archived, err := readAssetFiles(root, m.HistoryPath, id)
		if err != nil {
			return nil, nil, err
		}
		union := map[string][]byte{}
		for p, b := range active {
			union[p] = b
		}
		for p, b := range archived {
			if a, ok := union[p]; ok {
				if !bytes.Equal(a, b) {
					return nil, nil, fmt.Errorf("conflicting active and archived asset %s: restore the original bytes before recovery", p)
				}
				return nil, nil, fmt.Errorf("duplicate active and archived asset %s: remove the duplicate active file then rerun sdlc push --yes", p)
			}
			union[p] = b
		}
		bundle, err := issueassets.BuildBundle(id, issueAssetBody(body), union)
		if err != nil {
			return nil, nil, err
		}
		for p := range bundle.Files {
			src := filepath.Join(filepath.Dir(m.IssuePath), filepath.FromSlash(p))
			dst := filepath.Join(filepath.Dir(m.HistoryPath), filepath.FromSlash(p))
			if err := verifyAssetAttributes(ctx, root, filepath.ToSlash(dst)); err != nil {
				return nil, nil, err
			}
			expected[src], expected[dst] = true, true
			assets = append(assets, preparedArchiveMove{IssuePath: src, HistoryPath: dst})
		}
	}
	var other []string
	for _, e := range entries {
		if e.Orig != "" {
			if !expected[e.Orig] || !expected[e.Path] {
				other = append(other, statusLine(e))
			}
		} else if !expected[e.Path] || !(strings.Contains(e.XY, "D") || strings.Contains(e.XY, "A") || e.XY == "??") {
			other = append(other, statusLine(e))
		}
	}
	return assets, other, nil
}

// Called after staging and immediately before each legacy archive commit.
// Content-addressed names make this a byte proof even for recovered/untracked
// evidence: changed bytes cannot retain their original member name.
func verifyArchivedAssetMoves(ctx context.Context, root string, moves []preparedArchiveMove) error {
	ids := map[string]bool{}
	for _, m := range moves {
		parent := filepath.Dir(m.HistoryPath)
		id := filepath.Base(parent)
		if filepath.Base(filepath.Dir(parent)) == "assets" && issueIDPrefix(id+"-asset.md") == id {
			ids[id] = true
		}
	}
	if len(ids) == 0 {
		return nil
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	if canonical, err := filepath.EvalSymlinks(absRoot); err == nil {
		absRoot = canonical
	}
	for _, m := range moves {
		id := issueIDPrefix(filepath.Base(m.HistoryPath))
		if !ids[id] {
			continue
		}
		// The issue body, not a plan sidecar, is the bundle's index.
		if filepath.Base(filepath.Dir(m.HistoryPath)) != "issues" {
			continue
		}
		rel, err := gitx.InsideRoot(absRoot, m.HistoryPath)
		if err != nil {
			return err
		}
		body, err := os.ReadFile(filepath.Join(absRoot, filepath.FromSlash(rel)))
		if err != nil {
			return fmt.Errorf("read archived body %s: %w; restore that body before retrying", rel, err)
		}
		files, err := readAssetFiles(absRoot, rel, id)
		if err != nil {
			return err
		}
		for p := range files {
			if err := verifyAssetAttributes(ctx, absRoot, filepath.ToSlash(filepath.Join(filepath.Dir(rel), p))); err != nil {
				return err
			}
		}
		if _, err := issueassets.BuildBundle(id, issueAssetBody(body), files); err != nil {
			return fmt.Errorf("verify archived assets beside %s: %w; restore the original asset bytes before retrying", rel, err)
		}
		delete(ids, id)
	}
	if len(ids) > 0 {
		return fmt.Errorf("archived assets have no moved issue body; restore their issue body before retrying")
	}
	return nil
}
