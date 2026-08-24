# Pure Tokens Skills

This repository provides three independent Skills:

| Skill | What it does |
| --- | --- |
| `puretokens_balance` | Reads a current-balance snapshot only when the host exposes that read-only capability. |
| `puretokens_image` | Generates images through the current configured Pure Tokens Images API. |
| `puretokens_video` | Generates videos through the current configured Pure Tokens Videos API. |

Install the Skills you need into the host's Skill directory:

```bash
node bin/puretokens-skill.js install puretokens_balance --target ~/.codex/skills
node bin/puretokens-skill.js install puretokens_image --target ~/.codex/skills
node bin/puretokens-skill.js install puretokens_video --target ~/.codex/skills
```

## Connection contract

The host's current Pure Tokens connection owns the Base URL, authentication, and routing. CC Switch, Pure Tokens Desktop, or a manually configured host connection can provide it. The Skills do not read, scan, ask for, print, or store credentials or host configuration.

`puretokens_image` uses `POST /v1/images/generations`; the default is `gpt-image-2`, and every image request sends `async: true`. For another image model, it first verifies the exact ID and `image` capability with `GET /v1/media/models`.

`puretokens_video` first uses `GET /v1/media/models`, verifies an exact `video` model ID, then uses `POST /v1/videos`. Its default is `grok-imagine-video-1.5-preview`; it polls and delivers only the same task's native bytes.

The active connection must execute those requests and deliver native image or video bytes. If it cannot, the Skill stops before a billable submission and tells the user to check the existing Pure Tokens Base URL, authentication, and routing configuration. It never falls back to another provider or execution path. If the connection is not Pure Tokens, use https://puretokensx.com/.

## Image sizes and count

Images default to one result. An explicit `n` must be an integer from 1 through 6; a request is never split into several paid submissions. Supported `size` values are `1024x1024`, `1536x1024`, and `1024x1536`; supported `image_size` values are `1K`, `2K`, and `4K`.

Physical dimensions such as `200cm × 230cm` cannot be guaranteed and are never passed as `n` or `size`. The Skill explains the limitation and asks the user to choose one of the supported options.

## Model discovery

The README is for discovery only. At execution time, the current authenticated `GET /v1/media/models` response remains authoritative for non-default images and all videos. Capabilities are taken only from the base model catalog's explicit image/video declarations, never inferred from a model name.

<!-- media-model-catalog:start -->
## Media model catalog

Synchronized with the base model catalog: 2026-08-21T02:46:19.421Z.

This list is generated from Pure Tokens' base model catalog using explicit image/video capabilities. At execution time, the exact model and required capability must still appear in the current authenticated GET /v1/media/models response.

README is generated only from base-catalog models with explicit image/video capabilities; it never infers capability from a model name. Before release, run `npm run docs:sync-media-models-from-service` against the controlled base catalog; execution still uses the authenticated `GET /v1/media/models` response.

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
node bin/puretokens-skill.js upgrade puretokens_image --target ~/.codex/skills
```

For Claude Desktop, bundle and upload the required specialist Skill:

```bash
node bin/puretokens-skill.js bundle puretokens_image --format claude-desktop --out ./puretokens_image.zip
```
