<p align="center">
  <img src="./assets/brand/puretokens-skill-hero.png" alt="Pure Tokens Skill — one skill, every model" width="100%" />
</p>

<p align="center">
  <strong>English</strong> · <a href="./README.zh-CN.md">中文</a> · <a href="./CHANGELOG.md">Changelog</a>
</p>

# Pure Tokens Skills

`puretokens-skill` is the source repository for Pure Tokens Skills. It owns Skill instructions, versions, compatibility declarations, client installation instructions, and validation tools. It does not contain user credentials, Router configuration, or model-routing logic.

## Quick start

Choose the one path your host can actually execute:

| Host | Execution path | Setup |
| --- | --- | --- |
| Pure Tokens Desktop-managed client | Skill → managed MCP → local Router → service | Select the client groups, click **Verify and apply**, then restart the client and start a new chat. |
| GUI host with callable MCP tools | Skill → `puretokens-image` MCP → service | Install/configure the callable MCP delivery for that client. The GUI user never pastes a token into chat. |
| Terminal-capable code Agent | Skill → Direct Cloud → service | Install the Skill and inject `PURETOKENS_API_KEY` through the host's Secret/environment mechanism. No Desktop, Router, CLI sidecar, or MCP is required. |

Then start a new chat. Say `Generate a cute dog` for the default image model, or name a model such as `Use Nano Banana Pro to generate ...`.

> **Before using a specific model:** the exact model must appear in the authenticated `GET /v1/media/models` catalog for the selected execution path. Desktop-managed MCP uses only the selected client groups. Direct Cloud uses only the API key's permissions. Neither path can use every model mentioned in a public catalog.

Current Skill:

| Skill | Purpose |
| --- | --- |
| `puretokens_media` | Select an exact image/video model from the live catalog, submit one task, poll the same task, and deliver actual native media bytes plus local files. |

## Image models

These are registered model candidates, not a promise that every host can use them. Your current client/group or Direct Cloud token may show fewer models. The Skill uses a model only when the live `puretokens_list_media_models` or authenticated `GET /v1/media/models` response contains the exact ID and capability.

You do not need to type the full ID. Registered phrases such as `image2` and `Nano Banana Pro` are understood by the Skill and verified against the live catalog before a request is sent. If you simply ask for an image, the Skill uses `gpt-image-2`; if that exact model is unavailable in your current group, it stops and shows the available candidates instead of silently switching models.

| Model ID | You can also say | Good for | Real example |
| --- | --- | --- | --- |
| `gpt-image-2` | `image2`, `gpt image 2`, `openai image 2` | High-quality posters, product visuals, illustrations | `Use image2 to make a clean orange product launch poster.` |
| `gemini-3.0-pro-image` | `gemini pro image`, `nano banana pro` | Detailed concept art and polished marketing images | `Use Nano Banana Pro to create a premium cloud-computing hero image.` |
| `gemini-3.1-flash-lite-image` | `gemini flash lite image` | Fast thumbnails and social media drafts | `Use gemini flash lite image to make three bright social thumbnails.` |
| `gemini-3.1-flash-image` | `nano banana 2` | Faster Gemini image generation and conversational edits | `Use Nano Banana 2 to create a bright product social post.` |
| `grok-imagine-1.0` | exact ID only | Fast creative concepts and playful scenes | `Use grok-imagine-1.0 to draw a cheerful robot in a city park.` |
| `grok-imagine-image` | `grok image`, `grok imagine` | Social posts and everyday image generation | `Use grok-imagine-image to create a realistic café opening post.` |
| `grok-imagine-image-quality` | `grok quality image` | Sharper brand key visuals | `Use grok quality image to make a polished app-store banner.` |
| `wan2.7-image` | `wan image`, `wan 2.7 image` | Chinese posters and product creatives | `Use wan 2.7 image to make a Chinese New Year promotion poster.` |

The Skill requests one result by default. It passes a higher count only when the user explicitly asks for it and the selected execution contract supports that count; it never turns one request into several submissions. MCP calls `puretokens_generate_image` once, then polls the same task with `puretokens_image_result`. Direct Cloud accepts synchronous `data[].b64_json` and `data[].url` image results as well as asynchronous tasks, but reports success only after actual bytes are locally delivered.

`Nano Banana` by itself means the Gemini Nano Banana family. The current catalog uses `gemini-3.0-pro-image` for Nano Banana Pro and `gemini-3.1-flash-image` for Nano Banana 2. When both are available, the Skill asks which one you want; when only one is available, it uses that one. This keeps a named choice from turning into an invisible model substitution.

## Video models

These are registered video candidates. A model must have live capability `video` before it can be used.

| Model ID | You can also say | Good for | Real example |
| --- | --- | --- | --- |
| `grok-imagine-video` | `grok video`, `grok imagine video` | Short social clips and quick concepts | `Use grok-imagine-video to make a 5-second coffee ad.` |
| `grok-imagine-video-1.5` | `grok 1.5 video`, `grok video 1.5` | More polished short advertisements | `Use grok 1.5 video to make a 15-second 16:9 product ad.` |

The Skill calls `puretokens_generate_video` once, then polls the same task with `puretokens_video_result`. Direct Cloud videos always use the asynchronous task and `/content` delivery flow. If you simply ask for a video, it uses `grok-imagine-video-1.5`; if that exact model is unavailable in your current group, it stops and shows the available candidates.

If a model is missing, ambiguous, or does not have the requested capability, the Skill shows the live candidates and asks you to choose. It never silently changes models.

## What to say

| You want | Say this |
| --- | --- |
| Generate an image | `Generate a cute dog.` |
| Generate a video | `Generate a 15-second 16:9 product ad.` |
| Use Nano Banana | `Use Nano Banana Pro to create a premium product key visual.` |
| See available models | `List the image and video models I can use now.` |

## Boundaries

```text
Natural-language user request → Skill → (MCP → local Router → service | Direct Cloud → service)
```

- The Skill interprets requests such as “use image2” or “use Grok Video”, reads the live catalog, asks for clarification when the match is not unique, and selects the correct tool.
- MCP accepts only an exact model ID. It validates arguments, submits once, polls the result, and delivers local files. MCP never performs natural-language matching, guesses a model, or silently substitutes one.
- GUI clients use a callable `puretokens-image` MCP tool as their first execution path. A terminal-capable Agent may use Direct Cloud with a host-injected `PURETOKENS_API_KEY`; that mode does not require Pure Tokens Desktop, Router, an extra CLI, or MCP.
- The live catalog remains authoritative. Desktop Router and Direct Cloud both read the same authenticated `/v1/media/models` response with explicit `image` / `video` capabilities.

## Prerequisites

For the managed MCP path, complete these steps in order:

1. In Pure Tokens Desktop, open the configuration for the target client.
2. Select one or more groups that contain the target model.
3. Click **Verify and apply**.
4. Restart the target client and start a new chat.

Only models in the selected group or groups are available to the MCP path. If the target model is absent from the live media catalog, return to client configuration, select a group that contains it, and apply the configuration again. Desktop configures an MCP server named `puretokens-image` for supported clients. The Skill does not replace MCP configuration and never carries credentials.

For Direct Cloud, configure the host's normal **API Base URL** and **API Key** fields with `https://api.puretokensx.com` and a Pure Tokens API key. Hosts that use environment variables map those fields to `PURETOKENS_API_BASE_URL` and `PURETOKENS_API_KEY`; the Skill must never ask for, print, persist, or put that key in a prompt. Direct Cloud does not use the Desktop group selection UI; the API key's own permissions and the authenticated `/v1/media/models` catalog are authoritative.

## Install and update from GitHub

`puretokens_media` is the single behavior source for every supported client. Claude Desktop receives it as an uploadable ZIP; WorkBuddy receives a Desktop-managed, always-on generated delivery; terminal-capable Agents can use the same Skill with Direct Cloud. Always use this repository for the current shared-Skill installation instructions and files.

### Codex managed Skill

Here, **Codex** means a local Codex Agent that has a terminal and permission to write local files. It does not mean an ordinary ChatGPT conversation that happens to describe its runtime as Codex.

Use **Code → Download ZIP** on this repository page, or clone the repository. The manager requires Node.js 20 or later.

```bash
git clone https://github.com/yanyansay/puretokens-skill.git
cd puretokens-skill
```

For a Desktop-managed Codex client, select the target groups and click **Verify and apply** in Pure Tokens Desktop. Desktop atomically installs the generated shared Skill at `~/.codex/skills/puretokens_media` and configures the separate local `puretokens-image` MCP server. No Plugin Marketplace, Plugin unlock, or manual global configuration is required.

A terminal-capable Codex Agent that is not managed by Desktop can install the same Skill directly and use Direct Cloud when its host injects `PURETOKENS_API_KEY`:

```bash
node bin/puretokens-skill.js install puretokens_media --target ~/.codex/skills
```

Start a new Codex task after installation. The Skill never bundles, starts, or replaces the Desktop-managed MCP. When Pure Tokens Desktop has configured the callable `puretokens-image` MCP, the Skill uses it. Otherwise, Direct Cloud remains available when the Agent has HTTPS execution and the host-injected token.

### Claude Code, Gemini CLI, and OpenCode

Install into the target client's user Skill directory:

```bash

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
Install Pure Tokens Skill for the client I am using from https://github.com/yanyansay/puretokens-skill.

1. Before identifying a target client, confirm that this environment has both a local terminal and permission to write local files.
   - If this is a normal ChatGPT conversation, or either capability is unavailable, stop. Do not identify it as Codex merely because a model or runtime label says Codex. Do not clone or download the repository, write `~/.codex/skills`, or claim the Skill was installed. Tell me to use a terminal-capable local agent or have a local administrator install it.
2. Only after that check, identify whether this is Codex, Claude Code, Gemini CLI, or OpenCode.
3. Clone or download the repository into a temporary working directory.
4. Install only the matching Pure Tokens delivery:
   - Codex: `~/.codex/skills`
   - Claude Code: ~/.claude/skills
   - Gemini CLI: ~/.gemini/skills
   - OpenCode: ~/.config/opencode/skills
5. Do not overwrite any other Skill.
6. Do not read, request, print, or store API keys, cookies, passwords, Router tokens, or local authorization URLs.
7. Tell me the installed directory and whether the operation succeeded.

If this is Claude Desktop, do not claim it was installed automatically. Build the ZIP following the README, then tell me exactly where to upload and enable it. For WorkBuddy, tell me to use Pure Tokens Desktop's **Verify and apply** instead; do not manually create or replace its generated `puretokens_workbuddy_router` delivery.
```

Do not use that prompt as a self-install instruction inside a normal ChatGPT chat. Such a chat can be backed by a Codex runtime while still lacking access to the user's terminal and `~/.codex` directory; it must fail closed rather than pretend that a local Skill was installed.

To update a manually installed Codex Skill, pull the repository and run the matching upgrade command:

```bash
git pull
node bin/puretokens-skill.js upgrade puretokens_media --target ~/.codex/skills
```

Claude Code, Gemini CLI, and OpenCode use the matching `upgrade` command and target directory from the installation table above. Upgrade replaces only a Pure Tokens-managed directory containing matching `skill.json` and `SKILL.md`; it never overwrites another Skill.

### Windows PowerShell

```powershell
git clone https://github.com/yanyansay/puretokens-skill.git
Set-Location puretokens-skill
node .\bin\puretokens-skill.js install puretokens_media --target $HOME\.codex\skills
```

For the other clients, use `$HOME\.claude\skills`, `$HOME\.gemini\skills`, or `$HOME\.config\opencode\skills` with `node .\bin\puretokens-skill.js install puretokens_media --target ...`. If PowerShell cannot find `node`, install Node.js LTS from the official Node.js website and reopen PowerShell.

## Claude Desktop import and WorkBuddy routing

Copying a Skill to `~/.codex/skills` installs the Codex instruction layer only; its media execution still needs either a callable `puretokens-image` MCP server or host-injected Direct Cloud credentials.

Claude Desktop uses a graphical local Skill upload. Create the ZIP:

```bash
node bin/puretokens-skill.js bundle puretokens_media --format claude-desktop --out ./puretokens_media-0.4.4.zip
```

The ZIP has this layout:

```text
puretokens_media/
├── SKILL.md
├── skill.json
├── source-delivery.json
├── adapters/
│   └── workbuddy-execution.md
└── references/
    ├── behavior-scenarios.json
    ├── direct-cloud-contract.md
    ├── model-catalog-contract.md
    └── natural-language-aliases.json
```

In Claude Desktop, open **Settings → Features → Skills** (some builds show **Customize → Skills**), choose **Upload skill**, upload the ZIP, and enable `Pure Tokens Media`. If the installed build has no Skills entry, it cannot import custom Skills; that build can still use MCP tool descriptions, but it will not receive this Skill's deterministic model-selection and no-fallback policy.

For WorkBuddy, do not upload or enable a separate Skill yourself. Selecting a compatible group and clicking **Verify and apply** in Pure Tokens Desktop atomically renders and manages the always-on `puretokens_workbuddy_router` delivery from the shared `puretokens_media` source, along with the `puretokens-image` MCP entry and its reference files. Restart WorkBuddy or start a new chat afterwards. Bare image/video requests first discover the deferred MCP tools, then actually invoke them through `DeferExecuteTool`; a discovered tool or a rendered widget is not a media generation. An explicit request for WorkBuddy's built-in `ImageGen` or `VideoGen` still keeps that user choice.

To update Claude Desktop, get the new version from GitHub, generate a new ZIP, disable the old Skill, upload the new ZIP, and enable it. WorkBuddy regenerates the same shared media behavior on the next **Verify and apply**. Do not delete the MCP entry directly; Pure Tokens Desktop owns MCP configuration.

## Codex install and update

Pure Tokens Desktop manages Codex automatically during **Verify and apply**: it installs only its own `puretokens_media` directory and configures only its own `puretokens-image` MCP entry. It does not enable Codex Plugins, register a Marketplace, or change any other Skill.

For a standalone local Codex Agent, install and update only the managed skill directory:

```bash
node bin/puretokens-skill.js install puretokens_media --target ~/.codex/skills
node bin/puretokens-skill.js upgrade puretokens_media --target ~/.codex/skills
```

## Model-selection rules

`puretokens_media` must call `puretokens_list_media_models` first and match only fields returned in that response: `id`, `displayName`, `aliases`, `provider`, and `capabilities`. Generation calls must include the exact `model` and a stable `request_id`. One logical user request submits once, requests one result unless the user explicitly gives a count, and reuses the same `request_id` on a host retry; result polling always uses the same `task_id` and original model.

Completed media reports the exact model, the saved filename, and `Downloads/Pure Tokens`. Images are previewable only when MCP or the host returns native image content; Direct Cloud downloads returned `b64_json`, returned URLs, or completed `/content` bytes before it claims delivery. Videos may include a bounded native MCP resource for hosts that render it; larger videos remain successfully delivered as local MP4 files. An open-file/open-folder entry is shown only when the execution layer actually returned one.

Behavior scenarios for ambiguity, an empty catalog, unavailable MCP, task failure, and polling timeout are stored in `skills/puretokens_media/references/behavior-scenarios.json`. No error may trigger an automatic model switch or resubmission unless the user explicitly chooses a new model.

## Security boundary

The Skill does not contain or request cloud credentials, Router tokens, cookies, passwords, user configuration, group routing, payment data, local authorization addresses, media, task results, or prompt history.

Trae is currently not supported by this media Skill flow.
