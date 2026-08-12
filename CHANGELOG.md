# Changelog

## Unreleased — 2026-08-11

- Added `puretokens_get_balance`, a read-only Desktop-published balance snapshot with Chinese and English usage guidance. Balance checks do not inspect media models or any credential material.
- Made `README.md` the English-first entry point with a top-level English/中文 switch and added the Chinese README mirror.
- Added the Pure Tokens Skill brand hero asset generated with Image-2 and the official Pure Tokens icon.
- Added model-capability tables, usage examples, and copyable agent installation prompts in both languages.
- Reworked the README for first-time users with separate image/video model tables and plain-language examples.
- Added the Skill-owned natural-language alias registry, including `image2` → `gpt-image-2`, with live-catalog verification and no silent fallback.
- Added deterministic defaults: image requests without a model use `gpt-image-2`; video requests without a model use `grok-imagine-video-1.5`.
- Added the Nano Banana family aliases: `Nano Banana Pro` resolves to `gemini-3-pro-image-preview`; `Nano Banana 2` resolves to `gemini-3.1-flash-image-preview`; the unqualified family name asks for a choice when both are available.

## 0.2.0 — 2026-08-09

- Made the media Skill enforce catalog-first model selection, exact capability matching, stable request IDs, single submission, same-task polling, and no automatic fallback after errors or timeouts.
- Added Claude Desktop ZIP distribution guidance and a portable bundle command.
- Added safe upgrade and explicit-confirmation uninstall commands.
- Added behavior scenarios for ambiguous models, empty catalogs, unavailable MCP, failed tasks, and polling timeouts.

## 0.1.0 — 2026-08-09

- Created the Pure Tokens skill management repository.
- Added `puretokens_media`, the shared skill for selecting and invoking Pure Tokens image and video MCP tools.
- Added a versioned skill registry, machine-readable manifest, media model catalog schema, safe local installer and repository validation.
## 0.2.4

- Add exact model price lookup through `puretokens_get_model_price`.
- Report every selected group-specific price without inference or substitution.
