// transferverdict.go — the transfer guard's decision (#285), pure over the
// facts transferguard.go observes for one changed details file.
package main

type verdict int

const (
	verdictAccept  verdict = iota
	verdictForeign         // another workspace owns the issue: coordinate, or drop the edit
	verdictBehind          // merge main and resolve the prose
)

// detailsFacts are what a landing's change to one details file is judged on.
type detailsFacts struct {
	Published bool // main's history has touched the file
	Foreign   bool // another workspace is the card's claimant (#274)
	Based     bool // HEAD contains main's last commit to the file
}

// detailsVerdict is optimistic concurrency on published details: a change is
// refused only when another workspace holds the issue; this workspace's own
// issue and an unclaimed one (no owner to protect, #274) are editable, against
// main's latest version. (A branch that is based cannot conflict on the file:
// main has not touched it since.)
func detailsVerdict(f detailsFacts) verdict {
	switch {
	case !f.Published:
		return verdictAccept
	case f.Foreign:
		return verdictForeign
	case !f.Based:
		return verdictBehind
	}
	return verdictAccept
}
