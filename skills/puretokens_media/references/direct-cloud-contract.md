# Direct Cloud media contract

This contract is for code-oriented Agents that can execute HTTPS requests and
already receive the user's Pure Tokens access token through their own secret or
environment mechanism. It is not for ordinary GUI chats without a tool
execution channel.

## Credential boundary

- `PURETOKENS_ACCESS_TOKEN` is injected by the host. A Skill must never ask the
  user to paste it into chat, print it, write it into a file, or include it in
  a command transcript.
- `PURETOKENS_API_BASE_URL` defaults to `https://api.puretokensx.com` and must
  be an HTTPS origin without a path, query, fragment, or embedded credential.
- Desktop Router credentials (`PTS_ROUTER_TOKEN`, `PTS_ROUTER_BASE_URL`) are
  a separate transport. Direct Cloud must not read or reuse them.

## Catalog

Read the authenticated catalog before every new media task:

```text
GET /v1/media/models
Authorization: Bearer $PURETOKENS_ACCESS_TOKEN
```

The public gateway and Desktop Router return the same normalized `data[]`
shape. Only an entry whose exact `id` is returned may be selected, and only
its explicit `capabilities` array (`image` and/or `video`) grants media
capability. Provider wording and model-name substrings never grant a
capability.

## Submit and poll

Use one stable request ID for one user media request and retain it in the host
task state. The gateway does not currently provide a durable task idempotency
fence for these public media endpoints, so the host must make exactly one
submission and must never automatically retry an unknown outcome. A later
user-approved retry is a new logical task with a new request ID.

```text
POST /v1/images/generations
POST /v1/videos
Authorization: Bearer $PURETOKENS_ACCESS_TOKEN
Content-Type: application/json
```

Image bodies use the selected exact `model`, `prompt`, optional `size` and
`quality`, and `async: true`. Set `n` to `1` unless the user explicitly asks
for a different number of results. Pass that explicit count only when the
selected endpoint accepts it; do not turn one request into multiple submits.

An image generation response has exactly one of these successful result paths:

1. Every requested result is present in `data[]` with `b64_json`: decode the
   bytes and atomically write each local file.
2. Every requested result is present in `data[]` with `url`: download those
   exact returned URLs, atomically write each local file, and never expose the
   URLs in logs or user-facing output.
3. The response provides a task ID: poll `GET /v1/images/{task_id}`. After the
   service reports completion, retrieve `GET /v1/images/{task_id}/content` and
   atomically write the returned media bytes.

An empty `data[]`, an entry without either supported synchronous result field,
an unknown task state, a missing task ID, a failed download, or an incomplete
write is not success. Do not submit again or change models automatically.

Video bodies use the selected exact `model`, `prompt`, optional `seconds`,
`size`, `resolution`, and `aspect_ratio`. Videos are always asynchronous:
the response must provide a task ID; poll `GET /v1/videos/{task_id}`, then
retrieve `GET /v1/videos/{task_id}/content` only after completion. A completed
status without downloadable content bytes is not a completed video delivery.

Use bounded polling and stop as soon as the content bytes have been delivered
or the task enters a terminal error state. A timeout, an unrecognized status,
or a missing result field remains pending or failed; it never authorizes a
second submission.

The caller writes completed bytes atomically to `Downloads/Pure Tokens` and
reports each saved filename and directory. It may present an in-chat preview
only when the host actually supports that media type, and may offer an
open-file/open-folder control only when the host provides a real local entry.
It must never disclose an upstream signed URL or treat a task ID, status, HTML,
SVG, or text placeholder as a completed result.
