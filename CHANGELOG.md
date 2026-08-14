<p align="center">
  <strong>English</strong> · <a href="./CHANGELOG.zh-CN.md">中文</a>
</p>

# Changelog

## 0.3.1 — 2026-08-14

- Made `puretokens_media` the single source for Claude Desktop and WorkBuddy media behavior. WorkBuddy's always-on routing payload is now generated from that shared source, including the same catalog, exact-model, single-submission, polling, delivery, and failure rules.
- Removed the independent WorkBuddy router Skill source. Pure Tokens Desktop continues to install the generated adapter automatically after **Verify and apply**; users do not upload it manually.
- Removed stale balance and model-price tool claims from the media Skill. It now advertises only the five media MCP tools that the local sidecar actually exposes.

## 0.3.0 — 2026-08-14

- Added `puretokens_workbuddy_router`, a lightweight always-on WorkBuddy Skill that routes ordinary image and video requests to the configured Pure Tokens MCP before WorkBuddy's built-in media tools. It preserves an explicit user choice of a WorkBuddy built-in tool, keeps normal text/code requests untouched, and uses catalog-first exact model selection with no silent fallback.
- WorkBuddy's router is intended to be installed and upgraded by Pure Tokens Desktop as a managed integration; users do not upload or enable it manually.

## 0.2.7 — 2026-08-14

- Completed media reports now include the exact MCP-returned model ID, filename, and durable `Downloads/Pure Tokens` delivery state.
- Added bounded native MCP video resources for hosts that can render them. Large videos stay successful local MP4 deliveries without forcing an oversized stdio response.
- Result polling now carries the original exact model after an MCP restart rather than guessing a default route.
- Clarified that a target media model must be included in a group selected for the target client, followed by Verify and apply, client restart, and a new chat.

## 0.2.6 — 2026-08-13

- Completed image and video workflows now use the MCP-delivered local `Downloads/Pure Tokens` file as the durable delivery path, not a temporary preview URL.
- Image success requires actual native MCP image content before the Skill may claim that an in-chat preview is available; the Skill no longer invents an “image above” state.

## 0.2.4 — 2026-08-12

- Added `puretokens_get_balance`, a read-only Desktop-published balance snapshot with Chinese and English usage guidance. Balance checks do not inspect media models or any credential material.
- Made `README.md` the English-first entry point with a top-level English/中文 switch and added the Chinese README mirror.
- Added the Pure Tokens Skill brand hero asset generated with Image-2 and the official Pure Tokens icon.
- Added model-capability tables, usage examples, and copyable agent installation prompts in both languages.
- Reworked the README for first-time users with separate image/video model tables and plain-language examples.
- Added the Skill-owned natural-language alias registry, including `image2` → `gpt-image-2`, with live-catalog verification and no silent fallback.
- Added deterministic defaults: image requests without a model use `gpt-image-2`; video requests without a model use `grok-imagine-video-1.5`.
- Added the Nano Banana family aliases: `Nano Banana Pro` resolves to `gemini-3-pro-image-preview`; `Nano Banana 2` resolves to `gemini-3.1-flash-image-preview`; the unqualified family name asks for a choice when both are available.
- Added exact model price lookup through `puretokens_get_model_price`.
- Reports every selected group-specific price without inference or substitution.

## 0.2.0 — 2026-08-09

- Made the media Skill enforce catalog-first model selection, exact capability matching, stable request IDs, single submission, same-task polling, and no automatic fallback after errors or timeouts.
- Added Claude Desktop ZIP distribution guidance and a portable bundle command.
- Added safe upgrade and explicit-confirmation uninstall commands.
- Added behavior scenarios for ambiguous models, empty catalogs, unavailable MCP, failed tasks, and polling timeouts.

## 0.1.0 — 2026-08-09

- Created the Pure Tokens skill management repository.
- Added `puretokens_media`, the shared skill for selecting and invoking Pure Tokens image and video MCP tools.
- Added a versioned skill registry, machine-readable manifest, media model catalog schema, safe local installer and repository validation.
