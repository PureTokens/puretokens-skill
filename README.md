<p align="center">
  <img src="./assets/brand/puretokens-skill-hero.png" alt="Pure Tokens Official Skills" width="100%" />
</p>

# Pure Tokens Skills

This repository provides six independent Skills:

| Skill | What it does |
| --- | --- |
| `puretokens-balance` | Reads a current-balance snapshot from the fixed Pure Tokens API. |
| `puretokens-connection` | Checks the fixed Pure Tokens API identity declaration without exposing connection configuration. |
| `puretokens-models` | Reads the fixed API's authenticated media catalog and explains declared model capabilities, parameters, and media operations. |
| `puretokens-image` | Generates images and performs profile-gated image edits through the fixed Pure Tokens Images API. |
| `puretokens-video` | Generates videos and performs profile-gated image, video, or audio reference generation and video edits through the fixed Pure Tokens Videos API. |
| `puretokens-update` | Installs or safely upgrades local official Pure Tokens Skills. |

Install or update through a terminal-capable local Agent. It retrieves a fresh official checkout of this repository's `main` branch, validates it, then runs the repository's native source-only sync script for the current host. There is no product install ZIP, package manager, Node, npm, or user configuration step.

The source-only sync scripts are `runtime/puretokens-skill-install.sh` for macOS/Linux and `runtime/puretokens-skill-install.ps1` for Windows. They run from that fresh checkout (or receive it by absolute `--source` / `-Source` path), derive the documented host target, verify managed files before writing, and report the synchronized version.

## Agent-assisted installation

The client download page extracts the first fenced block under the fixed heading below. Keep that heading and its one `text` block stable; do not place another fenced block in this section.

### Copy this to a terminal-capable local agent

```text
Install or update the official Pure Tokens Skills from https://github.com/PureTokens/puretokens-skill.
```

## Host support

CC Switch is a connection-configuration tool, not a Skill host. The Skills always call the fixed public Pure Tokens API origin. The managed runtime narrowly resolves a matching Pure Tokens model credential only in memory for a fixed request; it never displays, copies, persists, or asks for the credential.

| Host | Current specialist-Skill delivery | Direct API execution | What the user does |
| --- | --- | --- | --- |
| Claude Code | Manual source install | Verified managed runtime | Install the required Skill into `~/.claude/skills`. |
| Codex | Manual source install | Verified managed runtime | Install the required Skill into `~/.agents/skills`. |
| WorkBuddy | Manual source install | Verified managed runtime | Install the required Skill into `~/.workbuddy/skills`. |
| Gemini CLI | Manual source install | Verified managed runtime | Install the required Skill into `~/.gemini/skills`. |
| Grok Build | Manual source install | Verified managed runtime | Install the required Skill into `~/.grok/skills`. |
| OpenCode | Manual source install | Verified managed runtime | Install the required Skill into `~/.config/opencode/skills`. |
| Trae | Manual source install | Manual credential setup | Install the required Skill into `~/.trae/skills`. |

The canonical matrix is `references/host-support.json`. The CLI intentionally never guesses a host directory.

## Direct API contract

The API-facing Skills call the fixed public API origin `https://api.puretokensx.com`. They use full URLs, never a user-selected Base URL or fallback endpoint. The managed `.puretokens-runtime/puretokens-direct-api.mjs` is installed beside the Skills. On Claude Code, Codex, WorkBuddy, Gemini CLI, Grok Build, and OpenCode, it reads only one matching Pure Tokens credential from that host's documented fixed configuration source. Sources normally target `https://api.puretokensx.com/v1` (or the required origin-only form); WorkBuddy may instead store its per-model resource URL beneath that same fixed origin as `/v1/...`, with no query or fragment. That source check binds the credential but is never reused as a request target. It retains the credential only in memory and sends it only as authorization for an allowed fixed Pure Tokens API path. It never prints, copies, persists, or asks for the key; it accepts no arbitrary URL and uses no MCP, proxy, or sidecar. `puretokens-update` is local-only and does not call API endpoints.

The direct contract requires authenticated full-URL HTTPS requests, JSON task responses, same-task status reads, and native-media byte delivery. Claude Code, Codex, WorkBuddy, Gemini CLI, Grok Build, and OpenCode use the managed local runtime for this binding. Every managed-runtime POST uses one bounded canonical Base64 argument rather than standard input; the runtime decodes it only in memory, rejects standard-input bodies immediately, enforces a 90-second total execution deadline, and returns a sanitized structured JSON failure envelope when it cannot complete. Trae remains a manual-configuration-only exception because it has no approved local credential resolver; the Skill stops safely there rather than reading or guessing its user state.

`puretokens-connection` makes exactly one read-only `GET https://api.puretokensx.com/v1` request. It confirms the fixed API only when that endpoint declares `status: "ok"`, `name: "Pure Tokens API"`, and `base_url: "/v1"`. The managed runtime may use a credential internally for this request, but the Skill never reveals or reports the user's configured Base URL or credential. This is not cryptographic anti-spoof verification.

`puretokens-models` makes exactly one read-only `GET https://api.puretokensx.com/v1/media/models` request. It exposes the authenticated catalog in a user-readable form: exact model IDs, returned capabilities, declared optional parameters, and declared media operations. It can shortlist models for an explicit technical requirement such as image-to-video, a reference medium, duration, aspect ratio, or resolution, but only when that requirement is explicitly declared by the live model profile. It never submits media work, retries the catalog request, falls back to the static README catalog, or ranks unreturned quality, price, speed, or availability information.

`puretokens-image` submits normal text-to-image work directly to `POST https://api.puretokensx.com/v1/images/generations` with `async: true`; it does not read the model catalog first. The installed versioned selection resolves the default `gpt-image-2` for normal generation and declared native-attached image edits, unique aliases, and known image parameters. The live catalog is read once only when the user explicitly asks for current model information, asks for an option or media operation absent from the installed selection, or needs a diagnosis after a model/parameter/capability rejection. A failed catalog read never blocks an otherwise valid core text-to-image submission, and it never causes automatic resubmission.

`puretokens-video` submits normal text-to-video work directly to `POST https://api.puretokensx.com/v1/videos`; it does not read the model catalog first. Its installed selection resolves the default `grok-imagine-video-1.5-preview`, unique aliases, and known parameters and media operations. The live catalog is read once only for explicit current-model discovery, an installed-profile gap needed to honor a requested option or media operation, or post-rejection diagnosis. A catalog failure never blocks an otherwise valid core text-to-video submission, and it never causes automatic resubmission. The Skill polls and delivers only the same task's native bytes.

The full direct-API contract is `references/direct-api-execution-contract.json`. If a direct request fails before a task is accepted, the Skill reports the returned failure and does not guess a Base URL, authentication, or routing cause. It never falls back to another execution path or identifies another relay.

## Model access groups

`GET https://api.puretokensx.com/v1/media/models` returns only models covered by the current managed key. If a requested exact model such as `minimax-h3` is not returned after the permitted diagnosis, the Skill says that the current connection did not return that model and does not submit or silently switch models. When the user expects access, it directs them to the Pure Tokens client configuration to select a group containing that exact model, create or select a managed key covering the selected groups, verify and apply it to the current host, then start a new host conversation and explicitly retry.

The Skill never guesses a group name or claims which group contains a model: the authenticated media catalog does not expose that mapping unless the API explicitly returns it.

## Balance

`puretokens-balance` makes exactly one read-only `GET https://api.puretokensx.com/api/product/desktop/account/balance` request through the configured-credential direct runtime where available. It reports only returned fields. If the direct request fails, it reports the returned result and directs the user to the Pure Tokens client balance view; it never guesses a balance, tries another endpoint, or asks for credentials.

## Skill updates

`puretokens-update` handles explicit requests to install, update, or synchronize local official Skills. The Agent first retrieves a fresh official checkout of this repository's `main` branch, then runs the matching source-only native sync script from that checkout with the current host ID. The script derives the documented global Skill directory, statically validates the checkout, and synchronizes in one local operation. It never downloads a custom install payload, runs an already-installed old updater, reuses user configuration, or makes a second package download. Before a Codex sync, it lists plugins and, when the exact legacy `puretokens-media` plugin is present, removes its exact selector and lists again to prove it is gone. If inspection, removal, or verification cannot complete, the sync stops without a success receipt; the user must remove it in Codex Plugins or through the workspace administrator and run the sync again. When it was removed, fully quit and restart Codex before opening a new test conversation: an already-open conversation retains its loaded legacy instructions. After all checks succeed, it removes verified retired managed Skill directories and old matching hidden retired backups, installs missing official Skills, and upgrades only managed matching Skill directories; an unmanaged current or retired same-name directory, or an unmanaged runtime directory, stops the whole sync before any target is changed. Only the exact versioned success receipt confirms completion. It never reads connection settings or credentials, and it never runs automatically during media work.

## Image sizes and count

An image request is never split into several paid submissions. Count and every size control are model-specific: `n`, `size`, `image_size`, `aspect_ratio`, `width`, and `height` may be sent only when the installed exact-model selection, or a permitted one-time live lookup for a missing requested option, declares that field and value. If `n` is not declared, the Skill does not invent it.

## Image request experience

`puretokens-image` distinguishes a new image, an image used as a visual reference, and an image to edit before submission. If a supplied URL or attachment has no stated role, it asks whether it is a reference or the edit target instead of guessing. A current local reference image uses the selected model's declared multipart `image_edit` operation while preserving its reference intent in the prompt; this uses no separate upload or rehosting path. `n` means variants of one complete brief, never several different assets: a request for a poster, avatar, and banner asks the user to choose the first asset rather than creating several paid tasks.

The Skill turns a natural-language request into a concise image brief while preserving stated purpose, subject, scene, style, composition, exact text, and constraints. It does not invent brands, copy, subjects, or a different operation. At completion, it reports an attachment or local path only when the active runtime confirms it; otherwise it reports native-byte delivery through that host without inventing a preview or download location.

Physical dimensions such as `200cm × 230cm` cannot be guaranteed and are never passed as `n`, `size`, or another API field. The Skill explains the limitation and lists the installed or on-demand-declared pixel or semantic-size choices. For a model that declares `width` plus `height` as a required pair, one dimension alone stops before submission. When a model declares an order for competing size expressions, the Skill submits only the highest-precedence expression the user supplied.

For `n` images, delivery reads exactly the zero-based indexes `0` through `n-1` from the same task, one index at a time and only after the task succeeds. A request is successful only when native bytes arrive for every requested index. A partial result names both delivered and missing indexes, then permits only another read of the missing content from that same task. The Skill never prefetches or re-downloads delivered content; it hands off each native result before reading the next.

## Model parameter profiles and receipts

Every normal image and video task uses the installed versioned selection; it never incurs a live-catalog preflight. Any requested optional field must be present with a compatible value in that selection. When the user asks for a current capability, an option or media operation absent from the selection, the Skill makes one on-demand catalog read; the same is allowed once after a model/parameter/capability rejection to explain it, never to retry automatically. Video prompt is required for normal text generation and may be omitted only when the installed or on-demand profile explicitly declares the exact single-reference exception.

The installed selection, or a permitted one-time on-demand profile lookup for an unsupported or incompatible requested media operation, controls media inputs. An explicitly supplied public HTTPS URL, file ID, or voice ID is sent only in its exact declared property and permitted transport; the Skill never downloads, probes, checks accessibility, rehosts, or rewrites it. A public-URL image edit additionally requires the exact declared `image_edit` JSON operation; a reference field alone is not an edit operation. A current local image reference or edit uses the declared multipart `image_edit` operation directly, with its exact path, field, and count limits; a local visual reference remains a reference in the prompt even though the transport operation is named `image_edit`. Native media explicitly attached in the current request uses only an advertised `multipart_file` operation and is sent with that one declared Images or Videos API request by the verified runtime. It accepts only explicit regular media files, limits attachment count and size, and never calls a separate upload API. Multiple native attachment types need an explicitly declared combined operation; multiple public URL/ID fields need no declared conflict. If a local reference image has no declared compatible `image_edit` multipart operation or the runtime cannot deliver its bytes, the Skill stops before submission and asks for its public HTTPS URL or a new explicit text-only request that ignores the attachment; it never inspects or converts that image into a prompt. For video, “first frame” selects the model's declared `image_to_video` operation and exact field; a supplied image with no stated first-frame/reference role is clarified before billing. The Skill never manufactures a URL or file ID or silently turns a media request into text generation.

On submit, continuation, reconciliation, completion, and failure, media Skills return a consistent receipt: exact model ID when returned, task ID when returned, current state, requested operation, requested count, requested size/parameters, delivered count on completion, and the next action. A failure contains its phase, `api_error_code` only when the public API explicitly returned that exact code (otherwise “not returned”), the returned HTTP status or “not returned”, and a safe user-facing message; HTTP 429 receipts include a valid `Retry-After` value when available. The receipt never exposes raw response bodies, upstream identifiers, internal URLs, stacks, request data, credentials, or user media. The receipt's `task_id` is normalized from the selected lifecycle's declared top-level ID field, or only top-level `task_id` / `id` when no lifecycle is declared; it is never inferred from a URL or nested response data. Missing task metadata is reported as not returned, never guessed.

If a submission POST began but the runtime's completion output is empty, truncated, malformed, or otherwise unusable, acceptance is unknown. The Skill returns one failure receipt and ends that response: it does not keep calling tools, polling, or reasoning about the same submission. It never sends another POST or replacement task merely to obtain an ID, and it never claims that no task was created, no charge occurred, or a refund happened. Without a task ID, a later “continue” or “retry” asks the user to explicitly confirm a new paid request; only that subsequent confirmation may create one new task.

## Asynchronous polling

Media polling begins only after submission returns a normalized `task_id` and runs only within the submission turn or a user turn that explicitly continues that same task ID. It first treats `reconciliation_required: true` as a terminal operational state regardless of its lifecycle status: ordinary polling stops, the same task ID is retained, and the Skill neither submits a replacement nor infers a refund. Otherwise it uses the selected model's declared lifecycle states when available; otherwise it recognizes only `pending`, `queued`, `running`, and `in_progress` as processing states. An unrecognized or missing state is reported as such and stops automatic polling. There is at most one status request in flight per task and never a background timer, queue, or worker. When a status read returns HTTP 429, the Skill honors a valid positive `Retry-After` and continues the same task only if it fits the remaining automatic-polling budget; otherwise it stops that window without resubmission. Otherwise an image task waits `3, 6, 12, 24, 30, 30` seconds before at most six same-task status reads; a video task waits `5, 10, 20, 40, 60, 60` seconds before at most seven reads. Each bounded window lasts at most 120 seconds for images or 300 seconds for videos. A non-429 5xx response, transport error, or timeout stops that window immediately. A still-processing task is reported with its task ID. When the user explicitly asks to continue it, the Skill opens one new bounded window for that same task only; it never treats a deadline or read error as failure and never submits a replacement task.

Media bytes are not cached in Skill state, prompts, or logs. Content is read only after terminal success, with one content read in flight; the active runtime hands off the native bytes before another read. It accepts only native image, video, audio, or octet-stream content responses and enforces a bounded delivery size; unsupported or oversized responses fail without a substitute URL or a replacement task. If it cannot hand off bytes without unbounded background work, duplicate reads, or cached copies, the Skill reports same-task delivery as unavailable instead of substituting a URL or submitting a new task. Confirmed delivery ends the current response: the Skill does not automatically inspect the output, search history or the workspace, invoke another Skill, or create another media task.

## Usage examples

- Connection: `Can this Pure Tokens Skill confirm its API?` The Skill checks only the fixed endpoint's `GET https://api.puretokensx.com/v1` declaration and does not reveal configuration.
- Models: `Show the video models currently available to me, their declared duration and aspect-ratio options, and which support image-to-video.` This is read-only; it does not submit a task.
- Image: `Use gpt-image-2 to generate a 2K, 16:9 illustration of a snowy village at dawn.`
- Another image model: `Use nano banana pro to generate a clean product poster.` The Skill resolves the unique installed alias and submits directly.
- Image reference URL: `Use gpt-image-2 with this public reference image URL: https://example.com/reference.png` The Skill uses the installed declared field and transport, or reads the catalog once only if the installed selection has no matching reference capability.
- Image edit: `Use gpt-image-2 to edit the attached image: replace the cloudy sky with a clear sunset.` The Skill sends the current attachment only through gpt-image-2's declared multipart image-edit operation, or reads the catalog once only if that operation is absent from the installed selection.
- Local image reference: `Use gpt-image-2 with my attached image as a composition reference: create a calm winter version.` The Skill uses the declared multipart image-edit transport while keeping the request as a reference-image brief; it does not upload or rehost the file.
- Video: `Use grok 1.5 video to generate a six-second cinematic sunrise over the ocean.`
- Image-to-video: `Use grok 1.5 video to animate this public image URL for six seconds: https://example.com/reference.png` The Skill uses the installed declared URL field and transport, or reads the catalog once only when that operation is absent.
- First-frame video: `Use wan3.0-video and use my attached image as the first frame.` The Skill selects the declared `image_to_video` operation and its `first_frame_image` field.
- Reference video: `Use seedance-2.5 to create a six-second video from my attached video.` The Skill uses the installed declared `reference_video` operation, or reads the catalog once only when it is absent.
- Reference audio: `Use minimax h3 to create a video from my attached audio.` The Skill uses the installed declared `reference_audio` operation, or reads the catalog once only when it is absent.
- Video edit: `Edit my attached video: turn daylight into night.` The Skill submits to `/v1/videos/edits` only when the installed or permitted on-demand profile declares `video_edit`.
- Existing task only: `Continue querying task <task_id>.` The Skill only reads that task; it never submits a replacement task automatically.
- Update: `Upgrade my Pure Tokens Skills.` The update Skill validates the official `main` checkout and safely synchronizes the local Skill directory.

## Model discovery

Use `puretokens-models` when the user asks what is actually available through Pure Tokens, which models support a media operation, or which models accept a particular declared parameter. Its authenticated `GET https://api.puretokensx.com/v1/media/models` response is the runtime source of truth: it reports exact model IDs, capabilities, optional parameter schema, and `input_schema.operations` without guessing missing fields. A compatibility shortlist is technical only; it matches declared capability, field/value, and operation metadata and does not make subjective quality or price claims.

The README is discovery-only. Capabilities are taken only from the base model catalog's explicit image/video declarations, never inferred from a model name. Each installed image/video Skill includes its capability-specific `references/model-selection.json`, generated from this same catalog; an alias is usable only when it resolves to one exact model ID.

<!-- media-model-catalog:start -->
## Media model catalog

Synchronized with the base model catalog: 2026-09-03T04:00:37.535Z.

This list is generated from Pure Tokens' base model catalog using explicit image/video capabilities. At execution time, the exact model and required capability must still appear in the authenticated GET https://api.puretokensx.com/v1/media/models response.

README is generated only from base-catalog models with explicit image/video capabilities; it never infers capability from a model name. The installed snapshot resolves normal generation models and known parameters; the live catalog is read on demand only for explicit discovery, an installed-profile gap, or post-rejection diagnosis. Before release, refresh from the controlled base catalog and run `npm run release:validate`; the release gate fails when the snapshot is over seven days old.

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
node bin/puretokens-skill.js upgrade puretokens-image --target ~/.agents/skills
```

Before publishing a release:

```bash
npm run docs:sync-media-models-from-service
npm run release:validate
```
