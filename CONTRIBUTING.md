# Maintainer Guide

This document is for maintainers and contributors. End users only need the installation and client import instructions in `README.md` or `README.zh-CN.md`.

## Local checks

Run these commands from the repository root before committing Skill changes:

```bash
npm run check
node bin/puretokens-skill.js validate
npm test
```

`install`, `upgrade`, and `bundle` also validate the selected Skill source before changing or creating files. The explicit commands above are the full maintainer gate and should not be added to the end-user README.

## Version consistency

When releasing a Skill update, keep these values aligned:

- `package.json`
- `skills/index.json`
- `skills/<skill-name>/skill.json`
- bundle filenames and README examples when the user-facing bundle name changes

Update `CHANGELOG.md` with the user-visible behavior change.

## Distribution policy

- This repository does not publish an npm package.
- This repository does not use GitHub Actions for automatic Skill publication.
- Codex, Claude Code, and Gemini CLI install the shared source into the documented global Skill directory. Claude Desktop uses a maintainer-built ZIP that the user uploads and enables. The authoritative support matrix is `references/host-support.json`; do not imply support from a Router or connection adapter alone.
- Do not put API keys, cookies, passwords, Router tokens, or local authorization URLs in the repository or a Skill bundle.
