# Maintainer Guide

End users follow README installation guidance and receive a native executable. Node and Go are repository development tools only.

## Validation

Run the authoritative engineering gate from the repository root:

```sh
npm run check
```

This gate must cover repository contracts, generated-document drift, executor behavior and actual request/receipt schema conformance. Test with isolated fixtures and local HTTP servers; never use real credentials or paid media requests as routine validation. Before release, additionally run `npm run release:validate` for catalog freshness. Fixture success does not complete real-host API, continuation or attachment-delivery acceptance; record those separately in `references/host-acceptance.json`.

When executor source changes, run `npm run executor:build` to regenerate all six artifacts and their checksums before packaging. `npm run release:package` prepares one-platform archives and a manifest pinned to the source commit. Rebuild the purposeful legacy migration archives with `npm run dist:build-legacy-migration-archive` when their installed contents change.

Before running the release-preparation workflow, rebuild and commit the executor artifacts, checksums and matching source inputs in the selected revision. That workflow tests and validates the revision, then packages its checksum-verified committed artifacts; it does not rebuild compiler-dependent binary bytes. Local packaging from modified or untracked source inputs produces draft candidates with `sourceCommit: null`, which cannot be promoted as pinned release assets.

## Contract changes

Keep request/receipt schemas, executable behavior, installed Skill instructions and command examples in the same change. Ordinary generation reads one selected profile and submits once; explicit validation, catalog filtering and diagnostics must not become mandatory preflights. Use concise user receipts while retaining safe machine metadata. Task records are optional user/workspace artifacts without prompts, credentials, reference URLs or media bytes.

Align package, Skill index, all six manifests, executor version/manifest and host-acceptance release versions. Refresh each manifest's sourceSha256 after editing SKILL.md. Update both changelogs. Preserve the exact first text installation block and its heading in both READMEs; the client download page extracts them.

## Distribution

This repository does not publish an npm package. Native fetch resolves official main to a commit, selects matching checksum-verified platform assets or the pinned source archive, then invokes native sync. GitHub Actions checks changes and prepares reviewable release candidates; publishing assets is a separate explicit action. Never delete unknown Skill directories, replace unrelated configuration or remove the historical migration bridge merely because it is old.

Exact media quotes, server idempotency and discovery of unknown submissions are server-side dependencies not implemented here. Do not claim these guarantees from validation, local records or catalog data. Keep keys, cookies and user media out of source, fixtures and release artifacts.
