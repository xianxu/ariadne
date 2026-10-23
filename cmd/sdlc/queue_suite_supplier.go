package main

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/pkg/workspace"
)

type suiteQueueSupplier struct {
	Checkout, Revision, Reason string
	Dirty                      bool
}
type suiteSupplierIO struct {
	resolve func(context.Context, string) (string, error)
	read    func(context.Context, string) suiteQueueSupplier
}

func suiteTicketLegacy(raw json.RawMessage) (bool, string) {
	r, err := readSuiteTicket(raw)
	return err != nil || r.Policy != suiteOrderingPolicy, r.Worktree
}

// annotateSuiteSuppliers is status-only. Cache both owner resolution and supplier
// reads per call; admission, grant and frozen JSON never invoke these probes.
func annotateSuiteSuppliers(ctx context.Context, observed suiteQueueObservation, io suiteSupplierIO) suiteQueueObservation {
	observed.Suppliers = map[int64]suiteQueueSupplier{}
	owners := map[string]suiteQueueSupplier{}
	suppliers := map[string]suiteQueueSupplier{}
	for _, ticket := range observed.Tickets {
		legacy, owner := suiteTicketLegacy(ticket.Record)
		if !legacy {
			continue
		}
		if prior, ok := owners[owner]; ok {
			observed.Suppliers[ticket.Number] = prior
			continue
		}
		item := suiteQueueSupplier{Reason: "ticket worktree unavailable or invalid"}
		switch {
		case ctx.Err() != nil:
			item.Reason = "status deadline exhausted"
		case !filepath.IsAbs(owner) || filepath.Clean(owner) != owner || len(owner) > 4096 || strings.ContainsAny(owner, "\x00\r\n"):
		default:
			supplier, err := io.resolve(ctx, owner)
			if err != nil {
				item.Reason = "worktree/workflow identity unavailable"
			} else if prior, ok := suppliers[supplier]; ok {
				item = prior
			} else {
				item = io.read(ctx, supplier)
				suppliers[supplier] = item
			}
		}
		owners[owner] = item
		observed.Suppliers[ticket.Number] = item
	}
	return observed
}

type suiteSupplierGit struct{ context.Context }

func (g suiteSupplierGit) GitInDir(dir string, args ...string) ([]byte, error) {
	out, _, err := gitx.RunGitBoundedContext(g.Context, dir, nil, 64<<10, args...)
	return out, err
}

func resolveQueueSupplier(ctx context.Context, owner string) (string, error) {
	g := suiteSupplierGit{ctx}
	v, err := workspace.NormalizeVantage(g, owner)
	if err != nil {
		return "", err
	}
	real, err := filepath.EvalSymlinks(owner)
	if err != nil || real != owner || real != v.WorktreeRoot {
		return "", fmt.Errorf("ticket does not name its exact checkout")
	}
	helper, err := workspace.SuiteHelper(owner, "")
	if err != nil {
		return "", err
	}
	return filepath.Dir(filepath.Dir(helper)), nil
}

func readQueueSupplier(ctx context.Context, supplier string) suiteQueueSupplier {
	item := suiteQueueSupplier{Checkout: supplier, Reason: "supplying Git revision/state unavailable"}
	g := suiteSupplierGit{ctx}
	top, err := g.GitInDir(supplier, "rev-parse", "--show-toplevel")
	if err != nil || strings.TrimSpace(string(top)) != supplier {
		return item
	}
	head, err := g.GitInDir(supplier, "rev-parse", "--verify", "HEAD^{commit}")
	if err != nil {
		return item
	}
	sha := strings.TrimSuffix(string(head), "\n")
	if (len(sha) != 40 && len(sha) != 64) || !isResolvedSHA(sha) {
		return item
	}
	status, err := g.GitInDir(supplier, "status", "--porcelain=v1", "-z", "--untracked-files=all", "--ignore-submodules=none")
	if err != nil {
		return item
	}
	after, err := g.GitInDir(supplier, "rev-parse", "--verify", "HEAD^{commit}")
	if err != nil || string(after) != string(head) {
		return item
	}
	item.Revision, item.Dirty, item.Reason = sha, len(status) != 0, ""
	return item
}

func (s suiteQueueSupplier) detail() string {
	if s.Reason != "" {
		return fmt.Sprintf("current supplying checkout %q, revision unavailable (%s); launch revision unknown", s.Checkout, s.Reason)
	}
	return fmt.Sprintf("current supplying checkout %q, revision %s, dirty=%t; launch revision unknown", s.Checkout, s.Revision, s.Dirty)
}
