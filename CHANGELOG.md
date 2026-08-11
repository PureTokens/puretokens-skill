# Changelog

## 0.2.0 — 2026-08-09

- Made the media Skill enforce catalog-first model selection, exact capability matching, stable request IDs, single submission, same-task polling, and no automatic fallback after errors or timeouts.
- Added Claude Desktop ZIP distribution guidance and a portable bundle command.
- Added safe upgrade and explicit-confirmation uninstall commands.
- Added behavior scenarios for ambiguous models, empty catalogs, unavailable MCP, failed tasks, and polling timeouts.

## 0.1.0 — 2026-08-09

- Created the private Pure Tokens skill management repository.
- Added `puretokens_media`, the shared skill for selecting and invoking Pure Tokens image and video MCP tools.
- Added a versioned skill registry, machine-readable manifest, media model catalog schema, safe local installer and repository validation.
