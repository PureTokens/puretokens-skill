# Pure Tokens Skills repository rules

This repository owns the six installable Skill instructions, model selections, contracts, documentation, validation, platform installers, and the small one-shot native API executor. It does not own, display, persist or manage user connection configuration or credentials.

## Direct API product contract

- API-facing Skills always use full fixed URLs under `https://api.puretokensx.com`; fixed requests never reuse a configured Base URL as their target.
- The managed `puretokens-api` native executor performs every fixed request directly. It runs only for one command, exits after its receipt, and never opens a port, proxy, sidecar, background queue, or desktop UI.
- A credential adapter may read only the exact effective connection files documented for the active host. It first verifies that the configured endpoint is Pure Tokens, then uses one matching bearer credential only in memory for the fixed request. It must never scan a home directory, probe third-party configuration, use provider labels, or fall back to arbitrary environment variables.
- Do not display, log, copy, persist, ask for, compare or report credentials, Base URLs, provider labels or host configuration. Do not add or require Node, npm, Python, MCP, a proxy, sidecar, separate upload service, or another authentication path.
- Computer Use, browser automation, desktop/mobile UI automation, and opening or clicking Pure Tokens Switch/Desktop are never API transports. Do not use them to submit a request, poll a task, retrieve media, or recover from an actual terminal/network failure.
- API-facing Skills must not invoke Imagen or another image/video Skill as a fallback. Never gate the fixed request on a separate “authenticated image/video interface” discovery step: use the active connection credential and issue the fixed API request through the current host's terminal or native HTTPS/API capability. Only an actual terminal, network, credential-resolution, attachment-byte, or API failure may produce the documented failure receipt.
- The fixed endpoints are `/v1`, `/v1/media/models`, `/v1/dashboard/billing/subscription`, `/v1/dashboard/billing/usage`, `/v1/images/generations`, `/v1/images/edits`, `/v1/videos`, and `/v1/videos/edits`, plus declared same-task status/content paths.
- JSON and declared multipart requests are sent only by the native executor. Current user attachments are sent only with their declared operation; never upload, rehost, convert them to a text prompt, or silently downgrade a media-reference request.
- If an active host has no verified credential adapter, if policy blocks the executor, if attachment transport is unavailable, or if native byte delivery fails, stop with the documented safe receipt. Do not suggest a fallback transport.

## Skill behavior

- Media submissions are asynchronous. Submit once and immediately return the receipt before any wait or download; use separate status/wait/content commands, retain only the same returned task ID, and never automatically resubmit after uncertain output, status failure, timeout or delivery failure.
- Do not make `/v1/media/models` a preflight for ordinary core generation. Read the small installed model index, then only the selected model profile; read the live catalog only for explicit discovery, an installed-profile gap, or post-rejection diagnosis.
- Treat physical dimensions as a request for guidance, not an API size value. Only declared model fields and values may be sent.
- Native bytes, not URLs, HTML, SVG, task IDs or status text, are required for successful delivery.
- Keep every user-facing failure sanitized. Never expose raw responses, request data, internal URLs, stacks, upstream identities, credentials or user media.

## Installation and updates

- Shell/PowerShell installers synchronize the six Skills and exactly one checksum-verified platform executor. They never install Node, npm, Python, a proxy, or a service.
- An update may replace only a verified managed `.puretokens-executor`, remove a verified old `puretokens-direct-api-runtime`, and remove verified retired official Skill directories. Never delete unknown directories.
- Preserve the exact first `text` install prompt under the required heading in both READMEs because the client download page extracts it.

## Validation

Run `npm run check`, `npm run release:validate`, and `go test ./...` from `runtime/executor` before release. Node is permitted for repository development and validation only; it is never a user runtime dependency for media work.

- Credential adapter fixture coverage is not real-host end-to-end acceptance; record the latter separately.
- User output files are explicit artifacts, never a hidden cache. Download one index per invocation; reuse a validated same-task file in the same directory and attach it before requesting the next index.
