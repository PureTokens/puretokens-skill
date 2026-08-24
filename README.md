# Pure Tokens Skills

This repository provides three independent Skills:

| Skill | What it does |
| --- | --- |
| `puretokens_balance` | Reads a current-balance snapshot only when the host exposes that read-only capability. |
| `puretokens_image` | Generates images through the current configured Pure Tokens Images API. |
| `puretokens_video` | Generates videos through the current configured Pure Tokens Videos API. |

Install the Skills you need into the supported host's documented global Skill directory:

```bash
# Codex
node bin/puretokens-skill.js install puretokens_balance --target ~/.agents/skills
node bin/puretokens-skill.js install puretokens_image --target ~/.agents/skills
node bin/puretokens-skill.js install puretokens_video --target ~/.agents/skills

# Claude Code
node bin/puretokens-skill.js install puretokens_balance --target ~/.claude/skills
node bin/puretokens-skill.js install puretokens_image --target ~/.claude/skills
node bin/puretokens-skill.js install puretokens_video --target ~/.claude/skills

# Gemini CLI
node bin/puretokens-skill.js install puretokens_balance --target ~/.gemini/skills
node bin/puretokens-skill.js install puretokens_image --target ~/.gemini/skills
node bin/puretokens-skill.js install puretokens_video --target ~/.gemini/skills
```

## Host support

CC Switch is a connection-configuration tool, not a Skill host. A supported host uses whichever current connection CC Switch, Pure Tokens Desktop, or the user has already configured.

| Host | Current specialist-Skill delivery | What the user does |
| --- | --- | --- |
| Codex | Manual source install | Install the required Skill into `~/.agents/skills`. |
| Claude Code | Manual source install | Install the required Skill into `~/.claude/skills`. |
| Claude Desktop | ZIP bundle | Bundle the required specialist Skill and upload/enable it in Claude Desktop Skills settings. |
| Gemini CLI | Manual source install | Install the required Skill into `~/.gemini/skills`. |
| WorkBuddy, Grok Build, OpenCode, Trae | Not currently distributed | Their Desktop Router/configuration adapters do not imply a compatible specialist-Skill delivery. |

The canonical matrix is `references/host-support.json`. The CLI intentionally never guesses a host directory.

## Connection contract

The host's current configured connection owns the Base URL, authentication, and routing. CC Switch, Pure Tokens Desktop, or a manually configured host connection can provide it. The Skills do not read, scan, ask for, print, or store credentials or host configuration; they do not inspect a provider label, Base URL, or service attribution.

`puretokens_image` uses `POST /v1/images/generations`; the default is `gpt-image-2`, and every image request sends `async: true`. For another image model, it first verifies the exact ID and `image` capability with `GET /v1/media/models`.

`puretokens_video` first uses `GET /v1/media/models`, verifies an exact `video` model ID, then uses `POST /v1/videos`. Its default is `grok-imagine-video-1.5-preview`; it polls and delivers only the same task's native bytes.

Every supported host is held to the same native-execution contract: authenticated relative-path HTTP, JSON task responses, native media-byte delivery, and continuation by the same task ID. The acceptance matrix is `references/host-native-execution-contract.json`; it does not grant the Skill access to a Base URL, API key, or host configuration.

The active connection must execute those requests and deliver native image or video bytes. If it cannot, the Skill stops before a billable submission and tells the user to check the existing Pure Tokens Base URL, authentication, and routing configuration. It never falls back to another execution path and does not identify or branch on other relay services.

## Balance

`puretokens_balance` makes exactly one read-only `GET /api/product/desktop/account/balance` request only when the host exposes the current connection's existing authenticated account session. It reports only returned fields. If that session is not exposed, it directs the user to the current connection's client balance view; it never guesses a balance, tries another endpoint, or asks for credentials.

## Image sizes and count

Images default to one result. An explicit `n` must be an integer from 1 through 6; a request is never split into several paid submissions. Supported `size` values are `1024x1024`, `1536x1024`, and `1024x1536`; supported `image_size` values are `1K`, `2K`, and `4K`.

Physical dimensions such as `200cm × 230cm` cannot be guaranteed and are never passed as `n` or `size`. The Skill explains the limitation and asks the user to choose one of the supported options.

For `n` images, delivery reads exactly the zero-based indexes `0` through `n-1` from the same task. A request is successful only when native bytes arrive for every requested index. A partial result names both delivered and missing indexes, then permits only another read of the missing content from that same task.

## Model parameter profiles and receipts

The default `gpt-image-2` uses the count and size values declared above. For another image model, a requested optional field such as `n`, `size`, or `image_size` must be present with a supported value in that model's authenticated live `input_schema`; prompt-only image requests do not need a profile. For video, every optional duration, aspect ratio, resolution, size, or other field likewise requires the selected model's live `input_schema`; prompt-only video requests remain valid without one. A missing or incompatible profile stops before submission and asks the user to remove the option or choose a model with a published profile.

On submit, continuation, completion, and failure, media Skills return a consistent receipt: exact model ID when returned, task ID when returned, current state, requested count, requested size/parameters, delivered count on completion, and the next action. Missing task metadata is reported as not returned, never guessed.

## Asynchronous polling

Media polling begins only after submission returns a `task_id`. If a task-status response includes a valid HTTP `Retry-After` delay, the Skill uses it. Otherwise it waits 2, 3, 5, and 8 seconds for the first four same-task status reads, then 15 seconds between subsequent reads. Automatic polling lasts at most 120 seconds for images and 300 seconds for videos, measured from the submission response. A still-pending task is reported as pending at that deadline; the user can explicitly ask to continue the same task, but the Skill never treats the deadline as failure or submits a replacement task.

## Usage examples

- Image: `Use gpt-image-2 to generate a 1024x1024 illustration of a snowy village at dawn.`
- Another image model: `Use nano banana pro to generate a clean product poster.` The Skill resolves a unique installed alias, then confirms the exact ID and image capability in the current authenticated catalog.
- Video: `Use grok 1.5 video to generate a six-second cinematic sunrise over the ocean.`
- Existing task only: `Continue querying task <task_id>.` The Skill only reads that task; it never submits a replacement task automatically.

## Model discovery

The README is for discovery only. At execution time, the current authenticated `GET /v1/media/models` response remains authoritative for non-default images and all videos. Capabilities are taken only from the base model catalog's explicit image/video declarations, never inferred from a model name. Each installed image/video Skill includes its capability-specific `references/model-selection.json`, generated from this same catalog; an alias is usable only when it resolves to one exact model ID.

<!-- media-model-catalog:start -->
## Media model catalog

Synchronized with the base model catalog: 2026-08-21T02:46:19.421Z.

This list is generated from Pure Tokens' base model catalog using explicit image/video capabilities. At execution time, the exact model and required capability must still appear in the current authenticated GET /v1/media/models response.

README is generated only from base-catalog models with explicit image/video capabilities; it never infers capability from a model name. The current catalog snapshot is discovery-only; the authenticated live model and its `input_schema` win at execution time. Before release, refresh from the controlled base catalog and run `npm run release:validate`; the release gate fails when the snapshot is over seven days old.

### Image models

| Model ID | Provider | You can also say | Good for | Example |
| --- | --- | --- | --- | --- |
| `gpt-image-2` | OpenAI | `image2` | Image generation | `Use image2 to generate an image.` |
| `grok-imagine-image` | xAI | `grok image` | Image generation | `Use grok-imagine-image to generate an image.` |
| `grok-imagine-image-quality` | xAI | Exact ID only | Image generation | `Use grok-imagine-image-quality to generate an image.` |
| `nano-banana-2` | Google | `nano banana 2` | Image generation | `Use nano-banana-2 to generate an image.` |
| `nano-banana-pro` | Google | `nano banana pro` | Image generation | `Use nano-banana-pro to generate an image.` |
| `qwen-image-2.0` | Qwen | Exact ID only | Image generation | `Use qwen-image-2.0 to generate an image.` |
| `qwen-image-2.0-pro` | Qwen | Exact ID only | Image generation | `Use qwen-image-2.0-pro to generate an image.` |
| `seedream-5.0-pro` | Doubao | Exact ID only | Image generation | `Use seedream-5.0-pro to generate an image.` |
| `wan2.7-image` | Qwen | Exact ID only | Image generation | `Use wan2.7-image to generate an image.` |
| `wan2.7-image-pro` | Qwen | Exact ID only | Image generation | `Use wan2.7-image-pro to generate an image.` |

### Video models

| Model ID | Provider | You can also say | Good for | Example |
| --- | --- | --- | --- | --- |
| `grok-imagine-video` | xAI | `grok video` | Video generation | `Use grok-imagine-video to generate a video.` |
| `grok-imagine-video-1.5-preview` | xAI | `grok 1.5 video` | Video generation | `Use grok-imagine-video-1.5-preview to generate a video.` |
| `minimax-h3` | MiniMax | `minimax h3` | Video generation | `Use minimax-h3 to generate a video.` |
| `seedance-2.0` | Doubao | Exact ID only | Video generation | `Use seedance-2.0 to generate a video.` |
| `seedance-2.0-fast` | Doubao | Exact ID only | Video generation | `Use seedance-2.0-fast to generate a video.` |
| `seedance-2.0-mini` | Doubao | Exact ID only | Video generation | `Use seedance-2.0-mini to generate a video.` |
| `seedance-2.5` | Doubao | Exact ID only | Video generation | `Use seedance-2.5 to generate a video.` |

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
