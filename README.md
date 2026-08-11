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
