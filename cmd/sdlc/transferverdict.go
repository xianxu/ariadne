// transferverdict.go — the transfer guard's decision (#285), pure over the
// facts transferguard.go observes for one changed details file.
package main

type verdict int

const (
	verdictAccept   verdict = iota
	verdictNotOwner         // restore main's version, or claim the issue to keep the edit
	verdictBehind           // the owner merges main and resolves the prose
)

// detailsFacts are what a landing's change to one details file is judged on.
type detailsFacts struct {
	Published bool // main's history has touched the file
	Owner     bool // this checkout is the card's claimant
	Based     bool // HEAD contains main's last commit to the file
}

// detailsVerdict is optimistic concurrency on published details: a change is
// the owner's to make, and only against main's latest version. (A branch that
// is based cannot conflict on the file: main has not touched it since.)
func detailsVerdict(f detailsFacts) verdict {
	switch {
	case !f.Published:
		return verdictAccept
	case !f.Owner:
		return verdictNotOwner
	case !f.Based:
		return verdictBehind
	}
	return verdictAccept
}
