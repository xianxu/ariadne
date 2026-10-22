package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

type queueHeadReader func(ctx context.Context, heads []string, selfRepo string) ([]queueHeadFact, error)
type queueHeadLimits struct{ Nodes, Bytes, Depth, Batch int }

// Safety ceilings, not latency targets: see #548's recorded initial basis.
var defaultQueueHeadLimits = queueHeadLimits{Nodes: 10000, Bytes: 32 << 20, Depth: 256, Batch: 128}

func gitQueueHeadReader(root string, mainRefs ...string) queueHeadReader {
	mainRef := "refs/remotes/origin/main"
	if len(mainRefs) > 0 {
		mainRef = mainRefs[0]
	}
	main := ""
	ancestry := map[string]bool{}
	limit := defaultQueueHeadLimits.Nodes
	return func(ctx context.Context, heads []string, selfRepo string) ([]queueHeadFact, error) {
		objects, err := gitx.ReadCommitObjects(ctx, root, heads)
		if err != nil {
			return nil, err
		}
		facts := make([]queueHeadFact, 0, len(objects))
		for _, obj := range objects {
			fact, err := parseQueueHeadObject(obj.OID, obj.Content)
			if err != nil {
				return nil, err
			}
			// Exact train authority, including malformed trains, takes precedence.
			if len(fact.Parents) > 1 && parseQueueHead(fact, selfRepo).Kind == queueHeadUnknown {
				if main == "" {
					out, diag, err := gitx.RunGitInContext(ctx, root, queueAncestryEnv(), "--no-replace-objects", "rev-parse", "--verify", mainRef+"^{commit}")
					if err != nil {
						return nil, fmt.Errorf("queue attribution cannot pin publication main: %w (%s)", err, strings.TrimSpace(string(diag)))
					}
					main = strings.TrimSpace(string(out))
					if !landingOIDValid(main) {
						return nil, fmt.Errorf("queue attribution has invalid publication main OID")
					}
				}
				fact.MainParents = true
				for _, parent := range fact.Parents[1:] {
					onMain, ok := ancestry[parent]
					if !ok {
						if len(ancestry) >= limit {
							return nil, fmt.Errorf("queue main ancestry exceeds node safety budget")
						}
						onMain, err = queueMainAncestor(ctx, root, parent, main)
						if err != nil {
							return nil, err
						}
						ancestry[parent] = onMain
					}
					if !onMain {
						fact.MainParents = false
						break
					}
				}
			}
			facts = append(facts, fact)
		}
		return facts, nil
	}
}

func queueAncestryEnv() []string {
	return []string{"GIT_NO_LAZY_FETCH=1", "GIT_GRAFT_FILE=" + os.DevNull, "GIT_SHALLOW_FILE=" + os.DevNull}
}

func queueMainAncestor(ctx context.Context, root, parent, main string) (bool, error) {
	_, diag, err := gitx.RunGitInContext(ctx, root, queueAncestryEnv(), "--no-replace-objects", "merge-base", "--is-ancestor", parent, main)
	if err == nil {
		return true, nil
	}
	var exit *exec.ExitError
	if ctx.Err() == nil && errors.As(err, &exit) && exit.ExitCode() == 1 {
		return false, nil
	}
	return false, fmt.Errorf("queue main ancestry for %s: %w (%s)", parent, err, strings.TrimSpace(string(diag)))
}

func readQueueHeadMemberships(ctx context.Context, heads []string, selfRepo string, read queueHeadReader) (map[string]queueHeadMembership, error) {
	return readQueueHeadMembershipsWithLimits(ctx, heads, selfRepo, read, defaultQueueHeadLimits)
}

// Discovery is a worklist: each node is activated once, and can pause once to
// read its parent fact. Train bases need classification, but never traversal.
func readQueueHeadMembershipsWithLimits(ctx context.Context, heads []string, selfRepo string, read queueHeadReader, limits queueHeadLimits) (map[string]queueHeadMembership, error) {
	budget := func(kind string) error { return fmt.Errorf("queue head attribution exceeds %s safety budget", kind) }
	if limits.Nodes <= 0 || limits.Bytes <= 0 || limits.Depth < 0 || limits.Batch <= 0 || limits.Batch > 128 {
		return nil, fmt.Errorf("invalid queue head attribution safety budgets")
	}
	steps := map[string]queueHeadStep{}
	active := map[string]bool{}
	dependencies := map[string]string{}
	waiting := map[string][]string{}
	requested := map[string]bool{}
	discoveryDepth := map[string]int{}
	var fetch, work []string
	request := func(head string, depth int) error {
		if !landingOIDValid(head) {
			return fmt.Errorf("invalid queue head identity %q", head)
		}
		if requested[head] {
			return nil
		}
		if depth > limits.Depth {
			return budget("depth")
		}
		if len(requested) >= limits.Nodes {
			return budget("node")
		}
		requested[head] = true
		discoveryDepth[head] = depth
		fetch = append(fetch, head)
		return nil
	}
	activate := func(head string) {
		if !active[head] {
			active[head] = true
			work = append(work, head)
		}
	}
	for _, head := range heads {
		if err := request(head, 0); err != nil {
			return nil, err
		}
		activate(head)
	}
	retained := 0
	for len(fetch) > 0 || len(work) > 0 {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if len(fetch) > 0 {
			n := min(len(fetch), limits.Batch)
			batch := append([]string(nil), fetch[:n]...)
			fetch = fetch[n:]
			got, err := read(ctx, batch, selfRepo)
			if err != nil {
				return nil, err
			}
			if len(got) != len(batch) {
				return nil, fmt.Errorf("incomplete queue head batch")
			}
			for i, fact := range got {
				if fact.OID != batch[i] || fact.Bytes < 0 {
					return nil, fmt.Errorf("unexpected queue head object identity or size")
				}
				if fact.Bytes > 1<<20 || fact.Bytes > limits.Bytes-retained {
					return nil, budget("byte")
				}
				retained += fact.Bytes
				steps[fact.OID] = parseQueueHead(fact, selfRepo)
				work = append(work, waiting[fact.OID]...)
				delete(waiting, fact.OID)
			}
		}
		for len(work) > 0 {
			head := work[0]
			work = work[1:]
			step, ok := steps[head]
			if !ok {
				waiting[head] = append(waiting[head], head)
				continue
			}
			if step.NeedParent == "" {
				continue
			}
			parent, ok := steps[step.NeedParent]
			if !ok {
				if err := request(step.NeedParent, discoveryDepth[head]+1); err != nil {
					return nil, err
				}
				waiting[step.NeedParent] = append(waiting[step.NeedParent], head)
				continue
			}
			if queueHeadNeedsParentMembership(step, parent) {
				dependencies[head] = step.NeedParent
				activate(step.NeedParent)
			}
		}
	}
	resolved := map[string]queueHeadMembership{}
	depths := map[string]int{}
	visiting := map[string]bool{}
	var resolve func(string) (queueHeadMembership, error)
	resolve = func(head string) (queueHeadMembership, error) {
		if err := ctx.Err(); err != nil {
			return queueHeadMembership{}, err
		}
		if result, ok := resolved[head]; ok {
			return result, nil
		}
		if visiting[head] {
			return queueHeadMembership{}, fmt.Errorf("cycle in queue head attribution")
		}
		visiting[head] = true
		step := steps[head]
		var parentStep *queueHeadStep
		var parentMembership *queueHeadMembership
		depth := 0
		if step.NeedParent != "" {
			parent := steps[step.NeedParent]
			parentStep = &parent
			depth = 1
		}
		if parent := dependencies[head]; parent != "" {
			result, err := resolve(parent)
			if err != nil {
				return queueHeadMembership{}, err
			}
			parentMembership = &result
			depth = depths[parent] + 1
		}
		if depth > limits.Depth {
			return queueHeadMembership{}, budget("depth")
		}
		result := selectQueueHeadMembership(step, parentStep, parentMembership)
		depths[head] = depth
		resolved[head] = result
		delete(visiting, head)
		return result, nil
	}
	out := map[string]queueHeadMembership{}
	for _, head := range heads {
		result, err := resolve(head)
		if err != nil {
			return nil, err
		}
		out[head] = result
	}
	return out, nil
}
