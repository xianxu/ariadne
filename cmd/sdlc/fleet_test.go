package main

import (
	"bytes"
	"encoding/json"
	"io"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/helptext"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/fleet"
)

type fleetTestGit struct{}

func (fleetTestGit) GitInDir(string, ...string) ([]byte, error) { return nil, nil }

func TestFleetCommandGrammarAndRegistration(t *testing.T) {
	cmd := newFleetCmd(fleetCommandDeps{})
	if cmd.Use != "fleet" || cmd.Args == nil {
		t.Fatalf("fleet command = use %q args %#v", cmd.Use, cmd.Args)
	}
	for _, name := range []string{"inventory", "policy"} {
		sub, _, err := cmd.Find([]string{name})
		if err != nil || sub.Name() != name {
			t.Fatalf("fleet %s missing: command=%v err=%v", name, sub, err)
		}
		if got := sub.Flags().Lookup("path").DefValue; got != "." {
			t.Fatalf("fleet %s --path default = %q, want .", name, got)
		}
		if sub.Flags().Lookup("json") == nil {
			t.Fatalf("fleet %s lacks --json", name)
		}
		if commandNeedsRepoLock(sub) {
			t.Fatalf("fleet %s is read-only and must not acquire the repository lock", name)
		}
		invocation := newFleetCmd(fleetCommandDeps{})
		invocation.SetArgs([]string{name, "unexpected"})
		if err := invocation.Execute(); err == nil {
			t.Fatalf("fleet %s accepted a positional argument", name)
		}
	}

	root := buildRoot()
	names := make([]string, 0, len(root.Commands()))
	for _, child := range root.Commands() {
		if !child.Hidden {
			names = append(names, child.Name())
		}
	}
	state := indexString(names, "state")
	if state < 0 || state+1 >= len(names) || names[state+1] != "fleet" {
		t.Fatalf("root order around state = %v, want fleet immediately after state", names)
	}
	if _, ok := defaultFleetCommandDeps().git.(execGitRunner); !ok {
		t.Fatalf("production fleet Git adapter = %T, want execGitRunner.GitInDir", defaultFleetCommandDeps().git)
	}
}

func TestFleetInventoryUsesNormalizedFleetAndSharedLoader(t *testing.T) {
	testRoot := t.TempDir()
	fleetRoot := filepath.Join(testRoot, "fleet")
	repoRoot := filepath.Join(fleetRoot, "repo")
	caller := filepath.Join(repoRoot, "caller")
	policyPath := fleet.PolicyDeclarationPath(repoRoot)
	git := fleetTestGit{}
	var normalizedPath, collectedRoot, loadedPath string
	resolveCalls := 0
	deps := fleetCommandDeps{
		git: git,
		normalizeVantage: func(g fleet.GitReader, path string) (fleet.Vantage, error) {
			if _, ok := g.(fleetTestGit); !ok {
				t.Fatalf("normalize Git = %T", g)
			}
			normalizedPath = path
			return fleet.Vantage{FleetRoot: fleetRoot, PrimaryRoot: repoRoot}, nil
		},
		loadPolicy: func(path string) fleet.PolicyCapability {
			loadedPath = path
			return fleet.PolicyCapability{Diagnostic: &fleet.PolicyDiagnostic{Code: fleet.DiagnosticMissingPolicy, Message: "missing", Path: path}}
		},
		collectInventory: func(root string, options fleet.InventoryOptions) (fleet.Inventory, error) {
			collectedRoot = root
			if options.Git == nil || options.LoadPolicy == nil {
				t.Fatal("inventory did not receive Git and shared policy loader")
			}
			options.LoadPolicy(policyPath)
			return fleet.Inventory{}, nil
		},
		resolvePolicy: func(fleet.PolicyCapabilityValue, fleet.CanonicalPaths) fleet.PolicyResult {
			resolveCalls++
			return fleet.PolicyResult{}
		},
	}
	stdout, _, err := executeFleetCommand(t, deps, "inventory", "--path", caller, "--json")
	if err != nil {
		t.Fatal(err)
	}
	if normalizedPath != caller || collectedRoot != fleetRoot || loadedPath != policyPath || resolveCalls != 0 {
		t.Fatalf("inventory routing normalize=%q root=%q policy=%q resolve=%d", normalizedPath, collectedRoot, loadedPath, resolveCalls)
	}
	var got fleet.Inventory
	if err := json.Unmarshal([]byte(stdout), &got); err != nil || got.Rows == nil || got.Diagnostics == nil {
		t.Fatalf("inventory JSON = %q, decode=%+v err=%v", stdout, got, err)
	}
}

func TestFleetPolicyCanonicalizesLoadsPrimaryAndAloneResolves(t *testing.T) {
	testRoot := t.TempDir()
	fleetRoot := filepath.Join(testRoot, "fleet")
	repoRoot := filepath.Join(fleetRoot, "repo")
	requested := filepath.Join(repoRoot, "future", "target")
	repoIdentity := filepath.Join(repoRoot, ".git")
	prospectiveArg := filepath.Join("future", "target")
	policyPath := fleet.PolicyDeclarationPath(repoRoot)
	digest := strings.Repeat("a", 64)
	value := fleet.PolicyCapabilityValue{PolicyVersion: 1, PolicyDigest: digest, KeyKind: "repo", Roots: []string{}, Capacity: fleet.Capacity{Kind: "unbounded"}}
	var canonicalInput, normalizedPath, loadedPath string
	var resolved fleet.CanonicalPaths
	deps := fleetCommandDeps{
		git: fleetTestGit{},
		canonicalProspectivePath: func(path string) (string, string, error) {
			canonicalInput = path
			return requested, repoRoot, nil
		},
		normalizeVantage: func(_ fleet.GitReader, path string) (fleet.Vantage, error) {
			normalizedPath = path
			return fleet.Vantage{RepoIdentity: repoIdentity, PrimaryRoot: repoRoot, WorktreeRoot: repoRoot}, nil
		},
		loadPolicy: func(path string) fleet.PolicyCapability {
			loadedPath = path
			return fleet.PolicyCapability{OK: true, Value: &value}
		},
		resolvePolicy: func(policy fleet.PolicyCapabilityValue, paths fleet.CanonicalPaths) fleet.PolicyResult {
			if !reflect.DeepEqual(policy, value) {
				t.Fatalf("resolved policy = %+v", policy)
			}
			resolved = paths
			return fleet.PolicyResult{OK: true, Value: &fleet.PolicyResultValue{PolicyVersion: 1, PolicyDigest: digest, RepoIdentity: paths.RepoIdentity, AdmissionKey: paths.RepoIdentity, Capacity: fleet.Capacity{Kind: "unbounded"}}}
		},
	}
	stdout, _, err := executeFleetCommand(t, deps, "policy", "--path", prospectiveArg, "--json")
	if err != nil {
		t.Fatal(err)
	}
	wantPaths := fleet.CanonicalPaths{RepoIdentity: repoIdentity, RepoRoot: repoRoot, WorktreeRoot: repoRoot, Requested: requested}
	if canonicalInput != prospectiveArg || normalizedPath != repoRoot || loadedPath != policyPath || !reflect.DeepEqual(resolved, wantPaths) {
		t.Fatalf("policy routing canonical=%q normalize=%q load=%q paths=%+v", canonicalInput, normalizedPath, loadedPath, resolved)
	}
	var got fleet.PolicyResult
	if err := json.Unmarshal([]byte(stdout), &got); err != nil || !got.OK {
		t.Fatalf("policy JSON = %q, decode=%+v err=%v", stdout, got, err)
	}
}

func TestFleetPolicyDiagnosticPreservesStdoutAndReturnsRefusal(t *testing.T) {
	repoRoot := filepath.Join(t.TempDir(), "repo")
	requested := filepath.Join(repoRoot, "requested")
	repoIdentity := filepath.Join(repoRoot, ".git")
	loadCalls, resolveCalls := 0, 0
	deps := fleetCommandDeps{
		git:                      fleetTestGit{},
		canonicalProspectivePath: func(string) (string, string, error) { return requested, repoRoot, nil },
		normalizeVantage: func(fleet.GitReader, string) (fleet.Vantage, error) {
			return fleet.Vantage{PrimaryRoot: repoRoot, RepoIdentity: repoIdentity, WorktreeRoot: repoRoot}, nil
		},
		loadPolicy: func(path string) fleet.PolicyCapability {
			loadCalls++
			return fleet.PolicyCapability{Diagnostic: &fleet.PolicyDiagnostic{Code: fleet.DiagnosticMissingPolicy, Message: "missing declaration", Path: path}}
		},
		resolvePolicy: func(fleet.PolicyCapabilityValue, fleet.CanonicalPaths) fleet.PolicyResult {
			resolveCalls++
			return fleet.PolicyResult{}
		},
	}
	stdout, stderr, err := executeFleetCommand(t, deps, "policy", "--json")
	if err == nil || loadCalls != 1 || resolveCalls != 0 {
		t.Fatalf("diagnostic result err=%v load=%d resolve=%d", err, loadCalls, resolveCalls)
	}
	if strings.Contains(stderr, "Usage:") {
		t.Fatalf("refusal printed usage: %q", stderr)
	}
	var got fleet.PolicyResult
	if decodeErr := json.Unmarshal([]byte(stdout), &got); decodeErr != nil || got.OK || got.Diagnostic == nil || got.Diagnostic.Code != fleet.DiagnosticMissingPolicy {
		t.Fatalf("refusal stdout = %q decoded=%+v err=%v", stdout, got, decodeErr)
	}
}

func TestFleetPolicyResolverDiagnosticPreservesStdoutAndReturnsRefusal(t *testing.T) {
	repoRoot := filepath.Join(t.TempDir(), "repo")
	requested := filepath.Join(repoRoot, "outside")
	repoIdentity := filepath.Join(repoRoot, ".git")
	digest := strings.Repeat("b", 64)
	resolveCalls := 0
	deps := fleetCommandDeps{
		git:                      fleetTestGit{},
		canonicalProspectivePath: func(string) (string, string, error) { return requested, repoRoot, nil },
		normalizeVantage: func(fleet.GitReader, string) (fleet.Vantage, error) {
			return fleet.Vantage{PrimaryRoot: repoRoot, RepoIdentity: repoIdentity, WorktreeRoot: repoRoot}, nil
		},
		loadPolicy: func(string) fleet.PolicyCapability {
			return fleet.PolicyCapability{OK: true, Value: &fleet.PolicyCapabilityValue{PolicyVersion: 1, PolicyDigest: digest, KeyKind: "declared-root", Roots: []string{"apps/*"}, Capacity: fleet.Capacity{Kind: "unbounded"}}}
		},
		resolvePolicy: func(_ fleet.PolicyCapabilityValue, paths fleet.CanonicalPaths) fleet.PolicyResult {
			resolveCalls++
			version := 1
			return fleet.PolicyResult{Diagnostic: &fleet.PolicyDiagnostic{Code: fleet.DiagnosticOutsideDeclaredScope, Message: "outside", Path: paths.Requested, PolicyVersion: &version}}
		},
	}
	stdout, stderr, err := executeFleetCommand(t, deps, "policy", "--json")
	if err == nil || resolveCalls != 1 || strings.Contains(stderr, "Usage:") {
		t.Fatalf("resolver refusal err=%v calls=%d stderr=%q", err, resolveCalls, stderr)
	}
	var got fleet.PolicyResult
	if decodeErr := json.Unmarshal([]byte(stdout), &got); decodeErr != nil || got.Diagnostic == nil || got.Diagnostic.Code != fleet.DiagnosticOutsideDeclaredScope {
		t.Fatalf("resolver refusal stdout=%q decoded=%+v err=%v", stdout, got, decodeErr)
	}
}

func TestFleetPolicyRejectsMalformedSuccessCapabilityBeforeResolve(t *testing.T) {
	for _, tt := range []struct {
		name       string
		capability fleet.PolicyCapability
	}{
		{"nil value", fleet.PolicyCapability{OK: true}},
		{"invalid value", fleet.PolicyCapability{OK: true, Value: &fleet.PolicyCapabilityValue{PolicyVersion: 1, PolicyDigest: "bad", KeyKind: "unknown", Roots: []string{}, Capacity: fleet.Capacity{Kind: "unbounded"}}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repoRoot := filepath.Join(t.TempDir(), "repo")
			requested := filepath.Join(repoRoot, "target")
			repoIdentity := filepath.Join(repoRoot, ".git")
			resolveCalls := 0
			deps := fleetCommandDeps{
				git:                      fleetTestGit{},
				canonicalProspectivePath: func(string) (string, string, error) { return requested, repoRoot, nil },
				normalizeVantage: func(fleet.GitReader, string) (fleet.Vantage, error) {
					return fleet.Vantage{PrimaryRoot: repoRoot, RepoIdentity: repoIdentity, WorktreeRoot: repoRoot}, nil
				},
				loadPolicy: func(string) fleet.PolicyCapability { return tt.capability },
				resolvePolicy: func(fleet.PolicyCapabilityValue, fleet.CanonicalPaths) fleet.PolicyResult {
					resolveCalls++
					return fleet.PolicyResult{}
				},
			}
			var err error
			var panicValue any
			func() {
				defer func() { panicValue = recover() }()
				_, _, err = executeFleetCommand(t, deps, "policy", "--json")
			}()
			if panicValue != nil {
				t.Fatalf("malformed capability panicked: %v", panicValue)
			}
			if err == nil || !strings.Contains(err.Error(), "fleet policy") || resolveCalls != 0 {
				t.Fatalf("malformed capability err=%v resolveCalls=%d", err, resolveCalls)
			}
		})
	}
}

func TestFleetPolicyRejectsMalformedResolverResultBeforeOutput(t *testing.T) {
	digest := strings.Repeat("c", 64)
	for _, result := range []struct {
		name  string
		value fleet.PolicyResult
	}{
		{"zero result", fleet.PolicyResult{}},
		{"invalid success value", fleet.PolicyResult{OK: true, Value: &fleet.PolicyResultValue{
			PolicyVersion: 1,
			PolicyDigest:  "bad",
			RepoIdentity:  "repo-identity",
			AdmissionKey:  "repo-identity",
			Capacity:      fleet.Capacity{Kind: "unbounded"},
		}}},
	} {
		for _, output := range []struct {
			name string
			args []string
		}{
			{"json", []string{"policy", "--json"}},
			{"human", []string{"policy"}},
		} {
			t.Run(result.name+"/"+output.name, func(t *testing.T) {
				repoRoot := filepath.Join(t.TempDir(), "repo")
				repoIdentity := filepath.Join(repoRoot, ".git")
				requested := filepath.Join(repoRoot, "target")
				resolveCalls, renderCalls := 0, 0
				deps := fleetCommandDeps{
					git:                      fleetTestGit{},
					canonicalProspectivePath: func(string) (string, string, error) { return requested, repoRoot, nil },
					normalizeVantage: func(fleet.GitReader, string) (fleet.Vantage, error) {
						return fleet.Vantage{PrimaryRoot: repoRoot, RepoIdentity: repoIdentity, WorktreeRoot: repoRoot}, nil
					},
					loadPolicy: func(string) fleet.PolicyCapability {
						return fleet.PolicyCapability{OK: true, Value: &fleet.PolicyCapabilityValue{
							PolicyVersion: 1,
							PolicyDigest:  digest,
							KeyKind:       "repo",
							Roots:         []string{},
							Capacity:      fleet.Capacity{Kind: "unbounded"},
						}}
					},
					resolvePolicy: func(fleet.PolicyCapabilityValue, fleet.CanonicalPaths) fleet.PolicyResult {
						resolveCalls++
						return result.value
					},
					renderPolicy: func(w io.Writer, _ fleet.PolicyResult) error {
						renderCalls++
						_, err := io.WriteString(w, "lenient output\n")
						return err
					},
				}

				var stdout string
				var err error
				var panicValue any
				func() {
					defer func() { panicValue = recover() }()
					stdout, _, err = executeFleetCommand(t, deps, output.args...)
				}()
				if panicValue != nil {
					t.Fatalf("malformed resolver result panicked: %v", panicValue)
				}
				if err == nil || !strings.Contains(err.Error(), "fleet policy: invalid resolved result") {
					t.Fatalf("malformed resolver result error = %v, want contextual validation error", err)
				}
				if stdout != "" || resolveCalls != 1 || renderCalls != 0 {
					t.Fatalf("malformed resolver result stdout=%q resolve=%d render=%d", stdout, resolveCalls, renderCalls)
				}
			})
		}
	}
}

func TestFleetInventoryAndPolicyHumanOutputUseTypedRenderers(t *testing.T) {
	testRoot := t.TempDir()
	fleetRoot := filepath.Join(testRoot, "fleet")
	repoRoot := filepath.Join(fleetRoot, "repo")
	repoIdentity := filepath.Join(repoRoot, ".git")
	renderedInventory, renderedPolicy := 0, 0
	deps := fleetCommandDeps{
		git: fleetTestGit{},
		normalizeVantage: func(fleet.GitReader, string) (fleet.Vantage, error) {
			return fleet.Vantage{FleetRoot: fleetRoot, PrimaryRoot: repoRoot, RepoIdentity: repoIdentity, WorktreeRoot: repoRoot}, nil
		},
		canonicalProspectivePath: func(string) (string, string, error) { return repoRoot, repoRoot, nil },
		collectInventory:         func(string, fleet.InventoryOptions) (fleet.Inventory, error) { return fleet.Inventory{}, nil },
		loadPolicy: func(string) fleet.PolicyCapability {
			return fleet.PolicyCapability{Diagnostic: &fleet.PolicyDiagnostic{Code: fleet.DiagnosticMissingPolicy, Message: "missing"}}
		},
		renderInventory: func(w io.Writer, _ fleet.Inventory) error {
			renderedInventory++
			_, err := io.WriteString(w, "human inventory\n")
			return err
		},
		renderPolicy: func(w io.Writer, _ fleet.PolicyResult) error {
			renderedPolicy++
			_, err := io.WriteString(w, "human policy\n")
			return err
		},
	}
	stdout, _, err := executeFleetCommand(t, deps, "inventory")
	if err != nil || stdout != "human inventory\n" || renderedInventory != 1 {
		t.Fatalf("human inventory = (%q, %v), calls=%d", stdout, err, renderedInventory)
	}
	stdout, _, err = executeFleetCommand(t, deps, "policy")
	if err == nil || stdout != "human policy\n" || renderedPolicy != 1 {
		t.Fatalf("human policy refusal = (%q, %v), calls=%d", stdout, err, renderedPolicy)
	}
}

func TestFleetHelpDocumentsDefaultPathAndOutput(t *testing.T) {
	for _, name := range []string{"fleet", "fleet-inventory", "fleet-policy"} {
		doc, ok := helptext.Get(name)
		if !ok {
			t.Fatalf("%s help is not embedded", name)
		}
		if !strings.Contains(doc, "--path") || !strings.Contains(doc, "defaults to `.`") {
			t.Fatalf("%s help does not state --path defaults to .:\n%s", name, doc)
		}
	}
	for _, args := range [][]string{{"fleet", "--help"}, {"fleet", "inventory", "--help"}, {"fleet", "policy", "--help"}} {
		root := buildRoot()
		var stdout bytes.Buffer
		root.SetOut(&stdout)
		root.SetArgs(args)
		if err := root.Execute(); err != nil {
			t.Fatalf("sdlc %v: %v", args, err)
		}
		if !strings.Contains(stdout.String(), "--path") || !strings.Contains(stdout.String(), "defaults to `.`") {
			t.Fatalf("sdlc %v help lacks default path:\n%s", args, stdout.String())
		}
	}
}

func executeFleetCommand(t *testing.T, deps fleetCommandDeps, args ...string) (string, string, error) {
	t.Helper()
	cmd := newFleetCmd(deps)
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func indexString(values []string, target string) int {
	for i, value := range values {
		if value == target {
			return i
		}
	}
	return -1
}
