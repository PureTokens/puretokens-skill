# Optional anonymous operations receipts

Updated 2026-10-01. Source candidate; not a published-release or real-host
acceptance claim.

Default is disabled. Only the explicit process environment value
`PTP_OPERATIONS_RECEIPTS=1` enables anonymous completion reports to
`https://console.puretokensx.com/api/product/analytics/client-receipts`.
No credential adapter, account session, persistent device ID or host
configuration is read for this request.

The native executor sends exactly six fixed fields: schemaVersion 1, random
event UUID, source `skill`, event, OS category and UTC timestamp. It sends no
bearer/cookie, host name, model name, prompt, result, media, paths or Key.
Redirects are rejected, deadline is one second, and failures have no retry,
queue, output or effect on the actual operation result.

The machine-readable declaration is `operationsReceipts` in
[`direct-api-execution-contract.json`](direct-api-execution-contract.json),
constrained by the matching schema and engineering contract test. It is an
anonymous reporting exception, not an extra inference or credential transport.
HTTP 202 reports a new event or UUID replay; every HTTP/network failure is
ignored without retry, so delivery can be lost and counts are not complete.

Shell and PowerShell emit `installed` after committed synchronization using
the newly installed executor. Same-version verified fast returns do not count
as new installations. Init emits `connection_verified` only after explicit
credential verification. The existing identity/catalog checks keep their
two-read, shared 20-second budget; optional receipt requests run outside that
budget and may add at most one second each when opted in.

Attempted `submit`, `evaluate` and `audio` commands emit
`execution_succeeded`/`execution_failed`. Success of async submit means accepted
submission, not completed or attached output. Polling, help, preflight,
doctor, models and local validation do not create execution counts.

`operations-receipt installed` is an installer-only helper. It performs no
host credential/configuration read and silently returns if reporting is off.
It never changes normal Skill authentication or runtime dependencies.
Operators must treat these as self-reported event counts, not unique users or
verified installations. Web migration 114 and matching BFF intake must be
released before receipt data appears. Existing releases remain unchanged until
a verified new Skill release is published.
