package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Stateful doubles keep the real files, per-file locks and published blobs;
// failures can happen after an effect, just as at the actual boundaries.
type workingdocSyncFake struct {
	SessionSync
	root       string
	registered map[string]*syncRegistration
	locks      map[string]bool
	stamps     int
}

func (f *workingdocSyncFake) RegistrationSnapshot() (map[string]*syncRegistration, error) {
	return f.registered, nil
}
func (f *workingdocSyncFake) Status(file string) (SyncStatus, error) {
	if f.locks[file] {
		return SyncStatus{Holder: "operator"}, nil
	}
	return SyncStatus{}, nil
}
func (f *workingdocSyncFake) Apply(file, records, hash string) error {
	if f.locks[file] {
		return ErrLocked
	}
	b, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	if fileHash(string(b)) != hash {
		return ErrStale
	}
	raw, err := os.ReadFile(records)
	if err != nil {
		return err
	}
	var recs []editRecord
	if err = json.Unmarshal(raw, &recs); err != nil {
		return err
	}
	next, err := applyRecords(string(b), recs)
	if err != nil {
		return err
	}
	if err = os.WriteFile(file, []byte(next), 0600); err == nil {
		f.stamps++
	}
	return err
}

type workingdocRepositoryFake struct {
	root          string
	blobs         map[settlementPointer][]byte
	current       map[string][]byte
	publishes     int
	baselineReads int
	afterPublish  func()
}

func (f *workingdocRepositoryFake) ValidateTarget(context.Context, string, string) error { return nil }
func (f *workingdocRepositoryFake) Prepare(context.Context, string, string) error        { return nil }

func (f *workingdocRepositoryFake) Mode(context.Context, string) (docPublicationMode, error) {
	return docProduct, nil
}
func (f *workingdocRepositoryFake) CommonDir(context.Context, string) (string, error) {
	return filepath.Join(f.root, ".git"), nil
}
func (f *workingdocRepositoryFake) ReadBaselines(_ context.Context, _ string, ps []settlementPointer) (map[settlementPointer][]byte, error) {
	f.baselineReads++
	out := map[settlementPointer][]byte{}
	for _, p := range ps {
		b, ok := f.blobs[p]
		if !ok {
			return nil, fmt.Errorf("missing baseline %s", p.Commit)
		}
		out[p] = b
	}
	return out, nil
}
func (f *workingdocRepositoryFake) Publish(_ context.Context, _ string, target string, payload []byte, expected string) (string, error) {
	sha := strings.Repeat("a", 40)
	if b, ok := f.current[target]; !ok || string(b) != string(payload) {
		if expected != "" && string(b) != string(f.blobs[settlementPointer{target, expected}]) {
			return "", fmt.Errorf("remote conflict")
		}
		f.publishes++
		sha = fmt.Sprintf("%040x", f.publishes)
		f.current[target] = append([]byte(nil), payload...)
	}
	f.blobs[settlementPointer{target, sha}] = append([]byte(nil), payload...)
	if f.afterPublish != nil {
		f.afterPublish()
	}
	return sha, nil
}
