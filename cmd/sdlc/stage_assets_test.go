package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestStageAndClearPreserveAssetBytes(t *testing.T) {
	for _, clear := range []bool{false, true} {
		for _, hook := range []bool{false, true} {
			name := map[bool]string{false: "stage", true: "clear"}[clear] + "/" + map[bool]string{false: "conversion", true: "hook"}[hook]
			t.Run(name, func(t *testing.T) {
				r, _ := moveFixture(t)
				t.Setenv("WF_ISSUES_DIR", "custom/issues")
				asset := "custom/issues/assets/000001/evidence.png"
				source := r[1]
				if clear {
					source = r[0]
				}
				procedureWrite(t, source, asset, "one\ntwo")
				testfix.Git(t, source, "add", ".")
				testfix.Git(t, source, "commit", "-qm", "evidence")
				if clear {
					if out, err := runStageTest(t, r[1], ":0", false); err != nil {
						t.Fatalf("prepare stage: %v\n%s", err, out)
					}
				}
				gitDir := strings.TrimSpace(testfix.Capture(t, r[0], "rev-parse", "--path-format=absolute", "--git-common-dir"))
				if hook {
					procedureWrite(t, gitDir, "hooks/post-checkout", "#!/bin/sh\n[ -f '"+asset+"' ] || exit 0\nprintf changed > '"+asset+"'\n")
					if err := os.Chmod(filepath.Join(gitDir, "hooks/post-checkout"), 0755); err != nil {
						t.Fatal(err)
					}
				} else {
					procedureWrite(t, gitDir, "info/attributes", "custom/issues/assets/** text eol=crlf\n")
				}
				before := slotState(t, r[0])
				var err error
				if clear {
					_, err = runUnstageTest(t, r[1], ":0")
				} else {
					_, err = runStageTest(t, r[1], ":0", false)
				}
				if err == nil || !strings.Contains(err.Error(), asset) {
					t.Fatalf("unsafe materialization accepted: %v", err)
				}
				if !hook && slotState(t, r[0]) != before {
					t.Fatal("conversion refusal changed the target slot")
				}
			})
		}
	}
}
