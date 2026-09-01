<p align="center">
  <strong>English</strong> · <a href="./CHANGELOG.zh-CN.md">中文</a>
</p>

# Changelog

## 0.13.0 — 2026-09-01

- Standardized the six installable Skill names and directories to Agent Skills kebab-case: `puretokens-balance`, `puretokens-connection`, `puretokens-models`, `puretokens-image`, `puretokens-video`, and `puretokens-update`.
- `sync` now moves verified retired managed Skill directories into a recoverable hidden backup before installing the current directories; it still refuses to alter an unmanaged conflict.
- Declared and validated the current seven supported hosts only: Claude Code, Codex, WorkBuddy, Gemini CLI, Grok Build, OpenCode, and Trae. Each uses a manual source install in its documented global Skill directory.
- Removed retired delivery mechanisms and combined-media-Skill descriptions from the active product surface.
- Made the direct execution contract explicit for every supported host: existing runtime authentication, fixed full-URL HTTPS requests, JSON task responses, same-task status reads, and native media-byte delivery. Images and videos have no alternate execution path.

## 0.12.3 — 2026-08-31

- Aligned image and video task handling with the public media contract: `reconciliation_required: true` takes precedence over lifecycle state, stops ordinary polling, retains the same task ID, and never triggers a replacement task or inferred refund.
- Status-read HTTP `429` honors a valid in-budget `Retry-After` and continues only the same task; non-429 server, transport, and timeout reads remain bounded stop conditions.
- Split profile-gated image editing by transport: public URL edits require an exact declared JSON `image_edit` operation, while native attachments require the exact declared multipart operation. A reference field alone never authorizes an edit or silent downgrade.

## 0.12.2 — 2026-08-31

- Standardized safe image and video failure receipts with phase, returned HTTP status when available, sanitized user-facing message, actionable next step, and rate-limit wait time when supplied.
- Explicitly prohibit raw error bodies, upstream identifiers, internal URLs, stacks, request data, credentials, and user media from failure output.

## 0.12.1 — 2026-08-31

- Improved image request interpretation: distinguish text-to-image, visual-reference, and edit requests; ask when an input image role is ambiguous; and preserve the user's stated intent when preparing a concise generation brief.
- Clarified that `n` is for variants of one brief only. Distinct requested assets are not silently packed into `n`, dropped, or split into multiple paid submissions.

## 0.12.0 — 2026-08-30

- Normal text-to-image and text-to-video submissions use installed versioned model selection and submit directly without a media-catalog preflight request.
- The live media catalog is read only on demand for explicit discovery, an installed-profile gap, or one post-rejection diagnosis. A failed catalog read cannot block valid core generation or trigger automatic resubmission.
- Hardened asynchronous task handling: normalize only declared top-level task IDs, stop on unknown states, preserve an uncertain submission outcome, and never automatically resubmit.
