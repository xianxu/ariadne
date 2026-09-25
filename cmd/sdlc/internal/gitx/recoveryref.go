package gitx

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
)

var (
	ErrRecoveryAbsent  = errors.New("recovery record absent")
	ErrRecoveryChanged = errors.New("recovery record changed since it was read")
)

const recoveryRefPrefix = "refs/sdlc/recovery/"

var recoveryToken = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

// RecoveryData is one operation's private recovery material. Document is the
// caller's serialized receipt; Objects are existing objects (source blobs,
// unpublished candidates) the record must keep reachable until finalization.
type RecoveryData struct {
	Document []byte
	Objects  []string
}

// RecoveryEntry is a value snapshot of one record; OID is the CAS generation.
type RecoveryEntry struct {
	Token string
	OID   string
	Data  RecoveryData
}

// RecoveryStore keeps operation receipts in local refs under
// refs/sdlc/recovery/<token>. A ref, not a string inside a JSON document, is
// what keeps pinned objects alive through gc: blobs/trees become tree entries
// and commits become parents of the record commit. Records are checkout-local
// and never pushed; the tracker's published provenance is the cross-clone
// record. Every write is a compare-and-swap against the generation read.
type RecoveryStore struct {
	ctx context.Context
	dir string
}

func NewRecoveryStore(ctx context.Context, dir string) (*RecoveryStore, error) {
	if ctx == nil || dir == "" {
		return nil, errors.New("recovery store requires context and repository")
	}
	return &RecoveryStore{ctx: ctx, dir: dir}, nil
}

func (s *RecoveryStore) git(input []byte, limit int, args ...string) ([]byte, []byte, error) {
	var in io.Reader
	if input != nil {
		in = bytes.NewReader(input)
	}
	return runGitBoundedInputContext(s.ctx, s.dir, recoveryIdentityEnv, in, limit, args...)
}

// Record commits are private bookkeeping; they must not depend on (or be
// refused by) the operator's identity or signing configuration.
var recoveryIdentityEnv = []string{
	"GIT_AUTHOR_NAME=sdlc recovery", "GIT_AUTHOR_EMAIL=sdlc@localhost",
	"GIT_COMMITTER_NAME=sdlc recovery", "GIT_COMMITTER_EMAIL=sdlc@localhost",
}

func recoveryRef(token string) (string, error) {
	if !recoveryToken.MatchString(token) {
		return "", fmt.Errorf("invalid recovery token %q", token)
	}
	return recoveryRefPrefix + token, nil
}

// Save creates (expected == "") or replaces (expected == current OID) a record.
func (s *RecoveryStore) Save(token, expected string, data RecoveryData) (string, error) {
	ref, err := recoveryRef(token)
	if err != nil {
		return "", err
	}
	if len(data.Document) == 0 || len(data.Document) > SnapshotBlobLimit {
		return "", errors.New("recovery document must be nonempty and within the blob limit")
	}
	if expected != "" {
		if _, err := parseObjectID([]byte(expected)); err != nil {
			return "", err
		}
	}
	var entries []string
	var parents []string
	objects := make([]string, 0, len(data.Objects))
	seen := map[string]bool{}
	for _, oid := range data.Objects {
		if !fullObjectID(oid) {
			return "", fmt.Errorf("recovery object %q is not a full object ID", oid)
		}
		if seen[oid] {
			continue
		}
		seen[oid] = true
		out, diag, err := s.git(nil, diagnosticOutputLimit, "cat-file", "-t", oid)
		if err != nil {
			if s.ctx.Err() != nil {
				return "", s.ctx.Err()
			}
			return "", fmt.Errorf("recovery object %s unavailable: %w: %s", oid, err, diag)
		}
		switch kind := strings.TrimSpace(string(out)); kind {
		case "blob":
			entries = append(entries, "100644 blob "+oid+"\t"+"pinned-"+oid)
		case "tree":
			entries = append(entries, "040000 tree "+oid+"\t"+"pinned-"+oid)
		case "commit":
			parents = append(parents, oid)
		default:
			return "", fmt.Errorf("recovery object %s has unsupported type %q", oid, kind)
		}
		objects = append(objects, oid)
	}
	doc, err := s.hashBlob(data.Document)
	if err != nil {
		return "", err
	}
	manifest, err := s.hashBlob([]byte(strings.Join(objects, "\n") + "\n"))
	if err != nil {
		return "", err
	}
	entries = append(entries, "100644 blob "+doc+"\tdocument", "100644 blob "+manifest+"\tobjects")
	sort.Slice(entries, func(i, j int) bool {
		return strings.SplitN(entries[i], "\t", 2)[1] < strings.SplitN(entries[j], "\t", 2)[1]
	})
	out, diag, err := s.git([]byte(strings.Join(entries, "\n")+"\n"), diagnosticOutputLimit, "mktree")
	if err != nil {
		return "", s.fail("write recovery tree", err, diag)
	}
	tree, err := parseObjectID(out)
	if err != nil {
		return "", err
	}
	args := []string{"commit-tree", tree, "-m", "sdlc recovery " + token}
	for _, p := range parents {
		args = append(args, "-p", p)
	}
	out, diag, err = s.git(nil, diagnosticOutputLimit, args...)
	if err != nil {
		return "", s.fail("write recovery commit", err, diag)
	}
	record, err := parseObjectID(out)
	if err != nil {
		return "", err
	}
	old := expected
	if old == "" {
		old = strings.Repeat("0", len(record))
	}
	if err := s.swap(ref, record, old, expected); err != nil {
		return "", err
	}
	return record, nil
}

// Delete removes a record only at the generation the caller read.
func (s *RecoveryStore) Delete(token, expected string) error {
	ref, err := recoveryRef(token)
	if err != nil {
		return err
	}
	if _, err := parseObjectID([]byte(expected)); err != nil {
		return errors.New("recovery delete requires the generation read")
	}
	_, diag, err := s.git(nil, diagnosticOutputLimit, "update-ref", "-d", ref, expected)
	if err != nil {
		return s.changedOr(ref, expected, "delete recovery record", err, diag)
	}
	return nil
}

func (s *RecoveryStore) swap(ref, record, old, expected string) error {
	_, diag, err := s.git(nil, diagnosticOutputLimit, "update-ref", "-m", "sdlc recovery", ref, record, old)
	if err != nil {
		return s.changedOr(ref, expected, "update recovery record", err, diag)
	}
	return nil
}

// A failed update-ref is attributed to a concurrent change only when the ref no
// longer holds the expected value; any other failure is reported as itself.
func (s *RecoveryStore) changedOr(ref, expected, what string, err error, diag []byte) error {
	if s.ctx.Err() != nil {
		return s.ctx.Err()
	}
	current, rerr := s.current(ref)
	if rerr == nil && current != expected {
		return fmt.Errorf("%w: %s", ErrRecoveryChanged, ref)
	}
	return s.fail(what, err, diag)
}

func (s *RecoveryStore) current(ref string) (string, error) {
	out, diag, err := s.git(nil, diagnosticOutputLimit, "rev-parse", "-q", "--verify", ref+"^{commit}")
	if err != nil {
		if s.ctx.Err() != nil {
			return "", s.ctx.Err()
		}
		if gitExitCode(err) == 1 && len(out) == 0 {
			return "", nil
		}
		return "", s.fail("read recovery ref", err, diag)
	}
	return parseObjectID(out)
}

func (s *RecoveryStore) Load(token string) (RecoveryEntry, error) {
	ref, err := recoveryRef(token)
	if err != nil {
		return RecoveryEntry{}, err
	}
	oid, err := s.current(ref)
	if err != nil {
		return RecoveryEntry{}, err
	}
	if oid == "" {
		return RecoveryEntry{}, fmt.Errorf("%w: %s", ErrRecoveryAbsent, token)
	}
	entry := RecoveryEntry{Token: token, OID: oid}
	if entry.Data.Document, err = s.readBlob(oid + ":document"); err != nil {
		return entry, err
	}
	manifest, err := s.readBlob(oid + ":objects")
	if err != nil {
		return entry, err
	}
	for _, line := range strings.Split(strings.TrimSuffix(string(manifest), "\n"), "\n") {
		if line == "" {
			continue
		}
		if !fullObjectID(line) {
			return entry, fmt.Errorf("malformed recovery manifest in %s", token)
		}
		entry.Data.Objects = append(entry.Data.Objects, line)
	}
	return entry, nil
}

// List reports every local record's token and generation, without contents.
func (s *RecoveryStore) List() ([]RecoveryEntry, error) {
	out, diag, err := s.git(nil, snapshotOutputLimit, "for-each-ref", "--format=%(objectname) %(refname)", recoveryRefPrefix)
	if err != nil {
		return nil, s.fail("list recovery records", err, diag)
	}
	var entries []RecoveryEntry
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 || !strings.HasPrefix(fields[1], recoveryRefPrefix) {
			return nil, errors.New("malformed recovery ref listing")
		}
		token := strings.TrimPrefix(fields[1], recoveryRefPrefix)
		oid, err := parseObjectID([]byte(fields[0]))
		if err != nil || !recoveryToken.MatchString(token) {
			return nil, fmt.Errorf("malformed recovery ref %q", fields[1])
		}
		entries = append(entries, RecoveryEntry{Token: token, OID: oid})
	}
	return entries, nil
}

func (s *RecoveryStore) hashBlob(content []byte) (string, error) {
	out, diag, err := s.git(content, diagnosticOutputLimit, "hash-object", "-w", "--no-filters", "--stdin")
	if err != nil {
		return "", s.fail("write recovery blob", err, diag)
	}
	return parseObjectID(out)
}

func (s *RecoveryStore) readBlob(spec string) ([]byte, error) {
	out, diag, err := s.git(nil, SnapshotBlobLimit, "cat-file", "blob", spec)
	if err != nil {
		return nil, s.fail("read recovery blob", err, diag)
	}
	return out, nil
}

func (s *RecoveryStore) fail(what string, err error, diag []byte) error {
	if s.ctx.Err() != nil {
		return s.ctx.Err()
	}
	return fmt.Errorf("%s: %w: %s", what, err, diag)
}

func fullObjectID(oid string) bool {
	if len(oid) != 40 && len(oid) != 64 {
		return false
	}
	_, err := hex.DecodeString(oid)
	return err == nil && strings.ToLower(oid) == oid && strings.Trim(oid, "0") != ""
}
