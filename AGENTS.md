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

- A **Skill** interprets the user's natural-language balance or media request. A balance request uses only a host-exposed, authenticated, read-only capability and never infers a value. A media request resolves an explicit model selection through the current configured connection's authenticated media catalog, requests only relative API paths and JSON bodies, and asks a clarification question when the request is ambiguous.
- The host owns the configured Base URL, authentication, routing, HTTP execution, and native-media delivery. The Skill never holds, reads, scans, asks for, displays, or stores credentials or host configuration.
- The **Pure Tokens API** remains authoritative for API-key group access and endpoint capabilities. The **Desktop Router** remains authoritative for Desktop-managed profile access and `openai_images` / `openai_video` routing. A Skill must never infer either from a model name.

Skills do not inspect or validate a connection's Base URL, provider label, service attribution, or credentials. They make no compatibility or fallback branch for another relay: they use the current configured connection's relative API paths once, or stop before a billable submission when that connection cannot execute the request or deliver native media bytes. Do not embed a Router token, cloud API key, localhost URL, or sidecar binary in this repository.

`references/media-model-catalog.json` is the sole source for published model aliases. `npm run docs:sync-media-models` derives the capability-specific `references/model-selection.json` files installed with image/video Skills and synchronizes both READMEs. Every specialist Skill includes an installed execution contract and behavior-scenario reference; update and test them with its `SKILL.md` in the same change. Do not create a combined always-apply media Skill for a host.

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
node bin/puretokens-skill.js install puretokens_image --target /private/tmp/puretokens-skill-test
```

Before committing, also run:

```bash
rg -n --hidden --glob '!node_modules/**' -- '(BEGIN [A-Z ]*PRIVATE|api[_-]?key|authorization:|bearer |pts-router-token|/Users/|127\\.0\\.0\\.1:)' .
git status --short
```

Do not commit credentials, cookies, local configuration, generated media, task IDs, user prompts, desktop build artifacts, or a license file. This is a proprietary repository.
