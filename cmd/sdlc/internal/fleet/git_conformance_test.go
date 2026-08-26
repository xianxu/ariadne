package fleet

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

const (
	contractInitialTime = "2026-01-02T03:04:05+00:00"
	contractMainTime    = "2026-01-03T03:04:05+00:00"
	contractMain2Time   = "2026-01-03T04:04:05+00:00"
	contractFeatureTime = "2026-01-04T03:04:05+00:00"
	contractMergeTime   = "2026-01-05T03:04:05+00:00"
)

type gitContractDriver struct {
	reader    GitReader
	primary   string
	linked    string
	prunable  string
	commonDir string
	roles     map[string]string
	diverge   func(*testing.T)
	merge     func(*testing.T)
	dirty     func(*testing.T)
}

type gitContractObservation struct {
	stage            string
	topLevel         string
	commonDir        string
	headRole         string
	committedAt      string
	mainRole         string
	originMainRole   string
	originMainExists bool
	worktrees        []contractWorktree
	status           []string
	behind           int
	ahead            int
}

type contractWorktree struct {
	pathRole string
	headRole string
	branch   string
	detached bool
	bare     bool
	locked   string
	prunable string
}

type gitModeObservation struct {
	topLevelExists bool
	commonDirRole  string
	headRole       string
	headExists     bool
	showExists     bool
	statusExists   bool
	worktrees      []contractWorktree
}

func TestFakeGitMutableContract(t *testing.T) {
	driver := newFakeGitContract(t)
	got := runGitContract(t, driver)
	assertGitContract(t, got)
}

func TestGitConformance(t *testing.T) {
	fake := runGitContract(t, newFakeGitContract(t))
	assertGitContract(t, fake)
	real := runGitContract(t, newRealGitContract(t))
	if !reflect.DeepEqual(fake, real) {
		t.Fatalf("fake/real Git observations differ\nfake: %#v\nreal: %#v", fake, real)
	}
}

func TestGitConformanceStatusBytes(t *testing.T) {
	fake := newFakeGitContract(t)
	real := newRealGitContract(t)
	for _, driver := range []*gitContractDriver{&fake, &real} {
		driver.diverge(t)
		driver.merge(t)
		driver.dirty(t)
	}
	fakeStatus, err := fake.reader.GitInDir(fake.linked, "status", "--porcelain=v1", "-z", "--untracked-files=all")
	if err != nil {
		t.Fatal(err)
	}
	realStatus, err := real.reader.GitInDir(real.linked, "status", "--porcelain=v1", "-z", "--untracked-files=all")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(fakeStatus, realStatus) {
		t.Fatalf("fake/live status bytes differ\nfake: %q\nreal: %q", fakeStatus, realStatus)
	}
	if !bytes.Contains(realStatus, []byte(" T typechange.txt\x00")) {
		t.Fatalf("live Git did not report the unstaged type change: %q", realStatus)
	}
}

func TestGitConformanceIgnoresAmbientRenameConfig(t *testing.T) {
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "status.renames")
	t.Setenv("GIT_CONFIG_VALUE_0", "false")

	fake := runGitContract(t, newFakeGitContract(t))
	real := runGitContract(t, newRealGitContract(t))
	if !reflect.DeepEqual(fake, real) {
		t.Fatalf("ambient Git config changed conformance observations\nfake: %#v\nreal: %#v", fake, real)
	}
}

func TestGitConformanceBareAndUnborn(t *testing.T) {
	t.Run("bare", func(t *testing.T) {
		fakeReader, fakeRoot, fakeRoles := newFakeBareContract(t)
		realReader, realRoot, realRoles := newRealBareContract(t)
		fake := observeGitMode(t, fakeReader, fakeRoot, fakeRoles)
		real := observeGitMode(t, realReader, realRoot, realRoles)
		if !reflect.DeepEqual(fake, real) {
			t.Fatalf("bare fake/real observations differ\nfake: %#v\nreal: %#v", fake, real)
		}
	})

	t.Run("unborn", func(t *testing.T) {
		fakeReader, fakeRoot := newFakeUnbornContract(t)
		realReader, realRoot := newRealUnbornContract(t)
		fake := observeGitMode(t, fakeReader, fakeRoot, nil)
		real := observeGitMode(t, realReader, realRoot, nil)
		if !reflect.DeepEqual(fake, real) {
			t.Fatalf("unborn fake/real observations differ\nfake: %#v\nreal: %#v", fake, real)
		}
	})
}

func observeGitMode(t *testing.T, reader GitReader, root string, roles map[string]string) gitModeObservation {
	t.Helper()
	role := func(raw []byte) string {
		sha := strings.TrimSpace(string(raw))
		if got, ok := roles[sha]; ok {
			return got
		}
		return sha
	}

	topLevel, topLevelErr := reader.GitInDir(root, "rev-parse", "--show-toplevel")
	commonRaw, commonErr := reader.GitInDir(root, "rev-parse", "--git-common-dir")
	if commonErr != nil {
		t.Fatal(commonErr)
	}
	common := strings.TrimSuffix(string(commonRaw), "\n")
	if !filepath.IsAbs(common) {
		common = filepath.Join(root, common)
	}
	commonRole := "unknown:" + common
	if canonicalContractPath(t, common) == canonicalContractPath(t, root) || canonicalContractPath(t, common) == canonicalContractPath(t, filepath.Join(root, ".git")) {
		commonRole = "common"
	}
	head, headErr := reader.GitInDir(root, "rev-parse", "HEAD")
	_, showErr := reader.GitInDir(root, "show", "-s", "--format=%cI", "HEAD")
	_, statusErr := reader.GitInDir(root, "status", "--porcelain=v1", "-z", "--untracked-files=all")
	porcelain, err := reader.GitInDir(root, "worktree", "list", "--porcelain", "-z")
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := gitx.ParseWorktrees(porcelain)
	if err != nil {
		t.Fatal(err)
	}
	worktrees := make([]contractWorktree, 0, len(parsed))
	for _, worktree := range parsed {
		pathRole := "unknown:" + worktree.Path
		if canonicalContractPath(t, worktree.Path) == canonicalContractPath(t, root) {
			pathRole = "root"
		}
		worktrees = append(worktrees, contractWorktree{
			pathRole: pathRole, headRole: role([]byte(worktree.HEAD)), branch: worktree.Branch,
			detached: worktree.Detached, bare: worktree.Bare,
		})
	}
	headRole := ""
	if headErr == nil {
		headRole = role(head)
	}
	return gitModeObservation{
		topLevelExists: topLevelErr == nil && len(topLevel) > 0,
		commonDirRole:  commonRole,
		headRole:       headRole, headExists: headErr == nil,
		showExists: showErr == nil, statusExists: statusErr == nil,
		worktrees: worktrees,
	}
}

func runGitContract(t *testing.T, driver gitContractDriver) []gitContractObservation {
	t.Helper()
	observations := []gitContractObservation{observeGitContract(t, driver, "initial", driver.primary)}
	driver.diverge(t)
	observations = append(observations, observeGitContract(t, driver, "diverged", driver.linked))
	driver.merge(t)
	observations = append(observations, observeGitContract(t, driver, "merged", driver.linked))
	driver.dirty(t)
	observations = append(observations, observeGitContract(t, driver, "dirty", driver.linked))
	return observations
}

func observeGitContract(t *testing.T, driver gitContractDriver, stage, dir string) gitContractObservation {
	t.Helper()
	command := func(args ...string) []byte {
		t.Helper()
		out, err := driver.reader.GitInDir(dir, args...)
		if err != nil {
			t.Fatalf("%s: git %s: %v\n%s", stage, strings.Join(args, " "), err, out)
		}
		return out
	}
	role := func(sha string) string {
		if got, ok := driver.roles[strings.TrimSpace(sha)]; ok {
			return got
		}
		return "unknown:" + strings.TrimSpace(sha)
	}

	topLevel := canonicalContractPath(t, strings.TrimSuffix(string(command("rev-parse", "--show-toplevel")), "\n"))
	commonRaw := strings.TrimSuffix(string(command("rev-parse", "--git-common-dir")), "\n")
	if !filepath.IsAbs(commonRaw) {
		commonRaw = filepath.Join(dir, commonRaw)
	}
	commonDir := canonicalContractPath(t, commonRaw)
	headRole := role(string(command("rev-parse", "HEAD")))
	committedAt := normalizeContractTime(t, strings.TrimSpace(string(command("show", "-s", "--format=%cI", "HEAD"))))
	mainRole := role(string(command("rev-parse", "--verify", "main")))

	originOut, originErr := driver.reader.GitInDir(dir, "rev-parse", "--verify", "--quiet", "origin/main")
	originRole := ""
	if originErr == nil {
		originRole = role(string(originOut))
	}

	worktreeRaw := command("worktree", "list", "--porcelain", "-z")
	parsed, err := gitx.ParseWorktrees(worktreeRaw)
	if err != nil {
		t.Fatalf("%s: ParseWorktrees: %v", stage, err)
	}
	worktrees := make([]contractWorktree, 0, len(parsed))
	for _, worktree := range parsed {
		path := filepath.Clean(worktree.Path)
		if path != filepath.Clean(driver.prunable) {
			path = canonicalContractPath(t, worktree.Path)
		}
		pathRole := "unknown:" + path
		switch path {
		case canonicalContractPath(t, driver.primary):
			pathRole = "primary"
		case canonicalContractPath(t, driver.linked):
			pathRole = "linked"
		case filepath.Clean(driver.prunable):
			pathRole = "prunable"
		}
		locked := ""
		if worktree.Locked != nil {
			locked = "present:" + *worktree.Locked
		}
		prunable := ""
		if worktree.Prunable != nil {
			prunable = "present:" + *worktree.Prunable
		}
		worktrees = append(worktrees, contractWorktree{
			pathRole: pathRole,
			headRole: role(worktree.HEAD),
			branch:   worktree.Branch,
			detached: worktree.Detached,
			bare:     worktree.Bare,
			locked:   locked,
			prunable: prunable,
		})
	}

	status := splitNUL(command("status", "--porcelain=v1", "-z", "--untracked-files=all"))
	behind, ahead := 0, 0
	if originErr == nil {
		counts := strings.Fields(string(command("rev-list", "--left-right", "--count", "origin/main...HEAD")))
		if len(counts) != 2 {
			t.Fatalf("%s: malformed rev-list counts %q", stage, counts)
		}
		if _, err := fmt.Sscan(counts[0], &behind); err != nil {
			t.Fatal(err)
		}
		if _, err := fmt.Sscan(counts[1], &ahead); err != nil {
			t.Fatal(err)
		}
	}

	return gitContractObservation{
		stage:            stage,
		topLevel:         pathRole(driver, topLevel),
		commonDir:        pathRole(driver, commonDir),
		headRole:         headRole,
		committedAt:      committedAt,
		mainRole:         mainRole,
		originMainRole:   originRole,
		originMainExists: originErr == nil,
		worktrees:        worktrees,
		status:           status,
		behind:           behind,
		ahead:            ahead,
	}
}

func assertGitContract(t *testing.T, got []gitContractObservation) {
	t.Helper()
	want := []gitContractObservation{
		{
			stage: "initial", topLevel: "primary", commonDir: "common", headRole: "initial",
			committedAt: "2026-01-02T03:04:05Z", mainRole: "initial",
			worktrees: []contractWorktree{{pathRole: "primary", headRole: "initial", branch: "main"}},
			status:    []string{},
		},
		{
			stage: "diverged", topLevel: "linked", commonDir: "common", headRole: "feature",
			committedAt: "2026-01-04T03:04:05Z", mainRole: "main2", originMainRole: "main2", originMainExists: true,
			worktrees: []contractWorktree{
				{pathRole: "primary", headRole: "main2", branch: "main"},
				{pathRole: "linked", headRole: "feature", branch: "feature"},
			},
			status: []string{}, behind: 2, ahead: 1,
		},
		{
			stage: "merged", topLevel: "linked", commonDir: "common", headRole: "merge",
			committedAt: "2026-01-05T03:04:05Z", mainRole: "main2", originMainRole: "main2", originMainExists: true,
			worktrees: []contractWorktree{
				{pathRole: "primary", headRole: "main2", branch: "main"},
				{pathRole: "linked", headRole: "merge", branch: "feature"},
			},
			status: []string{}, behind: 0, ahead: 2,
		},
		{
			stage: "dirty", topLevel: "linked", commonDir: "common", headRole: "merge",
			committedAt: "2026-01-05T03:04:05Z", mainRole: "main2", originMainRole: "main2", originMainExists: true,
			worktrees: []contractWorktree{
				{pathRole: "primary", headRole: "main2", branch: "main"},
				{pathRole: "linked", headRole: "merge", detached: true, locked: "present:contract lock"},
				{pathRole: "prunable", headRole: "main2", branch: "stale", prunable: "present:gitdir file points to non-existent location"},
			},
			status: []string{
				"C  copied\n file.txt", "copy source.txt",
				"M  copy source.txt",
				"R  renamed \nfile.txt", "rename\n source.txt",
				" M tracked.txt",
				" T typechange.txt",
				"??  untracked\nname.txt",
			},
			behind: 0, ahead: 2,
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("contract observations = %#v, want %#v", got, want)
	}
}

func newFakeGitContract(t *testing.T) gitContractDriver {
	t.Helper()
	root := t.TempDir()
	primary := filepath.Join(root, "repo")
	linked := filepath.Join(root, "linked")
	prunable := filepath.Join(root, "prunable")
	commonDir := filepath.Join(primary, ".git")
	for _, dir := range []string{primary, linked, commonDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	primary = canonicalContractPath(t, primary)
	linked = canonicalContractPath(t, linked)
	commonDir = canonicalContractPath(t, commonDir)
	prunable = filepath.Join(filepath.Dir(primary), "prunable")

	const initialSHA = "1111111111111111111111111111111111111111"
	const mainSHA = "2222222222222222222222222222222222222222"
	const main2SHA = "3333333333333333333333333333333333333333"
	const featureSHA = "4444444444444444444444444444444444444444"
	const mergeSHA = "5555555555555555555555555555555555555555"
	fake := NewFakeGit()
	repo := &FakeGitRepo{
		CommonDir:   commonDir,
		PrimaryRoot: primary,
		Worktrees:   []gitx.Worktree{{Path: primary, HEAD: initialSHA, Branch: "main"}},
		Refs:        map[string]string{"refs/heads/main": initialSHA},
		Commits: map[string]FakeGitCommit{
			initialSHA: {CommittedAt: mustContractTime(t, contractInitialTime)},
		},
		Dirty: map[string][]FakeGitStatusEntry{},
	}
	if err := fake.AddRepo(repo); err != nil {
		t.Fatal(err)
	}
	driver := gitContractDriver{
		reader: fake, primary: primary, linked: linked, prunable: prunable, commonDir: commonDir,
		roles: map[string]string{initialSHA: "initial", mainSHA: "main", main2SHA: "main2", featureSHA: "feature", mergeSHA: "merge"},
	}
	driver.diverge = func(t *testing.T) {
		mustFakeMutation(t, fake.AddCommit(commonDir, mainSHA, FakeGitCommit{Parents: []string{initialSHA}, CommittedAt: mustContractTime(t, contractMainTime)}))
		mustFakeMutation(t, fake.AddCommit(commonDir, main2SHA, FakeGitCommit{Parents: []string{mainSHA}, CommittedAt: mustContractTime(t, contractMain2Time)}))
		mustFakeMutation(t, fake.AddCommit(commonDir, featureSHA, FakeGitCommit{Parents: []string{initialSHA}, CommittedAt: mustContractTime(t, contractFeatureTime)}))
		mustFakeMutation(t, fake.SetRef(commonDir, "refs/heads/main", main2SHA))
		mustFakeMutation(t, fake.SetRef(commonDir, "refs/heads/feature", featureSHA))
		mustFakeMutation(t, fake.SetRef(commonDir, "refs/remotes/origin/main", main2SHA))
		mustFakeMutation(t, fake.AddWorktree(commonDir, gitx.Worktree{Path: linked, HEAD: featureSHA, Branch: "feature"}))
	}
	driver.merge = func(t *testing.T) {
		mustFakeMutation(t, fake.AddCommit(commonDir, mergeSHA, FakeGitCommit{Parents: []string{featureSHA, main2SHA}, CommittedAt: mustContractTime(t, contractMergeTime)}))
		mustFakeMutation(t, fake.SetRef(commonDir, "refs/heads/feature", mergeSHA))
	}
	driver.dirty = func(t *testing.T) {
		reason := "contract lock"
		prunableReason := "gitdir file points to non-existent location"
		mustFakeMutation(t, fake.DetachWorktree(commonDir, linked, mergeSHA))
		mustFakeMutation(t, fake.SetWorktreeAttributes(commonDir, linked, &reason, nil))
		mustFakeMutation(t, fake.SetRef(commonDir, "refs/heads/stale", main2SHA))
		mustFakeMutation(t, fake.AddWorktree(commonDir, gitx.Worktree{Path: prunable, HEAD: main2SHA, Branch: "stale", Prunable: &prunableReason}))
		mustFakeMutation(t, fake.SetDirty(linked, []FakeGitStatusEntry{
			{Code: " M", Path: "tracked.txt"},
			{Code: " T", Path: "typechange.txt"},
			{Code: "R ", Path: "renamed \nfile.txt", SourcePath: "rename\n source.txt"},
			{Code: "C ", Path: "copied\n file.txt", SourcePath: "copy source.txt"},
			{Code: "M ", Path: "copy source.txt"},
			{Code: "??", Path: " untracked\nname.txt"},
		}))
	}
	return driver
}

func newRealGitContract(t *testing.T) gitContractDriver {
	t.Helper()
	root := t.TempDir()
	primary := filepath.Join(root, "repo")
	linked := filepath.Join(root, "linked")
	prunable := filepath.Join(root, "prunable")
	runContractGit(t, root, nil, "init", "-b", "main", primary)
	primary = canonicalContractPath(t, primary)
	root = filepath.Dir(primary)
	linked = filepath.Join(root, "linked")
	prunable = filepath.Join(root, "prunable")
	runContractGit(t, primary, nil, "config", "user.email", "fleet-test@example.test")
	runContractGit(t, primary, nil, "config", "user.name", "Fleet Test")
	runContractGit(t, primary, nil, "config", "commit.gpgsign", "false")
	runContractGit(t, primary, nil, "config", "core.hooksPath", os.DevNull)
	runContractGit(t, primary, nil, "config", "status.renames", "copies")
	writeContractFile(t, primary, "tracked.txt", "initial\n")
	writeContractFile(t, primary, "typechange.txt", "regular\n")
	writeContractFile(t, primary, "rename\n source.txt", "rename\n")
	writeContractFile(t, primary, "copy source.txt", "copy\n")
	runContractGit(t, primary, nil, "add", "tracked.txt", "typechange.txt", "rename\n source.txt", "copy source.txt")
	runContractGit(t, primary, contractDateEnv(contractInitialTime), "commit", "-m", "initial")
	initialSHA := strings.TrimSpace(string(runContractGit(t, primary, nil, "rev-parse", "HEAD")))

	driver := gitContractDriver{
		reader: contractExecGitReader{}, primary: primary, linked: linked, prunable: prunable,
		commonDir: canonicalContractPath(t, filepath.Join(primary, ".git")), roles: map[string]string{initialSHA: "initial"},
	}
	driver.diverge = func(t *testing.T) {
		writeContractFile(t, primary, "main.txt", "main\n")
		runContractGit(t, primary, nil, "add", "main.txt")
		runContractGit(t, primary, contractDateEnv(contractMainTime), "commit", "-m", "main")
		mainSHA := strings.TrimSpace(string(runContractGit(t, primary, nil, "rev-parse", "HEAD")))
		writeContractFile(t, primary, "main2.txt", "main2\n")
		runContractGit(t, primary, nil, "add", "main2.txt")
		runContractGit(t, primary, contractDateEnv(contractMain2Time), "commit", "-m", "main2")
		main2SHA := strings.TrimSpace(string(runContractGit(t, primary, nil, "rev-parse", "HEAD")))
		runContractGit(t, primary, nil, "update-ref", "refs/remotes/origin/main", main2SHA)
		runContractGit(t, primary, nil, "worktree", "add", "-b", "feature", linked, initialSHA)
		writeContractFile(t, linked, "feature.txt", "feature\n")
		runContractGit(t, linked, nil, "add", "feature.txt")
		runContractGit(t, linked, contractDateEnv(contractFeatureTime), "commit", "-m", "feature")
		featureSHA := strings.TrimSpace(string(runContractGit(t, linked, nil, "rev-parse", "HEAD")))
		driver.roles[mainSHA] = "main"
		driver.roles[main2SHA] = "main2"
		driver.roles[featureSHA] = "feature"
	}
	driver.merge = func(t *testing.T) {
		runContractGit(t, linked, contractDateEnv(contractMergeTime), "merge", "--no-ff", "origin/main", "-m", "merge main")
		mergeSHA := strings.TrimSpace(string(runContractGit(t, linked, nil, "rev-parse", "HEAD")))
		driver.roles[mergeSHA] = "merge"
	}
	driver.dirty = func(t *testing.T) {
		writeContractFile(t, linked, "tracked.txt", "dirty\n")
		if err := os.Remove(filepath.Join(linked, "typechange.txt")); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink("tracked.txt", filepath.Join(linked, "typechange.txt")); err != nil {
			t.Fatal(err)
		}
		writeContractFile(t, linked, " untracked\nname.txt", "untracked\n")
		writeContractFile(t, linked, "copy source.txt", "changed\n")
		writeContractFile(t, linked, "copied\n file.txt", "copy\n")
		runContractGit(t, linked, nil, "add", "copy source.txt", "copied\n file.txt")
		runContractGit(t, linked, nil, "mv", "rename\n source.txt", "renamed \nfile.txt")
		runContractGit(t, linked, nil, "checkout", "--detach")
		runContractGit(t, primary, nil, "worktree", "lock", "--reason", "contract lock", linked)
		runContractGit(t, primary, nil, "worktree", "add", "-b", "stale", prunable, "main")
		if err := os.RemoveAll(prunable); err != nil {
			t.Fatal(err)
		}
	}
	return driver
}

func newFakeBareContract(t *testing.T) (GitReader, string, map[string]string) {
	t.Helper()
	bare := filepath.Join(t.TempDir(), "bare.git")
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
	return fake, canonicalContractPath(t, bare), map[string]string{head: "head"}
}

func newRealBareContract(t *testing.T) (GitReader, string, map[string]string) {
	t.Helper()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	bare := filepath.Join(root, "bare.git")
	runContractGit(t, root, nil, "init", "-b", "main", source)
	runContractGit(t, source, nil, "config", "user.email", "fleet-test@example.test")
	runContractGit(t, source, nil, "config", "user.name", "Fleet Test")
	runContractGit(t, source, nil, "config", "commit.gpgsign", "false")
	runContractGit(t, source, contractDateEnv(contractInitialTime), "commit", "--allow-empty", "-m", "initial")
	head := strings.TrimSpace(string(runContractGit(t, source, nil, "rev-parse", "HEAD")))
	runContractGit(t, root, nil, "clone", "--bare", source, bare)
	return contractExecGitReader{}, canonicalContractPath(t, bare), map[string]string{head: "head"}
}

func newFakeUnbornContract(t *testing.T) (GitReader, string) {
	t.Helper()
	primary := filepath.Join(t.TempDir(), "repo")
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
	return fake, canonicalContractPath(t, primary)
}

func newRealUnbornContract(t *testing.T) (GitReader, string) {
	t.Helper()
	primary := filepath.Join(t.TempDir(), "repo")
	runContractGit(t, filepath.Dir(primary), nil, "init", "-b", "main", primary)
	return contractExecGitReader{}, canonicalContractPath(t, primary)
}

func runContractGit(t *testing.T, dir string, extraEnv []string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = isolatedContractGitEnv(extraEnv)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git -C %q %s: %v\n%s", dir, strings.Join(args, " "), err, out)
	}
	return out
}

type contractExecGitReader struct{}

func (contractExecGitReader) GitInDir(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = isolatedContractGitEnv(nil)
	return cmd.CombinedOutput()
}

func isolatedContractGitEnv(extra []string) []string {
	env := make([]string, 0, len(os.Environ())+len(extra)+3)
	for _, entry := range os.Environ() {
		if strings.HasPrefix(entry, "GIT_") {
			continue
		}
		env = append(env, entry)
	}
	env = append(env,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL="+os.DevNull,
		"GIT_CONFIG_COUNT=0",
	)
	return append(env, extra...)
}

func mustFakeMutation(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func contractDateEnv(date string) []string {
	return []string{"GIT_AUTHOR_DATE=" + date, "GIT_COMMITTER_DATE=" + date}
}

func writeContractFile(t *testing.T, root, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustContractTime(t *testing.T, value string) time.Time {
	t.Helper()
	got, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func normalizeContractTime(t *testing.T, value string) string {
	t.Helper()
	got, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("parse Git commit timestamp %q: %v", value, err)
	}
	return got.UTC().Format(time.RFC3339)
}

func canonicalContractPath(t *testing.T, path string) string {
	t.Helper()
	got, err := canonicalPath(path)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func pathRole(driver gitContractDriver, path string) string {
	switch path {
	case driver.primary:
		return "primary"
	case driver.linked:
		return "linked"
	case driver.prunable:
		return "prunable"
	case driver.commonDir:
		return "common"
	default:
		return "unknown:" + path
	}
}

func splitNUL(raw []byte) []string {
	fields := bytes.Split(raw, []byte{0})
	result := make([]string, 0, len(fields))
	for _, field := range fields {
		if len(field) > 0 {
			result = append(result, string(field))
		}
	}
	return result
}
