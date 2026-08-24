<p align="center">
  <strong>English</strong> · <a href="./CHANGELOG.zh-CN.md">中文</a>
</p>

# Changelog

## Unreleased

- Fixed the `gpt-image-2` MCP branch: its generation call now terminates on native image content, so WorkBuddy does not make a spurious `puretokens_image_result` poll after a completed Image-2 response.
- Rebuilt the bilingual README media catalog from the global base model catalog rather than a local routing cache or an API-key-scoped response. `npm run docs:sync-media-models-from-base-catalog` accepts only explicit image/video capabilities, carries every configured model ID into the published list, and `GET /v1/media/models` remains the execution-time authorization check.

## 0.4.6 — 2026-08-20

- Kept Pure Tokens Desktop as an optional managed delivery for Codex and WorkBuddy while restoring independent Skill installation and Direct Cloud execution paths.
- Defined the third, host-native execution case for manually configured Pure Tokens media models: it preserves an explicit UI/tool selection only when the host proves the exact media capability and real delivery, otherwise falls back to the normal MCP or Direct Cloud decision tree without duplicate submission.
- Kept newly added media models immediately usable from the authenticated live catalog without waiting for a Skill update, while publishing the registered model catalog and convenience aliases so users can discover the Skill's capabilities.
- Defined Direct Cloud preflight delivery capability, API-key-scoped missing-model recovery, host-only request-ID handling, and zero-based multi-image `/content?index=N` retrieval.
- Made WorkBuddy rendering replace the managed delivery atomically so removed source files cannot leave stale behavior behind, preserved users' explicit built-in or manually configured model choices, and corrected the canonical repository URL.

## 0.4.5 — 2026-08-20

- Restored the Direct Cloud image-submit policy: image requests always send `async: true`. The execution layer still accepts a compatible synchronous `b64_json` or `url` response defensively, but the Skill never requests synchronous image generation.

## 0.4.4 — 2026-08-20

- Confirmed the Codex delivery is only the Desktop-managed generated Skill; the retired Plugin/Marketplace delivery is removed completely.
- Resolved the Direct Cloud result-mode contract: image requests no longer force `async: true`, so the execution layer classifies and delivers real synchronous `b64_json`/`url` responses or asynchronous task content according to the server response.
- Clarified that the Skill never performs I/O itself: MCP and the host Direct Cloud execution layer respectively download, write, preview, and expose genuine local-delivery evidence. When neither execution route is available, the Skill reports the missing capability without making Desktop or pasted credentials a prerequisite.
- Strengthened generated-delivery verification so Claude Desktop bundles are unpacked and compared byte-for-byte with the shared source, alongside the existing Codex and WorkBuddy provenance checks.

## 0.4.3 — 2026-08-20

- Standardized Direct Cloud credentials around the conventional **API Base URL** and **API Key** fields. Environment-backed hosts now use `PURETOKENS_API_BASE_URL` and `PURETOKENS_API_KEY`; the ambiguous `PURETOKENS_ACCESS_TOKEN` name is no longer part of the contract.

## 0.4.2 — 2026-08-20

- Replaced the Pure Tokens-specific Codex Plugin delivery with a generated managed Skill delivery. Pure Tokens Desktop now owns `~/.codex/skills/puretokens_media` and the separate `puretokens-image` MCP configuration; media generation no longer depends on the Codex Plugin feature, Marketplace setup, or Plugin unlock.
- Kept Direct Cloud independent of Desktop and clarified asynchronous multi-image delivery: completed image tasks with multiple declared content entries retrieve `/content` followed by bounded `/content?index=N` requests in response order.

## 0.4.1 — 2026-08-20

- Made `puretokens_media` request one result by default and permit a higher count only when the user explicitly provides it and the selected execution contract supports it. A request never becomes multiple generation submissions.
- Completed the Direct Cloud delivery contract: synchronous image `data[].b64_json` and `data[].url` results are decoded or downloaded into local files, asynchronous images and all videos retrieve `/content`, and no completed status is presented as success before media bytes are written. Open-file/open-folder entries and previews are now shown only when the execution layer actually returns them.
- Added a deterministic source provenance digest covering the shared Skill, manifest, agent entry, and references. The generated Codex Plugin and WorkBuddy delivery now carry that digest; Claude Desktop ZIP bundles include `source-delivery.json`, and tests verify each delivery against the shared source.

## 0.4.0 — 2026-08-20

- Clarified the execution boundary for `puretokens_media`: Skill owns natural-language policy, MCP owns typed local execution, Desktop Router remains a managed transport, and terminal-capable Agents can use the Direct Cloud contract without Desktop, Router, an extra CLI, or MCP. Desktop Router and Direct Cloud now use the same authenticated `GET /v1/media/models` response shape before calling the same image/video submit, status, and content endpoints. The Codex Plugin is a generated Skill delivery and does not bundle or start the Desktop-managed MCP.
- Added the official Codex Plugin delivery for `puretokens_media`. It uses the Desktop-managed `puretokens-image` MCP when callable, while remaining able to use Direct Cloud without that MCP.
- Made the copyable local-agent installation prompt fail closed in ordinary ChatGPT chats. A Codex runtime label alone no longer authorizes a local Skill install; the agent must first prove it has both a terminal and local file-write access.
- Updated the Nano Banana aliases to the current catalog IDs: `gemini-3.0-pro-image` and `gemini-3.1-flash-image`.

## 0.3.2 — 2026-08-14

- Hardened WorkBuddy media routing: `ToolSearch` now explicitly remains discovery-only, and the managed Skill requires `DeferExecuteTool` for every Pure Tokens catalog, generation, and result call.
- Prohibited SVG/HTML widgets, built-in media tools, search, and text claims from standing in for a successful Pure Tokens media request. A completion now requires MCP-returned model and native-result or local-delivery evidence.

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
