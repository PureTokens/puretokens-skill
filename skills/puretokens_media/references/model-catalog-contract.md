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

The Skill may also maintain an explicit user-facing phrase registry, such as `image2` → `gpt-image-2`. For MCP and Direct Cloud this registry only produces candidate canonical IDs; it never grants availability, invents capabilities, or bypasses the live catalog. The default-image and explicit `gpt-image-2` / `image2` exception is the Skill-defined Codex/CC Switch Pure Tokens Connection Images API: it calls `POST https://api.puretokensx.com/v1/images/generations` with `model: "gpt-image-2"` directly, through the active Pure Tokens connection and without consulting this catalog. For an explicitly selected non-`gpt-image-2` image model, that same Connection Images API may be used only when the host actually exposes the active connection as a callable authenticated HTTPS Images API executor that returns or delivers native image bytes. It must first read this catalog through that connection, then send the exact returned `image` `id` to `/v1/images/generations`. If that execution capability is absent, the Skill must tell the user before any fallback submission that the connection cannot directly execute the exact model and identify the verified same-model native, MCP, or Direct Cloud fallback; it stops if no same-model fallback exists. It must never call an upstream URL or silently change the model. For video, a Codex/CC Switch connection may use the Skill-defined Pure Tokens Connection Videos API only when the host actually exposes that connection as a callable authenticated HTTPS media executor and can deliver the returned bytes. That path must first read this catalog through the active connection, then send the exact returned `id` to `/v1/videos`; a configured chat API key or catalog visibility alone does not prove either execution capability. All other candidates are usable only when the current catalog returns the exact `id` with the requested capability.

## Required invariants

- `id` is unique in the active authorization scope: a Desktop client profile,
  a Direct Cloud API key, or an active connection with verified media execution.
- Any alias is unique within the active catalog after case and punctuation
  normalization.
- A model is listed only when an active route explicitly declares
  `image-generation` or `openai-video`.
- Removing a Desktop group, changing a route's media endpoint metadata, or
  removing a Direct Cloud API key permission removes the model from the next
  catalog response. The same applies to the current API key scope of a
  verified active connection.
- A text model is never returned as media merely because its name contains `image`, `video`, a provider name, or a marketing term.
