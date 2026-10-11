package testfix

import (
	"os"
	"testing"
)

// QuarantineEnv runs quarantined tests when set to "run".
const QuarantineEnv = "WF_TEST_QUARANTINED"

// Quarantine skips a test known to fail for a reason its owning issue tracks, so
// the suite stays a signal (#330). The skip names the issue; the issue's fix
// removes the call. Set WF_TEST_QUARANTINED=run to run it anyway, e.g. while
// fixing it. Never quarantine without an open owning issue.
func Quarantine(t testing.TB, issue, why string) {
	t.Helper()
	if os.Getenv(QuarantineEnv) == "run" {
		return
	}
	t.Skipf("QUARANTINED (%s): %s; %s=run runs it", issue, why, QuarantineEnv)
}
