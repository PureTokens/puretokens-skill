# Pure Tokens Skill Repository

This repository is the private source of truth for installable Pure Tokens Skills. It owns Skill instructions, manifests, catalog schemas, validation, and the safe local installer. It does **not** own client configuration, credentials, API keys, or model-routing code.

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

- An API-facing **Skill** interprets the user's natural-language balance, API-identity, model-discovery, or media request. It calls the fixed public Pure Tokens API origin, `https://api.puretokensx.com`; it never derives a request target from a host connection configuration. A connection check reads only that API's public `GET /v1` declaration. A model query reads its authenticated media catalog once and reports only declared capabilities, parameters, and operations. A normal media submission resolves the default model or a unique alias from its installed versioned model selection and calls the full Images or Videos API directly; it must not make a model-catalog preflight request. The live catalog is read only for an explicit model-discovery request, an installed-profile gap needed to honor a requested option or media operation, or one post-rejection diagnosis; it never gates a normal submission or causes an automatic retry. The separate `puretokens-update` Skill handles only explicit local Skill installation or upgrade: it validates a fresh official `main` checkout and runs the managed `sync` command without accessing credentials, host configuration, or media APIs.
- Direct API authentication is supplied by the active runtime's existing Pure Tokens request configuration. The Skill never holds, reads, scans, asks for, displays, copies, or stores credentials, Base URLs, or host configuration. It must never construct an authorization header, inspect a configuration file, or route through MCP, a local proxy, or a sidecar.
- The **Pure Tokens API** remains authoritative for API-key group access and endpoint capabilities. A Skill must never infer access or capability from a model name.

Skills never inspect a provider label, a user-configured Base URL, service attribution, or credentials, and they make no compatibility or fallback branch for another relay. They always call the fixed public Pure Tokens API origin. If the direct request fails before a task is accepted, they report the returned failure and do not invent a configuration diagnosis, retry a billable submission, or use another endpoint. For a declared `multipart_file` media operation, the Skill sends only the current request's explicit image, video, or audio attachment bytes in that one Images or Videos API request. The API owns any required internal attachment handling; the Skill never calls a separate upload path, reads internal storage settings, or exposes internal URLs. Do not embed a cloud API key, localhost URL, or sidecar binary in this repository.

`references/media-model-catalog.json` is the sole source for published model aliases. `npm run docs:sync-media-models` derives the capability-specific `references/model-selection.json` files installed with image/video Skills and synchronizes both READMEs. Every specialist Skill includes an installed execution contract and behavior-scenario reference; update and test them with its `SKILL.md` in the same change. Do not create a combined always-apply media Skill for a host.

## Conventions

- Skill directory and front-matter `name` use the Agent Skills standard: lowercase kebab-case.
- Every skill has both `SKILL.md` and `skill.json`.
- Update `skills/index.json`, `CHANGELOG.md`, the Chinese README and the English README in the same change when adding or changing a skill.
- Keep user-facing language concise and deterministic. If a model cannot be selected uniquely, the skill asks the user instead of guessing.
- Do not add a GitHub Actions workflow unless the user explicitly asks to enable Actions for this repository.

## Validation

```bash
npm run check
node bin/puretokens-skill.js list
node bin/puretokens-skill.js install puretokens-image --target /private/tmp/puretokens-skill-test
```

Before committing, also run:

```bash
rg -n --hidden --glob '!node_modules/**' -- '(BEGIN [A-Z ]*PRIVATE|api[_-]?key|authorization:|bearer |/Users/|127\\.0\\.0\\.1:)' .
git status --short
```

Do not commit credentials, cookies, local configuration, generated media, task IDs, user prompts, desktop build artifacts, or a license file. This is a proprietary repository.
