package main

import (
	"bytes"
	"errors"
	"fmt"
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

// issueArchiveDocs lists the documents an issue's archive moves that can
// carry a side folder: the issue file and each of its plan documents.
func issueArchiveDocs(issueSrc, issueDst, plansFull, recPlansDir, recHistoryDir string) []sideArchiveDoc {
	docs := []sideArchiveDoc{{issueSrc, issueDst}}
	base := filepath.Base(issueSrc)
	matches, _ := filepath.Glob(filepath.Join(plansFull, issueIDPrefix(base)+"-*.md"))
	sort.Strings(matches)
	for _, p := range matches {
		if planArtifactBelongsToIssue(base, filepath.Base(p)) {
			docs = append(docs, sideArchiveDoc{
				Src: filepath.Join(recPlansDir, filepath.Base(p)),
				Dst: archiveDestination(recHistoryDir, vocab.ArchivePlans, filepath.Base(p)),
			})
		}
	}
	return docs
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
		if names, _ := readDirNames(dstFull); len(names) > 0 {
			return nil, fmt.Errorf("archive destination occupied: %s holds %s; reconcile or remove that folder before retrying", dst, names[0])
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
func recoverSideArchiveMoves(moves []preparedArchiveMove, entries []gitx.StatusEntry, untracked func(string) bool) ([]preparedArchiveMove, []string, error) {
	deleted := map[string]bool{}
	for _, e := range entries {
		if strings.Contains(e.XY, "D") {
			deleted[e.Path] = true
		}
	}
	expected := map[string]bool{}
	var out []preparedArchiveMove
	for _, m := range moves {
		if !strings.HasSuffix(m.IssuePath, ".md") {
			continue
		}
		src, dst := sidedoc.SideDir(m.IssuePath), sidedoc.SideDir(m.HistoryPath)
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
				a, _ := os.ReadFile(filepath.Join(src, rel))
				b, _ := os.ReadFile(filepath.Join(dst, rel))
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
// archived side doc's parent: still resolves: the folder moved whole, beside
// its document.
func verifyArchivedSideMoves(root string, moves []preparedArchiveMove) error {
	for _, m := range moves {
		if !sidedoc.InSideDir(filepath.ToSlash(m.HistoryPath)) || !strings.HasSuffix(m.HistoryPath, ".md") {
			continue
		}
		full := underRoot(root, m.HistoryPath)
		raw, err := os.ReadFile(full)
		if err != nil {
			return fmt.Errorf("read archived side doc %s: %w; restore it before retrying", m.HistoryPath, err)
		}
		parent := sideParent(string(raw))
		if parent == "" {
			continue // not a side doc (a note kept in the folder); nothing to resolve
		}
		if _, err := os.Stat(filepath.Join(filepath.Dir(full), filepath.FromSlash(parent))); err != nil {
			return fmt.Errorf("archived side doc %s: its parent %s does not resolve; restore the folder beside its document before retrying", m.HistoryPath, parent)
		}
	}
	return nil
}

// sideParent reads a side doc's `parent:` from its frontmatter.
func sideParent(text string) string {
	rest, ok := strings.CutPrefix(text, "---\n")
	if !ok {
		return ""
	}
	fm, _, _ := strings.Cut(rest, "\n---")
	for _, l := range strings.Split(fm, "\n") {
		if v, ok := strings.CutPrefix(l, "parent:"); ok {
			return strings.Trim(strings.TrimSpace(v), `"`)
		}
	}
	return ""
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
	for _, d := range docs {
		src, dst := sidedoc.SideDir(d.Src), sidedoc.SideDir(d.Dst)
		live, err := view.Files(src)
		if err != nil {
			return err
		}
		files := live
		if kept != nil {
			files = nil
			for _, f := range kept {
				if strings.HasPrefix(f.Path, src+"/") {
					files = append(files, f)
				}
			}
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

// trunkSideReplace makes the archived issue's side folder the kept branch's
// (abandon's successor episode): kept members are written, archived members
// the branch no longer has are removed.
func trunkSideReplace(view *gitx.TrunkView, archivedDetail string, a mainArchive, w *gitx.TrunkWrite) error {
	live := path.Join(a.Dirs.Issues, path.Base(archivedDetail))
	keep := map[string]bool{}
	for _, f := range a.Sides {
		if t, ok := sidedoc.MoveMember(f.Path, live, archivedDetail); ok {
			if f.Mode == "120000" {
				return fmt.Errorf("%s is a symlink; side folders hold regular files only", f.Path)
			}
			w.Write[t] = f.Content
			keep[t] = true
		}
	}
	old, err := view.Files(sidedoc.SideDir(archivedDetail))
	if err != nil {
		return err
	}
	for _, f := range old {
		if !keep[f.Path] {
			w.Delete = append(w.Delete, f.Path)
		}
	}
	return nil
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
