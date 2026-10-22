The empty-pid.log excerpt is the real TestDocRepositoryCancellation failure from
recorded suite attempt 4638dd482de33746116a1788f13ba510 on
0ae0016dbbe6634bfbf59132045e683bfa76862a (2026-10-22).
The retained complete log hashes to
1776aed53279d84d6ecf1877f0646d63862a09edc9a18b9acba3fe9eeba42fef.

The native cancellation regression holds its writer after output-open and replays
this captured empty-content parse failure against the incomplete staging bytes.
The final readiness path must remain absent until the PID write closes and renames.
