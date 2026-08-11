<p align="center">
  <img src="./assets/brand/puretokens-skill-hero.png" alt="Pure Tokens Skill — one skill, every model" width="100%" />
</p>

<p align="center">
  <strong>English</strong> · <a href="./README.zh-CN.md">中文</a>
</p>

# Pure Tokens Skills

`puretokens-skill` is the source repository for Pure Tokens Skills. It owns Skill instructions, versions, compatibility declarations, client installation instructions, and validation tools. It does not contain user credentials, Router configuration, or model-routing logic.

Current Skill:

| Skill | Purpose |
| --- | --- |
| `puretokens_media` | Select an exact image or video model from the live catalog, submit one task through `puretokens-image`, and poll that same task. |

## Supported models and usage

The model catalog is live. The table below lists model families and current examples; a model is usable only when `puretokens_list_media_models` returns its exact `id` (or an exact returned alias) for the active client and group.

| Media | Catalog capability | Model IDs or aliases you may see | How to use |
| --- | --- | --- | --- |
| Image | `image` | `gpt-image-2`, `codex-gpt-image-2`, `gemini-3-pro-image-preview`, `gemini-3.1-flash-image-preview`, `grok-imagine-image`, `grok-imagine-image-lite`, `grok-imagine-image-pro` | “Use `gpt-image-2` to generate …” → `puretokens_generate_image` |
| Video | `video` | Any exact video ID returned by the catalog, such as `grok-imagine-video-1.5`, `seedance-2.0`, or `ltx-2.3-fat` when configured | “Use `grok-imagine-video-1.5` to make a 5-second ad” → `puretokens_generate_video` |
| Both | `image` + `video` | A model whose live entry declares both capabilities | Ask which medium is intended, then call only the matching tool |

Model names are not guessed from spelling. If a name is missing, ambiguous, or has no requested capability, the Skill shows the live candidates and asks you to choose. It never silently substitutes a model.

### Usage examples

| User request | Required behavior |
| --- | --- |
| `Use image2 to generate a cute dog.` | List the catalog → match an exact returned `id`/alias → submit once with `puretokens_generate_image` → poll the same task. |
| `Use gpt-image-2, square, high quality.` | Match the exact ID → pass `size` and `quality` if the tool supports them → poll the same task. |
| `Use Grok Video to make a 15-second 16:9 product ad.` | List the catalog → require one exact video match → call `puretokens_generate_video` with `seconds` and `aspect_ratio`. |
| `Show me the available image models.` | Call `puretokens_list_media_models` and show only entries whose live capabilities include `image`. |

## Boundaries

```text
Natural-language user request → Skill → Pure Tokens MCP → local Router → Pure Tokens service
```

- The Skill interprets requests such as “use image2” or “use Grok Video”, reads the live catalog, asks for clarification when the match is not unique, and selects the correct tool.
- MCP accepts only an exact model ID. It validates arguments, submits once, and polls the result. MCP never performs natural-language matching, guesses a model, or silently substitutes one.
- BFF and Router remain authoritative for model availability, group access, and media protocol.

## Prerequisites

The user must first complete “Verify and apply” for the current client in Pure Tokens Desktop and restart that client. Desktop configures an MCP server named `puretokens-image` for supported clients. The Skill does not replace MCP configuration and never carries credentials.

## Install and update from GitHub

Pure Tokens Desktop does not write Skill files into client directories and does not tie Skill contents to a Desktop release. Always use this repository for the current installation instructions and Skill files. Before installing, complete “Verify and apply” for the target client in Pure Tokens Desktop, then restart that client and start a new chat.

### Codex, Claude Code, Gemini CLI, and OpenCode

Use **Code → Download ZIP** on this repository page, or clone the repository. The manager requires Node.js 20 or later.

```bash
git clone https://github.com/yanyansay/puretokens-skill.git
cd puretokens-skill
node bin/puretokens-skill.js validate
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

### Copyable agent installation prompt

Paste this into an agent that can run local commands. Replace `<client>` with `Codex`, `Claude Code`, `Gemini CLI`, or `OpenCode`:

```text
Install Pure Tokens Skill for <client> from the public repository https://github.com/yanyansay/puretokens-skill.

1. Clone or download the repository into a temporary working directory.
2. Run `node bin/puretokens-skill.js validate`.
3. Install only `puretokens_media` into the correct user Skill directory:
   - Codex: ~/.codex/skills
   - Claude Code: ~/.claude/skills
   - Gemini CLI: ~/.gemini/skills
   - OpenCode: ~/.config/opencode/skills
4. Use the manager command `install` or `upgrade`; do not overwrite other Skills.
5. Do not read, request, print, or store API keys, cookies, passwords, Router tokens, or local authorization URLs.
6. Report the exact target directory and validation result. Do not claim Claude Desktop or WorkBuddy is installed by this prompt; those clients require the ZIP upload flow in the README.
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
node .\bin\puretokens-skill.js validate
node .\bin\puretokens-skill.js install puretokens_media --target "$HOME\.codex\skills"
```

Use `$HOME\.claude\skills`, `$HOME\.gemini\skills`, or `$HOME\.config\opencode\skills` for the other clients. If PowerShell cannot find `node`, install Node.js LTS from the official Node.js website and reopen PowerShell.

## Import into Claude Desktop and WorkBuddy

Copying a Skill to `~/.codex/skills` is not a Claude Desktop installation. That directory is only for Codex local Skills.

Claude Desktop and WorkBuddy use a graphical local Skill upload. Create the ZIP:

```bash
node bin/puretokens-skill.js validate
node bin/puretokens-skill.js bundle puretokens_media --format claude-desktop --out ./puretokens_media-0.2.0.zip
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

In WorkBuddy, open **Skills → Add Skill → Upload skill**, choose the same ZIP, confirm `Pure Tokens Media` appears in the installed list, enable it, and start a new chat. WorkBuddy owns the import and local configuration; Pure Tokens Desktop does not write to an undocumented WorkBuddy directory.

To update, get the new version from GitHub, generate a new ZIP, disable the old Skill, upload the new ZIP, and enable it in Claude Desktop or WorkBuddy. To uninstall, disable and delete the Skill from the client's Skills page. Do not delete the MCP entry directly; Pure Tokens Desktop owns MCP configuration.

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

`puretokens_media` must call `puretokens_list_media_models` first and match only fields returned in that response: `id`, `displayName`, `aliases`, `provider`, and `capabilities`. Generation calls must include the exact `model` and a stable `request_id`. One logical user request submits once; a host retry reuses the same `request_id`; result polling always uses the same `task_id`.

Behavior scenarios for ambiguity, an empty catalog, unavailable MCP, task failure, and polling timeout are stored in `skills/puretokens_media/references/behavior-scenarios.json`. No error may trigger an automatic model switch or resubmission unless the user explicitly chooses a new model.

## Security boundary

The Skill does not contain or request cloud credentials, Router tokens, cookies, passwords, user configuration, group routing, payment data, local authorization addresses, media, task results, or prompt history.

Trae is currently not supported by this media Skill flow.

## Validation

```bash
npm run check
node bin/puretokens-skill.js validate
npm test
```

This repository does not enable GitHub Actions, publish an npm package, or automatically publish a Claude Desktop Skill. It becomes active in Claude Desktop only after the generated bundle is uploaded and enabled.
