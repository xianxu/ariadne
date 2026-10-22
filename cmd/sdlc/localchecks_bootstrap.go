package main

import "context"

// withSubjectDefaultSuiteContext extends the main-authoritative policy only
// when main has neither a declaration nor a test target. The exact subject may
// supply a default make test, never its own declaration or CI substitutions.
// The boolean is diagnostic provenance, not a new persisted evidence format.
func withSubjectDefaultSuiteContext(ctx context.Context, git func(...string) (string, error), base, head string, decl localCheckDecl) (localCheckDecl, bool, error) {
	if err := ctx.Err(); err != nil {
		return localCheckDecl{}, false, err
	}
	main, err := withDefaultSuiteContext(ctx, git, base, decl)
	if err != nil {
		return localCheckDecl{}, false, err
	}
	if err := ctx.Err(); err != nil {
		return localCheckDecl{}, false, err
	}
	if _, hasSuite := main.check(suiteCheckName); decl.Version != 0 || hasSuite {
		return main, false, nil
	}
	subject, err := withDefaultSuiteContext(ctx, git, head, main)
	if err != nil {
		return localCheckDecl{}, false, err
	}
	_, bootstrap := subject.check(suiteCheckName)
	return subject, bootstrap, nil
}
