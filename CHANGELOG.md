<p align="center">
  <strong>English</strong> · <a href="./CHANGELOG.zh-CN.md">中文</a>
</p>

# Changelog

## 0.13.3 — 2026-09-01

- Extended the managed fixed-endpoint credential runtime to every supported host with an approved, auditable local connection contract: Claude Code, Codex, WorkBuddy, Gemini CLI, Grok Build, and OpenCode. Each resolver accepts only an exact Pure Tokens configuration match, keeps one unambiguous credential only in memory for an allowed fixed API request, and never displays, persists, or requests it.
- Updated all API-facing Skills, installed contracts, host matrix, and validation to use those six managed runtimes consistently for API identity, catalog, balance, Images, Videos, same-task status, and native-media delivery.
- Marked Trae explicitly as the single manual-credential-setup exception. Its product contract has no approved local credential resolver, so the Skill stops safely instead of scanning or inferring Trae user state.

## 0.13.2 — 2026-09-01

- Added the verified WorkBuddy managed direct-runtime binding. It narrowly resolves only a matching `https://api.puretokensx.com/v1` model credential from WorkBuddy's model configuration in memory, uses it only for the fixed allowed API paths, and never displays, persists, or requests it.
- Made the host matrix explicit: all seven hosts remain installable, while only Grok Build and WorkBuddy are verified for authenticated direct media execution. Other hosts now stop safely instead of implying that generic HTTP requests inherit their chat-model credentials.
- Added bounded declared multipart attachment submission for the verified runtime and hardened native media delivery: fixed request-path allowlists, regular-file-only attachments, attachment count and size limits, content-type checks, bounded output, no overwrite, and cleanup of partial failed downloads.
- Removed contradictory wording that said a Skill could not read configuration while its managed runtime performed the narrow in-memory credential binding. The Skill still never displays, compares, or reports host configuration.

## 0.13.1 — 2026-09-01

- Added the managed Grok Build direct API runtime. It narrowly resolves the configured Pure Tokens model credential only in memory, sends it only to fixed allowed Pure Tokens API paths, and never prints, stores, or requests the credential.
- `sync`, `install`, and `upgrade` now install or atomically upgrade the managed `.puretokens-runtime` beside the specialist Skills, refusing an unmanaged runtime conflict before changing Skill directories.
- Replaced the incorrect assumption that every host automatically injects a configured model credential into generic HTTP requests. Grok Build Skills now use the managed direct runtime rather than generic Fetch/WebFetch for API identity, media catalog, balance, Images, Videos, task status, and media content reads.

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
