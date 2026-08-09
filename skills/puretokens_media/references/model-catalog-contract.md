# Media model catalog contract

The Pure Tokens Router exposes `GET /v1/media/models` to the local MCP sidecar. The Skill treats this response as the only source of currently selectable media models.

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
  "id": "grok-imagine-video-1.5",
  "displayName": "Grok Imagine Video 1.5",
  "aliases": ["grok video", "grok 1.5 video"],
  "provider": "xAI",
  "kind": "video",
  "capabilities": ["video"]
}
```

The Router remains authoritative. A Skill may use only fields actually returned by the catalog. It must send `id`, never `displayName` or an alias, to generation tools.

## Required invariants

- `id` is unique in the active client profile.
- Any alias is unique within the active catalog after case and punctuation normalization.
- A model is listed only when an active route explicitly declares `openai_images` or `openai_video`.
- Removing a group or changing its media protocol removes the model from the next catalog response.
- A text model is never returned as media merely because its name contains `image`, `video`, a provider name, or a marketing term.
