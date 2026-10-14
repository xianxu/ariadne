package main

import (
	"context"
	"errors"
	"io"
	"os/exec"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestDeclaredSuiteAlwaysRequired(t *testing.T) {
	for _, required := range []string{"", `,"required":false`, `,"required":true`} {
		t.Run(required, func(t *testing.T) {
			decl, err := parseLocalChecks([]byte(`{"version":1,"checks":[{"name":"suite","command":["custom-test"],"weight":2` + required + `}]}`))
			if err != nil {
				t.Fatal(err)
			}
			for _, hasMake := range []bool{false, true} {
				got := effectiveLocalChecks(decl, hasMake, 1)
				c, ok := got.check("suite")
				if !ok || !c.Required || c.Weight != 2 || strings.Join(c.Command, " ") != "custom-test" {
					t.Fatalf("declared suite lost enforcement or overrides: %+v", got)
				}
			}
		})
	}
}

func TestRunCIGateDeclaredOptionalSuite(t *testing.T) {
	for _, required := range []string{"", `,"required":false`} {
		t.Run(required, func(t *testing.T) {
			dir := testfix.Repo(t, testfix.InitialCommit())
			writeRepoFile(t, dir, localChecksPath, `{"version":1,"checks":[{"name":"suite","command":["custom-test"],"weight":2`+required+`}]}`)
			testfix.Git(t, dir, "add", "-A")
			testfix.Git(t, dir, "commit", "-qm", "suite declaration")
			git := func(args ...string) (string, error) {
				out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
				return strings.TrimSpace(string(out)), err
			}
			head, _ := git("rev-parse", "HEAD")
			g := newCIFakeGH()
			g.setHead(3, head, prCheck{Name: "merge-check", Bucket: "pass"})
			ctx := context.Background()
			if err := runCIGate(ctx, g, "o/r", 3, head, head, git, nil, false, io.Discard); err == nil || !strings.Contains(err.Error(), "suite") {
				t.Fatalf("unrecorded suite must refuse: %v", err)
			}
			ran := 0
			run := func(ctx context.Context, c localCheck, sha string) (localRunRecord, error) {
				ran++
				if c.Weight != 2 || c.Command[0] != "custom-test" || !c.Required {
					t.Fatalf("lost overrides: %+v", c)
				}
				err := g.PostStatus(ctx, "o/r", sha, localStatusContext(c.Name), "success", "passed")
				return localRunRecord{Name: c.Name, SHA: sha, Exit: 0}, err
			}
			if err := runCIGate(ctx, g, "o/r", 3, head, head, git, run, false, io.Discard); err != nil || ran != 1 {
				t.Fatalf("suite must run before admission: ran=%d err=%v", ran, err)
			}
		})
	}
}

func TestLocalRecordsOnAncestorsProbeErrors(t *testing.T) {
	want := errors.New("git read failed")
	for _, failing := range []string{"parent", "walk"} {
		t.Run(failing, func(t *testing.T) {
			git := func(args ...string) (string, error) {
				if args[0] == "merge-base" {
					return "base", nil
				}
				if args[0] == "rev-parse" || (args[0] == "rev-list" && strings.Contains(strings.Join(args, " "), "--parents")) {
					if failing == "parent" {
						return "", want
					}
					return "head parent", nil
				}
				return "", want
			}
			_, err := localRecordsOnAncestors(context.Background(), newCIFakeGH(), "o/r", git, "head", "base", []string{"suite"})
			if !errors.Is(err, want) {
				t.Fatalf("%s error swallowed: %v", failing, err)
			}
		})
	}
}
