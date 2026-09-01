# Maintainer Guide

This document is for maintainers and contributors. End users only need the installation and client import instructions in `README.md` or `README.zh-CN.md`.

## Local checks

Run these commands from the repository root before committing Skill changes:

```bash
npm run check
node bin/puretokens-skill.js validate
npm test
```

`install`, `upgrade`, and `sync` also validate the selected Skill source before changing files. The explicit commands above are the full maintainer gate and should not be added to the end-user README.

## Version consistency

When releasing a Skill update, keep these values aligned:

- `package.json`
- `skills/index.json`
- `skills/<skill-name>/skill.json`
- README examples when the user-facing Skill name changes

Update `CHANGELOG.md` with the user-visible behavior change.

## Distribution policy

- This repository does not publish an npm package.
- This repository does not use GitHub Actions for automatic Skill publication.
- Claude Code, Codex, WorkBuddy, Gemini CLI, Grok Build, OpenCode, and Trae install the shared source into the documented global Skill directory. The authoritative support matrix is `references/host-support.json`.
- Do not put API keys, cookies, passwords, or local authorization URLs in the repository or an installable Skill.
