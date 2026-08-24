<p align="center">
  <img src="./assets/brand/puretokens-skill-hero.png" alt="Pure Tokens Skill — one skill, every model" width="100%" />
</p>

<p align="center">
  <strong>English</strong> · <a href="./README.zh-CN.md">中文</a> · <a href="./CHANGELOG.md">Changelog</a>
</p>

# Pure Tokens Skills

`puretokens-skill` is the source repository for Pure Tokens Skills. It owns Skill instructions, versions, compatibility declarations, client installation instructions, and validation tools. It does not contain user credentials, Router configuration, or model-routing logic.

## Provider scope

This Skill supports **Pure Tokens only**. It must not send media requests through another provider's API, MCP server, or native media tool, even when that provider is OpenAI-compatible or offers similarly named models. If the current connection is not verifiably Pure Tokens, it stops and tells the user to switch to Pure Tokens: https://puretokensx.com/

## Image dimensions

`n` is only for an explicit image count, such as “generate 3 images.” A physical dimension such as `200cm x 230cm` is never an image count and cannot be passed directly to `size`.

The currently supported image canvases are `1024x1024`, `1536x1024`, and `1024x1536`. An explicit `image_size` may be `1K`, `2K`, or `4K`; it is an output-resolution option, not a guaranteed physical print size. If a user requests centimetres, millimetres, metres, or inches, the Skill does not submit a request, guess DPI, convert it automatically, or choose a closest canvas. It clearly reports that the physical size cannot be guaranteed and lists the supported pixel canvases for the user to choose from.

## Quick start

Choose the one path your host can actually execute:

| Host | Execution path | Setup |
| --- | --- | --- |
| Codex or CC Switch with a Pure Tokens connection | Skill → Pure Tokens Connection API → service | Default-image and `gpt-image-2` requests directly use `POST https://api.puretokensx.com/v1/images/generations` with `gpt-image-2`. Video can use `GET /v1/media/models` and `POST /v1/videos` through the same connection only when the host exposes it as a callable authenticated HTTPS media executor with actual byte delivery. Neither path needs a global instruction, MCP, or `PURETOKENS_API_KEY` environment variable. |
| Pure Tokens Desktop-managed client | Skill → managed MCP → local Router → service | Optional convenience path: select the client groups, click **Verify and apply**, then restart the client and start a new chat. |
| Host with a selected native Pure Tokens media model | Skill → host-native media operation → service | The configured operation must expose an exact verified image/video model and actual media delivery. A general chat-model setting alone is not enough. |
| GUI host with callable MCP tools | Skill → `puretokens-image` MCP → service | Install/configure the callable MCP delivery for that client. The GUI user never pastes a token into chat. |
| HTTPS-capable Agent | Skill → Direct Cloud → service | Install the Skill and inject `PURETOKENS_API_KEY` through the host's Secret/environment mechanism. A CC Switch-connected host can use this path only when it can actually execute HTTPS and deliver media bytes locally. No Desktop, Router, CLI sidecar, or MCP is required. |

Then start a new chat. Say `Generate a cute dog` for the default image model, or name a model such as `Use Nano Banana Pro to generate ...`.

For Codex or CC Switch, the Skill itself defines the default Image-2 execution path: a default-image request or an explicit `gpt-image-2` / `image2` request calls `POST https://api.puretokensx.com/v1/images/generations` with `model: "gpt-image-2"` through the active Pure Tokens connection. When that host also explicitly exposes the current connection as a callable authenticated HTTPS video executor that can deliver the completed bytes, video first reads `GET https://api.puretokensx.com/v1/media/models`, then uses the exact returned video model with `POST https://api.puretokensx.com/v1/videos`, polls the same task, and retrieves `/content`. It does not depend on a system/developer/AGENTS instruction, `puretokens-image` MCP, or a second `PURETOKENS_API_KEY` environment variable. This is the Pure Tokens user endpoint, never an upstream endpoint. A chat connection that merely stores an API key but exposes no callable video execution and delivery is not enough; its user still needs a selected native executor, MCP, or Direct Cloud.

> **Before using a specific model:** every model except the Codex/CC Switch default `gpt-image-2` Images API path must appear in the authenticated `GET /v1/media/models` catalog for the selected execution path. A verified active-connection video executor uses its current connection scope; Desktop-managed MCP uses only the selected client groups; Direct Cloud uses only the API key's permissions. No path can use every model mentioned in a public catalog.

Current Skill:

| Skill | Purpose |
| --- | --- |
| `puretokens_media` | Select an exact image/video model from the live catalog, submit one task, poll the same task, and deliver actual native media bytes plus local files. |

<!-- media-model-catalog:start -->
## Media model catalog

Synchronized with the base model catalog: 2026-08-21T02:46:19.421Z.

This list is generated from Pure Tokens' base model catalog using explicit image/video capabilities. At execution time, the exact model and required capability must still appear in the current authenticated GET /v1/media/models response.

README is generated only from base-catalog models with explicit image/video capabilities; it never infers capability from a model name. Before release, run `npm run docs:sync-media-models-from-service` against the controlled base catalog; execution still uses the authenticated `GET /v1/media/models` response.

### Image models

| Model ID | Provider | You can also say | Good for | Example |
| --- | --- | --- | --- | --- |
| `gpt-image-2` | OpenAI | `image2`, `image 2`, `gpt image 2`, `openai image 2` | High-quality posters, product visuals, and illustrations | `Use image2 to make a clean orange product launch poster.` |
| `grok-imagine-image` | xAI | `grok image`, `grok imagine` | Social posts and everyday image generation | `Use grok-imagine-image to create a realistic cafe opening post.` |
| `grok-imagine-image-quality` | xAI | `grok quality image`, `grok high quality image` | Sharper brand key visuals | `Use grok quality image to make a polished app-store banner.` |
| `nano-banana-2` | Google | `nano banana`, `nano banana 2`, `nano banana two` | Fast visual exploration and social-media creatives | `Use Nano Banana 2 to create a bright product social post.` |
| `nano-banana-pro` | Google | `nano banana`, `nano banana pro`, `nano banana professional` | Polished marketing images and premium key visuals | `Use Nano Banana Pro to create a premium cloud-computing hero image.` |
| `qwen-image-2.0` | Qwen | `qwen image 2`, `qwen image 2.0` | General-purpose image creation and product creatives | `Use qwen-image-2.0 to make a clean ecommerce product scene.` |
| `qwen-image-2.0-pro` | Qwen | `qwen image 2 pro`, `qwen image 2.0 pro` | Higher-fidelity campaigns and product key visuals | `Use qwen-image-2.0-pro to make a premium product campaign visual.` |
| `seedream-5.0-pro` | Doubao | Exact ID only | Image generation | `Use seedream-5.0-pro to generate an image.` |
| `wan2.7-image` | Qwen | `wan image`, `wan 2.7 image` | Chinese posters and product creatives | `Use wan 2.7 image to make a Chinese New Year promotion poster.` |
| `wan2.7-image-pro` | Qwen | `wan 2.7 image pro` | Higher-fidelity Chinese posters and brand visuals | `Use wan2.7-image-pro to make a premium Chinese product launch poster.` |

### Video models

| Model ID | Provider | You can also say | Good for | Example |
| --- | --- | --- | --- | --- |
| `grok-imagine-video` | xAI | `grok video`, `grok imagine video` | Short social clips and quick concepts | `Use grok-imagine-video to make a 5-second coffee ad.` |
| `grok-imagine-video-1.5-preview` | xAI | `grok video`, `grok imagine video`, `grok 1.5 video`, `grok video 1.5`, `grok imagine video 1.5` | Video generation | `Use grok-imagine-video-1.5-preview to generate a short video.` |
| `minimax-h3` | MiniMax | `minimax h3`, `minimax h3 video` | Cinematic product clips and motion-led concepts | `Use minimax-h3 to make a 10-second product reveal video.` |
| `seedance-2.0` | Doubao | Exact ID only | Video generation | `Use seedance-2.0 to generate a short video.` |
| `seedance-2.0-fast` | Doubao | Exact ID only | Video generation | `Use seedance-2.0-fast to generate a short video.` |
| `seedance-2.0-mini` | Doubao | Exact ID only | Video generation | `Use seedance-2.0-mini to generate a short video.` |
| `seedance-2.5` | Doubao | Exact ID only | Video generation | `Use seedance-2.5 to generate a short video.` |

<!-- media-model-catalog:end -->

The Skill requests one result by default. It passes a higher count only when the user explicitly asks for it and the selected execution contract supports that count; it never turns one request into several submissions. MCP calls the selected generation tool once, then polls the same task. Direct Cloud image submissions always send `async: true`; its execution layer still defensively accepts compatible synchronous `data[].b64_json` and `data[].url` responses as well as asynchronous tasks, but reports success only after actual bytes are locally delivered.

## What to say

| You want | Say this |
| --- | --- |
| Generate an image | `Generate a cute dog.` |
| Generate a video | `Generate a 15-second 16:9 product ad.` |
| Use Nano Banana | `Use Nano Banana Pro to create a premium product key visual.` |
| See available models | `List the image and video models I can use now.` |

## Boundaries

```text
Natural-language user request → Skill → (active Connection API → service | MCP → local Router → service | Direct Cloud → service)
```

- The Skill interprets requests such as “use image2” or “use Grok Video”, reads the live catalog, asks for clarification when the match is not unique, and selects the correct tool.
- MCP accepts only an exact model ID. It validates arguments, submits once, polls the result, and delivers local files. MCP never performs natural-language matching, guesses a model, or silently substitutes one.
- A host-selected native Pure Tokens media operation takes precedence when it reports an exact verified model and real media delivery. Codex/CC Switch uses the active-connection Videos API only when the host exposes that verified execution-and-delivery capability; otherwise GUI clients use a callable `puretokens-image` MCP tool and an HTTPS-capable Agent may use Direct Cloud with a host-injected `PURETOKENS_API_KEY`. These independent paths do not require Pure Tokens Desktop, Router, an extra CLI, or MCP.
- The live catalog remains authoritative. A verified active connection, Desktop Router, and Direct Cloud all read the authenticated `/v1/media/models` response with explicit `image` / `video` capabilities.

## Prerequisites

For the Desktop-managed MCP path, complete these steps in order:

1. In Pure Tokens Desktop, open the configuration for the target client.
2. Select one or more groups that contain the target model.
3. Click **Verify and apply**.
4. Restart the target client and start a new chat.

Only models in the selected group or groups are available to this Desktop-managed MCP path. If the target model is absent from the live media catalog, return to client configuration, select a group that contains it, and apply the configuration again. Desktop configures an MCP server named `puretokens-image` for supported clients. A self-managed MCP through CC Switch or another provider does not require Desktop; its own authenticated `puretokens_list_media_models` response determines availability. The Skill does not replace MCP configuration and never carries credentials.

For a host-native manually configured Pure Tokens model, the selected operation must provide an exact verified `image` or `video` model and real media delivery. A generic text/chat model connection, an opaque model label, or a rendered widget does not meet that requirement. The same applies to a CC Switch connection configured with a Pure Tokens API key: it can use the Connection Videos API without a second credential only if the host exposes callable authenticated HTTPS video execution and actual byte delivery. For Direct Cloud, configure the host's normal **API Base URL** and **API Key** fields with `https://api.puretokensx.com` and a Pure Tokens API key. Hosts that use environment variables map those fields to `PURETOKENS_API_BASE_URL` and `PURETOKENS_API_KEY`; the Skill must never ask for, print, persist, or put that key in a prompt. Direct Cloud does not use the Desktop group selection UI; the API key's own permissions and the authenticated `/v1/media/models` catalog are authoritative.

## Install and update from GitHub

`puretokens_media` is the single behavior source for every supported client. Claude Desktop receives an uploadable ZIP; Codex can install the shared source directly; WorkBuddy uses a generated adapter when needed. Pure Tokens Desktop can manage Codex and WorkBuddy as an optional convenience path. Always use this repository for the current shared-Skill installation instructions and files.

### Codex

For Desktop-managed Codex, select the target groups and click **Verify and apply**. Desktop atomically installs the generated shared Skill at `~/.codex/skills/puretokens_media` and configures the separate local `puretokens-image` MCP server. No Plugin Marketplace or Plugin unlock is required.

For a standalone terminal-capable Codex Agent, install the same shared source directly and use its independently configured MCP or Direct Cloud capabilities:

```bash
node bin/puretokens-skill.js install puretokens_media --target ~/.codex/skills
```

Start a new Codex task after either installation path. The Skill never bundles, starts, or replaces a Desktop-managed MCP.

### Claude Code, Gemini CLI, and OpenCode

Clone the canonical repository, then install into the target client's user Skill directory:

```bash
git clone https://github.com/PureTokens/puretokens-skill.git
cd puretokens-skill

# Claude Code
node bin/puretokens-skill.js install puretokens_media --target ~/.claude/skills

# Gemini CLI
node bin/puretokens-skill.js install puretokens_media --target ~/.gemini/skills

# OpenCode
node bin/puretokens-skill.js install puretokens_media --target ~/.config/opencode/skills
```

### Copy this to a terminal-capable local agent

Paste this only into an agent that can run local commands and write local files:

```text
Install Pure Tokens Skill for the client I am using from https://github.com/PureTokens/puretokens-skill.

1. Before identifying a target client, confirm that this environment has both a local terminal and permission to write local files.
   - If this is a normal ChatGPT conversation, or either capability is unavailable, stop. Do not clone or download the repository or claim the Skill was installed. Tell me to use a terminal-capable local agent or have a local administrator install it.
2. Only after that check, identify whether this is Codex, Claude Code, Gemini CLI, or OpenCode.
3. Clone or download the repository into a temporary working directory.
4. Install only the matching Pure Tokens delivery:
   - Codex: ~/.codex/skills
   - Claude Code: ~/.claude/skills
   - Gemini CLI: ~/.gemini/skills
   - OpenCode: ~/.config/opencode/skills
5. Do not overwrite any other Skill.
6. Do not read, request, print, or store API keys, cookies, passwords, Router tokens, or local authorization URLs.
7. Tell me the installed directory and whether the operation succeeded.

If this is Claude Desktop, do not claim it was installed automatically. Build the ZIP following the README, then tell me exactly where to upload and enable it. For WorkBuddy, use the generated adapter command in the README only when the local WorkBuddy Skill directory is known; otherwise use Pure Tokens Desktop's **Verify and apply**.
```

Do not use that prompt as a self-install instruction inside a normal ChatGPT chat. It must fail closed rather than pretend that a local Skill was installed.

To update a manually installed Codex, Claude Code, Gemini CLI, or OpenCode Skill, pull the repository and run the matching upgrade command:

```bash
git pull
node bin/puretokens-skill.js upgrade puretokens_media --target ~/.codex/skills
```

Codex, Claude Code, Gemini CLI, and OpenCode use the matching `upgrade` command and target directory from the installation table above. Upgrade replaces only a Pure Tokens-managed directory containing matching `skill.json` and `SKILL.md`; it never overwrites another Skill.

### Windows PowerShell

```powershell
git clone https://github.com/PureTokens/puretokens-skill.git
Set-Location puretokens-skill
node .\bin\puretokens-skill.js install puretokens_media --target $HOME\.claude\skills
```

For the other clients, use `$HOME\.claude\skills`, `$HOME\.gemini\skills`, or `$HOME\.config\opencode\skills` with `node .\bin\puretokens-skill.js install puretokens_media --target ...`. If PowerShell cannot find `node`, install Node.js LTS from the official Node.js website and reopen PowerShell.

## Claude Desktop import and WorkBuddy routing

Claude Desktop uses a graphical local Skill upload. Create the ZIP:

```bash
node bin/puretokens-skill.js bundle puretokens_media --format claude-desktop --out ./puretokens_media-0.4.7.zip
```

The ZIP has this layout:

```text
puretokens_media/
├── SKILL.md
├── skill.json
├── source-delivery.json
└── references/
    ├── behavior-scenarios.json
    ├── direct-cloud-contract.md
    ├── model-catalog-contract.md
    └── natural-language-aliases.json
```

In Claude Desktop, open **Settings → Features → Skills** (some builds show **Customize → Skills**), choose **Upload skill**, upload the ZIP, and enable `Pure Tokens Media`. A Claude Desktop instance connected through CC Switch can use this same ZIP: independently configure a callable `puretokens-image` MCP tool through CC Switch or another local tool provider. If the host itself exposes authenticated HTTPS execution and local media delivery, it may use Direct Cloud instead. The ZIP intentionally excludes the WorkBuddy-only adapter.

For WorkBuddy, choose one installation path. Pure Tokens Desktop can atomically render and manage the always-on `puretokens_workbuddy_router` delivery from the shared `puretokens_media` source, along with the `puretokens-image` MCP entry and reference files: select a compatible group, click **Verify and apply**, then start a new chat. A self-managed WorkBuddy installation can render the same generated delivery from this repository when its local Skill directory is known:

```bash
node scripts/render-workbuddy-media-skill.mjs --out ~/.workbuddy/skills/puretokens_workbuddy_router
```

The self-managed path still needs either a callable `puretokens-image` MCP configured through WorkBuddy, CC Switch, or another tool provider, or the host's actual Direct Cloud capabilities. Bare image/video requests discover deferred MCP tools and invoke them with `DeferExecuteTool`; discovery or a rendered widget is not generation. A WorkBuddy `ImageGen`, `VideoGen`, or manually configured model selected in its UI/tool context is preserved. When that selection targets a verified Pure Tokens image/video operation, WorkBuddy's configured native execution runs it without a duplicate MCP submission. A general chat-model configuration or a model name written only in the message is not enough to bypass catalog-first selection.

To update Claude Desktop, get the new version from GitHub, generate a new ZIP, disable the old Skill, upload the new ZIP, and enable it. Desktop-managed WorkBuddy regenerates the shared media behavior on the next **Verify and apply**; self-managed WorkBuddy reruns the render command. Do not hand-edit generated deliveries.

## Codex install and update

Pure Tokens Desktop's **Verify and apply** atomically replaces only its own generated `puretokens_media` directory and configures only its own `puretokens-image` MCP entry. It does not enable Codex Plugins, register a Marketplace, or change any other Skill. A standalone Codex installation uses the shared source commands above and must independently provide callable MCP or Direct Cloud execution.

## Model-selection rules

In Codex or CC Switch, default-image and explicit `gpt-image-2` / `image2` requests bypass the media catalog and call the Skill-defined Pure Tokens Connection Images API once using `POST https://api.puretokensx.com/v1/images/generations` with `model: "gpt-image-2"`; they do not use MCP, Direct Cloud, an upstream endpoint, or polling. Video uses the Skill-defined Connection Videos API only when the host explicitly makes the active Pure Tokens connection callable for authenticated HTTPS execution and real byte delivery: it reads that connection's `/v1/media/models`, submits the exact returned video model once to `/v1/videos`, then polls and retrieves the same task's `/content`. A chat connection with a saved API key but no such capability must keep a selected native executor or use MCP/Direct Cloud. Other hosts and models must first call `puretokens_list_media_models` and match only fields returned in that response: `id`, `displayName`, `aliases`, `provider`, and `capabilities`. MCP generation calls include the exact `model` and a stable `request_id`; Direct Cloud retains that request ID in host task state because the public endpoint has no documented idempotency field. One logical user request submits once, requests one result unless the user explicitly gives a count, and reuses the same `request_id` on an MCP host retry. `gpt-image-2` returns native MCP image content from the generation call and must not be followed by `puretokens_image_result`; task-based image models poll only the same returned `task_id` and original model.

Completed media reports the exact model, the saved filename, and `Downloads/Pure Tokens`. Images are previewable only when MCP or the host returns native image content; Direct Cloud always requests asynchronous image generation and, as a compatibility fallback, downloads returned `b64_json`, returned URLs, or completed `/content` bytes before it claims delivery. For a completed multi-image task, it retrieves each declared result through the same `/content` endpoint with zero-based `index` values. Videos may include a bounded native MCP resource for hosts that render it; larger videos remain successfully delivered as local MP4 files. An open-file/open-folder entry is shown only when the execution layer actually returned one.

Behavior scenarios for ambiguity, an empty catalog, unavailable MCP, task failure, and polling timeout are stored in `skills/puretokens_media/references/behavior-scenarios.json`. No error may trigger an automatic model switch or resubmission unless the user explicitly chooses a new model.

## Security boundary

The Skill does not contain or request cloud credentials, Router tokens, cookies, passwords, user configuration, group routing, payment data, local authorization addresses, media, task results, or prompt history.

Trae is currently not supported by this media Skill flow.
