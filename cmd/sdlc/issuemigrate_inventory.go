package main

import (
	"errors"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

func migrationInventory(env *migrateEnv, mainView *gitx.TrunkView) (tracker.MigrationInput, error) {
	d := vocab.Issue().Discovery()
	roots := issue.IDDirs(d.Home, d.Archive)
	in := tracker.MigrationInput{Repository: env.target.Repository, ObjectFormat: env.format, Main: mainView.Ref(), Anchors: map[string][]tracker.MigrationAnchor{}}
	observation, err := observeIssueInventory(env, mainView.Ref())
	if err != nil {
		return in, err
	}
	in.Published, err = issueSnapshotFiles(env, mainView.Ref(), roots)
	if err != nil {
		return in, err
	}
	// Batch bodies by root, selected before cat-file so unrelated assets never
	// enter the body read. Archives have their own role and never enter Branches.
	for _, root := range roots {
		exists, err := mainView.Exists(root)
		if err != nil {
			return in, err
		}
		if !exists {
			continue
		}
		files, err := mainView.FilesSelected(root, func(f gitx.TreeFile) bool {
			_, _, ok := issue.IssueInventoryPath(f.Path, roots)
			return path.Dir(f.Path) == root && ok
		})
		if err != nil {
			return in, err
		}
		for _, f := range files {
			_, role, _ := issue.IssueInventoryPath(f.Path, roots)
			file := tracker.MigrationFile{Path: f.Path, Raw: f.Content}
			if role == issue.InventoryActive {
				in.Active = append(in.Active, file)
			} else {
				in.Archived = append(in.Archived, file)
			}
		}
	}
	groups := map[string][]string{}
	for _, ref := range observation.Tips {
		groups[ref.OID] = append(groups[ref.OID], ref.Name)
	}
	oids := make([]string, 0, len(groups))
	for oid := range groups {
		oids = append(oids, oid)
	}
	sort.Strings(oids)
	var anchorRefs []migrationBranchRef
	for _, oid := range oids {
		if oid == mainView.Ref() {
			continue
		}
		carriers := groups[oid]
		sort.Strings(carriers)
		files, err := issueSnapshotFiles(env, oid, roots)
		if err != nil {
			return in, fmt.Errorf("%s: %w", strings.Join(carriers, ", "), err)
		}
		related, err := env.gitTest("merge-base", oid, mainView.Ref())
		if err != nil {
			return in, err
		}
		ancestry := tracker.MigrationUnrelated
		if related {
			ancestry = tracker.MigrationRelated
		}
		in.Tips = append(in.Tips, tracker.MigrationTipEvidence{Tip: oid, Carriers: carriers, Ancestry: ancestry, Files: files})
		ref := migrationBranchRef{Name: inventoryCarrierLabel(carriers), OID: oid}
		if related {
			edits, err := branchIssueEdits(env, ref, mainView.Ref(), d.Home)
			if err != nil {
				return in, err
			}
			in.Branches = append(in.Branches, edits...)
		}
		anchorRefs = append(anchorRefs, ref)
	}
	if in.DirtyIssuePaths, err = dirtyIssuePaths(env, roots, observation.Worktrees); err != nil {
		return in, err
	}
	for _, f := range in.Active {
		fm, _, perr := issue.Parse(string(f.Raw))
		if status, _ := issue.GetField(fm, "status"); perr != nil || status != "codecomplete" {
			continue
		}
		id, _, _ := issue.ParseFilename(path.Base(f.Path))
		anchors, err := legacyCloseAnchors(env, mainView.Ref(), anchorRefs, f.Path)
		if err != nil {
			return in, err
		}
		in.Anchors[id] = anchors
	}
	// Let the pure planner choose duplicate generations once; gather proof only
	// for the source each resulting card actually imports.
	selected := tracker.PlanTrackerMigration(in)
	files := make(map[string]tracker.MigrationFile, len(in.Active)+len(in.Archived))
	for _, group := range [][]tracker.MigrationFile{in.Active, in.Archived} {
		for _, f := range group {
			files[f.Path] = f
		}
	}
	in.Legacy = map[string]tracker.MigrationLegacyClose{}
	for _, c := range selected.Cards {
		card, err := issue.ParseCard(c.Raw)
		if err != nil {
			return in, err
		}
		status, err := issue.FieldString(card.Frontmatter, "status")
		if err != nil {
			return in, err
		}
		actual, err := issue.FieldString(card.Frontmatter, "actual_hours")
		if err != nil {
			return in, err
		}
		if _, err := strconv.ParseFloat(actual, 64); status != "done" || err != nil {
			continue
		}
		proof, err := resolveLegacyCompletion(env.ctx, env.root, mainView.Ref(), env.target.Repository, files[c.Source], plansDir())
		if err != nil {
			return in, err
		}
		in.Legacy[c.Source] = tracker.MigrationLegacyClose{Completion: proof.Completion, HasCompletion: proof.HasCompletion, Adoption: proof.Adoption}
	}
	return in, nil
}

func inventoryCarrierLabel(carriers []string) string {
	labels := make([]string, 0, len(carriers))
	seen := map[string]bool{}
	for _, c := range carriers {
		c = strings.TrimPrefix(strings.TrimPrefix(c, "refs/heads/"), "refs/remotes/")
		if !seen[c] {
			seen[c] = true
			labels = append(labels, c)
		}
	}
	return strings.Join(labels, ", ")
}

// branchIssueEdits preserves actual merge-base changes. Comparing against main
// would fabricate edits in inherited active records on old branches.
func branchIssueEdits(env *migrateEnv, ref migrationBranchRef, main, home string) ([]tracker.MigrationBranchFile, error) {
	related, err := env.gitTest("merge-base", ref.OID, main)
	if err != nil || !related {
		return nil, err
	}
	base, err := env.git("merge-base", ref.OID, main)
	if err != nil {
		return nil, err
	}
	names, err := inventoryGit(env, "--literal-pathspecs", "diff", "--name-only", "--no-renames", "-z", base, ref.OID, "--", home)
	if err != nil {
		return nil, err
	}
	if len(names) > 0 && names[len(names)-1] != 0 {
		return nil, errors.New("truncated issue delta inventory")
	}
	changed := map[string]bool{}
	for _, p := range strings.Split(string(names), "\x00") {
		if _, _, ok := issue.IssueInventoryPath(p, []string{home}); ok {
			changed[p] = true
		}
	}
	if len(changed) == 0 {
		return nil, nil
	}
	// One listing plus one cat-file batch, independent of changed file count.
	present := map[string][]byte{}
	files, err := env.main.ViewOf(ref.OID).FilesSelected(home, func(f gitx.TreeFile) bool { return changed[f.Path] })
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		present[f.Path] = f.Content
	}
	paths := make([]string, 0, len(changed))
	for p := range changed {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	var result []tracker.MigrationBranchFile
	for _, p := range paths {
		result = append(result, tracker.MigrationBranchFile{Branch: inventoryCarrierLabel([]string{ref.Name}), Path: p, Raw: present[p]})
	}
	return result, nil
}

// dirtyIssuePaths consumes the already pinned worktree list; relisting would
// create a second observation scope inside the same migration inventory.
func dirtyIssuePaths(env *migrateEnv, roots []string, worktrees []gitx.Worktree) ([]string, error) {
	dirty := map[string]bool{}
	for _, wt := range worktrees {
		if wt.Bare || wt.Prunable != nil {
			continue
		}
		add := func(p string) {
			if _, _, ok := issue.IssueInventoryPath(p, roots); ok {
				dirty[wt.Path+": "+p] = true
			}
		}
		args := append([]string{"-C", wt.Path, "--literal-pathspecs", "ls-files", "--stage", "-v", "-z", "--"}, roots...)
		raw, err := inventoryGit(env, args...)
		if err != nil {
			return nil, err
		}
		entries, err := gitx.ParseIndexStageZ(raw)
		if err != nil {
			return nil, fmt.Errorf("worktree %s index: %w", wt.Path, err)
		}
		indexed := map[string]bool{}
		for _, e := range entries {
			if _, _, ok := issue.IssueInventoryPath(e.Path, roots); !ok {
				continue
			}
			indexed[e.Path] = true
			// Lower-case tags are assume-unchanged; S is skip-worktree. Neither
			// permits status to prove a clean file. Retain it and name the remedy.
			if e.Stage != 0 || e.Tag != strings.ToUpper(e.Tag) || strings.ToUpper(e.Tag) == "S" || (e.Mode != "100644" && e.Mode != "100755") {
				dirty[wt.Path+": "+e.Path+" (resolve index conflicts or clear assume-unchanged/skip-worktree flags)"] = true
			}
		}
		args = append([]string{"-C", wt.Path, "--literal-pathspecs", "status", "--porcelain=v1", "-z", "--untracked-files=all", "--"}, roots...)
		raw, err = inventoryGit(env, args...)
		if err != nil {
			return nil, err
		}
		status, err := gitx.ParseStatusZ(raw)
		if err != nil {
			return nil, err
		}
		for _, e := range status {
			add(e.Path)
			add(e.Orig)
		}
		paths, err := issueFilesystemFiles(env, wt.Path, roots)
		if err != nil {
			return nil, fmt.Errorf("worktree %s: %w", wt.Path, err)
		}
		for _, p := range paths {
			if !indexed[p] {
				add(p)
			}
		}
	}
	paths := make([]string, 0, len(dirty))
	for p := range dirty {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	return paths, nil
}
