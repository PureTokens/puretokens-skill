# Real-host acceptance

Run this guide on the selected release bytes, one host and OS at a time. Isolated
executor/installer tests are recorded separately and never mark real-host cells
passed. `references/host-acceptance.json` is the single acceptance record.

Before testing, record the release version, actual host version, OS version and
architecture, shell version, execution mode, executor platform and SHA-256,
date, and a workspace-relative evidence file.
Keep evidence free of credentials, connection configuration, prompts, reference
URLs and private media. Record case results and failure phase/code only; screenshots
may show only the relevant safe UI. An API task ID is not required in the checked-in
evidence.

Real-host installation changes the selected host's Skill directory; authenticated
checks use its current connection; media submissions are billable. Run each only
within authorization for that host and effect. Do not substitute a production
account, another host, or a mock result when an authorized environment is absent.
Never add paid tests to CI.

| Case | Action | Required evidence |
| --- | --- | --- |
| installation | Install or upgrade in the host, start a new conversation, invoke the loaded Skill | Actual loaded directory/version and executor hash; matching ownership inventory; unrelated files retained |
| authenticatedAPI | Query connection, models and balance once each | Separate identity/authentication outcomes; balance scope and currency; no configuration disclosed |
| image-generation | Generate one authorized image | One submission receipt; same-task completion; image opens after host attachment handoff |
| uploaded-image-edit | Edit a current local attachment | One declared edit submission; attachment preserved as input; resulting image opens |
| generated-image-edit | Explicitly edit the image just delivered | That image is the selected current attachment; one new edit task; result opens |
| multiple-images | Request variants with a model declaring `n` | Accepted count; each zero-based index handed off before fetching the next; all images open |
| video-generation | Generate one authorized video | Same task throughout; delivered video plays; no replacement submission |
| sameTaskResume | Resume an existing task after an interrupted wait/new conversation | Same ID, original operation/count retained; no new POST; reconciliation checked once on explicit resume |
| nativeAttachmentDelivery | Deliver image and video through the actual host | User can open the attached image and play the attached video; a local path or status text alone does not pass |

Use an existing safe task for continuation when possible. Network failure,
reconciliation and malformed media cases are fault-injected only into isolated
fixtures, never by changing production services. A real reconciliation outcome
can be recorded if naturally encountered; do not manufacture one with paid tasks.

For Windows, run the native installer fixtures with `powershell.exe` (5.1),
`pwsh`, and the 32-bit SysWOW64 PowerShell when present. The repository tests
these entrances on its Windows runner; this is not a record of real-host execution. That evidence is
separate from actual host discovery, authenticated API use and attachment handoff.

For each host/OS, attach an evidence entry with `host`, `os`, `hostVersion`,
`osVersion`, `osArchitecture` (`amd64` or `arm64`), `shellVersion`,
`executionMode` (`local`, `wsl`, `remote` or `sandbox`),
`executorPlatform`, `executorSha256`, `version`, `checkedAt`,
`artifact` and `cases`. `cases` uses the table's IDs and values `passed`, `failed`,
`pending` or `unavailable`. Set the four existing host/OS summary cells passed
only when a corresponding real-host case has passed. `authenticatedAPI` does
not imply media permissions or wallet coverage of subscriptions.

The host/OS summary cells summarize local evidence only; consult each evidence
entry for its exact architecture and versions. A pass on one architecture or
execution mode does not accept another. In WSL, record the Linux execution
environment, not the surrounding Windows desktop. Never copy desktop credentials
into a remote or sandbox environment to make a case pass.

Acceptance levels are derived from each individual evidence entry, not declared
manually or assembled from unrelated versions: `installation-verified` requires
the installation case; `api-verified` additionally requires authenticated API
checks; `media-verified` requires all nine cases, including playable video and
actual attachment handoff. No evidence is `unverified`. Registration and
credential fixtures do not establish any real-host level; Trae remains without
an API adapter.

Run `npm run acceptance:validate` after editing evidence. It rejects passed
summary cells without matching real-host evidence, mismatched release versions,
missing evidence artifacts and executor hashes that do not match the release.
The check reports pending cells; it does not convert them to passes.

## Platform matrix and counts

Each host has macOS, windows and linux summary slots. Pending means acceptance remains to be done, not proven client availability. This release schedules Linux targets for Claude Code, Codex, Gemini CLI, Grok Build and OpenCode; other Linux combinations are unavailable with an explicit reason until a client execution target is verified. Record Linux evidence with `os: linux` and a matching `linux-amd64` or `linux-arm64` executor. Do not substitute Windows/macOS hashes.

Four summary checks and the nine detailed cases above are distinct. The validator reports pending summary cells and the count of detailed case outcomes actually recorded; absence of evidence is never a pass.

For installation, retain separate outcomes for a clean install and an upgrade
from the preceding distributed version, default and explicit directories, spaces
and non-ASCII names, and a new conversation loading the updated instructions.
Include x86 PowerShell on amd64 Windows, native ARM64 and emulated processes as
distinct environments. For recovery, include interruption before handoff,
already-complete handoff, and a copied Windows/POSIX record with explicit output
rebinding. A copied record never supplies authentication on the new machine.

## Delivery latency measurements

Measure installation separately: stable manifest/selector resolution, archive
transfer, verification/synchronization, and the shared init deadline. Retain
same-version install and update cases with zero archive reads, zero installed
file changes and no init. A changed/missing file or unresolved transaction must
stop without a repair download. Include absent stable assets, selector/archive
checksum failures, a newer installed version and simultaneous updates.

Use only the current platform archive. Validate both Shell and PowerShell entry
paths and the absence of the other system's scripts in the archive. At least one
fresh install and repeat invocation must use an actual built ZIP rather than a
source-directory fixture. Init's two read-only checks share a 20-second deadline;
time out each phase separately without any retry or rollback of installed files.

For the selected authorized host/release, record elapsed milliseconds for:
request preparation to accepted receipt; accepted receipt to observed completed
status; observed completion to verified download; verified download to actual
openable/playable attachment handoff. Record status/content request counts and
whether the user had to repeat a continuation request. These are client-observed
intervals, not server generation-time estimates.

Keep timing evidence with the existing per-host acceptance artifact, using only
case names, durations, request counts, release/host identity and safe failure
codes. Do not store prompts, credentials, reference URLs or media bytes. Compare
single-image, local-image edit, video and interrupted-handoff recovery. Include a
normal first-window expiry, an actual network error and an interrupted process in
isolated fault-injection tests; never create production failures for measurement.
A recovery passes only when the same task is retained, no replacement POST is
sent, and the existing verified download is handed off without another content
GET. Local fixtures and virtual-clock timings remain separate from real-host
acceptance. Start with the user's two most-used accessible hosts; do not infer
which hosts or billable tasks are authorized from this guide.

## Stable publication

The manual stable publication gate requires every declared available host/OS
summary check to pass with matching real-host evidence. An API-capable target
also needs all nine detailed cases passed together in one local environment.
Unavailable targets must have an explicit reason; they are not advertised as
accepted. At least one complete media environment is required, so an empty or
all-unavailable matrix cannot authorize a stable release. These gates do not
run paid tests automatically and never fabricate missing evidence.
