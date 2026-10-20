package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/pkg/sidedoc"
)

// errInterrupted is a test's stop between two steps of an apply; a rerun
// resumes from the tree it left.
var errInterrupted = errors.New("interrupted")

// readMigration reads the snapshot planMigration needs: qa's entries, the
// side folder as it stands, the other Markdown under the session file's
// directory, and couch's registrations and locks for the files involved.
func readMigration(d deps, file, qa string) (migSnapshot, string, error) {
	file, err := filepath.Abs(file)
	if err != nil {
		return migSnapshot{}, "", err
	}
	root := filepath.Dir(file)
	text, err := d.ss.Peek(file)
	if err != nil {
		return migSnapshot{}, "", err
	}
	s := migSnapshot{
		Session: filepath.Base(file), SessionText: text, Dest: map[string][]byte{},
		Outside: map[string][]byte{}, OutsideTime: map[string]time.Time{},
		Registered: map[string]bool{}, Locked: map[string]bool{},
		Exists: func(rel string) bool {
			_, err := os.Lstat(filepath.Join(root, filepath.FromSlash(rel)))
			return err == nil
		},
	}
	if qa == "" {
		f, ok := sidedoc.Fields(text)
		if !ok || strings.Trim(f["operator"], `"`) == "" {
			return s, "", fmt.Errorf("refused: %s names no operator: in its frontmatter; pass --from qa/<operator>", s.Session)
		}
		qa = "qa/" + strings.Trim(f["operator"], `"`)
	}
	s.QA = filepath.ToSlash(filepath.Clean(qa))
	if filepath.IsAbs(qa) || s.QA == "." || s.QA == ".." || strings.HasPrefix(s.QA, "../") {
		return s, "", fmt.Errorf("refused: --from %s must be a folder inside %s's directory", qa, s.Session)
	}
	qaDir := filepath.Join(root, filepath.FromSlash(s.QA))
	if fi, err := os.Stat(qaDir); err != nil || !fi.IsDir() {
		return s, "", fmt.Errorf("refused: %s is not a directory beside %s", s.QA, s.Session)
	}
	err = filepath.WalkDir(qaDir, func(p string, e fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(qaDir, p)
		if e.IsDir() {
			// A dot directory is another tool's (the harness's .claude/),
			// even empty: listed as left, never entered, never removed.
			if p != qaDir && strings.HasPrefix(e.Name(), ".") {
				s.Entries = append(s.Entries, migEntry{Rel: filepath.ToSlash(rel)})
				return filepath.SkipDir
			}
			return nil
		}
		me := migEntry{Rel: filepath.ToSlash(rel), Regular: e.Type().IsRegular()}
		if me.Regular && classifyEntry(me.Rel, true) <= classAttachment {
			if me.Bytes, err = os.ReadFile(p); err != nil {
				return err
			}
			fi, err := e.Info()
			if err != nil {
				return err
			}
			me.Mtime = fi.ModTime()
		}
		s.Entries = append(s.Entries, me)
		return nil
	})
	if err != nil {
		return s, "", err
	}
	side := filepath.Join(root, sidedoc.SideDir(s.Session))
	err = filepath.WalkDir(root, func(p string, e fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		rel = filepath.ToSlash(rel)
		if e.IsDir() {
			if p == qaDir || (rel != "." && strings.HasPrefix(e.Name(), ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		if !e.Type().IsRegular() {
			return nil
		}
		inSide := strings.HasPrefix(p, side+string(filepath.Separator))
		if !inSide && (!strings.HasSuffix(rel, ".md") || rel == s.Session) {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if inSide {
			s.Dest[rel] = b
		}
		if strings.HasSuffix(rel, ".md") && rel != s.Session {
			fi, err := e.Info()
			if err != nil {
				return err
			}
			s.Outside[rel], s.OutsideTime[rel] = b, fi.ModTime()
		}
		return nil
	})
	if err != nil {
		return s, "", err
	}
	regs, err := d.ss.Registrations()
	if err != nil {
		return s, "", err
	}
	realRoot, _ := filepath.EvalSymlinks(root)
	for _, abs := range regs {
		rel, ok := underRoot(abs, root, realRoot)
		if !ok {
			continue
		}
		s.Registered[rel] = true
		_, outside := s.Outside[rel]
		// Locks matter wherever the run writes: the session file, qa's
		// files, their destinations (a stopped run's), and outside files
		// that name qa.
		if rel == s.Session || strings.HasPrefix(rel, s.QA+"/") || strings.HasPrefix(rel, sidedoc.SideDir(s.Session)+"/") ||
			(outside && strings.Contains(string(s.Outside[rel]), s.QA)) {
			st, err := d.ss.Status(filepath.Join(root, filepath.FromSlash(rel)))
			if err != nil {
				return s, "", err
			}
			s.Locked[rel] = st.Holder == "operator"
		}
	}
	s.Warnings = gitignoreGap(root, s.QA, sidedoc.SideDir(s.Session))
	return s, root, nil
}

// gitignoreGap warns when git ignores qa's files but would track the side
// folder's (the Spec: "Gitignore follows wherever qa/ was ignored"). Outside
// a git work tree, or when git can't answer, it says nothing: there is
// nothing to follow.
func gitignoreGap(root, qa, side string) []string {
	ignored := func(rel string) (bool, bool) {
		err := exec.Command("git", "-C", root, "check-ignore", "-q", "--", rel).Run()
		var exit *exec.ExitError
		switch {
		case err == nil:
			return true, true
		case errors.As(err, &exit) && exit.ExitCode() == 1:
			return false, true
		}
		return false, false
	}
	q, ok := ignored(qa + "/probe.md")
	if !ok || !q {
		return nil
	}
	if sd, ok := ignored(side + "/probe.md"); ok && !sd {
		return []string{fmt.Sprintf("git ignores %s/ but would track %s/: add %s/ to .gitignore before --apply, or the answers become tracked files", qa, side, side)}
	}
	return nil
}

// underRoot is abs as a slash path relative to root, through either spelling
// of root (couch stores resolved paths).
func underRoot(abs, root, realRoot string) (string, bool) {
	for _, r := range []string{root, realRoot} {
		if r == "" {
			continue
		}
		if rel, ok := strings.CutPrefix(abs, r+string(filepath.Separator)); ok {
			return filepath.ToSlash(rel), true
		}
	}
	return "", false
}

// runMigration applies a plan, one file at a time, in an order where a rerun
// after a stop anywhere resumes (#455's state table): create the new file
// exclusively, give it the old mtime, move couch's registration and its
// baseline, remove the old file; then the other files, then the session
// file's one apply, then the emptied directories. step is called at each
// boundary; a test returns errInterrupted from it.
func runMigration(d deps, s migSnapshot, root string, p migrationPlan, step func() error) error {
	abs := func(rel string) string { return filepath.Join(root, filepath.FromSlash(rel)) }
	sweepTemps(root)
	for _, m := range p.Moves {
		if !m.Resume {
			if err := createExclusive(abs(m.New), m.Bytes); err != nil {
				return fmt.Errorf("%s: %w", m.New, err)
			}
			if err := step(); err != nil {
				return err
			}
		}
		if err := os.Chtimes(abs(m.New), m.Mtime, m.Mtime); err != nil {
			return err
		}
		if m.Registered {
			if err := d.ss.Rename(abs(m.Old), abs(m.New)); err != nil {
				return fmt.Errorf("%s moved, its couch registration not yet: %w; rerun to finish", m.New, err)
			}
			if err := step(); err != nil {
				return err
			}
			if err := d.ss.MarkAllSeen(abs(m.New)); err != nil {
				return err
			}
		}
		if err := os.Remove(abs(m.Old)); err != nil {
			return err
		}
		if err := step(); err != nil {
			return err
		}
	}
	for _, f := range sortedKeys(p.Outside) {
		if recs, ok := p.OutsideEdits[f]; ok {
			if err := applyEdits(d, abs(f), recs, fileHash(string(s.Outside[f]))); err != nil {
				return fmt.Errorf("%s: %w", f, err)
			}
		} else if err := replaceIfUnchanged(abs(f), s.Outside[f], p.Outside[f]); err != nil {
			return fmt.Errorf("%s: %w", f, err)
		}
		if err := os.Chtimes(abs(f), s.OutsideTime[f], s.OutsideTime[f]); err != nil {
			return err
		}
		if err := step(); err != nil {
			return err
		}
	}
	if len(p.Session) > 0 {
		if err := applyEdits(d, abs(s.Session), p.Session, fileHash(s.SessionText)); err != nil {
			return fmt.Errorf("%s: %w", s.Session, err)
		}
		if err := step(); err != nil {
			return err
		}
	}
	removeEmpty(abs(s.QA))
	removeEmpty(filepath.Dir(abs(s.QA)))
	return nil
}

// tempPrefix names the migration's own temporary files, so a run that died
// mid-write leaves files the next run knows are its own.
const tempPrefix = ".tl-lead-migrate-"

// sweepTemps removes the temporary files a killed run left under root.
func sweepTemps(root string) {
	_ = filepath.WalkDir(root, func(p string, e fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return nil
		case e.IsDir() && p != root && strings.HasPrefix(e.Name(), "."):
			return filepath.SkipDir
		case e.Type().IsRegular() && strings.HasPrefix(e.Name(), tempPrefix):
			_ = os.Remove(p)
		}
		return nil
	})
}

// applyEdits sends records through couch's apply, naming the hash the plan
// read.
func applyEdits(d deps, file string, recs []editRecord, hash string) error {
	raw, err := json.Marshal(recs)
	if err != nil {
		return err
	}
	err = writeBody(string(raw), func(path string) error { return d.ss.Apply(file, path, hash) })
	switch {
	case errors.Is(err, ErrStale):
		return errors.New("it changed since the plan read it; rerun to rebuild its edits")
	case errors.Is(err, ErrLocked):
		return errors.New("the operator holds its lock; rerun after their submit")
	}
	return err
}

// createExclusive writes b to a temporary file beside p and links it into
// place, so p appears whole or not at all and never replaces a file.
func createExclusive(p string, b []byte) error {
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(p), tempPrefix+"*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return err
	}
	return os.Link(tmp.Name(), p)
}

// replaceIfUnchanged swaps p's bytes from before to after atomically, and
// refuses when p no longer holds before.
func replaceIfUnchanged(p string, before, after []byte) error {
	cur, err := os.ReadFile(p)
	if err != nil {
		return err
	}
	if string(cur) != string(before) {
		return errors.New("it changed since the plan read it; rerun to rebuild its edits")
	}
	tmp, err := os.CreateTemp(filepath.Dir(p), tempPrefix+"*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(after); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if fi, err := os.Stat(p); err == nil {
		_ = os.Chmod(tmp.Name(), fi.Mode().Perm())
	}
	return os.Rename(tmp.Name(), p)
}

// removeEmpty removes dir's empty directories, deepest first, then dir if
// it ends empty. A directory with anything left in it stays.
func removeEmpty(dir string) {
	var dirs []string
	_ = filepath.WalkDir(dir, func(p string, e fs.DirEntry, err error) error {
		if err != nil || !e.IsDir() {
			return nil
		}
		if p != dir && strings.HasPrefix(e.Name(), ".") {
			return filepath.SkipDir // another tool's, even when empty
		}
		dirs = append(dirs, p)
		return nil
	})
	slices.Reverse(dirs)
	for _, p := range dirs {
		_ = os.Remove(p) // fails, and keeps it, unless empty
	}
}

// renderPlan prints the plan for the TL: every move, what stays, each
// conflict, the edits, the mentions and dangling links, and the next step.
func renderPlan(w io.Writer, file string, p migrationPlan, s migSnapshot, applied bool) {
	mode := "dry run"
	if applied {
		mode = "applied"
	}
	fmt.Fprintf(w, "migrate-answers: %s → %s (%s)\n", s.QA, p.Side, mode)
	resumes := 0
	for _, m := range p.Moves {
		verb := "move"
		if m.Resume {
			verb, resumes = "resume", resumes+1
		}
		var tags []string
		if m.Registered {
			tags = append(tags, "registered")
		}
		if m.Marker != "" {
			tags = append(tags, "marker: "+m.Marker)
		}
		tail := ""
		if len(tags) > 0 {
			tail = "  [" + strings.Join(tags, ", ") + "]"
		}
		fmt.Fprintf(w, "%-8s %s → %s%s\n", verb, m.Old, m.New, tail)
	}
	for _, l := range p.Left {
		fmt.Fprintf(w, "left     %s (not content; stays, so %s is kept)\n", l, s.QA)
	}
	for _, c := range p.Conflicts {
		fmt.Fprintf(w, "conflict %s: %s\n", c.Path, c.Why)
	}
	for _, wn := range p.Warnings {
		fmt.Fprintf(w, "warning  %s\n", wn)
	}
	fmt.Fprintf(w, "session  %s: %d link and index edit(s), one apply\n", s.Session, len(p.Session))
	for _, f := range sortedKeys(p.Outside) {
		how := "rewritten in place"
		if _, ok := p.OutsideEdits[f]; ok {
			how = "through couch apply"
		}
		fmt.Fprintf(w, "outside  %s (%s)\n", f, how)
	}
	for _, n := range p.Mentions {
		fmt.Fprintf(w, "mention  %s:%s\n", n.Path, n.Text)
	}
	for _, n := range p.Dangling {
		fmt.Fprintf(w, "dangling %s: %s (resolves to nothing; left as written)\n", n.Path, n.Text)
	}
	fmt.Fprintf(w, "summary: %d moves (%d resumed), %d left, %d conflicts, %d session edits, %d outside files, %d mentions, %d dangling\n",
		len(p.Moves), resumes, len(p.Left), len(p.Conflicts), len(p.Session), len(p.Outside), len(p.Mentions), len(p.Dangling))
	switch {
	case len(p.Conflicts) > 0:
		fmt.Fprintln(w, "refused: resolve the conflicts above, then rerun; nothing was written")
	case !applied:
		fmt.Fprintf(w, "next: run it while the operator is not editing these docs: tl-lead status migrate-answers %s --apply\n", file)
	}
}

func newMigrateCmd(d deps) *cobra.Command {
	var from string
	var apply bool
	c := &cobra.Command{
		Use:   "migrate-answers [FILE] [--from qa/<operator>] [--apply]",
		Short: "One-time: move qa/<operator>/ answer docs into FILE's side folder, with links, index and couch registrations (#455); a dry run unless --apply",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return migrateAnswers(d, cmd.OutOrStdout(), args[0], from, apply, func() error { return nil })
		},
	}
	c.Flags().StringVar(&from, "from", "", "the answers folder, relative to FILE's directory (default: qa/<frontmatter operator:>)")
	c.Flags().BoolVar(&apply, "apply", false, "carry out the plan (default: print it)")
	return withDefaultStatusFile(c, 1)
}

func migrateAnswers(d deps, out io.Writer, file, from string, apply bool, step func() error) error {
	s, root, err := readMigration(d, file, from)
	if err != nil {
		return err
	}
	p := planMigration(s)
	if len(p.Conflicts) > 0 || !apply {
		renderPlan(out, file, p, s, false)
		if len(p.Conflicts) > 0 {
			return exitError{1, fmt.Errorf("refused: %d conflict(s)", len(p.Conflicts))}
		}
		return nil
	}
	if err := runMigration(d, s, root, p, step); err != nil {
		return err
	}
	renderPlan(out, file, p, s, true)
	return nil
}
