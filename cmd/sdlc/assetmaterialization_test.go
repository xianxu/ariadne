package main

import (
	"context"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestFastForwardRestRefusesAssetConversionBeforeEffects(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	original := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	testfix.Git(t, root, "switch", "-qc", "incoming")
	asset := "workshop/history/issues/assets/000001/evidence.png"
	procedureWrite(t, root, asset, "one\ntwo")
	procedureWrite(t, root, ".gitattributes", "workshop/history/issues/assets/** text eol=crlf\n")
	testfix.Git(t, root, "add", ".")
	testfix.Git(t, root, "commit", "-qm", "incoming evidence")
	target := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	testfix.Git(t, root, "switch", "-q", "main")
	reader, err := gitx.NewTrunkFileContext(context.Background(), root, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	env := &trackerEnv{ctx: context.Background(), root: root, main: reader, resting: "main"}
	warn := fastForwardRest(env, target)
	if !strings.Contains(warn, asset) || !strings.Contains(warn, ".gitattributes") {
		t.Fatalf("conversion accepted: %q", warn)
	}
	if strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD")) != original {
		t.Fatal("refusal fast forwarded HEAD")
	}
}
