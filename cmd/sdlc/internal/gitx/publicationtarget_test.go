package gitx

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestPublicationTargetUsesExplicitUpstreamAndPortableIdentity(t *testing.T) {
	repo, remote := trunkFixture(t, "source")
	testfix.Git(t, repo, "remote", "rename", "origin", "publishing")
	testfix.Git(t, repo, "config", "branch.main.remote", "publishing")
	testfix.Git(t, repo, "config", "branch.main.merge", "refs/heads/main")
	before := testfix.Capture(t, repo, "rev-parse", "HEAD")
	target, err := ResolvePublicationTarget(context.Background(), repo, "main")
	if err != nil {
		t.Fatal(err)
	}
	wantRemote, err := filepath.EvalSymlinks(remote)
	if err != nil {
		t.Fatal(err)
	}
	if target.Remote != "publishing" || target.Repository != "file:"+wantRemote || target.FetchURL != remote || target.PushURL != remote {
		t.Fatalf("wrong publication identity: %+v", target)
	}
	clone := filepath.Join(t.TempDir(), "clone")
	testfix.Git(t, "", "clone", "--quiet", "--origin", "elsewhere", remote, clone)
	other, err := ResolvePublicationTarget(context.Background(), clone, "main")
	if err != nil {
		t.Fatal(err)
	}
	if other.Repository != target.Repository {
		t.Fatalf("identity is checkout-dependent: %+v %+v", target, other)
	}
	if got := testfix.Capture(t, repo, "rev-parse", "HEAD"); got != before {
		t.Fatal("target read moved HEAD")
	}
}

func TestPublicationTargetRejectsAmbiguousOrDifferentDestinations(t *testing.T) {
	for _, name := range []string{"missing", "multiple-remotes", "local-upstream", "other-branch", "multiple-fetch", "multiple-push", "different-push", "unsafe-branch"} {
		t.Run(name, func(t *testing.T) {
			repo, _ := trunkFixture(t, "source")
			testfix.Git(t, repo, "config", "branch.main.remote", "origin")
			testfix.Git(t, repo, "config", "branch.main.merge", "refs/heads/main")
			branch := "main"
			switch name {
			case "missing":
				testfix.Git(t, repo, "config", "--unset-all", "branch.main.remote")
			case "multiple-remotes":
				testfix.Git(t, repo, "config", "--add", "branch.main.remote", "other")
			case "local-upstream":
				testfix.Git(t, repo, "config", "branch.main.remote", ".")
			case "other-branch":
				testfix.Git(t, repo, "config", "branch.main.merge", "refs/heads/develop")
			case "multiple-fetch":
				testfix.Git(t, repo, "config", "--add", "remote.origin.url", "https://github.com/example/other")
			case "multiple-push":
				testfix.Git(t, repo, "config", "--add", "remote.origin.pushurl", "https://github.com/example/one")
				testfix.Git(t, repo, "config", "--add", "remote.origin.pushurl", "https://github.com/example/two")
			case "different-push":
				testfix.Git(t, repo, "config", "remote.origin.pushurl", "https://github.com/example/other")
			case "unsafe-branch":
				branch = "main\nother"
			}
			if got, err := ResolvePublicationTarget(context.Background(), repo, branch); err == nil {
				t.Fatalf("accepted ambiguous target %+v", got)
			}
		})
	}
}

func TestPublicationTargetUsesEffectiveRewrittenURLs(t *testing.T) {
	repo, remote := trunkFixture(t, "source")
	testfix.Git(t, repo, "config", "branch.main.remote", "origin")
	testfix.Git(t, repo, "config", "branch.main.merge", "refs/heads/main")
	testfix.Git(t, repo, "config", "remote.origin.url", "https://github.com/example/repo.git")
	testfix.Git(t, repo, "config", "url."+remote+".insteadOf", "https://github.com/example/repo.git")
	target, err := ResolvePublicationTarget(context.Background(), repo, "main")
	if err != nil || target.FetchURL != remote {
		t.Fatalf("effective target %+v: %v", target, err)
	}
}

func TestPublicationTargetGitHubTransportIdentity(t *testing.T) {
	repo, _ := trunkFixture(t, "source")
	testfix.Git(t, repo, "config", "branch.main.remote", "origin")
	testfix.Git(t, repo, "config", "branch.main.merge", "refs/heads/main")
	testfix.Git(t, repo, "config", "remote.origin.url", "https://github.com/Example/Repo.git")
	testfix.Git(t, repo, "config", "remote.origin.pushurl", "git@github.com:example/repo.git")
	target, err := ResolvePublicationTarget(context.Background(), repo, "main")
	if err != nil || target.Repository != "github.com/example/repo" {
		t.Fatalf("portable identity %+v: %v", target, err)
	}
}

func TestPublicationTargetCancellationAndMissingContext(t *testing.T) {
	repo, _ := trunkFixture(t, "source")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := ResolvePublicationTarget(ctx, repo, "main"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled read: %v", err)
	}
	if _, err := ResolvePublicationTarget(nil, repo, "main"); err == nil {
		t.Fatal("accepted nil context")
	}
}
