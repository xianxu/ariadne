package fleet

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

// FakeGit persists the subset of Git state consumed by fleet collectors. Repo
// state is keyed by canonical Git common-dir identity, so calls from primary,
// nested, and linked-worktree directories observe the same mutations.
type FakeGit struct {
	repos map[string]*FakeGitRepo
}

var _ GitReader = (*FakeGit)(nil)

type FakeGitRepo struct {
	CommonDir   string
	PrimaryRoot string
	BareHEAD    string
	Worktrees   []gitx.Worktree
	Refs        map[string]string
	Commits     map[string]FakeGitCommit
	Dirty       map[string][]FakeGitStatusEntry
}

type FakeGitCommit struct {
	Parents     []string
	CommittedAt time.Time
}

type FakeGitStatusEntry struct {
	Code       string
	Path       string
	SourcePath string
}

func NewFakeGit() *FakeGit {
	return &FakeGit{repos: make(map[string]*FakeGitRepo)}
}

func (f *FakeGit) AddRepo(repo *FakeGitRepo) error {
	if repo == nil {
		return fmt.Errorf("fake git: nil repo")
	}
	owned, err := fakeGitOwnedRepo(repo)
	if err != nil {
		return err
	}
	if _, exists := f.repos[owned.CommonDir]; exists {
		return fmt.Errorf("fake git: duplicate common dir %q", owned.CommonDir)
	}
	f.repos[owned.CommonDir] = owned
	return nil
}

func fakeGitOwnedRepo(input *FakeGitRepo) (*FakeGitRepo, error) {
	commonDir, err := canonicalPath(input.CommonDir)
	if err != nil {
		return nil, fmt.Errorf("fake git: canonical common dir: %w", err)
	}
	primaryRoot, err := canonicalPath(input.PrimaryRoot)
	if err != nil {
		return nil, fmt.Errorf("fake git: canonical primary root: %w", err)
	}
	if len(input.Worktrees) == 0 {
		return nil, fmt.Errorf("fake git: repo %q has no worktrees", commonDir)
	}

	repo := &FakeGitRepo{
		CommonDir: commonDir, PrimaryRoot: primaryRoot, BareHEAD: input.BareHEAD,
		Worktrees: make([]gitx.Worktree, len(input.Worktrees)),
		Refs:      make(map[string]string, len(input.Refs)),
		Commits:   make(map[string]FakeGitCommit, len(input.Commits)),
		Dirty:     make(map[string][]FakeGitStatusEntry, len(input.Dirty)),
	}
	worktreePaths := make(map[string]bool, len(input.Worktrees))
	for i, inputWorktree := range input.Worktrees {
		worktree := cloneFakeGitWorktree(inputWorktree)
		path, err := fakeGitCanonicalListedPath(worktree.Path)
		if err != nil {
			return nil, fmt.Errorf("fake git: canonical worktree %q: %w", worktree.Path, err)
		}
		if worktreePaths[path] {
			return nil, fmt.Errorf("fake git: duplicate worktree %q", path)
		}
		worktree.Path = path
		repo.Worktrees[i] = worktree
		worktreePaths[path] = true
	}
	if repo.Worktrees[0].Path != primaryRoot {
		return nil, fmt.Errorf("fake git: first worktree %q is not primary %q", repo.Worktrees[0].Path, primaryRoot)
	}

	for sha, inputCommit := range input.Commits {
		if sha == "" || strings.ContainsRune(sha, 0) || inputCommit.CommittedAt.IsZero() {
			return nil, fmt.Errorf("fake git: invalid commit %q", sha)
		}
		commit := inputCommit
		commit.Parents = append([]string(nil), inputCommit.Parents...)
		repo.Commits[sha] = commit
	}
	for sha, commit := range repo.Commits {
		for _, parent := range commit.Parents {
			if strings.ContainsRune(parent, 0) {
				return nil, fmt.Errorf("fake git: commit %q has NUL in parent", sha)
			}
			if _, ok := repo.Commits[parent]; !ok {
				return nil, fmt.Errorf("fake git: commit %q has absent parent %q", sha, parent)
			}
		}
	}
	for ref, sha := range input.Refs {
		if !fakeGitValidRefName(ref) {
			return nil, fmt.Errorf("fake git: invalid ref %q", ref)
		}
		if _, ok := repo.Commits[sha]; !ok {
			return nil, fmt.Errorf("fake git: ref %q targets absent commit %q", ref, sha)
		}
		repo.Refs[ref] = sha
	}
	if err := validateFakeGitWorktrees(repo); err != nil {
		return nil, err
	}

	for path, inputEntries := range input.Dirty {
		canonical, err := canonicalPath(path)
		if err != nil {
			return nil, fmt.Errorf("fake git: canonical dirty worktree %q: %w", path, err)
		}
		if !worktreePaths[canonical] {
			return nil, fmt.Errorf("fake git: dirty state targets unconfigured worktree %q", path)
		}
		entries := append([]FakeGitStatusEntry(nil), inputEntries...)
		if _, err := fakeStatusPorcelain(entries); err != nil {
			return nil, err
		}
		repo.Dirty[canonical] = entries
	}
	return repo, nil
}

// AddWorktree mutates registered repo state while preserving the canonical
// identity established by AddRepo. Git continues to list missing linked
// worktrees only when their administrative state marks them locked or
// prunable.
func (f *FakeGit) AddWorktree(commonDir string, worktree gitx.Worktree) error {
	repo, err := f.repoByCommonDir(commonDir)
	if err != nil {
		return err
	}
	worktree = cloneFakeGitWorktree(worktree)
	path, err := fakeGitCanonicalListedPath(worktree.Path)
	if err != nil {
		return fmt.Errorf("fake git: canonical worktree %q: %w", worktree.Path, err)
	}
	for _, existing := range repo.Worktrees {
		if existing.Path == path {
			return fmt.Errorf("fake git: duplicate worktree %q", path)
		}
	}
	worktree.Path = path
	worktrees := append([]gitx.Worktree(nil), repo.Worktrees...)
	worktrees = append(worktrees, worktree)
	probe := *repo
	probe.Worktrees = worktrees
	if err := validateFakeGitWorktrees(&probe); err != nil {
		return err
	}
	repo.Worktrees = worktrees
	return nil
}

func (f *FakeGit) AddCommit(commonDir, sha string, commit FakeGitCommit) error {
	repo, err := f.repoByCommonDir(commonDir)
	if err != nil {
		return err
	}
	if sha == "" || commit.CommittedAt.IsZero() {
		return fmt.Errorf("fake git: invalid commit %q", sha)
	}
	if _, exists := repo.Commits[sha]; exists {
		return fmt.Errorf("fake git: duplicate commit %q", sha)
	}
	for _, parent := range commit.Parents {
		if _, ok := repo.Commits[parent]; !ok {
			return fmt.Errorf("fake git: commit %q has absent parent %q", sha, parent)
		}
	}
	commit.Parents = append([]string(nil), commit.Parents...)
	repo.Commits[sha] = commit
	return nil
}

func (f *FakeGit) SetRef(commonDir, ref, sha string) error {
	repo, err := f.repoByCommonDir(commonDir)
	if err != nil {
		return err
	}
	if !fakeGitValidRefName(ref) {
		return fmt.Errorf("fake git: invalid ref %q", ref)
	}
	if _, ok := repo.Commits[sha]; !ok {
		return fmt.Errorf("fake git: ref %q targets absent commit %q", ref, sha)
	}
	refs := make(map[string]string, len(repo.Refs)+1)
	for name, target := range repo.Refs {
		refs[name] = target
	}
	refs[ref] = sha
	worktrees := append([]gitx.Worktree(nil), repo.Worktrees...)
	const headsPrefix = "refs/heads/"
	if strings.HasPrefix(ref, headsPrefix) {
		branch := strings.TrimPrefix(ref, headsPrefix)
		for i := range worktrees {
			if worktrees[i].Branch == branch {
				worktrees[i].HEAD = sha
			}
		}
	}
	probe := *repo
	probe.Refs, probe.Worktrees = refs, worktrees
	if err := validateFakeGitWorktrees(&probe); err != nil {
		return err
	}
	repo.Refs, repo.Worktrees = refs, worktrees
	return nil
}

// DeleteRef mutates registered refs while rejecting a deletion that would
// leave any attached worktree without its branch tip.
func (f *FakeGit) DeleteRef(commonDir, ref string) error {
	repo, err := f.repoByCommonDir(commonDir)
	if err != nil {
		return err
	}
	if !fakeGitValidRefName(ref) {
		return fmt.Errorf("fake git: invalid ref %q", ref)
	}
	if _, exists := repo.Refs[ref]; !exists {
		return fmt.Errorf("fake git: ref %q is not configured", ref)
	}
	refs := make(map[string]string, len(repo.Refs)-1)
	for name, sha := range repo.Refs {
		if name != ref {
			refs[name] = sha
		}
	}
	probe := *repo
	probe.Refs = refs
	if err := validateFakeGitWorktrees(&probe); err != nil {
		return err
	}
	repo.Refs = refs
	return nil
}

func (f *FakeGit) DetachWorktree(commonDir, path, sha string) error {
	repo, worktree, err := f.worktreeByPath(commonDir, path)
	if err != nil {
		return err
	}
	if _, ok := repo.Commits[sha]; !ok {
		return fmt.Errorf("fake git: detached HEAD targets absent commit %q", sha)
	}
	original := *worktree
	worktree.HEAD, worktree.Branch = sha, ""
	worktree.Detached, worktree.Bare = true, false
	if err := validateFakeGitWorktrees(repo); err != nil {
		*worktree = original
		return err
	}
	return nil
}

func (f *FakeGit) SetWorktreeAttributes(commonDir, path string, locked, prunable *string) error {
	repo, worktree, err := f.worktreeByPath(commonDir, path)
	if err != nil {
		return err
	}
	original := *worktree
	worktree.Locked = cloneStringPointer(locked)
	worktree.Prunable = cloneStringPointer(prunable)
	if err := validateFakeGitWorktrees(repo); err != nil {
		*worktree = original
		return err
	}
	return nil
}

// SetDirty records status entries for the containing configured worktree. The
// caller may use any real directory vantage, including a symlink or subdir.
func (f *FakeGit) SetDirty(dir string, entries []FakeGitStatusEntry) error {
	repo, worktree, _, err := f.repoForDir(dir)
	if err != nil {
		return err
	}
	entries = append([]FakeGitStatusEntry(nil), entries...)
	if _, err := fakeStatusPorcelain(entries); err != nil {
		return err
	}
	repo.Dirty[worktree.Path] = entries
	return nil
}

func (f *FakeGit) repoByCommonDir(commonDir string) (*FakeGitRepo, error) {
	identity, err := canonicalPath(commonDir)
	if err != nil {
		return nil, fmt.Errorf("fake git: canonical common dir %q: %w", commonDir, err)
	}
	repo, ok := f.repos[identity]
	if !ok {
		return nil, fmt.Errorf("fake git: common dir %q is not configured", commonDir)
	}
	return repo, nil
}

func (f *FakeGit) worktreeByPath(commonDir, path string) (*FakeGitRepo, *gitx.Worktree, error) {
	repo, err := f.repoByCommonDir(commonDir)
	if err != nil {
		return nil, nil, err
	}
	identity, err := fakeGitCanonicalListedPath(path)
	if err != nil {
		return nil, nil, fmt.Errorf("fake git: canonical worktree %q: %w", path, err)
	}
	for i := range repo.Worktrees {
		if repo.Worktrees[i].Path == identity {
			return repo, &repo.Worktrees[i], nil
		}
	}
	return nil, nil, fmt.Errorf("fake git: worktree %q is not configured", path)
}

func cloneStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func cloneFakeGitWorktree(worktree gitx.Worktree) gitx.Worktree {
	worktree.Locked = cloneStringPointer(worktree.Locked)
	worktree.Prunable = cloneStringPointer(worktree.Prunable)
	return worktree
}

func validateFakeGitWorktrees(repo *FakeGitRepo) error {
	if len(repo.Worktrees) == 0 {
		return fmt.Errorf("fake git: repo has no worktrees")
	}
	primaryBare := repo.Worktrees[0].Bare
	if primaryBare {
		if repo.BareHEAD != "" {
			if fakeGitAllZeroObjectID(repo.BareHEAD) || strings.ContainsRune(repo.BareHEAD, 0) {
				return fmt.Errorf("fake git: invalid bare HEAD %q", repo.BareHEAD)
			}
			if _, ok := repo.Commits[repo.BareHEAD]; !ok {
				return fmt.Errorf("fake git: bare HEAD commit %q is absent from graph", repo.BareHEAD)
			}
		}
	} else if repo.BareHEAD != "" {
		return fmt.Errorf("fake git: non-bare primary has a bare HEAD")
	}

	for i, worktree := range repo.Worktrees {
		if err := validateFakeGitWorktree(repo, i, worktree); err != nil {
			return err
		}
	}
	porcelain, err := fakeWorktreePorcelain(repo)
	if err != nil {
		return err
	}
	parsed, err := gitx.ParseWorktrees(porcelain)
	if err != nil {
		return fmt.Errorf("fake git: generated worktree porcelain does not round-trip: %w", err)
	}
	if len(parsed) != len(repo.Worktrees) {
		return fmt.Errorf("fake git: generated %d worktrees, parsed %d", len(repo.Worktrees), len(parsed))
	}
	return nil
}

func validateFakeGitWorktree(repo *FakeGitRepo, index int, worktree gitx.Worktree) error {
	prefix := fmt.Sprintf("fake git: worktree %q", worktree.Path)
	for name, value := range map[string]string{
		"path": worktree.Path, "HEAD": worktree.HEAD, "branch": worktree.Branch,
	} {
		if strings.ContainsRune(value, 0) {
			return fmt.Errorf("%s has NUL in %s", prefix, name)
		}
	}
	for name, value := range map[string]*string{"locked": worktree.Locked, "prunable": worktree.Prunable} {
		if value != nil && strings.ContainsRune(*value, 0) {
			return fmt.Errorf("%s has NUL in %s reason", prefix, name)
		}
	}
	if _, err := os.Stat(worktree.Path); err != nil {
		missingAdministrativeWorktree := index > 0 && os.IsNotExist(err) && (worktree.Locked != nil || worktree.Prunable != nil)
		if !missingAdministrativeWorktree {
			return fmt.Errorf("%s path is unavailable: %w", prefix, err)
		}
	}

	if worktree.Bare {
		if index != 0 {
			return fmt.Errorf("%s is a non-primary bare worktree", prefix)
		}
		if worktree.HEAD != "" || worktree.Branch != "" || worktree.Detached || worktree.Locked != nil || worktree.Prunable != nil {
			return fmt.Errorf("%s combines bare with checkout state or attributes", prefix)
		}
		return nil
	}
	if (worktree.Branch != "") == worktree.Detached {
		return fmt.Errorf("%s must have exactly one of branch or detached", prefix)
	}
	if worktree.Branch != "" && !fakeGitValidBranch(worktree.Branch) {
		return fmt.Errorf("%s has invalid branch %q", prefix, worktree.Branch)
	}
	if worktree.HEAD == "" {
		return fmt.Errorf("%s has no displayed HEAD", prefix)
	}
	if fakeGitAllZeroObjectID(worktree.HEAD) {
		if worktree.Detached {
			return fmt.Errorf("%s has a detached zero HEAD", prefix)
		}
		if _, resolved := repo.Refs["refs/heads/"+worktree.Branch]; resolved {
			return fmt.Errorf("%s has a zero HEAD for a resolved branch", prefix)
		}
		return nil
	}
	if _, ok := repo.Commits[worktree.HEAD]; !ok {
		return fmt.Errorf("%s HEAD %q is absent from graph", prefix, worktree.HEAD)
	}
	if worktree.Branch != "" {
		resolved, ok := repo.Refs["refs/heads/"+worktree.Branch]
		if !ok {
			return fmt.Errorf("%s branch %q is unresolved with a nonzero HEAD", prefix, worktree.Branch)
		}
		if resolved != worktree.HEAD {
			return fmt.Errorf("%s displayed HEAD %q disagrees with branch target %q", prefix, worktree.HEAD, resolved)
		}
	}
	return nil
}

func fakeGitAllZeroObjectID(value string) bool {
	if len(value) != 40 && len(value) != 64 {
		return false
	}
	for _, r := range value {
		if r != '0' {
			return false
		}
	}
	return true
}

func fakeGitValidBranch(branch string) bool {
	if branch == "" || strings.HasPrefix(branch, "-") {
		return false
	}
	return fakeGitValidRefName("refs/heads/" + branch)
}

func fakeGitValidRefName(ref string) bool {
	if ref == "" || ref == "@" || strings.HasPrefix(ref, "/") || strings.HasSuffix(ref, "/") ||
		strings.HasSuffix(ref, ".") || strings.Contains(ref, "//") || strings.Contains(ref, "..") || strings.Contains(ref, "@{") {
		return false
	}
	for _, r := range ref {
		if r <= ' ' || r == 0x7f || strings.ContainsRune("~^:?*[\\", r) {
			return false
		}
	}
	for _, component := range strings.Split(ref, "/") {
		if component == "" || strings.HasPrefix(component, ".") || strings.HasSuffix(component, ".lock") {
			return false
		}
	}
	return true
}

func (f *FakeGit) GitInDir(dir string, args ...string) ([]byte, error) {
	repo, worktree, commandDir, err := f.repoForDir(dir)
	if err != nil {
		return nil, err
	}
	command := strings.Join(args, " ")
	switch {
	case slices.Equal(args, []string{"worktree", "list", "--porcelain", "-z"}):
		return fakeWorktreePorcelain(repo)
	case slices.Equal(args, []string{"rev-parse", "--show-toplevel"}):
		if worktree.Bare {
			return nil, fmt.Errorf("fake git: this operation must be run in a work tree")
		}
		return []byte(worktree.Path + "\n"), nil
	case slices.Equal(args, []string{"rev-parse", "--git-common-dir"}):
		rel, err := filepath.Rel(commandDir, repo.CommonDir)
		if err != nil {
			return nil, fmt.Errorf("fake git: git-common-dir from %q: %w", dir, err)
		}
		return []byte(rel + "\n"), nil
	case slices.Equal(args, []string{"rev-parse", "HEAD"}):
		sha, err := fakeResolveWorktreeHEAD(repo, *worktree)
		if err != nil {
			return nil, err
		}
		return []byte(sha + "\n"), nil
	case slices.Equal(args, []string{"show", "-s", "--format=%cI", "HEAD"}):
		sha, err := fakeResolveWorktreeHEAD(repo, *worktree)
		if err != nil {
			return nil, err
		}
		commit, ok := repo.Commits[sha]
		if !ok {
			return nil, fmt.Errorf("fake git: HEAD commit %q is absent", sha)
		}
		if commit.CommittedAt.IsZero() {
			return nil, fmt.Errorf("fake git: HEAD commit %q has no timestamp", sha)
		}
		return []byte(commit.CommittedAt.Format("2006-01-02T15:04:05-07:00") + "\n"), nil
	case slices.Equal(args, []string{"status", "--porcelain=v1", "-z", "--untracked-files=all"}):
		if worktree.Bare {
			return nil, fmt.Errorf("fake git: this operation must be run in a work tree")
		}
		return fakeStatusPorcelain(repo.Dirty[worktree.Path])
	}

	if (len(args) == 3 || len(args) == 4) && args[0] == "rev-parse" && args[1] == "--verify" && (len(args) == 3 || args[2] == "--quiet") {
		revision := args[len(args)-1]
		sha, ok := fakeResolveRevision(repo, *worktree, revision)
		if !ok {
			return nil, fakeGitExitError{err: fmt.Errorf("fake git: revision %q not found", revision), code: 1}
		}
		return []byte(sha + "\n"), nil
	}
	if len(args) == 4 && args[0] == "rev-list" && args[1] == "--left-right" && args[2] == "--count" {
		return fakeRevList(repo, *worktree, args[3])
	}
	return nil, fmt.Errorf("fake git: unsupported command in %q: git %s", dir, command)
}

type fakeGitExitError struct {
	err  error
	code int
}

func (e fakeGitExitError) Error() string { return e.err.Error() }
func (e fakeGitExitError) Unwrap() error { return e.err }
func (e fakeGitExitError) ExitCode() int { return e.code }

func (f *FakeGit) repoForDir(dir string) (*FakeGitRepo, *gitx.Worktree, string, error) {
	commandDir, err := canonicalPath(dir)
	if err != nil {
		return nil, nil, "", fmt.Errorf("fake git: canonical command directory %q: %w", dir, err)
	}
	var matchedRepo *FakeGitRepo
	var matchedWorktree *gitx.Worktree
	bestLength := -1
	for _, repo := range f.repos {
		for i := range repo.Worktrees {
			if fakeGitPathContains(repo.Worktrees[i].Path, commandDir) && len(repo.Worktrees[i].Path) > bestLength {
				matchedRepo = repo
				matchedWorktree = &repo.Worktrees[i]
				bestLength = len(repo.Worktrees[i].Path)
			}
		}
	}
	if matchedRepo == nil {
		return nil, nil, "", fmt.Errorf("fake git: %q is not in a configured worktree", dir)
	}
	return matchedRepo, matchedWorktree, commandDir, nil
}

func fakeWorktreePorcelain(repo *FakeGitRepo) ([]byte, error) {
	var out bytes.Buffer
	for _, worktree := range repo.Worktrees {
		out.WriteString("worktree ")
		out.WriteString(worktree.Path)
		out.WriteByte(0)
		if worktree.Bare {
			out.WriteString("bare")
			out.WriteByte(0)
		} else {
			head, err := fakeDisplayedWorktreeHEAD(repo, worktree)
			if err != nil {
				return nil, err
			}
			out.WriteString("HEAD ")
			out.WriteString(head)
			out.WriteByte(0)
			switch {
			case worktree.Branch != "":
				out.WriteString("branch refs/heads/")
				out.WriteString(worktree.Branch)
			case worktree.Detached:
				out.WriteString("detached")
			default:
				return nil, fmt.Errorf("fake git: worktree %q has no checkout state", worktree.Path)
			}
			out.WriteByte(0)
		}
		fakeGitWriteOptional(&out, "locked", worktree.Locked)
		fakeGitWriteOptional(&out, "prunable", worktree.Prunable)
		out.WriteByte(0)
	}
	return out.Bytes(), nil
}

func fakeGitWriteOptional(out *bytes.Buffer, name string, value *string) {
	if value == nil {
		return
	}
	out.WriteString(name)
	if *value != "" {
		out.WriteByte(' ')
		out.WriteString(*value)
	}
	out.WriteByte(0)
}

func fakeDisplayedWorktreeHEAD(repo *FakeGitRepo, worktree gitx.Worktree) (string, error) {
	if worktree.Branch != "" {
		if sha, ok := repo.Refs["refs/heads/"+worktree.Branch]; ok {
			return sha, nil
		}
	}
	if worktree.HEAD == "" {
		return "", fmt.Errorf("fake git: worktree %q has no HEAD", worktree.Path)
	}
	return worktree.HEAD, nil
}

func fakeResolveWorktreeHEAD(repo *FakeGitRepo, worktree gitx.Worktree) (string, error) {
	if worktree.Bare {
		if repo.BareHEAD == "" {
			return "", fmt.Errorf("fake git: bare repo %q has no HEAD", worktree.Path)
		}
		return repo.BareHEAD, nil
	}
	if worktree.Branch != "" {
		sha, ok := repo.Refs["refs/heads/"+worktree.Branch]
		if !ok {
			return "", fmt.Errorf("fake git: worktree %q has an unborn HEAD", worktree.Path)
		}
		return sha, nil
	}
	if worktree.HEAD == "" {
		return "", fmt.Errorf("fake git: worktree %q has no HEAD", worktree.Path)
	}
	return worktree.HEAD, nil
}

func fakeResolveRevision(repo *FakeGitRepo, worktree gitx.Worktree, revision string) (string, bool) {
	if revision == "HEAD" {
		sha, err := fakeResolveWorktreeHEAD(repo, worktree)
		return sha, err == nil
	}
	for _, candidate := range []string{revision, "refs/heads/" + revision, "refs/remotes/" + revision} {
		if sha, ok := repo.Refs[candidate]; ok {
			return sha, true
		}
	}
	_, ok := repo.Commits[revision]
	return revision, ok
}

func fakeStatusPorcelain(entries []FakeGitStatusEntry) ([]byte, error) {
	entries = append([]FakeGitStatusEntry(nil), entries...)
	sort.Slice(entries, func(i, j int) bool {
		iUntracked, jUntracked := entries[i].Code == "??", entries[j].Code == "??"
		if iUntracked != jUntracked {
			return !iUntracked
		}
		return entries[i].Path < entries[j].Path
	})
	var out bytes.Buffer
	for _, entry := range entries {
		if !validStatusCode(entry.Code) || entry.Path == "" || strings.ContainsRune(entry.Path, 0) {
			return nil, fmt.Errorf("fake git: invalid dirty entry %#v", entry)
		}
		isRename := strings.ContainsAny(entry.Code, "RC")
		if isRename != (entry.SourcePath != "") || strings.ContainsRune(entry.SourcePath, 0) {
			return nil, fmt.Errorf("fake git: invalid rename/copy entry %#v", entry)
		}
		out.WriteString(entry.Code)
		out.WriteByte(' ')
		out.WriteString(entry.Path)
		out.WriteByte(0)
		if isRename {
			out.WriteString(entry.SourcePath)
			out.WriteByte(0)
		}
	}
	return out.Bytes(), nil
}

func TestFakeStatusPorcelainValidatesExactXYStates(t *testing.T) {
	valid := []string{
		" A", " M", " T", " D",
		"M ", "MM", "MT", "MD",
		"T ", "TM", "TT", "TD",
		"A ", "AM", "AT", "AD",
		"D ",
		"R ", "RM", "RT", "RD",
		"C ", "CM", "CT", "CD",
		" R", " C",
		"DD", "AU", "UD", "UA", "DU", "AA", "UU",
		"??",
	}
	for _, code := range valid {
		t.Run("valid "+fmt.Sprintf("%q", code), func(t *testing.T) {
			entry := FakeGitStatusEntry{Code: code, Path: "target"}
			if strings.ContainsAny(code, "RC") {
				entry.SourcePath = "source"
			}
			if _, err := fakeStatusPorcelain([]FakeGitStatusEntry{entry}); err != nil {
				t.Fatalf("fakeStatusPorcelain rejected documented state %q: %v", code, err)
			}
		})
	}

	invalid := []FakeGitStatusEntry{
		{Code: "ZZ", Path: "path"},
		{Code: "R?", Path: "target", SourcePath: "source"},
		{Code: "!!", Path: "path"},
		{Code: "  ", Path: "path"},
		{Code: "? ", Path: "path"},
		{Code: " U", Path: "path"},
		{Code: "R ", Path: "target"},
		{Code: "??", Path: "path", SourcePath: "source"},
		{Code: "\x00?", Path: "path"},
		{Code: "??", Path: "bad\x00path"},
		{Code: "R ", Path: "target", SourcePath: "bad\x00source"},
	}
	for _, entry := range invalid {
		t.Run("invalid "+fmt.Sprintf("%q", entry.Code), func(t *testing.T) {
			if _, err := fakeStatusPorcelain([]FakeGitStatusEntry{entry}); err == nil {
				t.Fatalf("fakeStatusPorcelain accepted impossible state %#v", entry)
			}
		})
	}
}

func fakeRevList(repo *FakeGitRepo, worktree gitx.Worktree, rangeArg string) ([]byte, error) {
	parts := strings.Split(rangeArg, "...")
	if len(parts) != 2 || parts[0] == "" || parts[1] != "HEAD" {
		return nil, fmt.Errorf("fake git: unsupported rev-list range %q", rangeArg)
	}
	left, ok := fakeResolveRevision(repo, worktree, parts[0])
	if !ok {
		return nil, fmt.Errorf("fake git: revision %q not found", parts[0])
	}
	right, ok := fakeResolveRevision(repo, worktree, parts[1])
	if !ok {
		return nil, fmt.Errorf("fake git: revision %q not found", parts[1])
	}
	leftReachable, err := fakeReachable(repo, left)
	if err != nil {
		return nil, err
	}
	rightReachable, err := fakeReachable(repo, right)
	if err != nil {
		return nil, err
	}
	behind, ahead := 0, 0
	for sha := range leftReachable {
		if !rightReachable[sha] {
			behind++
		}
	}
	for sha := range rightReachable {
		if !leftReachable[sha] {
			ahead++
		}
	}
	return []byte(fmt.Sprintf("%d\t%d\n", behind, ahead)), nil
}

func fakeReachable(repo *FakeGitRepo, start string) (map[string]bool, error) {
	result := make(map[string]bool)
	stack := []string{start}
	for len(stack) > 0 {
		sha := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if result[sha] {
			continue
		}
		commit, ok := repo.Commits[sha]
		if !ok {
			return nil, fmt.Errorf("fake git: commit %q is absent from graph", sha)
		}
		result[sha] = true
		stack = append(stack, commit.Parents...)
	}
	return result, nil
}

func fakeGitCanonicalListedPath(path string) (string, error) {
	canonical, err := canonicalPath(path)
	if err == nil {
		return canonical, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	ancestor := filepath.Clean(abs)
	var missing []string
	for {
		resolved, resolveErr := filepath.EvalSymlinks(ancestor)
		if resolveErr == nil {
			for i := len(missing) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, missing[i])
			}
			return filepath.Clean(resolved), nil
		}
		if !errors.Is(resolveErr, os.ErrNotExist) {
			return "", resolveErr
		}
		parent := filepath.Dir(ancestor)
		if parent == ancestor {
			return "", resolveErr
		}
		missing = append(missing, filepath.Base(ancestor))
		ancestor = parent
	}
}

func fakeGitPathContains(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func TestFakeGitRejectsUnsupportedCommands(t *testing.T) {
	driver := newFakeGitContract(t)
	for _, args := range [][]string{
		{"branch", "--show-current"},
		{"rev-parse HEAD"},
		{"rev-list", "--left-right", "--count", "main...main"},
	} {
		if out, err := driver.reader.GitInDir(driver.primary, args...); err == nil || !strings.Contains(err.Error(), "unsupported") {
			t.Errorf("unsupported git %s = (%q, %v), want loud error", strings.Join(args, " "), out, err)
		}
	}
}

func TestFakeGitModelsBareAndUnbornHEAD(t *testing.T) {
	t.Run("bare", func(t *testing.T) {
		root := t.TempDir()
		bare := filepath.Join(root, "bare.git")
		if err := os.Mkdir(bare, 0o755); err != nil {
			t.Fatal(err)
		}
		const head = "1111111111111111111111111111111111111111"
		fake := NewFakeGit()
		if err := fake.AddRepo(&FakeGitRepo{
			CommonDir: bare, PrimaryRoot: bare, BareHEAD: head,
			Worktrees: []gitx.Worktree{{Path: bare, Bare: true}},
			Commits:   map[string]FakeGitCommit{head: {CommittedAt: mustContractTime(t, contractInitialTime)}},
		}); err != nil {
			t.Fatal(err)
		}

		for _, args := range [][]string{
			{"rev-parse", "--show-toplevel"},
			{"status", "--porcelain=v1", "-z", "--untracked-files=all"},
		} {
			if out, err := fake.GitInDir(bare, args...); err == nil {
				t.Errorf("bare git %s = %q, want error", strings.Join(args, " "), out)
			}
		}
		out, err := fake.GitInDir(bare, "worktree", "list", "--porcelain", "-z")
		if err != nil {
			t.Fatal(err)
		}
		got, err := gitx.ParseWorktrees(out)
		if err != nil {
			t.Fatal(err)
		}
		if want := []gitx.Worktree{{Path: canonicalContractPath(t, bare), Bare: true}}; !reflect.DeepEqual(got, want) {
			t.Fatalf("bare worktree list = %+v, want %+v", got, want)
		}
		if out, err := fake.GitInDir(bare, "rev-parse", "HEAD"); err != nil || strings.TrimSpace(string(out)) != head {
			t.Fatalf("bare rev-parse HEAD = (%q, %v), want %s", out, err, head)
		}
	})

	t.Run("unborn", func(t *testing.T) {
		root := t.TempDir()
		primary := filepath.Join(root, "repo")
		common := filepath.Join(primary, ".git")
		if err := os.MkdirAll(common, 0o755); err != nil {
			t.Fatal(err)
		}
		const zeroHEAD = "0000000000000000000000000000000000000000"
		fake := NewFakeGit()
		if err := fake.AddRepo(&FakeGitRepo{
			CommonDir: common, PrimaryRoot: primary,
			Worktrees: []gitx.Worktree{{Path: primary, HEAD: zeroHEAD, Branch: "main"}},
		}); err != nil {
			t.Fatal(err)
		}

		out, err := fake.GitInDir(primary, "worktree", "list", "--porcelain", "-z")
		if err != nil {
			t.Fatal(err)
		}
		got, err := gitx.ParseWorktrees(out)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 || got[0].HEAD != zeroHEAD || got[0].Branch != "main" {
			t.Fatalf("unborn worktree list = %+v, want zero displayed HEAD on main", got)
		}
		for _, args := range [][]string{{"rev-parse", "HEAD"}, {"show", "-s", "--format=%cI", "HEAD"}} {
			if out, err := fake.GitInDir(primary, args...); err == nil {
				t.Errorf("unborn git %s = %q, want error", strings.Join(args, " "), out)
			}
		}
	})
}

func TestFakeGitMutationAPIKeepsCanonicalIdentityAndMissingWorktrees(t *testing.T) {
	root := t.TempDir()
	primary := filepath.Join(root, "repo")
	common := filepath.Join(primary, ".git")
	linked := filepath.Join(root, "linked")
	nested := filepath.Join(linked, "nested")
	for _, dir := range []string{common, nested} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	linkedAlias := filepath.Join(root, "linked-alias")
	if err := os.Symlink(linked, linkedAlias); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	const head = "1111111111111111111111111111111111111111"
	fake := NewFakeGit()
	repo := &FakeGitRepo{
		CommonDir: common, PrimaryRoot: primary,
		Worktrees: []gitx.Worktree{{Path: primary, HEAD: head, Branch: "main"}},
		Refs:      map[string]string{"refs/heads/main": head},
		Commits:   map[string]FakeGitCommit{head: {CommittedAt: mustContractTime(t, contractInitialTime)}},
	}
	if err := fake.AddRepo(repo); err != nil {
		t.Fatal(err)
	}
	if err := fake.AddWorktree(common, gitx.Worktree{Path: linkedAlias, HEAD: head, Detached: true}); err != nil {
		t.Fatal(err)
	}
	if err := fake.SetDirty(linkedAlias, []FakeGitStatusEntry{{Code: "??", Path: "odd\n name.txt"}}); err != nil {
		t.Fatal(err)
	}
	if out, err := fake.GitInDir(nested, "rev-parse", "--show-toplevel"); err != nil || strings.TrimSpace(string(out)) != canonicalContractPath(t, linked) {
		t.Fatalf("nested linked top-level = (%q, %v), want canonical linked root", out, err)
	}
	if out, err := fake.GitInDir(nested, "status", "--porcelain=v1", "-z", "--untracked-files=all"); err != nil || !bytes.Equal(out, []byte("?? odd\n name.txt\x00")) {
		t.Fatalf("linked status = (%q, %v), want canonical-keyed dirty entry", out, err)
	}

	missingParent := filepath.Join(root, "missing-parent")
	if err := os.Mkdir(missingParent, 0o755); err != nil {
		t.Fatal(err)
	}
	missingParentAlias := filepath.Join(root, "missing-parent-alias")
	if err := os.Symlink(missingParent, missingParentAlias); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	missing := filepath.Join(missingParentAlias, "prunable")
	wantMissing := filepath.Join(canonicalContractPath(t, missingParent), "prunable")
	reason := "gitdir file points to non-existent location"
	if err := fake.SetRef(common, "refs/heads/stale", head); err != nil {
		t.Fatal(err)
	}
	if err := fake.AddWorktree(common, gitx.Worktree{Path: missing, HEAD: head, Branch: "stale", Prunable: &reason}); err != nil {
		t.Fatalf("add missing prunable worktree: %v", err)
	}
	out, err := fake.GitInDir(primary, "worktree", "list", "--porcelain", "-z")
	if err != nil {
		t.Fatal(err)
	}
	got, err := gitx.ParseWorktrees(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[2].Path != wantMissing || got[2].Prunable == nil || *got[2].Prunable != reason {
		t.Fatalf("worktrees after missing mutation = %+v", got)
	}
}

func TestFakeGitRegistrationRejectsMissingOrdinaryWorktrees(t *testing.T) {
	root := t.TempDir()
	primary := filepath.Join(root, "repo")
	common := filepath.Join(primary, ".git")
	if err := os.MkdirAll(common, 0o755); err != nil {
		t.Fatal(err)
	}
	const head = "1111111111111111111111111111111111111111"
	newRepo := func(primaryRoot string) *FakeGitRepo {
		return &FakeGitRepo{
			CommonDir: common, PrimaryRoot: primaryRoot,
			Worktrees: []gitx.Worktree{{Path: primaryRoot, HEAD: head, Branch: "main"}},
			Refs:      map[string]string{"refs/heads/main": head},
			Commits:   map[string]FakeGitCommit{head: {CommittedAt: mustContractTime(t, contractInitialTime)}},
		}
	}

	if err := NewFakeGit().AddRepo(newRepo(filepath.Join(root, "missing-primary"))); err == nil {
		t.Fatal("AddRepo accepted a nonexistent primary worktree")
	}

	fake := NewFakeGit()
	if err := fake.AddRepo(newRepo(primary)); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(root, "missing-linked")
	if err := fake.AddWorktree(common, gitx.Worktree{Path: missing, HEAD: head, Detached: true}); err == nil {
		t.Fatal("AddWorktree accepted a nonexistent ordinary linked worktree")
	}
	lockedReason := "administrative lock"
	if err := fake.AddWorktree(common, gitx.Worktree{Path: missing, HEAD: head, Detached: true, Locked: &lockedReason}); err != nil {
		t.Fatalf("AddWorktree rejected a missing locked worktree: %v", err)
	}
}

func TestFakeGitRegistrationOwnsCallerState(t *testing.T) {
	root := t.TempDir()
	primary := filepath.Join(root, "repo")
	common := filepath.Join(primary, ".git")
	linked := filepath.Join(root, "linked")
	for _, dir := range []string{common, linked} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	const parentSHA = "1111111111111111111111111111111111111111"
	const headSHA = "2222222222222222222222222222222222222222"
	parents := []string{parentSHA}
	locked, prunable := "registered lock", "registered prune"
	worktrees := []gitx.Worktree{{
		Path: primary, HEAD: headSHA, Branch: "main", Locked: &locked, Prunable: &prunable,
	}}
	refs := map[string]string{
		"refs/heads/main":          headSHA,
		"refs/remotes/origin/main": parentSHA,
	}
	commits := map[string]FakeGitCommit{
		parentSHA: {CommittedAt: mustContractTime(t, contractInitialTime)},
		headSHA:   {Parents: parents, CommittedAt: mustContractTime(t, contractMainTime)},
	}
	dirtyEntries := []FakeGitStatusEntry{{Code: "??", Path: "registered.txt"}}
	dirty := map[string][]FakeGitStatusEntry{primary: dirtyEntries}
	repo := &FakeGitRepo{
		CommonDir: common, PrimaryRoot: primary, Worktrees: worktrees,
		Refs: refs, Commits: commits, Dirty: dirty,
	}
	fake := NewFakeGit()
	if err := fake.AddRepo(repo); err != nil {
		t.Fatal(err)
	}

	linkedLocked, linkedPrunable := "linked lock", "linked prune"
	linkedInput := gitx.Worktree{
		Path: linked, HEAD: headSHA, Detached: true,
		Locked: &linkedLocked, Prunable: &linkedPrunable,
	}
	if err := fake.AddWorktree(common, linkedInput); err != nil {
		t.Fatal(err)
	}

	// Mutate every caller-owned reference after registration. FakeGit must keep
	// the validated snapshot it accepted, not observe any of these aliases.
	repo.CommonDir = primary
	worktrees[0].HEAD = "mutated"
	worktrees[0].Branch = "mutated"
	*worktrees[0].Locked = "mutated"
	*worktrees[0].Prunable = "mutated"
	refs["refs/heads/main"] = parentSHA
	delete(commits, headSHA)
	parents[0] = "missing"
	dirtyEntries[0].Path = "mutated.txt"
	*linkedInput.Locked = "mutated"
	*linkedInput.Prunable = "mutated"

	if out, err := fake.GitInDir(primary, "rev-parse", "HEAD"); err != nil || strings.TrimSpace(string(out)) != headSHA {
		t.Fatalf("registered HEAD after caller mutation = (%q, %v), want %s", out, err, headSHA)
	}
	if out, err := fake.GitInDir(primary, "rev-list", "--left-right", "--count", "origin/main...HEAD"); err != nil || string(out) != "0\t1\n" {
		t.Fatalf("registered graph after caller mutation = (%q, %v), want 0\\t1", out, err)
	}
	if out, err := fake.GitInDir(primary, "status", "--porcelain=v1", "-z", "--untracked-files=all"); err != nil || !bytes.Equal(out, []byte("?? registered.txt\x00")) {
		t.Fatalf("registered dirty state after caller mutation = (%q, %v)", out, err)
	}
	out, err := fake.GitInDir(primary, "worktree", "list", "--porcelain", "-z")
	if err != nil {
		t.Fatal(err)
	}
	got, err := gitx.ParseWorktrees(out)
	if err != nil {
		t.Fatal(err)
	}
	want := []gitx.Worktree{
		{Path: canonicalContractPath(t, primary), HEAD: headSHA, Branch: "main", Locked: stringPointer("registered lock"), Prunable: stringPointer("registered prune")},
		{Path: canonicalContractPath(t, linked), HEAD: headSHA, Detached: true, Locked: stringPointer("linked lock"), Prunable: stringPointer("linked prune")},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("registered worktrees after caller mutation = %+v, want %+v", got, want)
	}
	commonOut, err := fake.GitInDir(primary, "rev-parse", "--git-common-dir")
	if err != nil {
		t.Fatal(err)
	}
	resolvedCommon := filepath.Join(primary, strings.TrimSpace(string(commonOut)))
	if canonicalContractPath(t, resolvedCommon) != canonicalContractPath(t, common) {
		t.Fatalf("registered common dir after caller mutation = %q, want %q", resolvedCommon, common)
	}
}

func stringPointer(value string) *string { return &value }

func TestFakeGitRegistrationRejectsInvalidState(t *testing.T) {
	root := t.TempDir()
	primary := filepath.Join(root, "repo")
	common := filepath.Join(primary, ".git")
	other := filepath.Join(root, "other")
	for _, dir := range []string{common, other} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	const head = "1111111111111111111111111111111111111111"
	validCommit := FakeGitCommit{CommittedAt: mustContractTime(t, contractInitialTime)}
	validRepo := func() *FakeGitRepo {
		return &FakeGitRepo{
			CommonDir: common, PrimaryRoot: primary,
			Worktrees: []gitx.Worktree{{Path: primary, HEAD: head, Branch: "main"}},
			Refs:      map[string]string{"refs/heads/main": head},
			Commits:   map[string]FakeGitCommit{head: validCommit},
		}
	}

	tests := []struct {
		name   string
		mutate func(*FakeGitRepo)
	}{
		{"dangling ref", func(repo *FakeGitRepo) { repo.Refs["refs/heads/main"] = "missing" }},
		{"missing parent", func(repo *FakeGitRepo) {
			repo.Commits[head] = FakeGitCommit{Parents: []string{"missing"}, CommittedAt: validCommit.CommittedAt}
		}},
		{"malformed worktree", func(repo *FakeGitRepo) { repo.Worktrees[0].Branch = "" }},
		{"invalid dirty entry", func(repo *FakeGitRepo) {
			repo.Dirty = map[string][]FakeGitStatusEntry{primary: {{Code: "?", Path: "bad"}}}
		}},
		{"dirty key outside worktree", func(repo *FakeGitRepo) {
			repo.Dirty = map[string][]FakeGitStatusEntry{other: {{Code: "??", Path: "bad"}}}
		}},
		{"bare with head", func(repo *FakeGitRepo) {
			repo.Worktrees[0].Bare, repo.Worktrees[0].Branch = true, ""
		}},
		{"bare with branch", func(repo *FakeGitRepo) { repo.Worktrees[0].Bare = true }},
		{"bare with detached", func(repo *FakeGitRepo) {
			repo.Worktrees[0].Bare, repo.Worktrees[0].Detached = true, true
			repo.Worktrees[0].Branch, repo.Worktrees[0].HEAD = "", ""
		}},
		{"bare with attributes", func(repo *FakeGitRepo) {
			repo.Worktrees[0] = gitx.Worktree{Path: primary, Bare: true, Locked: stringPointer("lock")}
		}},
		{"branch and detached", func(repo *FakeGitRepo) { repo.Worktrees[0].Detached = true }},
		{"invalid branch ref grammar", func(repo *FakeGitRepo) {
			delete(repo.Refs, "refs/heads/main")
			repo.Refs["refs/heads/bad..name"] = head
			repo.Worktrees[0].Branch = "bad..name"
		}},
		{"branch with NUL", func(repo *FakeGitRepo) {
			delete(repo.Refs, "refs/heads/main")
			repo.Refs["refs/heads/bad\x00name"] = head
			repo.Worktrees[0].Branch = "bad\x00name"
		}},
		{"HEAD with NUL", func(repo *FakeGitRepo) {
			repo.Worktrees[0] = gitx.Worktree{Path: primary, HEAD: head + "\x00", Detached: true}
		}},
		{"resolved HEAD outside graph", func(repo *FakeGitRepo) {
			repo.Worktrees[0] = gitx.Worktree{Path: primary, HEAD: "missing", Detached: true}
		}},
		{"optional attribute with NUL", func(repo *FakeGitRepo) {
			repo.Worktrees[0].Locked = stringPointer("bad\x00lock")
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := validRepo()
			tt.mutate(repo)
			if err := NewFakeGit().AddRepo(repo); err == nil {
				t.Fatal("AddRepo accepted invalid state")
			}
		})
	}

	t.Run("mutation APIs", func(t *testing.T) {
		fake := NewFakeGit()
		if err := fake.AddRepo(validRepo()); err != nil {
			t.Fatal(err)
		}
		if err := fake.SetDirty(primary, []FakeGitStatusEntry{{Code: "?", Path: "bad"}}); err == nil {
			t.Fatal("SetDirty accepted invalid state")
		}
		for _, worktree := range []gitx.Worktree{
			{Path: other, HEAD: head},
			{Path: other, HEAD: head, Branch: "main", Detached: true},
			{Path: other, HEAD: "missing", Detached: true},
			{Path: other, HEAD: head, Detached: true, Prunable: stringPointer("bad\x00reason")},
		} {
			if err := fake.AddWorktree(common, worktree); err == nil {
				t.Fatalf("AddWorktree accepted invalid state %+v", worktree)
			}
		}
	})
}
