package main

import "fmt"

// preCloseRunCommand is the run every pre-close instruction names (#578), so
// the repository's rule is said once, by preCloseCheck, and never in prose.
const preCloseRunCommand = "sdlc local-check run --pre-close"

// preCloseCheck is the check a branch runs before close (#578). Where trains
// are on and main declares the affected suite, it is affected: the train's
// suite on its prefix is the landing's one full run. Everywhere else it is
// the suite, which merge reuses when code has not changed since. Pure.
func preCloseCheck(trainsOn, affectedDeclared bool) (name, why string) {
	switch {
	case trainsOn && affectedDeclared:
		return affectedCheckName, "trains are on and main declares " + affectedConfigPath + ": the train runs the full suite"
	case trainsOn:
		return suiteCheckName, "trains are on, but main declares no " + affectedConfigPath
	case affectedDeclared:
		return suiteCheckName, "trains are off (" + landingConfigPath + "): merge lands on a full suite"
	}
	return suiteCheckName, "trains are off and main declares no " + affectedConfigPath
}

// resolvePreCloseCheck reads main's landing and affected config at rev and
// decides. A file that cannot be read or trusted degrades to the suite, as
// the car gate's stand-in does (#473 M2), and says why.
func resolvePreCloseCheck(git func(args ...string) (string, error), rev string) (name, why string) {
	s, err := readLandingSchedule(git, rev)
	if err != nil {
		return suiteCheckName, fmt.Sprintf("%s unreadable (%v)", landingConfigPath, err)
	}
	_, declared, err := readAffectedConfig(git, rev)
	if err != nil {
		return suiteCheckName, fmt.Sprintf("%s unreadable (%v)", affectedConfigPath, err)
	}
	return preCloseCheck(!s.Off(), declared)
}
