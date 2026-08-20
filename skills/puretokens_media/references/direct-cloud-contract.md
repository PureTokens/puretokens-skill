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
`quality`, and `async: true`. A synchronous `data[].b64_json` or `data[].url`
is a completed image. Otherwise poll `GET /v1/images/{task_id}`, then retrieve
`GET /v1/images/{task_id}/content` only after completion.

Video bodies use the selected exact `model`, `prompt`, optional `seconds`,
`size`, `resolution`, and `aspect_ratio`. Videos are always asynchronous:
poll `GET /v1/videos/{task_id}`, then retrieve
`GET /v1/videos/{task_id}/content` only after completion.

The caller writes completed bytes atomically to `Downloads/Pure Tokens` and
reports the saved file path. It may present an in-chat preview only when the
host actually supports that media type. It must never disclose an upstream
signed URL or treat a task ID as a completed result.
