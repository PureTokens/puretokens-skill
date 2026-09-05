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

## What is installed

| Skill | Purpose |
| --- | --- |
| `puretokens-connection` | Verify the fixed Pure Tokens API identity without exposing configuration. |
| `puretokens-balance` | Read the current balance snapshot. |
| `puretokens-models` | Read the authenticated current model catalog and declared capabilities. |
| `puretokens-image` | Generate images and perform declared image edits. |
| `puretokens-video` | Generate videos and use declared image/video/audio references or video edits. |
| `puretokens-update` | Initialize, show usage guidance, and safely synchronize these official Skills. |

## How direct API generation works

The media Skills always use full fixed URLs under `https://api.puretokensx.com`: Images uses `/v1/images/generations`; Videos uses `/v1/videos`. They do not send media work to an arbitrary configured Base URL, MCP server, local proxy, sidecar, or a second endpoint.

Each request is performed by the single-file native executor installed with the Skills. It runs only for that command, connects to the fixed API, then exits; it does not start a port, background service, proxy, sidecar, or desktop automation. The executor reads only the active connection record at the supported host's documented configuration path, verifies that record targets Pure Tokens, and keeps one matching credential in memory for the fixed request. This works with Pure Tokens Switch, CC Switch, or a manual connection that uses that host's documented configuration shape. Neither the Skill nor the user passes a key or Base URL. It never scans home directories, checks a provider label, asks for, displays, stores, logs, or reports an API key, Base URL, or host configuration.

No Node, npm, Python, Go, Pure Tokens Desktop, MCP, or upload relay is required to generate media. The installer verifies and places the platform-native executor, which performs fixed requests, multipart attachments, same-task polling, and bounded native-media delivery. Only an actual executor, network, verified credential-adapter, attachment-path, attachment-byte, or API failure may stop a request before billing.

Computer Use, browser automation, and opening or clicking Pure Tokens Switch/Desktop are not fallback execution paths. The Skills never use them to find a visible generation interface, obtain a credential, submit media, or deliver a result; they also never invoke another image or video Skill as a fallback.

## Media routing

When the current host uses a Pure Tokens connection, ordinary image requests select `puretokens-image` before generic `imagegen`, Imagen, or other image Skills; ordinary video requests select `puretokens-video` before generic video Skills. This applies even when the user does not explicitly say “Pure Tokens.” Once selected, the Pure Tokens specialist either executes its fixed API path or stops safely; it never falls back to a generic media Skill.

The priority is carried by the installed Skill metadata and the host's current connection context. The Skill may resolve only one matching current-connection credential privately in memory for a fixed request; it never exposes or reports connection configuration. A host that ignores installed-Skill selection metadata must correct its own selection policy; a Skill cannot force another host or third-party Skill to change priority at runtime.

## Supported hosts

| Host | Global Skill directory | Direct execution |
| --- | --- | --- |
| Claude Code | `~/.claude/skills` | Verified native-executor credential adapter |
| Codex | `~/.agents/skills` | Verified native-executor credential adapter |
| WorkBuddy | `~/.workbuddy/skills` | Verified native-executor credential adapter |
| Gemini CLI | `~/.gemini/skills` | Verified native-executor credential adapter |
| Grok Build | `~/.grok/skills` | Verified native-executor credential adapter |
| OpenCode | `~/.config/opencode/skills` | Verified native-executor credential adapter |
| Trae | `~/.trae/skills` | Installable; no Switch-managed credential record exists, so requests stop safely |

The repository contract defines the same seven hosts in `references/host-support.json`. It does not infer support from a provider-name field.

## Images and videos

Normal text generation first reads a small installed model index, then only the selected model profile; it does not load every model or fetch the catalog before every task. `puretokens-image` defaults to `gpt-image-2`; `puretokens-video` uses its installed default. The live catalog is read only when a user explicitly asks for current models, requests an option or media operation absent from the selected profile, or needs a post-rejection diagnosis.

Image and video tasks are asynchronous. Each submission creates at most one POST. Once a returned top-level task ID is known, the Skill polls and retrieves only that task. If a POST may have started but the task ID is missing, acceptance is unknown: it never submits a duplicate task or claims that no charge occurred. Image content is delivered sequentially from zero-based indexes `0..n-1`; video content is delivered only after terminal success.

For a local image, video, or audio attachment, the Skill sends the current attachment only in the exact declared multipart Images or Videos API operation. It never uploads to a separate service, rehosts media, turns attachment bytes into a prompt, or silently downgrades an image/video reference request to text generation. Public HTTPS URLs and declared IDs are passed only in profile-declared fields.

Model capabilities, optional parameters, operations, lifecycle states and access are facts returned by the authenticated `GET https://api.puretokensx.com/v1/media/models` catalog. A model not returned by the current connection is not submitted; the user should select a Pure Tokens group/key that covers that model and start a new host conversation.

<!-- media-model-catalog:start -->
## Media model catalog

Synchronized with the base model catalog: 2026-09-03T04:00:37.535Z.

This list is generated from Pure Tokens' base model catalog using explicit image/video capabilities. At execution time, the exact model and required capability must still appear in the authenticated GET https://api.puretokensx.com/v1/media/models response.

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

Every media result gives the exact model when returned, task ID when returned, state, requested parameters, and next action. Failures include a safe phase, public API code only when explicitly returned, HTTP status when returned, a sanitized message, and an action the user can take. The Skills never expose raw response bodies, request headers/bodies, internal URLs, credentials, or user media.

## Updating

`puretokens-update` obtains a fresh official `main` checkout and runs the source sync script for the active host. The installer synchronizes all six Skills and the SHA-256-verified platform-native executor, preserves unmanaged directories, removes only verified retired official Skill directories, and removes the verified retired Node runtime if present. The versioned success receipt is the only confirmation that an update completed.

The source sync scripts are `runtime/puretokens-skill-install.sh` for macOS/Linux and `runtime/puretokens-skill-install.ps1` for Windows. They install, verify, and place the platform executor; users do not need Node, npm, Python, Go, or a package manager.

After every successful installation or update, the installer automatically runs `init`. It performs one non-billable fixed `/v1` identity check without displaying credentials or host configuration, then prints the current usage guide and examples. If verification does not complete, it reports a sanitized reason such as no active matching connection, missing credential, API rejection with its HTTP status, network failure, or an unconfirmed API identity; it never prints the configured URL, provider, or key. To run it again later, ask the host Agent to initialize Pure Tokens Skills or check the current Pure Tokens connection; it must invoke the installed executor's `init` command and show the guide without modifying configuration.

## Development validation

Maintainers can run:

```bash
npm run check
npm run release:validate
```
