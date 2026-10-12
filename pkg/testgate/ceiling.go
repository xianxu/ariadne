package testgate

import "time"

// HangCeiling bounds a test's wait for an event that a correct program produces
// promptly. It catches a hang, not a slow machine: with several slots' suites
// running at once (load ~50), spawning a binary or running git is many times
// slower, and a deadline of a few seconds turns contention into failures
// (#325). Assert ordering or concurrency with event evidence; this ceiling only
// keeps a broken run from stalling until go test's own timeout. It lives here,
// beside the load gate, so every module package's tests share one value.
const HangCeiling = 2 * time.Minute
