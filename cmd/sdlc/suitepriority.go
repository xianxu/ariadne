package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

const suitePriorityProtocolLimit = 64 << 10

type suitePrioritySubject struct {
	CommonDir string `json:"common_dir"`
	Issue     int    `json:"issue"`
	Issues    []int  `json:"issues,omitempty"` // exact train prefix; older helpers see issue zero (FIFO)
}

type suitePrioritySource struct {
	CommonDir string             `json:"common_dir"`
	Deadlines map[string]float64 `json:"deadlines"`
	RefreshAt float64            `json:"refresh_at"`
	Error     string             `json:"error,omitempty"`
}

// suiteSubject binds identity to the original checkout, never the temporary
// worktree used by a local check. A different tested revision cannot borrow
// this checkout's issue grant.
func suiteSubject(root, testedSHA string) (suitePrioritySubject, error) {
	r := execGitRunner{}
	common, err := landingGit(r, root, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return suitePrioritySubject{}, err
	}
	subject := suitePrioritySubject{CommonDir: common}
	head, err := landingGit(r, root, "rev-parse", "HEAD")
	if err != nil {
		return subject, nil
	} // unborn checkout has no tested issue head
	if testedSHA != "" && head != testedSHA {
		return subject, nil
	}
	branch, err := landingGit(r, root, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		return subject, nil
	} // detached or unborn: ordinary FIFO
	subject.Issue = issue.IDFromFilename(branch)
	return subject, nil
}

// suiteGrantSnapshot is a read-only projection of the one grant authority.
// Missing logs are normal. Any damaged evidence confers no priority, because
// a skipped entry could have been a revoke of an earlier grant.
func suiteGrantSnapshot(subjects []suitePrioritySubject, now time.Time) []suitePrioritySource {
	sources := []suitePrioritySource{}
	requested := map[string]map[int]bool{}
	for _, s := range subjects {
		if _, ok := requested[s.CommonDir]; !ok {
			requested[s.CommonDir] = map[int]bool{}
			sources = append(sources, suitePrioritySource{CommonDir: s.CommonDir, Deadlines: map[string]float64{}})
		}
		requested[s.CommonDir][s.Issue] = true
	}
	for i := range sources {
		source := &sources[i]
		entries, skipped, err := landingline.ReadGrantEntries(filepath.Join(source.CommonDir, "sdlc", "landing"))
		if err != nil {
			source.Error = "cannot read grants: " + err.Error()
			continue
		}
		if skipped != 0 {
			source.Error = fmt.Sprintf("grant log has %d malformed entries", skipped)
			continue
		}
		for id, g := range landingline.ValidGrants(entries, now) {
			if id > 0 && requested[source.CommonDir][id] {
				source.Deadlines[strconv.Itoa(id)] = suitePriorityUnix(g.Expires)
			}
		}
		// Conservatively refresh at every future temporal boundary, including
		// future-dated grants which are not active in ValidGrants yet.
		for _, g := range entries {
			for _, at := range []time.Time{g.Granted, g.Expires} {
				if at.After(now) && (source.RefreshAt == 0 || suitePriorityUnix(at) < source.RefreshAt) {
					source.RefreshAt = suitePriorityUnix(at)
				}
			}
		}
	}
	return sources
}

func suitePriorityUnix(at time.Time) float64 {
	return float64(at.Unix()) + float64(at.Nanosecond())/1e9
}

func runSuitePriority(out io.Writer, input io.Reader, describe bool, repo string, now time.Time) error {
	var response any
	if describe {
		subject, err := suiteSubject(repo, "")
		if err != nil {
			return err
		}
		response = struct {
			Version int                  `json:"version"`
			Subject suitePrioritySubject `json:"subject"`
		}{1, subject}
	} else {
		data, err := io.ReadAll(io.LimitReader(input, suitePriorityProtocolLimit+1))
		if err != nil {
			return err
		}
		if len(data) > suitePriorityProtocolLimit {
			return errors.New("suite priority request exceeds 64 KiB")
		}
		var request struct {
			Version  int                    `json:"version"`
			Subjects []suitePrioritySubject `json:"subjects"`
		}
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&request); err != nil {
			return fmt.Errorf("suite priority request: %w", err)
		}
		if err := decoder.Decode(new(any)); err != io.EOF {
			return errors.New("suite priority request must be one JSON object")
		}
		if request.Version != 1 || request.Subjects == nil {
			return errors.New("suite priority request requires version 1 and subjects array")
		}
		for _, s := range request.Subjects {
			if !filepath.IsAbs(s.CommonDir) || filepath.Clean(s.CommonDir) != s.CommonDir || len(s.CommonDir) > 4096 || s.Issue < 0 {
				return errors.New("suite priority subject requires an absolute canonical common_dir and nonnegative issue")
			}
		}
		response = struct {
			Version int                   `json:"version"`
			Sources []suitePrioritySource `json:"sources"`
		}{1, suiteGrantSnapshot(request.Subjects, now)}
	}
	data, err := json.Marshal(response)
	if err != nil {
		return err
	}
	if len(data)+1 > suitePriorityProtocolLimit {
		return errors.New("suite priority response exceeds 64 KiB")
	}
	_, err = out.Write(append(data, '\n'))
	return err
}

func newSuitePriorityCmd(repo *string, now func() time.Time) *cobra.Command {
	var describe bool
	cmd := &cobra.Command{Use: "suite-priority", Hidden: true, Args: cobra.NoArgs, SilenceUsage: true, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSuitePriority(cmd.OutOrStdout(), cmd.InOrStdin(), describe, *repo, now())
		}}
	cmd.Flags().BoolVar(&describe, "describe", false, "describe this checkout's suite subject")
	return cmd
}
