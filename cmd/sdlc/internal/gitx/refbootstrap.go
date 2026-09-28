package gitx

import (
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"
)

// BootstrapOutcome records publication knowledge, not the desired ref state.
type BootstrapOutcome uint8

const (
	BootstrapNotPublished BootstrapOutcome = iota
	BootstrapExisting
	BootstrapCreated
	BootstrapRejected
	BootstrapUncertain
)

var (
	ErrBootstrapExists   = errors.New("bootstrap destination already exists")
	ErrBootstrapRejected = errors.New("bootstrap expected-absence reservation rejected")
)

// BootstrapResult is a value snapshot. RemoteTip is the exact tip observed before
// refusal, or the candidate when creation was acknowledged; it is not a live ref.
type BootstrapResult struct {
	Candidate string
	RemoteTip string
	Outcome   BootstrapOutcome
}

// Bootstrap creates an orphan branch containing only files (ordinary 100644
// blobs). It never changes caller refs, index or worktree. The caller supplies a
// unique operation token in message and durably saves the candidate in beforePush.
// Existing destinations always refuse, including identical trees. An unknown
// push stays unknown: candidate reachability alone cannot establish which caller
// created the ref. Recovery belongs to the caller holding the durable receipt.
func (t *TrunkFile) Bootstrap(files map[string][]byte, message string, beforePush func(BootstrapResult) error) (BootstrapResult, error) {
	result := BootstrapResult{Outcome: BootstrapNotPublished}
	if beforePush == nil || strings.TrimSpace(message) == "" {
		return result, errors.New("bootstrap requires a receipt callback and operation-token message")
	}
	paths, err := bootstrapPaths(files)
	if err != nil {
		return result, err
	}
	// Own the bytes before any callbacks can run.
	snapshot := make(map[string][]byte, len(files))
	for _, p := range paths {
		snapshot[p] = append([]byte(nil), files[p]...)
	}
	ref := t.localRef()
	if _, diag, err := t.run(nil, "check-ref-format", ref); err != nil {
		return result, fmt.Errorf("invalid bootstrap ref: %w\n%s", err, diag)
	}
	out, diag, err := t.run(nil, "ls-remote", "--refs", "--exit-code", "--", t.remote, ref)
	if err == nil {
		fields := strings.Fields(string(out))
		if len(fields) != 2 || fields[1] != ref {
			return result, errors.New("malformed bootstrap remote-ref response")
		}
		result.RemoteTip, err = parseObjectID([]byte(fields[0]))
		if err != nil {
			return result, err
		}
		result.Outcome = BootstrapExisting
		return result, ErrBootstrapExists
	}
	if gitExitCode(err) != 2 || len(out) != 0 {
		return result, fmt.Errorf("inspect bootstrap destination: %w\n%s", err, diag)
	}
	result.Candidate, err = t.bootstrapCommit(paths, snapshot, message)
	if err != nil {
		return result, err
	}
	if err := beforePush(result); err != nil {
		return result, fmt.Errorf("save bootstrap receipt: %w", err)
	}
	if err := t.operationContext().Err(); err != nil {
		return result, err
	}
	out, diag, err = t.run(nil, "push", "--porcelain", "--force-with-lease="+ref+":", "--", t.remote, result.Candidate+":"+ref)
	observation := observedPush(err, append(out, diag...))
	if t.operationContext().Err() != nil {
		// Cancellation can leave partial porcelain output; it is not a completed
		// rejection predicate and grants no permission to repeat the reservation.
		observation = pushUnknown
	}
	switch observation {
	case pushAccepted:
		result.Outcome, result.RemoteTip = BootstrapCreated, result.Candidate
		return result, nil
	case pushRejected:
		result.Outcome = BootstrapRejected
		return result, fmt.Errorf("%w: %v\n%s%s", ErrBootstrapRejected, err, out, diag)
	default:
		result.Outcome = BootstrapUncertain
		return result, fmt.Errorf("%w: bootstrap candidate %s: %v\n%s%s", ErrPublicationUncertain, result.Candidate, err, out, diag)
	}
}

func bootstrapPaths(files map[string][]byte) ([]string, error) {
	if len(files) == 0 {
		return nil, errors.New("bootstrap requires a nonempty file snapshot")
	}
	paths := make([]string, 0, len(files))
	for p := range files {
		if p == "." || path.IsAbs(p) || path.Clean(p) != p || strings.ContainsAny(p, "\\\x00\r\n:*?[") {
			return nil, fmt.Errorf("invalid bootstrap path %q", p)
		}
		for _, component := range strings.Split(p, "/") {
			if component == "" || component == ".." || strings.EqualFold(component, ".git") {
				return nil, fmt.Errorf("invalid bootstrap path %q", p)
			}
		}
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		for parent := path.Dir(p); parent != "."; parent = path.Dir(parent) {
			if _, exists := files[parent]; exists {
				return nil, fmt.Errorf("overlapping bootstrap paths %q and %q", parent, p)
			}
		}
	}
	return paths, nil
}

// An orphan has no parent tree, so commitSet's parent-based operation cannot be
// used. Reuse its private index/blob files and signing policy, fixing every mode
// to regular file and disabling checkout filters to preserve snapshot bytes.
func (t *TrunkFile) bootstrapCommit(paths []string, files map[string][]byte, message string) (string, error) {
	sign, err := t.signs()
	if err != nil {
		return "", err
	}
	idx, cleanup, err := tempIndexPath()
	defer cleanup()
	if err != nil {
		return "", err
	}
	env := []string{"GIT_INDEX_FILE=" + idx}
	if _, diag, err := t.run(env, "read-tree", "--empty"); err != nil {
		return "", fmt.Errorf("initialize bootstrap tree: %w\n%s", err, diag)
	}
	for _, p := range paths {
		file, remove, err := writeTemp(files[p])
		if err != nil {
			remove()
			return "", err
		}
		out, diag, err := t.run(env, "hash-object", "-w", "--no-filters", "--", file)
		remove()
		if err != nil {
			return "", fmt.Errorf("write bootstrap blob: %w\n%s", err, diag)
		}
		blob, err := parseObjectID(out)
		if err != nil {
			return "", err
		}
		if _, diag, err := t.run(env, "update-index", "--add", "--cacheinfo", "100644,"+blob+","+p); err != nil {
			return "", fmt.Errorf("index bootstrap blob: %w\n%s", err, diag)
		}
	}
	out, diag, err := t.run(env, "write-tree")
	if err != nil {
		return "", fmt.Errorf("write bootstrap tree: %w\n%s", err, diag)
	}
	tree, err := parseObjectID(out)
	if err != nil {
		return "", err
	}
	args := []string{"commit-tree", tree, "-m", message}
	if sign {
		args = append(args, "-S")
	}
	out, diag, err = t.run(nil, args...)
	if err != nil {
		return "", fmt.Errorf("create bootstrap commit: %w\n%s", err, diag)
	}
	return parseObjectID(out)
}
