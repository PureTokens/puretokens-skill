# Pure Tokens Skill Repository

This repository is the private source of truth for installable Pure Tokens skills. It owns skill instructions, manifests, catalog schemas, validation and the safe local installer. It does **not** own the desktop MCP binary, Router, cloud credentials, API Keys, or model routing code.

## Structure

```text
skills/index.json                         # installable skill registry
skills/<skill-name>/SKILL.md              # Codex skill instruction entry point
skills/<skill-name>/skill.json            # machine-readable manifest
skills/<skill-name>/references/           # bounded, versioned reference material
schemas/                                  # catalog and manifest schemas
bin/puretokens-skill.js                   # list/install/validate CLI
scripts/                                  # repository validation helpers
```

## Product boundary

- A **Skill** interprets the user's natural-language request, resolves an explicit model selection through an authenticated media catalog, and asks a clarification question when the request is ambiguous. It selects an execution branch but does not itself hold credentials or execute HTTP.
- The **Pure Tokens MCP** exposes strict typed tools, accepts only an exact model ID, submits and polls tasks, and never performs natural-language matching or silent model fallback. It is required only for hosts without native Shell/HTTPS tool execution.
- The **Pure Tokens API** remains authoritative for API-key group access and endpoint capabilities. The **Desktop Router** remains authoritative for Desktop-managed profile access and `openai_images` / `openai_video` routing. A skill must never infer either from a model name.

The media skill supports a capability branch: use the registered `puretokens-image` MCP tools when available; otherwise an Agent with HTTPS execution and a host-injected `PURETOKENS_ACCESS_TOKEN` may use the documented Direct Cloud contract. Do not embed a Router token, a cloud access token, a localhost URL, or a sidecar binary in this repository.

## Conventions

- Skill directory and front-matter `name` use `snake_case`.
- Every skill has both `SKILL.md` and `skill.json`.
- Update `skills/index.json`, `CHANGELOG.md`, the Chinese README and the English README in the same change when adding or changing a skill.
- Keep user-facing language concise and deterministic. If a model cannot be selected uniquely, the skill asks the user instead of guessing.
- Do not add a GitHub Actions workflow unless the user explicitly asks to enable Actions for this repository.

## Validation

```bash
npm run check
node bin/puretokens-skill.js list
node bin/puretokens-skill.js install puretokens_media --target /private/tmp/puretokens-skill-test
```

Before committing, also run:

```bash
rg -n --hidden --glob '!node_modules/**' -- '(BEGIN [A-Z ]*PRIVATE|api[_-]?key|authorization:|bearer |pts-router-token|/Users/|127\\.0\\.0\\.1:)' .
git status --short
```

Do not commit credentials, cookies, local configuration, generated media, task IDs, user prompts, desktop build artifacts, or a license file. This is a proprietary repository.
