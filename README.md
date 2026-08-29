<p align="center">
  <img src="./assets/brand/puretokens-skill-hero.png" alt="Pure Tokens Official Skills" width="100%" />
</p>

# Pure Tokens Skills

This repository provides six independent Skills:

| Skill | What it does |
| --- | --- |
| `puretokens_balance` | Reads a current-balance snapshot from the fixed Pure Tokens API. |
| `puretokens_connection` | Checks the fixed Pure Tokens API identity declaration, without reading connection configuration. |
| `puretokens_models` | Reads the fixed API's authenticated media catalog and explains declared model capabilities, parameters, and media operations. |
| `puretokens_image` | Generates images and performs profile-gated image edits through the fixed Pure Tokens Images API. |
| `puretokens_video` | Generates videos and performs profile-gated image, video, or audio reference generation and video edits through the fixed Pure Tokens Videos API. |
| `puretokens_update` | Installs or safely upgrades local official Pure Tokens Skills. |

Install the Skills you need into the supported host's documented global Skill directory:

```bash
# Codex
node bin/puretokens-skill.js install puretokens_balance --target ~/.agents/skills
node bin/puretokens-skill.js install puretokens_connection --target ~/.agents/skills
node bin/puretokens-skill.js install puretokens_models --target ~/.agents/skills
node bin/puretokens-skill.js install puretokens_image --target ~/.agents/skills
node bin/puretokens-skill.js install puretokens_video --target ~/.agents/skills
node bin/puretokens-skill.js install puretokens_update --target ~/.agents/skills

# Claude Code
node bin/puretokens-skill.js install puretokens_balance --target ~/.claude/skills
node bin/puretokens-skill.js install puretokens_connection --target ~/.claude/skills
node bin/puretokens-skill.js install puretokens_models --target ~/.claude/skills
node bin/puretokens-skill.js install puretokens_image --target ~/.claude/skills
node bin/puretokens-skill.js install puretokens_video --target ~/.claude/skills
node bin/puretokens-skill.js install puretokens_update --target ~/.claude/skills

# Gemini CLI
node bin/puretokens-skill.js install puretokens_balance --target ~/.gemini/skills
node bin/puretokens-skill.js install puretokens_connection --target ~/.gemini/skills
node bin/puretokens-skill.js install puretokens_models --target ~/.gemini/skills
node bin/puretokens-skill.js install puretokens_image --target ~/.gemini/skills
node bin/puretokens-skill.js install puretokens_video --target ~/.gemini/skills
node bin/puretokens-skill.js install puretokens_update --target ~/.gemini/skills
```

## Agent-assisted installation

The client download page extracts the first fenced block under the fixed heading below. Keep that heading and its one `text` block stable; do not place another fenced block in this section.

### Copy this to a terminal-capable local agent

```text
Install or update the official Pure Tokens Skills for this local agent host from the only authorized source: `https://github.com/PureTokens/puretokens-skill.git` (branch `main`). Do the work; do not only describe it. Do not substitute a package, mirror, fork, or similarly named repository.

1. Identify the current host from this runtime:
   - Codex: `$env:USERPROFILE\.agents\skills` on Windows PowerShell; `~/.agents/skills` on macOS/Linux.
   - Claude Code: `~/.claude/skills`.
   - Gemini CLI: `~/.gemini/skills`.
   If the current host is not one of these or cannot be identified, ask me which host I use before changing files.

2. If a terminal, PowerShell, exec, or shell tool is available, use it before replying. Create a new temporary working directory without deleting an existing directory. Clone `https://github.com/PureTokens/puretokens-skill.git` from branch `main`, enter it, and run `npm run check`.

3. If `npm run check` fails, report the failure and stop. Do not change any Skill directory.

4. Sync all six Skills — `puretokens_balance`, `puretokens_connection`, `puretokens_models`, `puretokens_image`, `puretokens_video`, `puretokens_update` — in the selected installation root with `node bin/puretokens-skill.js sync --target <installation-root>`. It installs missing official Skills and upgrades only matching managed Skills. If any same-name existing destination is not a managed Skill, it must stop before changing anything, leave the conflict untouched, and report it.

5. Run `npm run check` again. Verify every requested destination has `SKILL.md` and `skill.json`, and each manifest name matches its directory name. Report installed or upgraded paths and any conflicts. Tell me to start a new host conversation before testing.

Safety: never read, display, copy, change, or ask for API keys, Base URLs, authentication files, model settings, MCP settings, AI-client configuration, Pure Tokens connection configuration, environment variables, shell profiles, or system proxies. Do not use third-party package mirrors or delete files.

If terminal execution is unavailable or denied, say so and provide the next copyable official command for my operating system and terminal. Only check for Node.js LTS if `npm run check` or the installer actually requires it. Advance one verifiable step per reply.
```

## Host support

CC Switch is a connection-configuration tool, not a Skill host. CC Switch, Pure Tokens Desktop, or a manual client setup can make existing Pure Tokens authentication available to the active runtime. The Skills themselves always call the fixed public Pure Tokens API origin; they never read the user's connection configuration.

| Host | Current specialist-Skill delivery | What the user does |
| --- | --- | --- |
| Codex | Manual source install | Install the required Skill into `~/.agents/skills`. |
| Claude Code | Manual source install | Install the required Skill into `~/.claude/skills`. |
| Claude Desktop | ZIP bundle | Bundle the required specialist Skill and upload/enable it in Claude Desktop Skills settings. |
| Gemini CLI | Manual source install | Install the required Skill into `~/.gemini/skills`. |
| WorkBuddy, Grok Build, OpenCode, Trae | Not currently distributed | Their Desktop Router/configuration adapters do not imply a compatible specialist-Skill delivery. |

The canonical matrix is `references/host-support.json`. The CLI intentionally never guesses a host directory.

## Direct API contract

The API-facing Skills call the fixed public API origin `https://api.puretokensx.com`. They use full URLs, never a configured Base URL or a relative-path host executor. The active runtime supplies its existing Pure Tokens request authentication automatically; Skills never read, scan, ask for, print, copy, or store API keys, Base URLs, authentication files, provider labels, or client configuration. They never construct an authorization header, use MCP, a local proxy, a sidecar, or another endpoint. `puretokens_update` is local-only and does not call API endpoints.

`puretokens_connection` makes exactly one read-only `GET https://api.puretokensx.com/v1` request. It confirms the fixed API only when that endpoint declares `status: "ok"`, `name: "Pure Tokens API"`, and `base_url: "/v1"`. This does not reveal or validate the user's configured Base URL, and it is not cryptographic anti-spoof verification.

`puretokens_models` makes exactly one read-only `GET https://api.puretokensx.com/v1/media/models` request. It exposes the authenticated catalog in a user-readable form: exact model IDs, returned capabilities, declared optional parameters, and declared media operations. It can shortlist models for an explicit technical requirement such as image-to-video, a reference medium, duration, aspect ratio, or resolution, but only when that requirement is explicitly declared by the live model profile. It never submits media work, retries the catalog request, falls back to the static README catalog, or ranks unreturned quality, price, speed, or availability information.

`puretokens_image` reads `GET https://api.puretokensx.com/v1/media/models` before every new task, including its default `gpt-image-2`, then uses `POST https://api.puretokensx.com/v1/images/generations` with `async: true`. Every non-core image field — count, pixel size, semantic image size, aspect ratio, reference field, and strength — must be declared by that exact authenticated profile. An explicitly supplied public HTTPS reference URL may be sent only in a profile property whose declared transport allows it. An attached native image may be sent only when `input_schema.operations.image_edit` declares `POST` to `/v1/images/generations` or `/v1/images/edits`, multipart input, fields, counts, and transport; the Skill combines that declared path with the fixed API origin.

`puretokens_video` first uses `GET https://api.puretokensx.com/v1/media/models`, verifies an exact `video` model ID, then uses `POST https://api.puretokensx.com/v1/videos`. Its default is `grok-imagine-video-1.5-preview`; it polls and delivers only the same task's native bytes. Prompt requirement and every optional field come from the exact live profile. When `constraints.resolution_by_mode` is declared, the selected text, image, or reference operation must use that mode's resolution set rather than the broader `resolution` property. An explicitly supplied public HTTPS media URL, file ID, or voice ID may use a declared profile property and declared transport; native attached media instead requires the corresponding declared multipart operation (`image_to_video`, `reference_image_video`, `reference_video`, `reference_audio`, or `video_edit`). Video editing uses `POST https://api.puretokensx.com/v1/videos/edits` only when the live profile declares it.

The full direct-API contract is `references/direct-api-execution-contract.json`. If a direct request fails before a task is accepted, the Skill reports the returned failure and does not guess a Base URL, authentication, or routing cause. It never falls back to another execution path or identifies another relay.

## Balance

`puretokens_balance` makes exactly one read-only `GET https://api.puretokensx.com/api/product/desktop/account/balance` request with the active runtime's existing Pure Tokens account authentication. It reports only returned fields. If the direct request is not authenticated or fails, it reports the returned result and directs the user to the Pure Tokens client balance view; it never guesses a balance, tries another endpoint, or asks for credentials.

## Skill updates

`puretokens_update` handles explicit requests to install, update, or synchronize local official Skills. On Codex, Claude Code, and Gemini CLI it validates a fresh `main` checkout, then runs `node bin/puretokens-skill.js sync --target <installation-root>`. The command installs missing official Skills and upgrades only managed matching Skill directories; an unmanaged same-name directory stops the whole sync before any target is changed. On Claude Desktop it builds new ZIP bundles and guides the user through uploading and enabling them. It never reads connection settings or credentials, and it never runs automatically during media work.

## Image sizes and count

An image request is never split into several paid submissions. Count and every size control are model-specific: `n`, `size`, `image_size`, `aspect_ratio`, `width`, and `height` may be sent only when the selected model's current authenticated profile declares that exact field and value. If `n` is not declared, the Skill does not invent it.

Physical dimensions such as `200cm × 230cm` cannot be guaranteed and are never passed as `n`, `size`, or another API field. The Skill explains the limitation and lists the exact model's declared pixel or semantic-size choices.

For `n` images, delivery reads exactly the zero-based indexes `0` through `n-1` from the same task, one index at a time and only after the task succeeds. A request is successful only when native bytes arrive for every requested index. A partial result names both delivered and missing indexes, then permits only another read of the missing content from that same task. The Skill never prefetches or re-downloads delivered content; it hands off each native result before reading the next.

## Model parameter profiles and receipts

Every new image and video task uses the selected model's authenticated live `input_schema`; the static selection list is only for aliases. Any requested optional field must be present with a compatible value. Video prompt is required when the profile says so; it may be omitted only for the exact single-reference exception explicitly declared by that profile. A missing or incompatible profile stops before submission and asks the user to remove the option or choose a model with a published profile.

The same live profile controls media inputs. An explicitly supplied public HTTPS URL, file ID, or voice ID is sent only in its exact declared property and permitted transport; the Skill never downloads, probes, checks accessibility, rehosts, or rewrites it. Native media explicitly attached in the current request uses only an advertised `multipart_file` operation. The Skill sends those bytes with the one declared Images or Videos API request; the Pure Tokens gateway then performs short-lived internal R2 staging, verifies its provider-facing HTTPS URL, and does not return that URL. Multiple native attachment types need an explicitly declared combined operation; multiple public URL/ID fields need no declared conflict. The Skill never manufactures a URL or file ID, calls a separate upload API, or silently turns a media request into text generation.

On submit, continuation, completion, and failure, media Skills return a consistent receipt: exact model ID when returned, task ID when returned, current state, requested operation, requested count, requested size/parameters, delivered count on completion, and the next action. Missing task metadata is reported as not returned, never guessed.

## Asynchronous polling

Media polling begins only after submission returns a `task_id` and runs only within the submission turn or a user turn that explicitly continues that same task ID. It has at most one status request in flight per task and never creates a background timer, queue, or worker. A valid positive HTTP `Retry-After` is used only while time remains in the automatic-polling budget. Otherwise an image task waits `3, 6, 12, 24, 30, 30` seconds before at most six same-task status reads; a video task waits `5, 10, 20, 40, 60, 60` seconds before at most seven reads. Each bounded window lasts at most 120 seconds for images or 300 seconds for videos. A rate limit, 5xx response, transport error, or timeout stops that window immediately. A still-pending task is reported as pending with its task ID. When the user explicitly asks to continue it, the Skill opens one new bounded window for that same task only; it never treats a deadline or read error as failure and never submits a replacement task.

Media bytes are not cached in Skill state, prompts, or logs. Content is read only after terminal success, with one content read in flight; the active runtime hands off the native bytes before another read. If it cannot do that without unbounded background work, duplicate reads, or cached copies, the Skill reports same-task delivery as unavailable instead of substituting a URL or submitting a new task.

## Usage examples

- Connection: `Can this Pure Tokens Skill confirm its API?` The Skill checks only the fixed endpoint's `GET https://api.puretokensx.com/v1` declaration and does not reveal configuration.
- Models: `Show the video models currently available to me, their declared duration and aspect-ratio options, and which support image-to-video.` This is read-only; it does not submit a task.
- Image: `Use gpt-image-2 to generate a 2K, 16:9 illustration of a snowy village at dawn.`
- Another image model: `Use nano banana pro to generate a clean product poster.` The Skill resolves a unique installed alias, then confirms the exact ID and image capability in the current authenticated catalog.
- Image reference URL: `Use gpt-image-2 with this public reference image URL: https://example.com/reference.png` The Skill first confirms a matching profile field and declared URL transport.
- Image edit: `Use grok-imagine-image to edit the attached image: replace the cloudy sky with a clear sunset.` The Skill first confirms the authenticated image-edit profile and direct multipart attachment delivery.
- Video: `Use grok 1.5 video to generate a six-second cinematic sunrise over the ocean.`
- Image-to-video: `Use grok 1.5 video to animate this public image URL for six seconds: https://example.com/reference.png` The Skill uses it only when the authenticated profile declares the matching URL field and transport.
- Reference video: `Use seedance-2.5 to create a six-second video from my attached video.` The Skill uses it only if the authenticated profile publishes `reference_video`.
- Reference audio: `Use minimax h3 to create a video from my attached audio.` The Skill uses it only if the authenticated profile publishes `reference_audio`.
- Video edit: `Edit my attached video: turn daylight into night.` The Skill submits to `/v1/videos/edits` only if the authenticated profile publishes `video_edit`.
- Existing task only: `Continue querying task <task_id>.` The Skill only reads that task; it never submits a replacement task automatically.
- Update: `Upgrade my Pure Tokens Skills.` The update Skill validates the official `main` checkout and safely synchronizes the local Skill directory.

## Model discovery

Use `puretokens_models` when the user asks what is actually available through Pure Tokens, which models support a media operation, or which models accept a particular declared parameter. Its authenticated `GET https://api.puretokensx.com/v1/media/models` response is the runtime source of truth: it reports exact model IDs, capabilities, optional parameter schema, and `input_schema.operations` without guessing missing fields. A compatibility shortlist is technical only; it matches declared capability, field/value, and operation metadata and does not make subjective quality or price claims.

The README is discovery-only. Capabilities are taken only from the base model catalog's explicit image/video declarations, never inferred from a model name. Each installed image/video Skill includes its capability-specific `references/model-selection.json`, generated from this same catalog; an alias is usable only when it resolves to one exact model ID.

<!-- media-model-catalog:start -->
## Media model catalog

Synchronized with the base model catalog: 2026-08-29T05:03:13.833Z.

This list is generated from Pure Tokens' base model catalog using explicit image/video capabilities. At execution time, the exact model and required capability must still appear in the current authenticated GET /v1/media/models response.

README is generated only from base-catalog models with explicit image/video capabilities; it never infers capability from a model name. The current catalog snapshot is discovery-only; the authenticated live model and its `input_schema` win at execution time. Before release, refresh from the controlled base catalog and run `npm run release:validate`; the release gate fails when the snapshot is over seven days old.

### Image models

| Model ID | Provider | You can also say | Good for | Example |
| --- | --- | --- | --- | --- |
| `gpt-image-2` | OpenAI | `image2` | Image generation | `Use gpt-image-2 to generate an image.` |
| `grok-imagine-image` | xAI | `grok image` | Image generation | `Use grok-imagine-image to generate an image.` |
| `grok-imagine-image-2.0` | xAI | `grok image 2.0` | Image generation | `Use grok-imagine-image-2.0 to generate an image.` |
| `grok-imagine-image-quality` | xAI | Exact ID only | Image generation | `Use grok-imagine-image-quality to generate an image.` |
| `nano-banana-2` | Google | `nano banana 2` | Image generation | `Use nano-banana-2 to generate an image.` |
| `nano-banana-2-lite` | Google | Exact ID only | Image generation | `Use nano-banana-2-lite to generate an image.` |
| `nano-banana-pro` | Google | `nano banana pro` | Image generation | `Use nano-banana-pro to generate an image.` |
| `seedream-5.0-pro` | ByteDance | Exact ID only | Image generation | `Use seedream-5.0-pro to generate an image.` |

### Video models

| Model ID | Provider | You can also say | Good for | Example |
| --- | --- | --- | --- | --- |
| `grok-imagine-video` | xAI | `grok video` | Video generation | `Use grok-imagine-video to generate a video.` |
| `grok-imagine-video-1.5` | xAI | Exact ID only | Video generation | `Use grok-imagine-video-1.5 to generate a short video.` |
| `grok-imagine-video-1.5-preview` | xAI | `grok 1.5 video` | Video generation | `Use grok-imagine-video-1.5-preview to generate a video.` |
| `minimax-h3` | MiniMax | `minimax h3` | Video generation | `Use minimax-h3 to generate a video.` |
| `seedance-2.0` | ByteDance | Exact ID only | Video generation | `Use seedance-2.0 to generate a video.` |
| `seedance-2.0-fast` | ByteDance | Exact ID only | Video generation | `Use seedance-2.0-fast to generate a video.` |
| `seedance-2.0-mini` | ByteDance | Exact ID only | Video generation | `Use seedance-2.0-mini to generate a video.` |
| `seedance-2.5` | ByteDance | Exact ID only | Video generation | `Use seedance-2.5 to generate a video.` |
| `wan3.0-video` | Qwen | `wan3 video`, `wan 3 video` | Video generation | `Use wan3.0-video to generate a short video.` |
| `wan3.0-video-prime` | Qwen | `wan3 video prime`, `wan 3 video prime` | Video generation | `Use wan3.0-video-prime to generate a short video.` |

<!-- media-model-catalog:end -->

## Updating

Pull the current repository and run the matching command for each installed Skill:

```bash
node bin/puretokens-skill.js upgrade puretokens_image --target ~/.agents/skills
```

For Claude Desktop, bundle and upload the required specialist Skill:

```bash
node bin/puretokens-skill.js bundle puretokens_image --format claude-desktop --out ./puretokens_image.zip
```

Before publishing a release:

```bash
npm run docs:sync-media-models-from-service
npm run release:validate
```
