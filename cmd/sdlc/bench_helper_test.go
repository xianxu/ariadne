package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBenchSuiteUsesToolSupplierInsteadOfReplay(t *testing.T) {
	c := benchContextFixture(t)
	prepareTestWrite(t, c.WorkRoot, "Makefile.workflow", "# historical replay\n", 0600)
	prepareTestWrite(t, c.WorkRoot, "scripts/test-lock.py", fakeSuiteLockHelper("raise Exception('historical admission must not run')\n"), 0755)
	supplier := t.TempDir()
	helper := filepath.Join(supplier, "scripts", "test-lock.py")
	body := []byte(fakeSuiteLockHelper("print('acquired')\n"))
	prepareTestWrite(t, supplier, "scripts/test-lock.py", string(body), 0755)
	sum := sha256.Sum256(body)
	raw, err := os.ReadFile(filepath.Join(c.WorkRoot, benchPublicContextPath))
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	fields["suite_helper"] = map[string]string{"path": helper, "sha256": hex.EncodeToString(sum[:])}
	raw, err = json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(c.WorkRoot, benchPublicContextPath), raw, 0600); err != nil {
		t.Fatal(err)
	}
	got, err := prepareSuiteHelper(context.Background(), c.WorkRoot, "")
	if err != nil || got != helper {
		t.Fatalf("admission supplier = %s, %v; want current tool %s", got, err, helper)
	}
}

func benchHelperFixture(t *testing.T) benchSuiteHelper {
	t.Helper()
	root := t.TempDir()
	prepareTestWrite(t, root, "scripts/test-lock.py", fakeSuiteLockHelper("print('acquired')\n"), 0755)
	h, err := captureBenchSuiteHelper(root)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestBenchSuiteRefusesChangedOrMissingSupplier(t *testing.T) {
	for _, kind := range []string{"changed", "missing", "symlink", "unbound"} {
		t.Run(kind, func(t *testing.T) {
			c := benchContextFixture(t)
			prepareTestWrite(t, c.WorkRoot, "Makefile.workflow", "# must not fall back\n", 0600)
			prepareTestWrite(t, c.WorkRoot, "scripts/test-lock.py", fakeSuiteLockHelper("print('acquired')\n"), 0755)
			switch kind {
			case "changed":
				if err := os.WriteFile(c.SuiteHelper.Path, []byte("print('changed')\n"), 0600); err != nil {
					t.Fatal(err)
				}
			case "missing":
				if err := os.Remove(c.SuiteHelper.Path); err != nil {
					t.Fatal(err)
				}
			case "symlink":
				moved := c.SuiteHelper.Path + "-original"
				if err := os.Rename(c.SuiteHelper.Path, moved); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(moved, c.SuiteHelper.Path); err != nil {
					t.Fatal(err)
				}
			case "unbound":
				c.SuiteHelper = benchSuiteHelper{}
				writeBenchContextFixture(t, c.WorkRoot, c)
			}
			if _, err := prepareSuiteHelper(context.Background(), c.WorkRoot, ""); err == nil {
				t.Fatal("unavailable bound supplier fell back to historical admission")
			}
		})
	}
}

func TestBenchContextRefusesOldAdmissionContract(t *testing.T) {
	c := benchContextFixture(t)
	c.Version = 1
	writeBenchContextFixture(t, c.WorkRoot, c)
	if _, err := loadBenchContext(context.Background(), c.WorkRoot); err == nil {
		t.Fatal("old bench context admitted without the new helper contract")
	}
}

func TestBenchScoreRefusesChangedSupplierBeforeAdmission(t *testing.T) {
	c := benchContextFixture(t)
	prepareTestWrite(t, c.WorkRoot, "Makefile.workflow", "# legacy supplier must not rescue this run\n", 0600)
	prepareTestWrite(t, c.WorkRoot, "scripts/test-lock.py", fakeSuiteLockHelper("print('acquired')\n"), 0755)
	helper := benchHelperFixture(t)
	if err := os.WriteFile(helper.Path, []byte("print('changed')\n"), 0600); err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(t.TempDir(), "score.log")
	result, err := runBenchScoreCommand(context.Background(), benchScoreCommand{SuiteHelper: helper, Root: c.WorkRoot, WorkRoot: c.WorkRoot, Role: "parent", LogPath: log, Argv: []string{"true"}})
	if err == nil || result.Exit != -1 || !result.Quiet {
		t.Fatalf("unavailable supplier admitted: %+v %v", result, err)
	}
	if _, err := os.Stat(log); !os.IsNotExist(err) {
		t.Fatalf("refused scorer created a command log: %v", err)
	}
}
