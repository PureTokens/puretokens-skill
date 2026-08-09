# Pure Tokens Skills

`puretokens-skill` is the private source of truth for installable Pure Tokens skills. It owns skill instructions, versions, compatibility declarations, and local validation tools. It does not contain API keys, Router tokens, user configuration, or routing logic.

The initial skill is `puretokens_media`, which lets MCP-capable clients select and invoke Pure Tokens image and video models.

## Boundaries

```text
Natural-language user request → Skill → Pure Tokens MCP → local Router → Pure Tokens service
```

- The Skill interprets the request, queries the media catalog, resolves an unambiguous user-selected model, and asks for clarification when necessary.
- MCP accepts an exact model ID, validates it, submits one task, and polls its result. MCP never performs natural-language matching, guesses a model, or silently substitutes a model.
- The BFF and Router remain authoritative for group access and media protocol availability.

## Local installation

```bash
git clone https://github.com/yanyansay/puretokens-skill.git
cd puretokens-skill
node bin/puretokens-skill.js list
node bin/puretokens-skill.js install puretokens_media
```

The default install target is `~/.codex/skills/puretokens_media/`. Use `--target .codex/skills` to install into a project.

The media skill requires a client that was configured through Pure Tokens Desktop. Desktop installs the `puretokens-image` MCP entry for Claude Code, Claude Desktop, ChatGPT/Codex, WorkBuddy, Gemini CLI, Grok Build, and OpenCode. Trae is not automatically configured for this MCP flow.

Run `npm run check` before committing. This repository intentionally has no GitHub Actions workflow and is not published to npm.
