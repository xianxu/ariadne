package fleet

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

func TestCollectFacts_FakeAndRealGitConformance(t *testing.T) {
	fake := newFakeGitContract(t)
	real := newRealGitContract(t)
	for _, driver := range []*gitContractDriver{&fake, &real} {
		driver.diverge(t)
		driver.merge(t)
		driver.dirty(t)
	}

	fakeFacts := CollectFacts(fake.reader, fake.linked)
	realFacts := CollectFacts(real.reader, real.linked)
	if got, want := observeFacts(t, fakeFacts, fake.roles), observeFacts(t, realFacts, real.roles); got != want {
		t.Fatalf("fake/real facts differ\nfake: %#v\nreal: %#v", got, want)
	}
	if got := observeFacts(t, fakeFacts, fake.roles); got != (factsObservation{
		Available: true, Head: "merge", Timestamp: "2026-01-05T03:04:05Z",
		BaseAvailable: true, BaseRef: "origin/main", Ahead: 2, Behind: 0, DirtyCount: 6,
	}) {
		t.Fatalf("CollectFacts() = %#v, want measured contract facts", got)
	}
}

func TestCollectFacts_FallsBackToMainAndReportsNoBaseExplicitly(t *testing.T) {
	driver := newFakeGitContract(t)
	facts := CollectFacts(driver.reader, driver.primary)
	got := observeFacts(t, facts, driver.roles)
	if !got.Available || !got.BaseAvailable || got.BaseRef != "main" || got.Ahead != 0 || got.Behind != 0 {
		t.Fatalf("CollectFacts(main fallback) = %#v, want available main base with zero divergence", got)
	}

	fake := driver.reader.(*FakeGit)
	mustFakeMutation(t, fake.DetachWorktree(driver.commonDir, driver.primary, "1111111111111111111111111111111111111111"))
	mustFakeMutation(t, fake.DeleteRef(driver.commonDir, "refs/heads/main"))
	noBase := CollectFacts(driver.reader, driver.primary)
	if !noBase.Available || noBase.BaseAvailable || noBase.BaseRef != "" || noBase.BaseError == "" || noBase.Ahead != nil || noBase.Behind != nil {
		t.Fatalf("CollectFacts(no base) = %#v, want explicit unavailable base without zero divergence", noBase)
	}
}

func TestCollectFacts_FakeAndRealGitAgreeOnMainFallback(t *testing.T) {
	fake := newFakeGitContract(t)
	real := newRealGitContract(t)
	fakeFacts := CollectFacts(fake.reader, fake.primary)
	realFacts := CollectFacts(real.reader, real.primary)
	if got, want := observeFacts(t, fakeFacts, fake.roles), observeFacts(t, realFacts, real.roles); got != want {
		t.Fatalf("fake/real main fallback differs\nfake: %#v\nreal: %#v", got, want)
	}
	if got := observeFacts(t, realFacts, real.roles); !got.Available || !got.BaseAvailable || got.BaseRef != "main" || got.Ahead != 0 || got.Behind != 0 {
		t.Fatalf("CollectFacts(real main fallback) = %#v, want available main base", got)
	}
}

func TestCollectFacts_FakeAndRealGitAgreeOnExplicitNoBase(t *testing.T) {
	fake := newFakeGitContract(t)
	fakeGit := fake.reader.(*FakeGit)
	mustFakeMutation(t, fakeGit.DetachWorktree(fake.commonDir, fake.primary, "1111111111111111111111111111111111111111"))
	mustFakeMutation(t, fakeGit.DeleteRef(fake.commonDir, "refs/heads/main"))

	real := newRealGitContract(t)
	runContractGit(t, real.primary, nil, "checkout", "--detach")
	runContractGit(t, real.primary, nil, "branch", "-D", "main")

	fakeFacts := CollectFacts(fake.reader, fake.primary)
	realFacts := CollectFacts(real.reader, real.primary)
	if got, want := normalizeFactsAvailability(fakeFacts), normalizeFactsAvailability(realFacts); got != want {
		t.Fatalf("fake/real no-base availability differs\nfake: %#v\nreal: %#v", got, want)
	}
	for _, facts := range []MeasuredFacts{fakeFacts, realFacts} {
		if !facts.Available || facts.BaseAvailable || facts.BaseError == "" || facts.BaseRef != "" || facts.Ahead != nil || facts.Behind != nil {
			t.Fatalf("CollectFacts(no base) = %#v, want explicit unavailable base", facts)
		}
	}
}

func TestCollectFacts_FakeAndRealGitAgreeOnUnbornCommandError(t *testing.T) {
	fakeReader, fakeRoot := newFakeUnbornContract(t)
	realReader, realRoot := newRealUnbornContract(t)
	fakeFacts := CollectFacts(fakeReader, fakeRoot)
	realFacts := CollectFacts(realReader, realRoot)
	if got, want := normalizeFactsAvailability(fakeFacts), normalizeFactsAvailability(realFacts); got != want {
		t.Fatalf("fake/real unborn availability differs\nfake: %#v\nreal: %#v", got, want)
	}
	for _, facts := range []MeasuredFacts{fakeFacts, realFacts} {
		if facts.Available || facts.Error == "" || facts.Head != "" || facts.CommitTimestamp != "" || facts.DirtyCount != nil {
			t.Fatalf("CollectFacts(unborn) = %#v, want explicit command-error facts", facts)
		}
	}
}

func TestCollectFacts_PreservesGitFailure(t *testing.T) {
	driver := newFakeGitContract(t)
	facts := CollectFacts(failingFactsReader{GitReader: driver.reader, fail: "status --porcelain=v1 -z --untracked-files=all"}, driver.primary)
	if facts.Available || facts.Error == "" || facts.DirtyCount != nil {
		t.Fatalf("CollectFacts(status failure) = %#v, want explicit unavailable error without dirty zero", facts)
	}
	if !strings.Contains(facts.Error, "status --porcelain=v1 -z --untracked-files=all") || !strings.Contains(facts.Error, "injected failure") {
		t.Fatalf("CollectFacts(status failure) error = %q, want command and Git output", facts.Error)
	}
}

func TestCollectFacts_PartialFailuresMarshalAsCoherentPrefixes(t *testing.T) {
	driver := newFakeGitContract(t)
	for _, failure := range []string{
		"rev-parse HEAD",
		"show -s --format=%cI HEAD",
		"status --porcelain=v1 -z --untracked-files=all",
	} {
		t.Run(failure, func(t *testing.T) {
			facts := CollectFacts(failingFactsReader{GitReader: driver.reader, fail: failure}, driver.primary)
			row := validTreeRow()
			row.Facts = facts
			if _, err := json.Marshal(row); err != nil {
				t.Fatalf("CollectFacts(%s) partial facts %#v rejected by JSON contract: %v", failure, facts, err)
			}
		})
	}
}

func TestCollectFacts_OnlyFallsBackForMissingRefExitCode(t *testing.T) {
	driver := newFakeGitContract(t)
	reader := &probeRecordingReader{GitReader: driver.reader, originErr: fmt.Errorf("repository corruption")}
	facts := CollectFacts(reader, driver.primary)
	if !facts.Available || facts.BaseAvailable || facts.BaseRef != "" || !strings.Contains(facts.BaseError, "repository corruption") {
		t.Fatalf("CollectFacts(unexpected origin error) = %#v, want preserved operational error", facts)
	}
	if reader.mainProbed {
		t.Fatal("CollectFacts fell back to main after an operational origin/main error")
	}

	_, err := driver.reader.GitInDir(driver.primary, "rev-parse", "--verify", "--quiet", "origin/main")
	var exitCoder interface{ ExitCode() int }
	if !errors.As(err, &exitCoder) || exitCoder.ExitCode() != 1 {
		t.Fatalf("FakeGit missing ref error = %v, want exit code 1", err)
	}
}

func TestCountStatusEntries_NULSafeRenameCopyAndControlPaths(t *testing.T) {
	status := []byte(" M tracked\x00R  renamed\npath\x00source\npath\x00C  copied\rpath\x00copy\rsource\x00?? untracked\npath\x00")
	count, err := countStatusEntries(status)
	if err != nil || count != 4 {
		t.Fatalf("countStatusEntries() = (%d, %v), want (4, nil)", count, err)
	}
}

func TestCountStatusEntries_RejectsTruncatedAndInvalidPorcelain(t *testing.T) {
	tests := [][]byte{
		[]byte(" M missing-final-nul"),
		[]byte("ZZ unknown\x00"),
		[]byte("!  ignored\x00"),
		[]byte("R  target\x00"),
		[]byte("R  target\x00\x00"),
		[]byte(" M ordinary\x00\x00"),
		[]byte(" M ordinary\x00source\x00"),
	}
	for _, raw := range tests {
		if _, err := countStatusEntries(raw); err == nil {
			t.Errorf("countStatusEntries(%q) succeeded, want parse error", raw)
		}
	}
}

func TestParseDivergence_RequiresExactGitFraming(t *testing.T) {
	behind, ahead, err := parseDivergence([]byte("12\t34\n"))
	if err != nil || behind != 12 || ahead != 34 {
		t.Fatalf("parseDivergence(valid) = (%d, %d, %v)", behind, ahead, err)
	}
	for _, raw := range [][]byte{
		[]byte("12 34\n"), []byte(" 12\t34\n"), []byte("12\t34"),
		[]byte("12\t34\n\n"), []byte("+12\t34\n"), []byte("-12\t34\n"),
		[]byte("12\t34\t56\n"), []byte("999999999999999999999999\t1\n"),
	} {
		if _, _, err := parseDivergence(raw); err == nil {
			t.Errorf("parseDivergence(%q) succeeded, want strict framing error", raw)
		}
	}
}

func TestCountStatusEntries_GeneratedValidPorcelain(t *testing.T) {
	rng := rand.New(rand.NewSource(200))
	codes := []string{" M", "M ", "R ", " C", "??", "UU"}
	for n := 0; n < 200; n++ {
		entries := make([]FakeGitStatusEntry, 1+rng.Intn(12))
		for i := range entries {
			code := codes[rng.Intn(len(codes))]
			entries[i] = FakeGitStatusEntry{
				Code: code,
				Path: arbitraryNonNULBytes(rng, 1+rng.Intn(64)),
			}
			if strings.ContainsAny(code, "RC") {
				entries[i].SourcePath = arbitraryNonNULBytes(rng, 1+rng.Intn(64))
			}
		}
		porcelain, err := fakeStatusPorcelain(entries)
		if err != nil {
			t.Fatal(err)
		}
		count, err := countStatusEntries(porcelain)
		if err != nil || count != len(entries) {
			t.Fatalf("countStatusEntries generated %q = (%d, %v), want (%d, nil)", porcelain, count, err, len(entries))
		}
	}
}

func FuzzCountStatusEntries(f *testing.F) {
	f.Add([]byte(" M tracked\x00R  renamed\x00source\x00?? untracked\x00"))
	f.Fuzz(func(t *testing.T, raw []byte) {
		entries := make([]FakeGitStatusEntry, 1+len(raw)%9)
		for i := range entries {
			code := []string{" M", "R ", "C ", "??"}[i%4]
			entries[i] = FakeGitStatusEntry{Code: code, Path: fuzzNonNUL(raw, i)}
			if strings.ContainsAny(code, "RC") {
				entries[i].SourcePath = fuzzNonNUL(raw, i+len(entries))
			}
		}
		porcelain, err := fakeStatusPorcelain(entries)
		if err != nil {
			t.Fatal(err)
		}
		count, err := countStatusEntries(porcelain)
		if err != nil || count != len(entries) {
			t.Fatalf("countStatusEntries generated %q = (%d, %v), want (%d, nil)", porcelain, count, err, len(entries))
		}
	})
}

func FuzzCountStatusEntriesRaw(f *testing.F) {
	f.Add([]byte("R  target\x00source\x00"))
	f.Fuzz(func(t *testing.T, raw []byte) {
		count, err := countStatusEntries(raw)
		if err == nil && count < 0 {
			t.Fatalf("countStatusEntries(%q) = %d", raw, count)
		}
	})
}

func FuzzParseDivergence(f *testing.F) {
	f.Add([]byte("0\t1\n"))
	f.Fuzz(func(t *testing.T, raw []byte) {
		behind, ahead, err := parseDivergence(raw)
		if err == nil && (behind < 0 || ahead < 0) {
			t.Fatalf("parseDivergence(%q) = (%d, %d, nil)", raw, behind, ahead)
		}
	})
}

type factsObservation struct {
	Available     bool
	Head          string
	Timestamp     string
	BaseAvailable bool
	BaseRef       string
	Ahead         int
	Behind        int
	DirtyCount    int
}

type factsAvailability struct {
	Available     bool
	HasError      bool
	HasHead       bool
	HasTimestamp  bool
	HasDirtyCount bool
	BaseAvailable bool
	HasBaseError  bool
	HasBaseRef    bool
	HasDivergence bool
}

func normalizeFactsAvailability(facts MeasuredFacts) factsAvailability {
	return factsAvailability{
		Available: facts.Available, HasError: facts.Error != "", HasHead: facts.Head != "",
		HasTimestamp: facts.CommitTimestamp != "", HasDirtyCount: facts.DirtyCount != nil,
		BaseAvailable: facts.BaseAvailable, HasBaseError: facts.BaseError != "", HasBaseRef: facts.BaseRef != "",
		HasDivergence: facts.Ahead != nil || facts.Behind != nil,
	}
}

func observeFacts(t *testing.T, facts MeasuredFacts, roles map[string]string) factsObservation {
	t.Helper()
	result := factsObservation{Available: facts.Available, BaseAvailable: facts.BaseAvailable, BaseRef: facts.BaseRef}
	if role, ok := roles[facts.Head]; ok {
		result.Head = role
	} else {
		result.Head = facts.Head
	}
	result.Timestamp = normalizeContractTime(t, facts.CommitTimestamp)
	if facts.Ahead != nil {
		result.Ahead = *facts.Ahead
	}
	if facts.Behind != nil {
		result.Behind = *facts.Behind
	}
	if facts.DirtyCount != nil {
		result.DirtyCount = *facts.DirtyCount
	}
	return result
}

type failingFactsReader struct {
	GitReader
	fail string
}

type probeRecordingReader struct {
	GitReader
	originErr  error
	mainProbed bool
}

func (r *probeRecordingReader) GitInDir(dir string, args ...string) ([]byte, error) {
	if strings.Join(args, " ") == "rev-parse --verify --quiet origin/main" {
		return []byte("operational output"), r.originErr
	}
	if strings.Join(args, " ") == "rev-parse --verify --quiet main" {
		r.mainProbed = true
	}
	return r.GitReader.GitInDir(dir, args...)
}

func (r failingFactsReader) GitInDir(dir string, args ...string) ([]byte, error) {
	if strings.Join(args, " ") == r.fail {
		return []byte("injected failure"), fmt.Errorf("injected failure")
	}
	return r.GitReader.GitInDir(dir, args...)
}

func arbitraryNonNULBytes(rng *rand.Rand, length int) string {
	raw := make([]byte, length)
	for i := range raw {
		raw[i] = byte(1 + rng.Intn(255))
	}
	return string(raw)
}

func fuzzNonNUL(raw []byte, salt int) string {
	if len(raw) == 0 {
		return fmt.Sprintf("path-%d", salt)
	}
	path := make([]byte, len(raw))
	for i, value := range raw {
		if value == 0 {
			path[i] = byte(1 + (i+salt)%255)
			continue
		}
		path[i] = value
	}
	return string(path)
}
