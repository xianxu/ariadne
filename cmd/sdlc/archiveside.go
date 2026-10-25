package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/pkg/sidedoc"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// A document's side docs live in its `<name>.side/` folder, and the folder
// moves with the document (#424; xx-universal-review's collaboration-modes.md
// §5). Archiving keeps the base name, so the folder moves as it is: every link
// inside it is relative, and no file is rewritten. These are the folder's
// counterparts of the asset archive's stages (archiveassets.go): prepare
// before anything moves, finish with the shared member finisher, recover an
// interrupted archive from the moved documents, and verify before the commit.

// sideArchiveDoc is one document an issue's archive moves, as its caller
// records paths (root-relative, or absolute on the merge path's rename leg).
type sideArchiveDoc struct{ Src, Dst string }

// sideArchiveDocsForIssue projects the side owners belonging to an issue from
// an existing path inventory. A side root confers an owner even when its .md
// document is absent; this projection never creates that document. Archived
// plan paths normalize to the active owner so interrupted moves share one key.
func sideArchiveDocsForIssue(issueSrc, issueDst string, paths []string, plansDir, historyDir string) []sideArchiveDoc {
	docs := []sideArchiveDoc{{issueSrc, issueDst}}
	seen := map[string]bool{filepath.Clean(issueSrc): true}
	plansDir = filepath.Clean(plansDir)
	archivedPlans := filepath.Clean(vocab.ArchiveSubdir(historyDir, vocab.ArchivePlans))
	base := filepath.Base(issueSrc)
	for _, candidate := range paths {
		owner, side := sidedoc.Owner(filepath.ToSlash(candidate))
		if !side {
			owner = candidate
		}
		owner = filepath.Clean(filepath.FromSlash(owner))
		if filepath.Dir(owner) != plansDir && filepath.Dir(owner) != archivedPlans {
			continue
		}
		if !strings.HasSuffix(owner, ".md") || !planArtifactBelongsToIssue(base, filepath.Base(owner)) {
			continue
		}
		src := filepath.Join(plansDir, filepath.Base(owner))
		if seen[src] {
			continue
		}
		seen[src] = true
		docs = append(docs, sideArchiveDoc{src, archiveDestination(historyDir, vocab.ArchivePlans, filepath.Base(owner))})
	}
	sort.Slice(docs[1:], func(i, j int) bool { return docs[i+1].Src < docs[j+1].Src })
	return docs
}

// issueArchiveDocs inventories top-level plan documents and side roots. The
// recorded roots may differ from their filesystem locations on merge's leg.
func issueArchiveDocs(issueSrc, issueDst, plansFull, recPlansDir, recHistoryDir string) []sideArchiveDoc {
	matches, _ := filepath.Glob(filepath.Join(plansFull, issueIDPrefix(filepath.Base(issueSrc))+"-*"))
	paths := make([]string, 0, len(matches))
	for _, p := range matches {
		paths = append(paths, filepath.Join(recPlansDir, filepath.Base(p)))
	}
	return sideArchiveDocsForIssue(issueSrc, issueDst, paths, recPlansDir, recHistoryDir)
}

// prepareSideArchive lists the member moves for each document's side folder,
// refusing before anything moves: a symlink in a folder (never followed), a
// folder that is not a regular directory, or a destination folder that
// already holds anything.
func prepareSideArchive(root string, docs []sideArchiveDoc, untracked func(string) bool) ([]preparedArchiveMove, error) {
	var moves []preparedArchiveMove
	for _, d := range docs {
		src, dst := sidedoc.SideDir(d.Src), sidedoc.SideDir(d.Dst)
		srcFull, dstFull := underRoot(root, src), underRoot(root, dst)
		info, err := os.Lstat(srcFull)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("side folder %s is not a regular directory; replace it with one before archiving", src)
		}
		if dinfo, err := os.Lstat(dstFull); err == nil {
			if !dinfo.IsDir() || dinfo.Mode()&os.ModeSymlink != 0 {
				return nil, fmt.Errorf("archive destination occupied: %s is not a folder; reconcile or remove it before retrying", dst)
			}
			names, err := readDirNames(dstFull)
			if err != nil && !errors.Is(err, io.EOF) {
				return nil, fmt.Errorf("read archive destination %s: %w; repair it before retrying", dst, err)
			}
			if len(names) > 0 {
				return nil, fmt.Errorf("archive destination occupied: %s holds %s; reconcile or remove that folder before retrying", dst, names[0])
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		members, err := sideFiles(srcFull)
		if err != nil {
			return nil, fmt.Errorf("side folder %s: %w", src, err)
		}
		for _, rel := range members {
			s, t := filepath.Join(src, rel), filepath.Join(dst, rel)
			moves = append(moves, preparedArchiveMove{IssuePath: s, HistoryPath: t, SourceUntracked: untracked != nil && untracked(s)})
		}
	}
	return moves, nil
}

// finishSideArchive moves the prepared members through the shared finisher,
// then removes the emptied source folders.
func finishSideArchive(root string, docs []sideArchiveDoc, moves []preparedArchiveMove) error {
	if err := finishArchiveMemberMoves(root, moves); err != nil {
		return err
	}
	for _, d := range docs {
		removeEmptyDirs(underRoot(root, sidedoc.SideDir(d.Src)))
	}
	return nil
}

// recoverSideArchiveMoves resumes the side folders of an interrupted archive.
// A moved document is the authority: its members are the union of its active
// and archived folders, and the same member in both refuses (a conflict, or a
// duplicate to remove). Status entries inside no moved document's folders are
// returned as other, so the resume refuses them.
func recoverSideArchiveMoves(moves []preparedArchiveMove, entries []gitx.StatusEntry, issuesDir, plansDir, historyDir string, untracked func(string) bool) ([]preparedArchiveMove, []string, error) {
	deleted := map[string]bool{}
	for _, e := range entries {
		if strings.Contains(e.XY, "D") {
			deleted[e.Path] = true
		}
	}
	// Inventory both roots once, then group candidates by issue ID. Even an
	// unchanged active member must finish when its issue already moved.
	byID := map[string][]string{}
	addPath := func(p string) {
		owner, ok := sidedoc.Owner(filepath.ToSlash(p))
		if !ok {
			owner = p
		}
		id := issueIDPrefix(filepath.Base(owner))
		byID[id] = append(byID[id], p)
	}
	for _, dir := range []string{plansDir, vocab.ArchiveSubdir(historyDir, vocab.ArchivePlans)} {
		names, err := os.ReadDir(dir)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, nil, err
		}
		for _, name := range names {
			addPath(filepath.Join(dir, name.Name()))
		}
	}
	for _, e := range entries {
		addPath(e.Path)
		if e.Orig != "" {
			addPath(e.Orig)
		}
	}
	docs := map[string]sideArchiveDoc{}
	for _, m := range moves {
		if !strings.HasSuffix(m.IssuePath, ".md") {
			continue
		}
		docs[m.IssuePath] = sideArchiveDoc{m.IssuePath, m.HistoryPath}
		if isIssuePath(m.IssuePath, issuesDir) {
			for _, d := range sideArchiveDocsForIssue(m.IssuePath, m.HistoryPath, byID[issueIDPrefix(filepath.Base(m.IssuePath))], plansDir, historyDir) {
				docs[d.Src] = d
			}
		}
	}
	ordered := make([]sideArchiveDoc, 0, len(docs))
	for _, d := range docs {
		ordered = append(ordered, d)
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Src < ordered[j].Src })
	expected := map[string]bool{}
	var out []preparedArchiveMove
	for _, d := range ordered {
		src, dst := sidedoc.SideDir(d.Src), sidedoc.SideDir(d.Dst)
		active, err := sideFilesIfAny(src)
		if err != nil {
			return nil, nil, err
		}
		archived, err := sideFilesIfAny(dst)
		if err != nil {
			return nil, nil, err
		}
		inActive := map[string]bool{}
		for _, rel := range active {
			inActive[rel] = true
		}
		union := append([]string{}, active...)
		for _, rel := range archived {
			if inActive[rel] {
				a, err := os.ReadFile(filepath.Join(src, rel))
				if err != nil {
					return nil, nil, err
				}
				b, err := os.ReadFile(filepath.Join(dst, rel))
				if err != nil {
					return nil, nil, err
				}
				if !bytes.Equal(a, b) {
					return nil, nil, fmt.Errorf("conflicting active and archived side doc %s: restore the original bytes before recovery", filepath.Join(src, rel))
				}
				return nil, nil, fmt.Errorf("duplicate active and archived side doc %s: remove the active copy then rerun sdlc push --yes", filepath.Join(src, rel))
			}
			union = append(union, rel)
		}
		for _, rel := range union {
			s, t := filepath.Join(src, rel), filepath.Join(dst, rel)
			expected[s], expected[t] = true, true
			// An active member is tracked if git knows it; a moved one was
			// tracked if its deletion shows, else it was untracked (#154).
			wasUntracked := !deleted[s]
			if inActive[rel] {
				wasUntracked = untracked != nil && untracked(s)
			}
			out = append(out, preparedArchiveMove{IssuePath: s, HistoryPath: t, SourceUntracked: wasUntracked})
		}
	}
	var other []string
	for _, e := range entries {
		if e.Orig != "" {
			if !expected[e.Orig] || !expected[e.Path] {
				other = append(other, statusLine(e))
			}
		} else if !expected[e.Path] {
			other = append(other, statusLine(e))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].IssuePath < out[j].IssuePath })
	return out, other, nil
}

// verifyArchivedSideMoves checks, before the archive commit, that every
// side-folder member arrived: present at its destination as a regular file,
// gone from its source. Nothing here judges a side doc's content: archiving
// keeps the base name and moves the folder whole, so a parent: resolves after
// the move exactly when it did before, and a check that could only fail after
// every move would strand the archive halfway.
func verifyArchivedSideMoves(root string, moves []preparedArchiveMove) error {
	for _, m := range moves {
		if !sidedoc.InSideDir(filepath.ToSlash(m.HistoryPath)) {
			continue
		}
		if info, err := os.Lstat(underRoot(root, m.HistoryPath)); err != nil || !info.Mode().IsRegular() {
			return fmt.Errorf("archived side doc %s is missing or not a regular file; rerun sdlc push --yes to finish the archive", m.HistoryPath)
		}
		if _, err := os.Lstat(underRoot(root, m.IssuePath)); err == nil {
			return fmt.Errorf("side doc %s is still at its source; remove the duplicate active copy, then rerun sdlc push --yes", m.IssuePath)
		}
	}
	return nil
}

func underRoot(root, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(root, p)
}

// sideFiles lists a folder's files, relative to it, refusing a symlink.
func sideFiles(dir string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("%s is a symlink; side folders hold regular files only", p)
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		out = append(out, rel)
		return nil
	})
	sort.Strings(out)
	return out, err
}

func sideFilesIfAny(dir string) ([]string, error) {
	if _, err := os.Lstat(dir); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	return sideFiles(dir)
}

func readDirNames(dir string) ([]string, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return f.Readdirnames(1)
}

// removeEmptyDirs removes dir and its subdirectories when they hold no files,
// deepest first; a folder that still holds anything stays.
func removeEmptyDirs(dir string) {
	var dirs []string
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err == nil && d.IsDir() {
			dirs = append(dirs, p)
		}
		return nil
	})
	for i := len(dirs) - 1; i >= 0; i-- {
		_ = os.Remove(dirs[i]) // fails, harmlessly, on a non-empty folder
	}
}

// trunkSideArchive adds docs' side folders to a trunk archive write (#424):
// each member moves under its document's archived folder, bytes unchanged.
// kept, when not nil, is the kept branch's side files at their live paths
// (abandon archives the branch's work): they are archived instead of main's,
// and main's live copies are removed. A destination folder holding anything,
// or a symlink, refuses.
func trunkSideArchive(view *gitx.TrunkView, docs []sideArchiveDoc, w *gitx.TrunkWrite, kept []gitx.TreeFile) error {
	keptByOwner := map[string][]gitx.TreeFile{}
	for _, f := range kept {
		if owner, ok := sidedoc.Owner(f.Path); ok {
			keptByOwner[owner] = append(keptByOwner[owner], f)
		}
	}
	for _, d := range docs {
		src, dst := sidedoc.SideDir(d.Src), sidedoc.SideDir(d.Dst)
		live, err := view.Files(src)
		if err != nil {
			return err
		}
		files := live
		if kept != nil {
			files = keptByOwner[d.Src]
		}
		if len(files)+len(live) == 0 {
			continue
		}
		if taken, err := trunkFirstFile(view, dst); err != nil {
			return err
		} else if taken != "" {
			return fmt.Errorf("archive destination occupied: %s; reconcile or remove that archived side doc before retrying", taken)
		}
		for _, f := range files {
			if f.Mode == "120000" {
				return fmt.Errorf("%s is a symlink; side folders hold regular files only", f.Path)
			}
			if t, ok := sidedoc.MoveMember(f.Path, d.Src, d.Dst); ok {
				w.Write[t] = f.Content
			}
		}
		for _, f := range live {
			w.Delete = append(w.Delete, f.Path)
		}
	}
	return nil
}

// trunkSideReplace makes the archived documents' side folders the kept
// branch's (abandon's successor episode), for the issue and every plan the
// episode rewrites: kept members are written, archived members the branch no
// longer has are removed.
func trunkSideReplace(view *gitx.TrunkView, archivedDetail string, a mainArchive, w *gitx.TrunkWrite) error {
	issueSrc := path.Join(a.Dirs.Issues, path.Base(archivedDetail))
	oldPlans, err := view.Files(vocab.ArchiveSubdir(a.Dirs.History, vocab.ArchivePlans))
	if err != nil {
		return err
	}
	oldIssue, err := view.Files(sidedoc.SideDir(archivedDetail))
	if err != nil {
		return err
	}
	paths := make([]string, 0, len(oldPlans)+len(a.Sides)+len(a.Plans))
	for _, f := range oldPlans {
		paths = append(paths, f.Path)
	}
	for name := range a.Plans {
		paths = append(paths, path.Join(a.Dirs.Plans, name))
	}
	keptByOwner := map[string][]gitx.TreeFile{}
	for _, f := range a.Sides {
		paths = append(paths, f.Path)
		if owner, ok := sidedoc.Owner(f.Path); ok {
			keptByOwner[owner] = append(keptByOwner[owner], f)
		}
	}
	oldByOwner := map[string][]gitx.TreeFile{}
	for _, files := range [][]gitx.TreeFile{oldPlans, oldIssue} {
		for _, f := range files {
			if owner, ok := sidedoc.Owner(f.Path); ok {
				oldByOwner[owner] = append(oldByOwner[owner], f)
			}
		}
	}
	docs := sideArchiveDocsForIssue(issueSrc, archivedDetail, paths, a.Dirs.Plans, a.Dirs.History)
	for _, d := range docs {
		keep := map[string]bool{}
		for _, f := range keptByOwner[d.Src] {
			if t, ok := sidedoc.MoveMember(f.Path, d.Src, d.Dst); ok {
				if f.Mode == "120000" {
					return fmt.Errorf("%s is a symlink; side folders hold regular files only", f.Path)
				}
				w.Write[t] = f.Content
				keep[t] = true
			}
		}
		for _, f := range oldByOwner[d.Dst] {
			if !keep[f.Path] {
				w.Delete = append(w.Delete, f.Path)
			}
		}
	}
	return nil
}

// trunkSideResidue names an active side-doc file of the issue or of one of
// its plans ("" when none): what a retry must refuse once the details moved.
func trunkSideResidue(view *gitx.TrunkView, detailRel, plansDir string) (string, error) {
	if r, err := trunkFirstFile(view, sidedoc.SideDir(detailRel)); err != nil || r != "" {
		return r, err
	}
	base, first := path.Base(detailRel), ""
	_, err := view.FilesSelected(plansDir, func(f gitx.TreeFile) bool {
		if folder, ok := sidedoc.Folder(f.Path); ok && first == "" && path.Dir(folder) == plansDir && planArtifactBelongsToIssue(base, path.Base(folder)) {
			first = f.Path
		}
		return false
	})
	return first, err
}

// trunkFirstFile names one file under dir on the trunk ("" when none),
// without fetching any blob.
func trunkFirstFile(view *gitx.TrunkView, dir string) (string, error) {
	first := ""
	_, err := view.FilesSelected(dir, func(f gitx.TreeFile) bool {
		if first == "" {
			first = f.Path
		}
		return false
	})
	return first, err
}
