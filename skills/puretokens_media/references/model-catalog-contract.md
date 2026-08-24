# Media model catalog contract

The authenticated Pure Tokens Router and public gateway both expose `GET
/v1/media/models`. MCP and Direct Cloud receive the same normalized response;
the Skill treats the response for its current authorization scope as the only
source of selectable media models.

## Current minimum response

```json
{
  "object": "list",
  "data": [
    {
      "id": "exact-model-id",
      "object": "model",
      "capabilities": ["image"]
    }
  ]
}
```

`capabilities` contains one or both of `image` and `video`.

## Forward-compatible fields

To support a stable natural-language skill without embedding a model-name table in the Skill repository, the BFF/Router may add these optional fields:

```json
{
  "id": "grok-imagine-video-1.5-preview",
  "displayName": "Grok Imagine Video 1.5 Preview",
  "aliases": ["grok video", "grok 1.5 video"],
  "provider": "xAI",
  "kind": "video",
  "capabilities": ["video"]
}
```

The authenticated media catalog remains authoritative. A Skill may use only
fields actually returned by the catalog. It must send `id`, never `displayName`
or an alias, to generation tools.

The Skill may also maintain an explicit user-facing phrase registry, such as `image2` → `gpt-image-2`. For MCP and Direct Cloud this registry only produces candidate canonical IDs; it never grants availability, invents capabilities, or bypasses the live catalog. The only exception is the Skill-defined Codex/CC Switch Pure Tokens Connection Images API: a default-image or explicit `gpt-image-2` / `image2` request there calls `POST https://api.puretokensx.com/v1/images/generations` with `model: "gpt-image-2"` directly, through the active Pure Tokens connection and without consulting this catalog. It must never call an upstream URL. All other candidates are usable only when the current catalog returns the exact `id` with the requested capability.

## Required invariants

- `id` is unique in the active authorization scope: a Desktop client profile or
  a Direct Cloud API key.
- Any alias is unique within the active catalog after case and punctuation
  normalization.
- A model is listed only when an active route explicitly declares
  `image-generation` or `openai-video`.
- Removing a Desktop group, changing a route's media endpoint metadata, or
  removing a Direct Cloud API key permission removes the model from the next
  catalog response.
- A text model is never returned as media merely because its name contains `image`, `video`, a provider name, or a marketing term.
