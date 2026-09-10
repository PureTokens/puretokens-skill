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

This repository does not publish an npm package. Native fetch resolves official main to a commit, selects matching checksum-verified platform assets or the pinned source archive, then invokes native sync. GitHub Actions checks changes and prepares reviewable release candidates; publishing assets is a separate explicit action. Never delete unknown Skill directories or replace unrelated configuration. Retired migration archives and Node installation commands are not maintained.

Exact media quotes, server idempotency and discovery of unknown submissions are server-side dependencies not implemented here. Do not claim these guarantees from validation, local records or catalog data. Keep keys, cookies and user media out of source, fixtures and release artifacts.

Build identity is recorded in `runtime/executor/build-proof.json` and embedded in each executable. `npm run validate` checks inputs and artifact identities; `npm run executor:verify` independently rebuilds and compares all six outputs. Current-platform Go tests also execute the packaged binary offline. These development checks add no user runtime dependency.
