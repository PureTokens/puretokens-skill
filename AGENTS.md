# Pure Tokens Skills repository rules

This repository owns the six installable Skill instructions, model selections, contracts, documentation, validation and source-only installers. It does not own, display, persist or manage user connection configuration or credentials.

## Direct API product contract

- API-facing Skills always use full fixed URLs under `https://api.puretokensx.com`; fixed requests never reuse a configured Base URL as their target.
- The current host's terminal or native HTTPS/API capability performs the fixed request directly. It may resolve exactly one current matching credential from the active host connection in memory for that request, whether the connection was configured by Pure Tokens Switch, a third-party CC Switch, or manual host configuration.
- Do not use provider labels to identify the connection. Do not display, log, copy, persist, ask for, compare or report credentials, Base URLs, provider labels or host configuration. Resolving one matching credential in memory for the fixed request is required execution work, not a user-facing connection inspection.
- Do not add or require a Node runtime, npm, Python, MCP, proxy, sidecar, local media helper, separate upload service, or another authentication path for balance, model discovery, image generation, editing, video generation, or video editing.
- Computer Use, browser automation, desktop/mobile UI automation, and opening or clicking Pure Tokens Switch/Desktop are never API transports. Do not use them to submit a request, poll a task, retrieve media, or recover from an actual terminal/network failure.
- API-facing Skills must not invoke Imagen or another image/video Skill as a fallback. Never gate the fixed request on a separate “authenticated image/video interface” discovery step: use the active connection credential and issue the fixed API request through the current host's terminal or native HTTPS/API capability. Only an actual terminal, network, credential-resolution, attachment-byte, or API failure may produce the documented failure receipt.
- The fixed endpoints are `/v1`, `/v1/media/models`, `/api/product/desktop/account/balance`, `/v1/images/generations`, `/v1/images/edits`, `/v1/videos`, and `/v1/videos/edits`, plus declared same-task status/content paths.
- JSON and declared multipart requests are sent through the host-native API capability. Current user attachments are sent only with their declared operation; never upload, rehost, convert them to a text prompt, or silently downgrade a media-reference request.
- If host policy blocks network/API execution, attachment transport, or native byte delivery, stop before a billable request and return the documented safe failure receipt. Do not suggest a fallback transport.

## Skill behavior

- Media submissions are asynchronous. Submit once, retain only the same returned task ID, and never automatically resubmit after uncertain output, status failure, timeout or delivery failure.
- Do not make `/v1/media/models` a preflight for ordinary core generation. Read the small installed model index, then only the selected model profile; read the live catalog only for explicit discovery, an installed-profile gap, or post-rejection diagnosis.
- Treat physical dimensions as a request for guidance, not an API size value. Only declared model fields and values may be sent.
- Native bytes, not URLs, HTML, SVG, task IDs or status text, are required for successful delivery.
- Keep every user-facing failure sanitized. Never expose raw responses, request data, internal URLs, stacks, upstream identities, credentials or user media.

## Installation and updates

- Source-only shell/PowerShell installers synchronize the six Skills. They are installation tools only; they do not participate in API execution and must not install an API runtime.
- An update may remove only a verified old `puretokens-direct-api-runtime` directory and verified retired official Skill directories. Never delete unknown directories.
- Preserve the exact first `text` install prompt under the required heading in both READMEs because the client download page extracts it.

## Validation

Run `npm run check` and `npm run release:validate` before release. Node is permitted for repository development and validation only; it is never a user runtime dependency for media work.
