<p align="center">
  <img src="./assets/brand/puretokens-skill-hero.png" alt="Pure Tokens Skill — one skill, every model" width="100%" />
</p>

<p align="center">
  <strong>English</strong> · <a href="./README.zh-CN.md">中文</a> · <a href="./CHANGELOG.md">Changelog</a>
</p>

# Pure Tokens Skills

`puretokens-skill` is the source repository for Pure Tokens Skills. It owns Skill instructions, versions, compatibility declarations, client installation instructions, and validation tools. It does not contain user credentials, Router configuration, or model-routing logic.

## 3-step quick start

1. In Pure Tokens Desktop, select a group for your client that contains the image or video model you plan to use, then click **Verify and apply**.
2. For WorkBuddy, **Verify and apply** automatically installs the always-on delivery generated from the same `puretokens_media` source. For other clients, install `puretokens_media` using the client table or the copyable agent prompt below.
3. Start a new chat. Say `Generate a cute dog` for the default image model, or name a model such as `Use Nano Banana Pro to generate ...`.

> **Before using a specific model:** the model must belong to at least one group selected for that client. For example, select a group containing `gpt-image-2` before asking the Skill to use `image2`. After changing groups, click **Verify and apply**, restart the target client, and start a new chat. The Skill can use only models returned from the currently selected group or groups; it cannot use every model in the public catalog.

Current Skill:

| Skill | Purpose |
| --- | --- |
| `puretokens_media` | Select an exact image/video model from the live catalog, submit one task, poll the same task, and deliver its native result plus the local file. |

## Image models

These are the image models currently shown by the public catalog. Your client/group may show fewer models. The Skill uses a model only when the live `puretokens_list_media_models` response contains the exact ID or alias.

You do not need to type the full ID. Registered phrases such as `image2` and `Nano Banana Pro` are understood by the Skill and verified against the live catalog before a request is sent. If you simply ask for an image, the Skill uses `gpt-image-2`; if that exact model is unavailable in your current group, it stops and shows the available candidates instead of silently switching models.

| Model ID | You can also say | Good for | Real example |
| --- | --- | --- | --- |
| `gpt-image-2` | `image2`, `gpt image 2`, `openai image 2` | High-quality posters, product visuals, illustrations | `Use image2 to make a clean orange product launch poster.` |
| `gemini-3.0-pro-image` | `gemini pro image` | Detailed concept art and polished marketing images | `Use gemini pro image to create a premium cloud-computing hero image.` |
| `gemini-3.1-flash-lite-image` | `gemini flash lite image` | Fast thumbnails and social media drafts | `Use gemini flash lite image to make three bright social thumbnails.` |
| `gemini-3-pro-image-preview` | `nano banana pro` | Detailed, professional Gemini image work | `Use Nano Banana Pro to create a premium product key visual.` |
| `gemini-3.1-flash-image-preview` | `nano banana 2` | Faster Gemini image generation and conversational edits | `Use Nano Banana 2 to create a bright product social post.` |
| `grok-imagine-1.0` | `grok image`, `grok imagine` | Fast creative concepts and playful scenes | `Use grok-imagine-1.0 to draw a cheerful robot in a city park.` |
| `grok-imagine-image` | `grok image`, `grok imagine` | Social posts and everyday image generation | `Use grok-imagine-image to create a realistic café opening post.` |
| `grok-imagine-image-quality` | `grok quality image` | Sharper brand key visuals | `Use grok quality image to make a polished app-store banner.` |
| `wan2.7-image` | `wan image`, `wan 2.7 image` | Chinese posters and product creatives | `Use wan 2.7 image to make a Chinese New Year promotion poster.` |

The Skill calls `puretokens_generate_image` once, then polls the same task with `puretokens_image_result`.

`Nano Banana` by itself means the Gemini Nano Banana family. When both `Nano Banana Pro` and `Nano Banana 2` are available, the Skill asks which one you want. When only one is available, it uses that one. This keeps a named choice from turning into an invisible model substitution.

## Video models

These are the video models currently shown by the public catalog. A model must have live capability `video` before it can be used.

| Model ID | You can also say | Good for | Real example |
| --- | --- | --- | --- |
| `grok-imagine-video` | `grok video`, `grok imagine video` | Short social clips and quick concepts | `Use grok-imagine-video to make a 5-second coffee ad.` |
| `grok-imagine-video-1.5` | `grok 1.5 video`, `grok video 1.5` | More polished short advertisements | `Use grok 1.5 video to make a 15-second 16:9 product ad.` |

The Skill calls `puretokens_generate_video` once, then polls the same task with `puretokens_video_result`. If you simply ask for a video, it uses `grok-imagine-video-1.5`; if that exact model is unavailable in your current group, it stops and shows the available candidates.

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
Natural-language user request → Skill → Pure Tokens MCP → local Router → Pure Tokens service
```

- The Skill interprets requests such as “use image2” or “use Grok Video”, reads the live catalog, asks for clarification when the match is not unique, and selects the correct tool.
- MCP accepts only an exact model ID. It validates arguments, submits once, and polls the result. MCP never performs natural-language matching, guesses a model, or silently substitutes one.
- BFF and Router remain authoritative for model availability, group access, and media protocol.

## Prerequisites

Before using a specific image or video model, complete these steps in order:

1. In Pure Tokens Desktop, open the configuration for the target client.
2. Select one or more groups that contain the target model.
3. Click **Verify and apply**.
4. Restart the target client and start a new chat.

Only models in the selected group or groups are available to the Skill. If the target model is absent from the live media catalog, return to client configuration, select a group that contains it, and apply the configuration again. Desktop configures an MCP server named `puretokens-image` for supported clients. The Skill does not replace MCP configuration and never carries credentials.

## Install and update from GitHub

`puretokens_media` is the single behavior source for every supported client. Claude Desktop receives it as an uploadable ZIP; WorkBuddy receives a Desktop-managed, always-on generated delivery after **Verify and apply**, so generic image/video requests use the configured Pure Tokens MCP. Always use this repository for the current shared-Skill installation instructions and files. Before installing a shared Skill, complete **Verify and apply** for the target client, then restart that client and start a new chat.

### Codex, Claude Code, Gemini CLI, and OpenCode

Use **Code → Download ZIP** on this repository page, or clone the repository. The manager requires Node.js 20 or later.

```bash
git clone https://github.com/yanyansay/puretokens-skill.git
cd puretokens-skill
```

Install into the target client's user Skill directory:

```bash
# Codex
node bin/puretokens-skill.js install puretokens_media --target ~/.codex/skills

# Claude Code
node bin/puretokens-skill.js install puretokens_media --target ~/.claude/skills

# Gemini CLI
node bin/puretokens-skill.js install puretokens_media --target ~/.gemini/skills

# OpenCode
node bin/puretokens-skill.js install puretokens_media --target ~/.config/opencode/skills
```

### Copy this to your agent

Paste this into an agent that can run local commands:

```text
Install Pure Tokens Skill for the client I am using from https://github.com/yanyansay/puretokens-skill.

1. Identify whether this is Codex, Claude Code, Gemini CLI, or OpenCode.
2. Clone or download the repository into a temporary working directory.
3. Install or upgrade only `puretokens_media` in the matching user Skill directory:
   - Codex: ~/.codex/skills
   - Claude Code: ~/.claude/skills
   - Gemini CLI: ~/.gemini/skills
   - OpenCode: ~/.config/opencode/skills
5. Do not overwrite any other Skill.
6. Do not read, request, print, or store API keys, cookies, passwords, Router tokens, or local authorization URLs.
7. Tell me the installed directory and whether the operation succeeded.

If this is Claude Desktop, do not claim it was installed automatically. Build the ZIP following the README, then tell me exactly where to upload and enable it. For WorkBuddy, tell me to use Pure Tokens Desktop's **Verify and apply** instead; do not manually create or replace its generated `puretokens_workbuddy_router` delivery.
```

To update, pull or download the latest repository contents and run the matching `upgrade` command:

```bash
git pull
node bin/puretokens-skill.js upgrade puretokens_media --target ~/.codex/skills
```

Replace `~/.codex/skills` with the target directory for the client you are updating. Upgrade replaces only a Pure Tokens-managed directory containing matching `skill.json` and `SKILL.md`; it never overwrites another Skill.

### Windows PowerShell

```powershell
git clone https://github.com/yanyansay/puretokens-skill.git
Set-Location puretokens-skill
node .\bin\puretokens-skill.js install puretokens_media --target "$HOME\.codex\skills"
```

Use `$HOME\.claude\skills`, `$HOME\.gemini\skills`, or `$HOME\.config\opencode\skills` for the other clients. If PowerShell cannot find `node`, install Node.js LTS from the official Node.js website and reopen PowerShell.

## Claude Desktop import and WorkBuddy routing

Copying a Skill to `~/.codex/skills` is not a Claude Desktop installation. That directory is only for Codex local Skills.

Claude Desktop uses a graphical local Skill upload. Create the ZIP:

```bash
node bin/puretokens-skill.js bundle puretokens_media --format claude-desktop --out ./puretokens_media-0.2.9.zip
```

The ZIP has this layout:

```text
puretokens_media/
├── SKILL.md
├── skill.json
└── references/
    ├── behavior-scenarios.json
    └── model-catalog-contract.md
```

In Claude Desktop, open **Settings → Features → Skills** (some builds show **Customize → Skills**), choose **Upload skill**, upload the ZIP, and enable `Pure Tokens Media`. If the installed build has no Skills entry, it cannot import custom Skills; that build can still use MCP tool descriptions, but it will not receive this Skill's deterministic model-selection and no-fallback policy.

For WorkBuddy, do not upload or enable a separate Skill yourself. Selecting a compatible group and clicking **Verify and apply** in Pure Tokens Desktop atomically renders and manages the always-on `puretokens_workbuddy_router` delivery from the shared `puretokens_media` source, along with the `puretokens-image` MCP entry and its reference files. Restart WorkBuddy or start a new chat afterwards. Bare image/video requests then prefer Pure Tokens; an explicit request for WorkBuddy's built-in `ImageGen` or `VideoGen` still keeps that user choice.

To update Claude Desktop, get the new version from GitHub, generate a new ZIP, disable the old Skill, upload the new ZIP, and enable it. WorkBuddy regenerates the same shared media behavior on the next **Verify and apply**. Do not delete the MCP entry directly; Pure Tokens Desktop owns MCP configuration.

## Codex local install, upgrade, and uninstall

The default Codex Skill directory is:

```bash
node bin/puretokens-skill.js list
node bin/puretokens-skill.js install puretokens_media
node bin/puretokens-skill.js upgrade puretokens_media
node bin/puretokens-skill.js uninstall puretokens_media --yes
```

An explicit project directory is also supported:

```bash
node bin/puretokens-skill.js install puretokens_media --target .codex/skills
node bin/puretokens-skill.js upgrade puretokens_media --target .codex/skills
node bin/puretokens-skill.js uninstall puretokens_media --target .codex/skills --yes
```

Upgrade atomically moves the old managed directory to a temporary backup and cleans it up only after replacement succeeds. Uninstall requires explicit `--yes` and removes only a managed directory containing a matching `skill.json` and `SKILL.md`.

## Model-selection rules

`puretokens_media` must call `puretokens_list_media_models` first and match only fields returned in that response: `id`, `displayName`, `aliases`, `provider`, and `capabilities`. Generation calls must include the exact `model` and a stable `request_id`. One logical user request submits once; a host retry reuses the same `request_id`; result polling always uses the same `task_id` and original model.

Completed media reports the exact model returned by MCP, the saved filename, and `Downloads/Pure Tokens`. Images are previewable only when MCP returns native `image` content. Videos may include a bounded native MCP resource for hosts that render it; larger videos remain successfully delivered as local MP4 files and open from the same Downloads folder.

Behavior scenarios for ambiguity, an empty catalog, unavailable MCP, task failure, and polling timeout are stored in `skills/puretokens_media/references/behavior-scenarios.json`. No error may trigger an automatic model switch or resubmission unless the user explicitly chooses a new model.

## Security boundary

The Skill does not contain or request cloud credentials, Router tokens, cookies, passwords, user configuration, group routing, payment data, local authorization addresses, media, task results, or prompt history.

Trae is currently not supported by this media Skill flow.
