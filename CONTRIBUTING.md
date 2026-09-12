# Maintainer Guide

End users follow README installation guidance and receive a native executable. Node and Go are repository development tools only.

Use CHANGELOG.md for shipped and candidate changes, and references/host-acceptance.json for outstanding real-host acceptance. Keep local handoff notes separate from public release documentation.

## Validation

Run the authoritative engineering gate from the repository root:

```sh
npm run check
```

This gate must cover repository contracts, generated-document drift, executor behavior and actual request/receipt schema conformance. Test with isolated fixtures and local HTTP servers; never use real credentials or paid media requests as routine validation. Before release, additionally run `npm run release:validate` for catalog freshness and reproducible six-platform binary verification. Fixture success does not complete real-host API, continuation or attachment-delivery acceptance; record those separately in `references/host-acceptance.json`.

After changing executor source, dependencies, build inputs or package metadata, use the toolchain in `runtime/executor/build-config.json` and run `npm run executor:build` to regenerate all six artifacts and their checksums before packaging. `npm run release:package` prepares one-platform archives and a manifest pinned to the source commit.

Before running the release-preparation workflow, rebuild and commit the executor artifacts, checksums and matching source inputs in the selected revision. That workflow tests and validates the revision, then packages its checksum-verified committed artifacts; it rebuilds all six platforms into temporary directories with the exact toolchain in `runtime/executor/build-config.json` and compares bytes before packaging the committed artifacts. Local packaging from modified or untracked source inputs produces draft candidates with `sourceCommit: null`, which cannot be promoted as pinned release assets.

Installation ownership uses complete per-directory inventories, including empty
directories. Added/modified/missing entries and symlinks block replacement and
recovery. An unmarked directory may be adopted only when its full contents match the current official source. No historical snapshots or retired-installation migration are maintained.
Inventory markers detect accidental edits, not malicious local tampering.

`npm run acceptance:validate` checks the real-host record and release hashes;
follow `references/host-acceptance-guide.md` for actual host acceptance. Pending
cells remain pending without real evidence. Windows installer fixtures already
invoke both Windows PowerShell 5.1 and PowerShell 7; preserve both entrances.

For parser changes, run the relevant bounded Go fuzz targets from
`runtime/executor`, for example `go test -run '^$' -fuzz FuzzMediaContainers
-fuzztime 10s -parallel 2`. Seed cases also run in ordinary Go tests. Do not
install a codec runtime for users: container checks are bounded, and actual video
playback remains a host acceptance item.

## Contract changes

Keep request/receipt schemas, executable behavior, installed Skill instructions and command examples in the same change. Ordinary generation reads one selected profile and submits once; explicit validation, catalog filtering and diagnostics must not become mandatory preflights. Use concise user receipts while retaining safe machine metadata. Task records are optional user/workspace artifacts without prompts, credentials, reference URLs or media bytes.

Align package, Skill index, all six manifests, executor version/manifest and host-acceptance release versions. Refresh each manifest's sourceSha256 after editing SKILL.md. Update both changelogs. Preserve the exact first text installation block and its heading in both READMEs; the client download page extracts them.

## Distribution

This repository does not publish an npm package. Native fetch resolves the latest stable release manifest, pins its version/source commit and verifies its selector and current-platform ZIP. No ordinary install follows main or falls back to a source archive. Explicit local source sync remains available for maintainers. Keep six shared Skills; do not fork payloads per client or create a global shared-executor service.

`npm run release:package` produces six OS/architecture-only archives, four bootstrap/selector script assets and the schema-v2 release manifest. `npm run release:verify` checks all archive members against current source, checksums, lengths, platform-only contents and script hashes; `--publishable` additionally requires clean committed distribution sources. Dirty candidates cannot be promoted. The same checks run on the downloaded draft assets immediately before publication.

Release preparation remains read-only with respect to GitHub releases. The separate manual `Publish verified stable release` workflow accepts an existing matching version tag, tests the sources, prepares and verifies all assets, creates a draft, uploads and re-downloads the complete asset set, and only then promotes it to latest. Failure leaves the previous public release unchanged and any new draft for maintainer review; never overwrite an existing tag/release or silently downgrade installations. Publishing is an explicit action, not a side effect of pushing main. Never delete unknown Skill directories or replace unrelated configuration.

From 0.18.1 onward, each distributed version is immutable. Increment the version before changing installed Skills, profiles, executors or installation/build inputs; same-version fast verification and downgrade protection rely on this. `release:validate` checks complete first-parent Git history, including same-version changes that were later reverted, and includes untracked/deleted distribution files. A version-bumped dirty tree is only a candidate, never proof of publishable assets. Executor test-only changes do not require a new version. Keep the gate before push and publication; main changes are no longer the default installation source.

Stable promotion also runs `node scripts/validate-host-acceptance.mjs --stable`: every declared available target needs real evidence, API-capable targets need all nine cases passed in one local environment, and an empty/all-unavailable matrix is rejected. Regular development gates still accept honest pending evidence. Run real-host acceptance only with explicit authorization; do not turn it into paid CI traffic.

Acceptance classification must match the credential adapter state in `references/host-support.json`. Hosts without an adapter keep installation acceptance separate and mark API-dependent summaries and detailed cases `unavailable` with an explicit reason, not `pending` or `passed`. Credential fixtures never substitute for real-host evidence.

Exact media quotes, server idempotency and discovery of unknown submissions are server-side dependencies not implemented here. Do not claim these guarantees from validation, local records or catalog data. Keep keys, cookies and user media out of source, fixtures and release artifacts.

Build identity is recorded in `runtime/executor/build-proof.json` and embedded in each executable. `npm run validate` checks inputs and artifact identities; `npm run executor:verify` independently rebuilds and compares all six outputs. Current-platform Go tests also execute the packaged binary offline. These development checks add no user runtime dependency.
