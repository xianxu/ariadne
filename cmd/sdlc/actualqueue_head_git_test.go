package main

import (
	"context"
	"fmt"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"strings"
	"testing"
)

func queueGraphOID(i int) string { return fmt.Sprintf("%040x", i+1) }
func queueGraphFact(i int, message string, parents ...string) queueHeadFact {
	return queueHeadFact{OID: queueGraphOID(i), Message: message, Parents: parents, Bytes: len(message) + 200}
}

type queueGraphFake struct {
	facts map[string]queueHeadFact
	reads map[string]int
	calls int
}

func (f *queueGraphFake) read(ctx context.Context, heads []string, _ string) ([]queueHeadFact, error) {
	f.calls++
	var out []queueHeadFact
	for _, head := range heads {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		f.reads[head]++
		fact, ok := f.facts[head]
		if !ok {
			return nil, fmt.Errorf("missing object %s", head)
		}
		out = append(out, fact)
	}
	return out, nil
}
func TestReadQueueHeadMembershipsCountsSharedGraphOnce(t *testing.T) {
	for _, n := range []int{8, 16} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			f := &queueGraphFake{facts: map[string]queueHeadFact{}, reads: map[string]int{}}
			f.facts[queueGraphOID(0)] = queueGraphFact(0, "#548: root")
			f.facts[queueGraphOID(1)] = queueGraphFact(1, "issue-sync: update issues", queueGraphOID(0))
			var heads []string
			for i := 2; i < n+2; i++ {
				f.facts[queueGraphOID(i)] = queueGraphFact(i, "issue-sync: update issues", queueGraphOID(1))
				heads = append(heads, queueGraphOID(i), queueGraphOID(i))
			}
			got, err := readQueueHeadMemberships(context.Background(), heads, "ariadne", f.read)
			if err != nil || len(got) != n {
				t.Fatalf("memberships=%+v err=%v", got, err)
			}
			for head, count := range f.reads {
				if count != 1 {
					t.Fatalf("%s read %d times", head, count)
				}
			}
			if len(f.reads) != n+2 || f.calls > 3 {
				t.Fatalf("reads=%d calls=%d n=%d", len(f.reads), f.calls, n)
			}
		})
	}
}
func TestReadQueueHeadMembershipsSafetyBudgets(t *testing.T) {
	for _, kind := range []string{"node", "byte", "depth"} {
		t.Run(kind, func(t *testing.T) {
			f := &queueGraphFake{facts: map[string]queueHeadFact{}, reads: map[string]int{}}
			for i := 0; i < 4; i++ {
				msg := "issue-sync: update issues"
				var parents []string
				if i == 0 {
					msg = "#548: root"
				} else {
					parents = []string{queueGraphOID(i - 1)}
				}
				f.facts[queueGraphOID(i)] = queueGraphFact(i, msg, parents...)
			}
			limits := queueHeadLimits{Nodes: 4, Bytes: 10000, Depth: 3, Batch: 1}
			if _, err := readQueueHeadMembershipsWithLimits(context.Background(), []string{queueGraphOID(3)}, "ariadne", f.read, limits); err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "node":
				limits.Nodes = 3
			case "byte":
				limits.Bytes = 600
			case "depth":
				limits.Depth = 2
			}
			f.reads = map[string]int{}
			got, err := readQueueHeadMembershipsWithLimits(context.Background(), []string{queueGraphOID(3), queueGraphOID(0)}, "ariadne", f.read, limits)
			if err == nil || !strings.Contains(err.Error(), kind+" safety budget") || got != nil {
				t.Fatalf("partial budget proof=%+v err=%v", got, err)
			}
		})
	}
}
func TestReadQueueHeadMembershipsMissingAndCancelled(t *testing.T) {
	f := &queueGraphFake{facts: map[string]queueHeadFact{}, reads: map[string]int{}}
	if _, err := readQueueHeadMemberships(context.Background(), []string{queueGraphOID(0)}, "ariadne", f.read); err == nil {
		t.Fatal("missing object accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := readQueueHeadMemberships(ctx, []string{queueGraphOID(0)}, "ariadne", f.read); err == nil {
		t.Fatal("cancelled reader accepted")
	}
}

func TestReadQueueHeadMembershipsDepthBoundsDiscovery(t *testing.T) {
	f := &queueGraphFake{facts: map[string]queueHeadFact{}, reads: map[string]int{}}
	for i := 0; i < 1000; i++ {
		msg := "issue-sync: update issues"
		var parents []string
		if i == 0 {
			msg = "#548: root"
		} else {
			parents = []string{queueGraphOID(i - 1)}
		}
		f.facts[queueGraphOID(i)] = queueGraphFact(i, msg, parents...)
	}
	_, err := readQueueHeadMemberships(context.Background(), []string{queueGraphOID(999)}, "ariadne", f.read)
	if err == nil || !strings.Contains(err.Error(), "depth safety budget") {
		t.Fatalf("expected depth refusal: %v", err)
	}
	if len(f.reads) > 257 {
		t.Fatalf("depth ceiling did not bound discovery: read %d nodes", len(f.reads))
	}
}

func TestReadQueueHeadMembershipsRealGitTrainAndIntegration(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	base := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	tree := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD^{tree}"))
	commit := func(message string, parents ...string) string {
		args := []string{"commit-tree", tree, "-m", message}
		for _, parent := range parents {
			args = append(args, "-p", parent)
		}
		return strings.TrimSpace(testfix.Capture(t, root, args...))
	}
	car1 := commit("#1: first car", base)
	car2 := commit("#2: second car", base)
	main := commit("#99: unrelated main", base)
	sync := commit("issue-sync: update issues", main)
	first := commit("Landing train 213: #1 (PR #101)\n\nLanding-Train: 213\nLanding-PR: 101", sync, car1)
	second := commit("Landing train 213: #2 (PR #102)\n\nLanding-Train: 213\nLanding-PR: 102", first, car2)
	main2 := commit("#98: advanced main", main)
	train := commit(fmt.Sprintf("Merge %s into landing train 213 (sdlc merge, #449)\n\nIntegrated-Main: %s\nLanding-Train: 213", main2, main2), second, main2)
	own := commit(fmt.Sprintf("Merge %s into issue-one (sdlc merge, #332)\n\nIntegrated-Main: %s", main2, main2), car1, main2)
	got, err := readQueueHeadMemberships(context.Background(), []string{train, first, own}, "ariadne", gitQueueHeadReader(root))
	if err != nil {
		t.Fatal(err)
	}
	if !got[train].Known || fmt.Sprint(got[train].Issues) != "[2 1]" || got[train].Train != 213 {
		t.Fatalf("train borrowed held/main cars: %+v", got[train])
	}
	if !got[first].Known || fmt.Sprint(got[first].Issues) != "[1]" {
		t.Fatalf("prefix borrowed base: %+v", got[first])
	}
	if !got[own].Known || fmt.Sprint(got[own].Issues) != "[1]" || got[own].Train != 0 {
		t.Fatalf("integration borrowed second parent: %+v", got[own])
	}
}
