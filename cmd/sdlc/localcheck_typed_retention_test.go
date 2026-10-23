package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTypedResumeRetentionRequiresCompatibleParentContract(t *testing.T) {
	for _, mode := range []string{"same", "different", "legacy child", "legacy parent"} {
		t.Run(mode, func(t *testing.T) {
			root, link, now := retentionFixture(t, true)
			load := func(id string) SuiteAttempt {
				b, err := os.ReadFile(filepath.Join(root, "attempts", id+".json"))
				if err != nil {
					t.Fatal(err)
				}
				var a SuiteAttempt
				if err := json.Unmarshal(b, &a); err != nil {
					t.Fatal(err)
				}
				return a
			}
			parent, child := load(link.RerunOf), load(link.ChildID)
			parent.Version, child.Version = 4, 4
			parent.RunnerContractSHA256, child.RunnerContractSHA256 = strings.Repeat("f", 64), strings.Repeat("f", 64)
			switch mode {
			case "different":
				child.RunnerContractSHA256 = strings.Repeat("e", 64)
			case "legacy child":
				child.Version = 2
				child.RunnerContractSHA256 = ""
			case "legacy parent":
				parent.Version = 2
				parent.RunnerContractSHA256 = ""
			}
			link.ParentRecordSHA256 = resumeRecordDigest(retentionWrite(t, root, parent.ID, parent))
			retentionWrite(t, root, child.ID, child)
			retentionWrite(t, root, link.ID, link)
			if _, err := pruneResumeBundles(root, now); err != nil {
				t.Fatal(err)
			}
			for _, id := range []string{parent.ID, child.ID, link.ID} {
				_, err := os.Stat(filepath.Join(root, "attempts", id+".json"))
				if os.IsNotExist(err) != (mode == "same") {
					t.Fatalf("%s survival=%v for %s", id, err, mode)
				}
			}
		})
	}
}
