package reviewpolicy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// ConfigPath is a repo's review configuration (#310), repo-local like
// workshop/issue-tracker.json: the base layer never carries it.
const ConfigPath = "workshop/review.json"

// config is the file's shape.
type config struct {
	BoundaryReviewer Mode                   `json:"boundary_reviewer"`
	ReviewerPairs    map[Agent]Agent        `json:"reviewer_pairs"`
	ReviewerOrder    map[ChangeKind][]Agent `json:"reviewer_order"`  // #409: preference per kind under `other`
	ReviewerModels   map[Agent]string       `json:"reviewer_models"` // #409: the model each reviewer is asked for
}

// Load reads root's reviewer policy. An absent file is DefaultMode; an
// unreadable or invalid one is an error naming the file.
func Load(root string) (Policy, error) {
	path := filepath.Join(root, filepath.FromSlash(ConfigPath))
	raw, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Policy{}, nil
	}
	if err != nil {
		return Policy{}, fmt.Errorf("read %s: %w", ConfigPath, err)
	}
	var c config
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&c); err != nil {
		return Policy{}, fmt.Errorf("parse %s: %w", ConfigPath, err)
	}
	p := Policy{Mode: c.BoundaryReviewer, Pairs: c.ReviewerPairs, Orders: c.ReviewerOrder, Models: c.ReviewerModels}
	if err := p.Validate(); err != nil {
		return Policy{}, fmt.Errorf("%s: %w", ConfigPath, err)
	}
	return p, nil
}
