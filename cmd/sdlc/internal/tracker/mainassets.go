package tracker

import (
	"bytes"
	"fmt"
	"path"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issueassets"
)

// AssetBundleFromView reads one bounded issue inventory; Markdown is its index.
func AssetBundleFromView(view *gitx.TrunkView, detailRel, id string, body []byte) (issueassets.Bundle, error) {
	home := path.Dir(detailRel)
	assets := path.Join(home, "assets")
	dir := path.Join(assets, id)
	// Include ancestor entries: a symlink replacing assets/ or its ID folder
	// is a blob at that exact path and is not listed by a trailing-slash query.
	files, err := view.FilesSelected(home, func(f gitx.TreeFile) bool {
		return f.Path == assets || f.Path == dir || strings.HasPrefix(f.Path, dir+"/")
	})
	if err != nil {
		return issueassets.Bundle{}, fmt.Errorf("cannot read asset bundle %s: %w; restore this bundle from its last valid commit before retrying", dir, err)
	}
	inventory := map[string][]byte{}
	for _, f := range files {
		if f.Path == assets || f.Path == dir {
			return issueassets.Bundle{}, fmt.Errorf("asset directory %s is a file or symlink: replace it with a regular directory", f.Path)
		}
		if f.Mode != "100644" {
			return issueassets.Bundle{}, fmt.Errorf("asset %s is not an ordinary file: restore a regular asset file", f.Path)
		}
		inventory[strings.TrimPrefix(f.Path, path.Dir(detailRel)+"/")] = f.Content
	}
	return issueassets.BuildBundle(id, assetBody(body), inventory)
}
func assetBody(raw []byte) []byte {
	if _, body, err := issue.Parse(string(raw)); err == nil {
		return []byte(body)
	}
	return raw
}

// pinnedDetail validates immutable source pins before any publication or replay.
func pinnedDetail(view *gitx.TrunkView, spec ReceiptSpec) ([]byte, issueassets.Bundle, error) {
	detail, err := view.ReadBlob(spec.SourceBlob)
	if err != nil {
		return nil, issueassets.Bundle{}, err
	}
	files := map[string][]byte{}
	total := 0
	for _, entry := range spec.AssetEntries() {
		b, err := view.ReadBlob(entry.Blob)
		if err != nil {
			return nil, issueassets.Bundle{}, err
		}
		total += len(b)
		if total > issueassets.MaxBundleBytes {
			return nil, issueassets.Bundle{}, fmt.Errorf("pinned asset bundle exceeds 8 MiB")
		}
		files[strings.TrimPrefix(entry.Path, path.Dir(spec.SourcePath)+"/")] = b
	}
	bundle, err := issueassets.BuildBundle(spec.IssueID, assetBody(detail), files)
	return detail, bundle, err
}

func validateMainCandidate(main *gitx.TrunkFile, spec ReceiptSpec, candidate string) error {
	view := main.ViewOf(candidate)
	detail, want, err := pinnedDetail(view, spec)
	if err != nil {
		return err
	}
	docs, err := view.FilesSelected(path.Dir(spec.DestinationPath), func(f gitx.TreeFile) bool { return f.Path == spec.DestinationPath })
	if err != nil {
		return err
	}
	if len(docs) != 1 || docs[0].Mode != "100644" || !bytes.Equal(docs[0].Content, detail) {
		return fmt.Errorf("candidate details do not match the pinned source body")
	}
	got, err := AssetBundleFromView(view, spec.DestinationPath, spec.IssueID, docs[0].Content)
	if err != nil {
		return err
	}
	if len(got.Files) != len(want.Files) {
		return fmt.Errorf("candidate asset members do not match the pinned source bundle")
	}
	for p, b := range want.Files {
		if actual, ok := got.Files[p]; !ok || !bytes.Equal(actual, b) {
			return fmt.Errorf("candidate asset %s does not match the pinned source bundle", p)
		}
	}
	return nil
}

func checkedMainPublisher(main *gitx.TrunkFile, root string, spec ReceiptSpec, candidate string, before func(ReceiptSpec, *gitx.TrunkView) error) (mainPublisher, error) {
	if err := validateMainCandidate(main, spec, candidate); err != nil {
		return mainPublisher{}, blockedMainReceipt(root, spec, candidate, err)
	}
	return mainPublisher{t: main, spec: spec, before: before}, nil
}

func blockedMainReceipt(root string, spec ReceiptSpec, candidate string, err error) error {
	return fmt.Errorf("issue #%s operation %s candidate %s blocked: %w; saved receipt %s at refs/sdlc/recovery/%s:document is retained; inspect with sdlc issue recovery list and ask the TL/hub to repair this legacy receipt (ariadne#352), then retry sdlc issue recovery reconcile --issue %s", spec.IssueID, spec.Token, candidate, err, root, spec.Token, issue.CLIRef(spec.IssueID))
}
