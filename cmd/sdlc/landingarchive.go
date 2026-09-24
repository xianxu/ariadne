package main

import (
	"bytes"
	"fmt"
	"path"
	"reflect"
	"sort"
	"strings"
	"time"

	"go.yaml.in/yaml/v3"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/pkg/vocab"
)

const landingArchiveLimit = 10000

var landingArchiveRunner gitRunner = execGitRunner{}
var newLandingArchivePublisher = func(root, remote string) (trunkPublisher, error) { return gitx.NewTrunkFile(root, remote, "main") }

type landingArchiveRoots struct{ issues, plans, history string }
type landingArchiveBlob struct {
	mode, oid string
	content   []byte
}
type landingArchiveSnapshot map[string]landingArchiveBlob
type landingOwnedIssue struct {
	path, frontmatter, body, anchor string
	requiredPlans                   []string
}
type landingArchiveMove struct{ source, destination string }
type landingArchivePlan struct {
	write gitx.TrunkWrite
	moves []landingArchiveMove
}

func archiveRead(root string, args ...string) ([]byte, error) {
	if root == "" {
		return nil, fmt.Errorf("archive repository directory is required")
	}
	out, err := landingArchiveRunner.GitInDir(root, args...)
	if err != nil {
		return nil, fmt.Errorf("archive git %s: %w\n%s", args[0], err, out)
	}
	return out, nil
}
func resolveArchiveRoots(root, issuesDir, plansDir, historyDir string) (landingArchiveRoots, error) {
	dirs := []string{issuesDir, plansDir, historyDir}
	for i, dir := range dirs {
		rel, err := gitx.InsideRoot(root, dir)
		if err != nil {
			return landingArchiveRoots{}, err
		}
		if rel == "." || strings.Contains(rel, "\\") || strings.ContainsRune(rel, 0) {
			return landingArchiveRoots{}, fmt.Errorf("unsafe archive root %q", dir)
		}
		for _, part := range strings.Split(rel, "/") {
			if part == ".git" {
				return landingArchiveRoots{}, fmt.Errorf("unsafe archive root %q", dir)
			}
		}
		dirs[i] = rel
	}
	for i, a := range dirs {
		for j, b := range dirs {
			if i != j && (a == b || strings.HasPrefix(a, b+"/")) {
				return landingArchiveRoots{}, fmt.Errorf("archive roots overlap: %s and %s", a, b)
			}
		}
	}
	return landingArchiveRoots{dirs[0], dirs[1], dirs[2]}, nil
}
func validateArchivePR(repo string, pr landingPR) error {
	if !landingRepoValid(repo) || repo != pr.Repo || pr.Number <= 0 || pr.State != "MERGED" || pr.BaseRef != "main" || !landingOIDValid(pr.HeadOID) || !landingOIDValid(pr.BaseOID) || !landingOIDValid(pr.MergeOID) {
		return fmt.Errorf("archive requires exact merged PR identity and immutable commits")
	}
	return nil
}
func archiveTree(root, ref string, dirs landingArchiveRoots) (landingArchiveSnapshot, error) {
	if !landingOIDValid(ref) {
		return nil, fmt.Errorf("archive snapshot requires a full object ID")
	}
	args := []string{"--literal-pathspecs", "ls-tree", "-r", "-z", "--full-tree", ref, "--"}
	for _, dir := range []string{dirs.issues, dirs.plans, dirs.history} {
		if dir != "" {
			args = append(args, dir)
		}
	}
	out, err := archiveRead(root, args...)
	if err != nil {
		return nil, err
	}
	result := landingArchiveSnapshot{}
	records := bytes.Split(out, []byte{0})
	if len(records)-1 > landingArchiveLimit {
		return nil, fmt.Errorf("archive tree exceeds %d entries", landingArchiveLimit)
	}
	for _, record := range records {
		if len(record) == 0 {
			continue
		}
		parts := bytes.SplitN(record, []byte{'\t'}, 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("malformed archive tree entry")
		}
		fields := strings.Fields(string(parts[0]))
		name := string(parts[1])
		if len(fields) != 3 || !landingOIDValid(fields[2]) || path.Clean(name) != name || path.IsAbs(name) || strings.HasPrefix(name, "../") {
			return nil, fmt.Errorf("malformed archive tree entry %q", name)
		}
		if _, exists := result[name]; exists {
			return nil, fmt.Errorf("duplicate archive tree path %q", name)
		}
		result[name] = landingArchiveBlob{mode: fields[0], oid: fields[2]}
	}
	return result, nil
}
func archiveBlob(root string, b landingArchiveBlob) ([]byte, error) {
	if b.mode != "100644" {
		return nil, fmt.Errorf("archive requires ordinary non-executable blob, got mode %s", b.mode)
	}
	return archiveRead(root, "cat-file", "blob", b.oid)
}
func parseLandingIssue(name string, content []byte) (landingOwnedIssue, error) {
	fm, body, err := issue.Parse(string(content))
	if err != nil {
		return landingOwnedIssue{}, fmt.Errorf("parse %s: %w", name, err)
	}
	id, ok := issue.GetField(fm, "id")
	want := issueIDPrefix(path.Base(name))
	if !ok || want == "" || id != want {
		return landingOwnedIssue{}, fmt.Errorf("issue identity mismatch at %s", name)
	}
	return landingOwnedIssue{path: name, frontmatter: fm, body: body}, nil
}
func selectLandingIssues(root string, pr landingPR, issuesDir string) ([]landingOwnedIssue, error) {
	out, err := archiveRead(root, "rev-list", "--max-count=10001", pr.HeadOID, "--not", pr.BaseOID, "--")
	if err != nil {
		return nil, err
	}
	commits := strings.Fields(string(out))
	if len(commits) > landingArchiveLimit {
		return nil, fmt.Errorf("archive ownership exceeds %d commits", landingArchiveLimit)
	}
	own := map[string]bool{}
	for _, oid := range commits {
		if !landingOIDValid(oid) {
			return nil, fmt.Errorf("malformed archive ownership commit")
		}
		own[oid] = true
	}
	dirs := landingArchiveRoots{issues: issuesDir}
	tree, err := archiveTree(root, pr.HeadOID, dirs)
	if err != nil {
		return nil, err
	}
	var selected []landingOwnedIssue
	names := []string{}
	for name := range tree {
		if path.Dir(name) == dirs.issues && issueFilename(path.Base(name)) {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		content, err := archiveBlob(root, tree[name])
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", name, err)
		}
		ref, err := parseLandingIssue(name, content)
		if err != nil {
			return nil, err
		}
		status, _ := issue.GetField(ref.frontmatter, "status")
		if status != "codecomplete" {
			continue
		}
		anchor, err := codecompleteAnchorCommitAt(pr.HeadOID, name, func(args ...string) ([]byte, error) { return archiveRead(root, args...) })
		if err != nil {
			return nil, err
		}
		if anchor == "" {
			return nil, fmt.Errorf("codecomplete issue %s has no close anchor", name)
		}
		if own[anchor] {
			ref.anchor = anchor
			selected = append(selected, ref)
		}
	}
	return selected, nil
}

// Archive selection additionally pins the minimum artifact membership at the PR head.
func selectArchiveLandingIssues(root string, pr landingPR, dirs landingArchiveRoots) ([]landingOwnedIssue, error) {
	selected, err := selectLandingIssues(root, pr, dirs.issues)
	if err != nil || len(selected) == 0 {
		return selected, err
	}
	tree, err := archiveTree(root, pr.HeadOID, landingArchiveRoots{plans: dirs.plans})
	if err != nil {
		return nil, err
	}
	for i := range selected {
		for name, b := range tree {
			if path.Dir(name) == dirs.plans && planArtifactBelongsToIssue(path.Base(selected[i].path), path.Base(name)) {
				if b.mode != "100644" {
					return nil, fmt.Errorf("unsafe PR plan mode at %s", name)
				}
				selected[i].requiredPlans = append(selected[i].requiredPlans, name)
			}
		}
		sort.Strings(selected[i].requiredPlans)
	}
	return selected, nil
}

// Bodies and the editorial updated date may evolve; lifecycle metadata may not.
func sameLandingGeneration(original, current string) error {
	decode := func(fm string) (map[string]any, error) {
		var fields map[string]any
		if err := yaml.Unmarshal([]byte(fm), &fields); err != nil {
			return nil, err
		}
		delete(fields, "updated")
		delete(fields, "status")
		return fields, nil
	}
	a, err := decode(original)
	if err != nil {
		return err
	}
	b, err := decode(current)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(a, b) {
		return fmt.Errorf("stable frontmatter differs from reviewed PR")
	}
	return nil
}

func archiveRelated(name string, selected []landingOwnedIssue, dirs landingArchiveRoots) bool {
	for _, ref := range selected {
		base := path.Base(ref.path)
		parent := path.Dir(name)
		if (parent == dirs.issues || parent == dirs.history || parent == vocab.ArchiveSubdir(dirs.history, vocab.ArchiveIssues)) && issueIDPrefix(path.Base(name)) == issueIDPrefix(base) {
			return true
		}
		if (parent == dirs.plans || parent == vocab.ArchiveSubdir(dirs.history, vocab.ArchivePlans)) && planArtifactBelongsToIssue(base, path.Base(name)) {
			return true
		}
	}
	return false
}
func archiveSnapshot(root, ref string, selected []landingOwnedIssue, dirs landingArchiveRoots) (landingArchiveSnapshot, error) {
	tree, err := archiveTree(root, ref, dirs)
	if err != nil {
		return nil, err
	}
	for name, b := range tree {
		if archiveRelated(name, selected, dirs) {
			b.content, err = archiveBlob(root, b)
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", name, err)
			}
			tree[name] = b
		}
	}
	return tree, nil
}

// planLandingArchive is pure: derive moves from fresh bytes, preserving bodies
// while requiring the same issue identity and the codecomplete transition.
func planLandingArchive(selected []landingOwnedIssue, current landingArchiveSnapshot, dirs landingArchiveRoots, date string) (landingArchivePlan, error) {
	plan := landingArchivePlan{write: gitx.TrunkWrite{Write: map[string][]byte{}}}
	add := func(src, dst string, body []byte) error {
		if _, exists := current[dst]; exists {
			return fmt.Errorf("archive destination occupied: %s", dst)
		}
		if _, exists := plan.write.Write[dst]; exists {
			return fmt.Errorf("duplicate archive destination: %s", dst)
		}
		plan.write.Write[dst] = body
		plan.write.Delete = append(plan.write.Delete, src)
		plan.moves = append(plan.moves, landingArchiveMove{src, dst})
		return nil
	}
	for _, owned := range selected {
		b, present := current[owned.path]
		if !present {
			return plan, fmt.Errorf("active issue missing without confirmed archive: %s", owned.path)
		}
		if b.mode != "100644" {
			return plan, fmt.Errorf("unsafe issue mode at %s", owned.path)
		}
		ref, err := parseLandingIssue(owned.path, b.content)
		if err != nil {
			return plan, err
		}
		if err := sameLandingGeneration(owned.frontmatter, ref.frontmatter); err != nil {
			return plan, fmt.Errorf("issue generation changed at %s: %w", owned.path, err)
		}
		for _, name := range owned.requiredPlans {
			if _, present := current[name]; !present {
				return plan, fmt.Errorf("PR-owned plan artifact missing: %s", name)
			}
		}
		for name := range current {
			parent := path.Dir(name)
			if (parent == dirs.issues || parent == dirs.history || parent == vocab.ArchiveSubdir(dirs.history, vocab.ArchiveIssues)) && issueIDPrefix(path.Base(name)) == issueIDPrefix(path.Base(owned.path)) && name != owned.path {
				return plan, fmt.Errorf("conflicting issue identity/archive: %s", name)
			}
		}
		done, err := publishedIssueContent(ref.frontmatter, ref.body, date)
		if err != nil {
			return plan, fmt.Errorf("%s: %w", owned.path, err)
		}
		if err := add(owned.path, archiveDestination(dirs.history, vocab.ArchiveIssues, path.Base(owned.path)), done); err != nil {
			return plan, err
		}
		names := []string{}
		for name := range current {
			if path.Dir(name) == dirs.plans && planArtifactBelongsToIssue(path.Base(owned.path), path.Base(name)) {
				names = append(names, name)
			}
		}
		sort.Strings(names)
		for _, name := range names {
			b := current[name]
			if b.mode != "100644" {
				return plan, fmt.Errorf("unsafe plan mode at %s", name)
			}
			if err := add(name, archiveDestination(dirs.history, vocab.ArchivePlans, path.Base(name)), b.content); err != nil {
				return plan, err
			}
		}
	}
	return plan, nil
}
func archiveProvenance(repo string, pr landingPR) string {
	return fmt.Sprintf("Landing-PR: %s#%d\nLanding-Head: %s", repo, pr.Number, pr.HeadOID)
}

func archiveLandingPR(root, remote, repo string, pr landingPR, issuesDir, plansDir, historyDir string) error {
	if err := validateArchivePR(repo, pr); err != nil {
		return err
	}
	dirs, err := resolveArchiveRoots(root, issuesDir, plansDir, historyDir)
	if err != nil {
		return err
	}
	selected, err := selectArchiveLandingIssues(root, pr, dirs)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		return nil
	}
	pub, err := newLandingArchivePublisher(root, remote)
	if err != nil {
		return err
	}
	msg := archiveCommitMessage + "\n\n" + archiveProvenance(repo, pr)
	return pub.UpdateMany(msg, func(v *gitx.TrunkView) (gitx.TrunkWrite, error) {
		complete, err := confirmLandingArchive(root, v.Ref(), repo, pr, dirs, selected)
		if err != nil {
			return gitx.TrunkWrite{}, err
		}
		if complete {
			return gitx.TrunkWrite{}, nil
		}
		snapshot, err := archiveSnapshot(root, v.Ref(), selected, dirs)
		if err != nil {
			return gitx.TrunkWrite{}, err
		}
		plan, err := planLandingArchive(selected, snapshot, dirs, time.Now().Format("2006-01-02"))
		return plan.write, err
	})
}

// landingArchiveComplete is read-only. In particular, missing-local-ref recovery
// must not create an archive and then use that new effect to justify completion.
func landingArchiveComplete(root, remoteMainOID, repo string, pr landingPR, issuesDir, plansDir, historyDir string) (bool, error) {
	if err := validateArchivePR(repo, pr); err != nil {
		return false, err
	}
	dirs, err := resolveArchiveRoots(root, issuesDir, plansDir, historyDir)
	if err != nil {
		return false, err
	}
	selected, err := selectArchiveLandingIssues(root, pr, dirs)
	if err != nil {
		return false, err
	}
	if len(selected) == 0 {
		return true, nil
	}
	return confirmLandingArchive(root, remoteMainOID, repo, pr, dirs, selected)
}
func confirmLandingArchive(root, tip, repo string, pr landingPR, dirs landingArchiveRoots, selected []landingOwnedIssue) (bool, error) {
	if !landingOIDValid(tip) {
		return false, fmt.Errorf("archive proof requires a full main object ID")
	}
	if _, err := archiveRead(root, "merge-base", "--is-ancestor", pr.MergeOID, tip); err != nil {
		return false, fmt.Errorf("PR integration is not confirmed on fresh main: %w", err)
	}
	out, err := archiveRead(root, "log", "--max-count=10001", "--format=%H%x00%(trailers:key=Landing-PR,valueonly)%x00%(trailers:key=Landing-Head,valueonly)%x00", tip, "--not", pr.BaseOID, "--")
	if err != nil {
		return false, err
	}
	fields := strings.Split(string(out), "\x00")
	if (len(fields)-1)%3 != 0 {
		return false, fmt.Errorf("malformed archive provenance response")
	}
	if (len(fields)-1)/3 > landingArchiveLimit {
		return false, fmt.Errorf("archive provenance exceeds %d commits", landingArchiveLimit)
	}
	var matching []string
	for i := 0; i+2 < len(fields); i += 3 {
		oid := strings.TrimSpace(fields[i])
		if !landingOIDValid(oid) {
			return false, fmt.Errorf("malformed archive provenance commit")
		}
		if strings.TrimSpace(fields[i+1]) == fmt.Sprintf("%s#%d", repo, pr.Number) && strings.TrimSpace(fields[i+2]) == pr.HeadOID {
			matching = append(matching, oid)
		}
	}
	current, err := archiveSnapshot(root, tip, selected, dirs)
	if err != nil {
		return false, err
	}
	if len(matching) == 0 {
		// Validate absence as a genuine not-yet-archived state, not partial success.
		_, err := planLandingArchive(selected, current, dirs, time.Now().Format("2006-01-02"))
		return false, err
	}
	if len(matching) != 1 {
		return false, fmt.Errorf("ambiguous archive provenance for PR #%d", pr.Number)
	}
	archiveOID := matching[0]
	parents, err := archiveRead(root, "rev-list", "--parents", "-n", "1", archiveOID, "--")
	if err != nil {
		return false, err
	}
	p := strings.Fields(string(parents))
	if len(p) != 2 || !landingOIDValid(p[1]) {
		return false, fmt.Errorf("archive provenance is not one ordinary commit")
	}
	before, err := archiveSnapshot(root, p[1], selected, dirs)
	if err != nil {
		return false, err
	}
	after, err := archiveSnapshot(root, archiveOID, selected, dirs)
	if err != nil {
		return false, err
	}
	firstDest := archiveDestination(dirs.history, vocab.ArchiveIssues, path.Base(selected[0].path))
	b, present := after[firstDest]
	if !present {
		return false, fmt.Errorf("archive provenance missing %s", firstDest)
	}
	fm, _, err := issue.Parse(string(b.content))
	if err != nil {
		return false, err
	}
	date, _ := issue.GetField(fm, "updated")
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return false, fmt.Errorf("archive generation has invalid updated date")
	}
	plan, err := planLandingArchive(selected, before, dirs, date)
	if err != nil {
		return false, err
	}
	changed, err := archiveRead(root, "diff-tree", "--no-commit-id", "--name-only", "--no-renames", "-r", "-z", p[1], archiveOID, "--")
	if err != nil {
		return false, err
	}
	expected := map[string]bool{}
	for _, move := range plan.moves {
		expected[move.source] = true
		expected[move.destination] = true
	}
	paths := bytes.Split(changed, []byte{0})
	actual := map[string]bool{}
	for _, name := range paths {
		if len(name) > 0 {
			actual[string(name)] = true
		}
	}
	if len(actual) != len(expected) {
		return false, fmt.Errorf("archive provenance has incomplete or extra changes")
	}
	for name := range expected {
		if !actual[name] {
			return false, fmt.Errorf("archive provenance missing move %s", name)
		}
	}
	for _, move := range plan.moves {
		if _, exists := after[move.source]; exists {
			return false, fmt.Errorf("archive commit retained active source %s", move.source)
		}
		archived, exists := after[move.destination]
		if !exists || !bytes.Equal(archived.content, plan.write.Write[move.destination]) {
			return false, fmt.Errorf("archive generation differs at %s", move.destination)
		}
		currentBlob, exists := current[move.destination]
		if !exists || currentBlob.oid != archived.oid || currentBlob.mode != archived.mode {
			return false, fmt.Errorf("archived generation changed at %s", move.destination)
		}
	}
	for name := range current {
		if !archiveRelated(name, selected, dirs) {
			continue
		}
		if !expected[name] || path.Dir(name) == dirs.issues || path.Dir(name) == dirs.plans {
			return false, fmt.Errorf("active or conflicting archive artifact %s", name)
		}
	}
	return true, nil
}
