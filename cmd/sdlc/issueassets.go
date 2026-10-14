package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issueassets"
)

// assetDirectories validates every existing child beneath the repository root.
// Creation is an effect and is used only after attachment preflight succeeds.
func assetDirectories(root, rel string, create bool) error {
	if path.IsAbs(rel) || path.Clean(rel) != rel || strings.HasPrefix(rel, "../") || strings.Contains(rel, "\\") {
		return fmt.Errorf("unsafe asset directory %q; use the issue's repository directory", rel)
	}
	current := root
	for _, part := range strings.Split(rel, "/") {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if os.IsNotExist(err) && create {
			if err = os.Mkdir(current, 0755); err != nil && !os.IsExist(err) {
				return err
			}
			info, err = os.Lstat(current)
		}
		if err != nil {
			return err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("unsafe directory %s; replace it with a regular directory", current)
		}
	}
	return nil
}

func readAssetFiles(root, detailRel, id string) (map[string][]byte, error) {
	dir := path.Join(path.Dir(filepath.ToSlash(detailRel)), "assets", id)
	files := map[string][]byte{}
	if err := assetDirectories(root, dir, false); os.IsNotExist(err) {
		return files, nil
	} else if err != nil {
		return nil, err
	}
	r, err := os.OpenRoot(filepath.Join(root, filepath.FromSlash(dir)))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	d, err := r.Open(".")
	if err != nil {
		return nil, err
	}
	defer d.Close()
	entries, err := d.ReadDir(issueassets.MaxFiles + 1)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("read asset directory %s: %w; repair that directory before retrying", dir, err)
	}
	if len(entries) > issueassets.MaxFiles {
		return nil, fmt.Errorf("issue %s has more than 32 assets; remove an asset and its body link", id)
	}
	total := 0
	for _, entry := range entries {
		rel := path.Join(dir, entry.Name())
		raw, err := readAssetRegular(root, filepath.FromSlash(rel), issueassets.MaxFileBytes)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w; repair that asset before retrying", rel, err)
		}
		total += len(raw)
		if total > issueassets.MaxBundleBytes {
			return nil, fmt.Errorf("issue %s assets exceed 8 MiB; remove an asset and its body link", id)
		}
		files[path.Join("assets", id, entry.Name())] = raw
	}
	return files, nil
}

// attachmentStore is the effect boundary. Tests wrap the real portable folder
// store to inject failures without replacing the transition or its observations.
type attachmentStore interface {
	Install(string, []byte) (bool, error)
	LinkBody([]byte, []byte) error
	RemoveOwned(string, []byte) error
	Body() ([]byte, error)
}
type fileAttachmentStore struct{ root, detailRel, id string }

// Interrupted body staging has no durable ownership proof. Report it rather
// than deleting another request's bytes or silently leaving evidence behind.
func refuseAttachmentTemps(root, dir string) error {
	if err := assetDirectories(root, dir, false); err != nil {
		return fmt.Errorf("read issue directory: %w; repair that directory before attaching", err)
	}
	r, err := os.OpenRoot(root)
	if err != nil {
		return err
	}
	defer r.Close()
	d, err := r.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	for {
		entries, readErr := d.ReadDir(128)
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".sdlc-attach-") {
				return fmt.Errorf("interrupted attachment temporary %s; remove that temporary before retrying", path.Join(dir, entry.Name()))
			}
		}
		if readErr == io.EOF {
			return nil
		}
		if readErr != nil {
			return fmt.Errorf("read issue directory: %w; repair that directory before attaching", readErr)
		}
	}
}

func (s fileAttachmentStore) Body() ([]byte, error) {
	return readAssetRegular(s.root, filepath.FromSlash(s.detailRel), 1<<20)
}
func (s fileAttachmentStore) Install(rel string, data []byte) (bool, error) {
	dest := path.Join(path.Dir(s.detailRel), rel)
	if err := assetDirectories(s.root, path.Dir(dest), true); err != nil {
		return false, err
	}
	if existing, err := readAssetRegular(s.root, filepath.FromSlash(dest), issueassets.MaxFileBytes); err == nil {
		if bytes.Equal(existing, data) {
			return false, nil
		}
		return false, fmt.Errorf("asset collision at %s; restore the original bytes", dest)
	} else if !os.IsNotExist(err) {
		return false, err
	}
	r, err := os.OpenRoot(s.root)
	if err != nil {
		return false, err
	}
	defer r.Close()
	temp := path.Join(path.Dir(dest), ".sdlc-attach-"+operationToken("bytes"))
	if err = writeAttachTemp(r, temp, data); err != nil {
		return false, err
	}
	defer r.Remove(temp)
	if err = r.Link(temp, dest); err != nil {
		if os.IsExist(err) {
			current, e := readAssetRegular(s.root, filepath.FromSlash(dest), issueassets.MaxFileBytes)
			if e == nil && bytes.Equal(current, data) {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}
func writeAttachTemp(r *os.Root, name string, data []byte) error {
	f, err := r.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	_, writeErr := f.Write(data)
	syncErr := f.Sync()
	closeErr := f.Close()
	if writeErr != nil {
		_ = r.Remove(name)
		return writeErr
	}
	if syncErr != nil {
		_ = r.Remove(name)
		return syncErr
	}
	if closeErr != nil {
		_ = r.Remove(name)
	}
	return closeErr
}
func (s fileAttachmentStore) LinkBody(expected, next []byte) error {
	current, err := s.Body()
	if err != nil {
		return err
	}
	if !bytes.Equal(current, expected) {
		return fmt.Errorf("issue body changed; rerun attachment using the current body")
	}
	r, err := os.OpenRoot(s.root)
	if err != nil {
		return err
	}
	defer r.Close()
	temp := path.Join(path.Dir(s.detailRel), ".sdlc-attach-"+operationToken("body"))
	if err = writeAttachTemp(r, temp, next); err != nil {
		return err
	}
	defer r.Remove(temp)
	current, err = s.Body()
	if err != nil {
		return err
	}
	if !bytes.Equal(current, expected) {
		return fmt.Errorf("issue body changed; rerun attachment using the current body")
	}
	return r.Rename(temp, s.detailRel)
}
func (s fileAttachmentStore) RemoveOwned(rel string, data []byte) error {
	dest := path.Join(path.Dir(s.detailRel), rel)
	current, err := readAssetRegular(s.root, filepath.FromSlash(dest), issueassets.MaxFileBytes)
	if err != nil {
		return err
	}
	if !bytes.Equal(current, data) {
		return fmt.Errorf("asset %s changed; reconcile it before retrying", dest)
	}
	body, err := s.Body()
	if err != nil {
		return err
	}
	refs, err := issueassets.ParseRefs(s.id, issueAssetBody(body))
	if err != nil {
		return err
	}
	for _, ref := range refs {
		if ref == rel {
			return fmt.Errorf("asset %s is now referenced; rerun attachment to reconcile", dest)
		}
	}
	r, err := os.OpenRoot(s.root)
	if err != nil {
		return err
	}
	defer r.Close()
	return r.Remove(dest)
}

func applyAttachment(store attachmentStore, state issueassets.AttachState, next []byte) error {
	advance := func(event issueassets.AttachEvent) (issueassets.Effect, error) {
		var effect issueassets.Effect
		var err error
		state, effect, err = issueassets.NextAttach(state, event)
		return effect, err
	}
	effect, err := advance(issueassets.Retry)
	if err != nil {
		return err
	}
	if effect != issueassets.InstallNoReplace {
		return fmt.Errorf("attachment state is not prepared; rerun attachment")
	}
	created, err := store.Install(state.Destination(), state.Source())
	if err != nil {
		_, _ = advance(issueassets.InstallFailed)
		return fmt.Errorf("asset install failed: %w; rerun attachment", err)
	}
	event := issueassets.InstallExisting
	if created {
		event = issueassets.InstallNew
	}
	if _, err = advance(event); err != nil {
		return err
	}
	err = store.LinkBody(state.ExpectedBody(), next)
	if err == nil {
		_, err = advance(issueassets.LinkSucceeded)
		return err
	}
	// A lost/ambiguous response is reconciled against exact durable bytes before
	// rollback: successful body replacement must never cause evidence deletion.
	if current, readErr := store.Body(); readErr == nil && bytes.Equal(current, next) {
		_, err = advance(issueassets.LinkSucceeded)
		return err
	}
	effect, transitionErr := advance(issueassets.LinkFailed)
	if transitionErr != nil {
		return transitionErr
	}
	if effect == issueassets.RemoveOwnedBytes {
		if cleanupErr := store.RemoveOwned(state.Destination(), state.Source()); cleanupErr == nil {
			_, _ = advance(issueassets.RemovalConfirmed)
		} else {
			_, _ = advance(issueassets.RemovalFailed)
			return fmt.Errorf("body update failed (%v); retained %s (%v); reconcile that asset then rerun attachment", err, state.Destination(), cleanupErr)
		}
	}
	return fmt.Errorf("body update failed: %w; rerun attachment with the same source to reconcile %s", err, state.Destination())
}
