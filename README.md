<p align="center">
  <img src="./assets/brand/puretokens-skill-hero.png" alt="Pure Tokens Official Skills" width="100%" />
</p>

# Pure Tokens Skills

Official Skills for checking a Pure Tokens connection, balance and model catalog, and generating images or videos.

## Agent-assisted installation

The client download page reads the first `text` block below this heading. Keep this heading and block unchanged.

### Copy this to a terminal-capable local agent

```text
Install or update the official Pure Tokens Skills from https://github.com/PureTokens/puretokens-skill.
```

The agent downloads the official `runtime/puretokens-skill-fetch.sh` (Windows: `.ps1`) to a local file and runs install with the current host ID. Installed copies in `.puretokens-executor` provide check-update and update. Remote content is not piped into a shell.

## What is installed

| Skill | Purpose |
| --- | --- |
| `puretokens-connection` | Verify the fixed Pure Tokens API identity without exposing configuration. |
| `puretokens-balance` | Check account wallet balance or the current Key's remaining allowance in USD. |
| `puretokens-models` | Read the authenticated current model catalog and declared capabilities. |
| `puretokens-image` | Generate images and perform declared image edits. |
| `puretokens-video` | Generate videos and use declared image/video/audio references or video edits. |
| `puretokens-update` | Initialize, show usage guidance, and safely synchronize these official Skills. |

## How direct API generation works

The media Skills always use full fixed URLs under `https://api.puretokensx.com`: Images uses `/v1/images/generations`; Videos uses `/v1/videos`. They do not send media work to an arbitrary configured Base URL, MCP server, local proxy, sidecar, or a second endpoint.

Each request is performed by the single-file native executor installed with the Skills. It runs only for that command, connects to the fixed API, then exits; it does not start a port, background service, proxy, sidecar, or desktop automation. The executor reads only the active connection record at the supported host's documented configuration path, verifies that record targets Pure Tokens, and keeps one matching credential in memory for the fixed request. Pure Tokens Switch, CC Switch and manual configurations can use the exact records listed in `references/credential-adapters.md`; unpersisted project/session overrides are not verified. Neither the Skill nor the user passes a key or Base URL. It never scans home directories, checks a provider label, asks for, displays, stores, logs, or reports an API key, Base URL, or host configuration.

No Node, npm, Python, Go, Pure Tokens Desktop, MCP, or upload relay is required to generate media. The installer verifies and places the platform-native executor, which performs fixed requests, multipart attachments, same-task polling, and bounded native-media delivery. If validation or attachment preparation fails before POST starts, no task was submitted. Once POST may have started, a network or unreadable-response failure leaves the outcome unknown; the Skill preserves that uncertainty and never resubmits automatically.

Computer Use, browser automation, and opening or clicking Pure Tokens Switch/Desktop are not fallback execution paths. The Skills never use them to find a visible generation interface, obtain a credential, submit media, or deliver a result; they also never invoke another image or video Skill as a fallback.

## Media routing

When the current host uses a Pure Tokens connection, ordinary image requests select `puretokens-image` before generic `imagegen`, Imagen, or other image Skills; ordinary video requests select `puretokens-video` before generic video Skills. This applies even when the user does not explicitly say “Pure Tokens.” Once selected, the Pure Tokens specialist either executes its fixed API path or stops safely; it never falls back to a generic media Skill.

The priority is carried by the installed Skill metadata and the host's current connection context. Only the native executor resolves one matching current-connection credential privately in memory for a fixed request; it never exposes or reports connection configuration. A host that ignores installed-Skill selection metadata must correct its own selection policy; a Skill cannot force another host or third-party Skill to change priority at runtime.

## Supported hosts

| Host | Global Skill directory | Direct execution |
| --- | --- | --- |
| Claude Code | `~/.claude/skills` | Credential fixtures tested; host end-to-end acceptance pending |
| Codex | `~/.agents/skills` | Credential fixtures tested; host end-to-end acceptance pending |
| WorkBuddy | `~/.workbuddy/skills` | Credential fixtures tested; host end-to-end acceptance pending |
| Gemini CLI | `~/.gemini/skills` | Credential fixtures tested; host end-to-end acceptance pending |
| Grok Build | `~/.grok/skills` | Credential fixtures tested; host end-to-end acceptance pending |
| OpenCode | `~/.config/opencode/skills` | Credential fixtures tested; host end-to-end acceptance pending |
| Trae | `~/.trae/skills` | Installable; no Switch-managed credential record exists, so requests stop safely |

`references/host-support.json` defines these seven hosts. The table shows defaults; Claude/WorkBuddy honor explicit configuration-directory overrides. Gemini updates an existing higher-priority `.agents/skills` installation and reports managed duplicates. Provider labels never determine support.

## Images and videos

Normal text generation first reads a small installed model index, then only the selected model profile; it does not load every model or fetch the catalog before every task. `puretokens-image` defaults to `gpt-image-2`; `puretokens-video` uses its installed default. The live catalog is read only when a user explicitly asks for current models, requests an option or media operation absent from the selected profile, or needs a post-rejection diagnosis.

Image and video tasks are asynchronous. Each submission creates at most one POST. Once a returned top-level task ID is known, the Skill polls and retrieves only that task. If a POST may have started but the task ID is missing, acceptance is unknown: it never submits a duplicate task or claims that no charge occurred. Image content is delivered sequentially from zero-based indexes `0..n-1`; video content is delivered only after terminal success.

For a local image, video, or audio attachment, the Skill sends the current attachment only in the exact declared multipart Images or Videos API operation. It never uploads to a separate service, rehosts media, turns attachment bytes into a prompt, or silently downgrades an image/video reference request to text generation. Public HTTPS URLs are passed only in profile-declared fields; file/voice IDs are not supported.

The selected installed profile governs ordinary requests. An explicit authenticated catalog query reports current declarations; catalog membership is not a required preflight or a substitute for the media API’s access decision. If the media API denies access, report that rejection and guide the user to check the connection’s model permissions without automatically resubmitting.

<!-- media-model-catalog:start -->
## Media model catalog

Synchronized with the base model catalog: 2026-09-03T04:00:37.535Z.

This list is an installed selection aid from the base catalog, not a per-request authorization check. Ordinary generation uses only the selected installed profile without a live-catalog preflight; explicit discovery, a requested profile gap or rejection diagnosis may read the authenticated catalog once. The media API decides access.

README is generated only from base-catalog models with explicit image/video capabilities; it never infers capability from a model name. The installed model index selects a model and only that model's profile carries known parameters; the live catalog is read on demand only for explicit discovery, an installed-profile gap, or post-rejection diagnosis. Before release, refresh from the controlled base catalog and run `npm run release:validate`; the release gate fails when the snapshot is over seven days old.

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

## Errors and receipts

Machine receipts retain available model, task ID, original operation, state, safe parameters and progress. User-facing replies show only the useful status, actual attachment or actionable failure. Failures include a safe phase, public API code only when explicitly returned, HTTP status when returned, a sanitized message, and an action the user can take. The Skills never expose raw response bodies, request headers/bodies, internal URLs, credentials, or user media.

## Updating

`puretokens-update` resolves official main to an exact commit and version. Its native fetch wrapper checks versions without installation, or installs matching checksum-verified platform assets. When those assets are unavailable, it retrieves the same pinned official source archive and invokes native sync. The installer synchronizes all six Skills and the SHA-256-verified platform-native executor, preserves unmanaged directories, removes only verified retired official Skill directories, and removes the verified retired Node runtime if present. The versioned success receipt is the only confirmation that an update completed.

The source sync scripts are `runtime/puretokens-skill-install.sh` for macOS/Linux and `runtime/puretokens-skill-install.ps1` for Windows. They install, verify, and place the platform executor; users do not need Node, npm, Python, Go, or a package manager.

After every successful installation or update, the installer automatically runs `init`. It performs a non-billable fixed `/v1` identity check followed by one authenticated `/v1/media/models` request without displaying credentials or host configuration, then prints the current usage guide and examples. If verification does not complete, it reports a sanitized reason such as no active matching connection, missing credential, API rejection with its HTTP status, network failure, or an unconfirmed API identity; it never prints the configured URL, provider, or key. To run it again later, ask the host Agent to initialize Pure Tokens Skills or check the current Pure Tokens connection; it must invoke the installed executor's `init` command and show the guide without modifying configuration.

## Development validation

Maintainers can run:

```bash
npm run check
npm run release:validate
```

## Executor receipts and acceptance

The 0.17 command flow is submit → immediate task ID → bounded wait/status → one-index content download → host attachment delivery. No per-generation balance, init, or catalog preflight is added. A downloaded file is not yet a delivered attachment. Request JSON files replace interactive stdin; see each media Skill’s executor-usage reference.

Balance uses the same API-key route as the official CC Switch integration: `GET https://console.puretokensx.com/api/product/console/api-keys/usage`, followed by one public `/api/product/console/status` read for the USD conversion ratio. No browser login is required. An unlimited Key returns account wallet balance; a limited Key returns its own remaining allowance. The default reply is one concise amount with its scope, without subscription quotas. Legacy billing placeholders are never used as money. Queries stop after at most two GETs within 30 seconds; failures show an actionable reason without inventing an amount.

`references/host-acceptance.json` distinguishes tested credential fixtures from real host acceptance. Local automated checks do not prove Windows/macOS host attachment delivery.

Explicit request checks use `preflight` without submitting media. `doctor` combines local installation diagnostics with read-only connection checks; help-only questions read the installed usage guide without network access. Neither is a routine generation preflight. Optional task records in a user/workspace location support same-task resume and delivered-index tracking; they contain no credentials, prompt, reference URLs or media bytes.

Exact media quotations, server idempotency guarantees and lookup of unknown submissions require server capabilities not implemented here. Local validation and task records do not provide them.
