<p align="center">
  <strong>English</strong> · <a href="./CHANGELOG.zh-CN.md">中文</a>
</p>

# Changelog

## 0.13.16 — 2026-09-03

- Added a Codex-only legacy-plugin migration to `puretokens-update`. When synchronizing Codex's `~/.agents/skills`, it uses only the official Codex plugin interface to detect and remove the exact installed `puretokens-media` plugin after Skill synchronization. If the plugin cannot be inspected or removed, including workspace-managed cases, it reports the required Plugins or administrator action instead of claiming success.

## 0.13.15 — 2026-09-03

- A successful `puretokens-update` now reports the exact synchronized official Skill version after listing its installation, upgrade, and retired-Skill cleanup results.

## 0.13.14 — 2026-09-03

- Changed `puretokens-update` and both native installers to remove verified retired managed Skill directories after all current Skills have synchronized successfully. It also removes hidden retired backups left by earlier installers, so hosts cannot keep discovering obsolete `puretokens_media` instructions. Any unmanaged current, retired, or backup directory remains a preflight conflict and is never removed.

## 0.13.13 — 2026-09-03

- Replaced the user-facing Node-based Skill updater with official native platform installers: Shell on macOS/Linux and PowerShell on Windows. Installation and upgrades now download and statically validate the official `main` payload before synchronization, without requiring users to install Node, npm, Git, a package manager, or dependencies.

## 0.13.11 — 2026-09-03

- Made a current local visual-reference image use the selected model's declared multipart `image_edit` operation directly, while preserving the user's reference-image intent in the prompt. It never uploads, rehosts, invents a URL, or turns an unavailable attachment into text-only generation.
- Enforced declared image size-expression constraints: required pairs such as `width` plus `height` must be complete, and competing size expressions submit only the highest-precedence user-supplied form.
- Made video operation selection explicit for first-frame requests (`image_to_video` and its declared field, including Wan `first_frame_image`) and for declared `generate_audio` intent. An attached image with no stated first-frame/reference role now asks before billing.

## 0.13.10 — 2026-09-03

- Synchronized the installed media profiles with the current Pure Tokens base catalog. `gpt-image-2`, `seedream-5.0-pro`, `nano-banana-2`, `nano-banana-2-lite`, and `nano-banana-pro` now expose their declared native `image_edit` operations.
- Made `gpt-image-2` the default for an unspecified current local image-edit request as well as normal image generation. The Skill still sends attachments only through the declared operation and never switches models automatically.

## 0.13.9 — 2026-09-02

- Made an untransportable current local reference image fail closed. If the selected model does not declare a compatible native multipart reference transport, the Skill stops before any paid request and requires either the user's public HTTPS URL or a new explicit text-only request that ignores the attachment. It never inspects, summarizes, recreates, uploads, searches for, or silently converts that image into a text prompt.
- Made confirmed native-media delivery a hard turn boundary. After delivering the requested result, the Skill returns one completion receipt and stops; it does not inspect the output, search history or the workspace, invoke unrelated Skills, or start another media task without a new explicit user request.

## 0.13.8 — 2026-09-01

- Closed the WorkBuddy standard-input bypass in the direct runtime: every WorkBuddy POST now rejects `--json-stdin` and `--multipart-stdin` at argument validation, so only bounded Base64 request-body arguments are accepted.
- Made native video attachments route-locked. A declared image/video/audio operation must use its declared multipart path and field; it cannot degrade into a text-only JSON video request. For Seedance 2.0 Mini reference-image video, the required route is `reference_image_video` with 1–9 `reference_images` multipart files. If the host session cannot provide the current attachment bytes, the Skill stops before billing.

## 0.13.7 — 2026-09-01

- Fixed WorkBuddy media submissions that stalled before any API call: its Bash route now passes each bounded JSON request body or multipart descriptor as a canonical Base64 argument, not through standard input. The runtime decodes it only in memory and continues using the same fixed authenticated API path.

## 0.13.6 — 2026-09-01

- Made an unknown media submission a hard response boundary: after a started POST has unusable host output and no verified task ID, the Skill returns one receipt and ends the response without additional tools, status reads, polling, or prolonged follow-up work. A later continuation without an ID now requires explicit confirmation of a new paid request.

## 0.13.5 — 2026-09-01

- Simplified the client-download Agent prompt to one copyable sentence pointing to the official Pure Tokens Skill repository.
- Hardened the installation contract: the freshly cloned official `main` checkout is read-only. If validation fails or the checkout becomes dirty, installation stops; it must never run generators, docs sync, formatters, repair commands, Git writes, or any other source-mutating command to make the checkout pass.
- Prevented duplicate paid media submissions after a started POST has an empty, truncated, malformed, or otherwise unusable runtime result. The Skill now treats that outcome as unknown, retains only a verified returned task ID, and never creates a replacement task automatically.
- Fixed WorkBuddy credential resolution for its per-model Pure Tokens `/v1/...` resource URLs. The runtime still accepts only the fixed API origin with no query or fragment, requires one unambiguous credential, and sends requests only to fixed allowed API paths.

## 0.13.4 — 2026-09-01

- Added a deterministic model-access-group recovery path. When an exact requested image or video model is not returned by the current authenticated catalog after the permitted diagnosis, the Skill directs the user to select a group containing that model in Pure Tokens client configuration, ensure the managed key covers the selected groups, verify and apply it, then start a new host conversation and explicitly retry.
- The Skill does not guess a group name or claim model-to-group membership unless an authenticated API response explicitly provides that mapping.

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
