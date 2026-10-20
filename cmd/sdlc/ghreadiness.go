package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type landingReadinessGH interface {
	LandingReadiness(context.Context, string, int) (landingReadiness, error)
}

type landingReadiness struct {
	Number                                                                  int
	Repo, State, HeadRef, HeadOID, BaseRef, BaseTargetOID, MergeStateStatus string
}

const landingReadinessQuery = `query($owner: String!, $name: String!, $number: Int!) { repository(owner: $owner, name: $name) { nameWithOwner pullRequest(number: $number) { number state headRefName headRefOid baseRefName baseRef { name target { oid } } headRepository { nameWithOwner } isCrossRepository mergeStateStatus } } }`

// The fixed query selects one PR; cap both request identity and response bytes.
const landingReadinessLimit = 64 * 1024

func parseLandingReadiness(data []byte, repo string, number int) (landingReadiness, error) {
	fail := func(why string) (landingReadiness, error) {
		return landingReadiness{}, fmt.Errorf("landing readiness: %s", why)
	}
	if !landingRepoValid(repo) || len(repo) > 256 || number <= 0 {
		return fail("invalid repository or PR number")
	}
	if len(data) > landingReadinessLimit {
		return fail("response exceeds byte limit")
	}
	var raw struct {
		Errors json.RawMessage
		Data   *struct {
			Repository *struct {
				NameWithOwner string
				PullRequest   *struct {
					Number                                      int
					State, HeadRefName, HeadRefOID, BaseRefName string
					BaseRef                                     *struct {
						Name   string
						Target *struct{ OID string }
					}
					HeadRepository    *struct{ NameWithOwner string }
					IsCrossRepository *bool
					MergeStateStatus  *string
				}
			}
		}
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fail("malformed GraphQL response: " + err.Error())
	}
	if len(raw.Errors) > 0 {
		var errs []json.RawMessage
		if err := json.Unmarshal(raw.Errors, &errs); err != nil || len(errs) > 0 || string(raw.Errors) == "null" {
			return fail("GraphQL errors or incomplete evidence")
		}
	}
	if raw.Data == nil || raw.Data.Repository == nil || raw.Data.Repository.PullRequest == nil {
		return fail("missing repository or PR")
	}
	repository := raw.Data.Repository
	pr := repository.PullRequest
	if !strings.EqualFold(repository.NameWithOwner, repo) || pr.Number != number {
		return fail("repository or PR identity mismatch")
	}
	if pr.State != "OPEN" {
		return fail("PR is not OPEN")
	}
	if pr.HeadRepository == nil || !strings.EqualFold(pr.HeadRepository.NameWithOwner, repo) || pr.IsCrossRepository == nil || *pr.IsCrossRepository {
		return fail("missing or cross-repository head identity")
	}
	if !landingReadinessBranchValid(pr.HeadRefName) || pr.BaseRefName != "main" || pr.BaseRef == nil || pr.BaseRef.Name != pr.BaseRefName {
		return fail("invalid head or main base branch")
	}
	if !landingOIDValid(pr.HeadRefOID) || pr.BaseRef.Target == nil || !landingOIDValid(pr.BaseRef.Target.OID) {
		return fail("missing or malformed head/live base OID")
	}
	status := ""
	if pr.MergeStateStatus != nil {
		status = *pr.MergeStateStatus
	}
	return landingReadiness{Number: number, Repo: repo, State: pr.State, HeadRef: pr.HeadRefName, HeadOID: pr.HeadRefOID, BaseRef: pr.BaseRefName, BaseTargetOID: pr.BaseRef.Target.OID, MergeStateStatus: status}, nil
}

func landingReadinessBranchValid(branch string) bool {
	if branch == "" || len(branch) > 1024 || branch == "@" || strings.HasPrefix(branch, "-") || strings.HasPrefix(branch, "/") || strings.HasSuffix(branch, "/") || strings.HasSuffix(branch, ".") || strings.Contains(branch, "..") || strings.Contains(branch, "//") || strings.Contains(branch, "@{") {
		return false
	}
	for _, c := range branch {
		if c <= 32 || c == 127 || strings.ContainsRune("~^:?*[\\", c) {
			return false
		}
	}
	for _, part := range strings.Split(branch, "/") {
		if strings.HasPrefix(part, ".") || strings.HasSuffix(part, ".lock") {
			return false
		}
	}
	return true
}

func (realGH) LandingReadiness(ctx context.Context, repo string, number int) (landingReadiness, error) {
	if !landingRepoValid(repo) || len(repo) > 256 || number <= 0 {
		return landingReadiness{}, fmt.Errorf("landing readiness: invalid repository or PR number")
	}
	parts := strings.Split(repo, "/")
	data, err := runLandingGHWithLimit(ctx, landingReadinessLimit, "api", "graphql", "--hostname", "github.com", "-f", "query="+landingReadinessQuery, "-f", "owner="+parts[0], "-f", "name="+parts[1], "-F", "number="+strconv.Itoa(number))
	if err != nil {
		return landingReadiness{}, err
	}
	return parseLandingReadiness(data, repo, number)
}

var _ landingReadinessGH = realGH{}
