package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestQueueHeadQualifiedAuthorityContext(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	base := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	tree := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD^{tree}"))
	commit := func(message string, parents ...string) string {
		args := []string{"commit-tree", tree, "-m", message}
		for _, p := range parents {
			args = append(args, "-p", p)
		}
		return strings.TrimSpace(testfix.Capture(t, root, args...))
	}
	own := commit("#10: own", base)
	main := commit("#99: main", base)
	foreign := commit("#20: foreign", base)
	testfix.Git(t, root, "update-ref", "refs/remotes/origin/main", main)
	reader := gitQueueHeadReader(root)
	for _, repo := range []string{"ariadne", "echo", "repo.with-dash"} {
		for _, zero := range []string{"0", "00", "000000"} {
			ref := repo + "#" + zero
			for i, subject := range []string{ref, ref + ", #10", "#10, " + ref, "#10, " + ref + ", #11", "#10, #11, " + ref} {
				t.Run(fmt.Sprintf("%s/%s/%d", repo, zero, i), func(t *testing.T) {
					head := commit(subject+": invalid authority", own, main)
					got, err := readQueueHeadMemberships(t.Context(), []string{head}, repo, reader)
					if err != nil || !got[head].Known || fmt.Sprint(got[head].Issues) != "[10]" {
						t.Fatalf("proof selection disagreed with %q authority for %q: %+v err=%v", repo, subject, got, err)
					}
				})
			}
		}
		for _, tc := range []struct{ subject, want string }{
			{repo + "#20: definite issue", "[20]"},
			{"other#0: foreign authority", "[]"},
			{"#10, other#0: local and foreign authority", "[10]"},
		} {
			head := commit(tc.subject, own, main)
			got, err := readQueueHeadMemberships(t.Context(), []string{head}, repo, reader)
			if err != nil || !got[head].Known || fmt.Sprint(got[head].Issues) != tc.want {
				t.Fatalf("definite authority changed for %s: %+v err=%v", tc.subject, got, err)
			}
		}
		head := commit(repo+"#0: foreign integration", own, foreign)
		got, err := readQueueHeadMemberships(t.Context(), []string{head}, repo, reader)
		if err != nil || got[head].Known {
			t.Fatalf("invalid lead excused a foreign parent: %+v err=%v", got, err)
		}
	}
}
