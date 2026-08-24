<system_reminder>
## WorkBuddy media routing

This is the WorkBuddy delivery of the shared Pure Tokens Media Skill. Apply
these routing rules only when the user asks to generate or create an image or
video, asks for a media model, or explicitly mentions Pure Tokens media. Do
not affect normal text, coding, analysis, image understanding, or editing
requests.

1. Preserve an explicit WorkBuddy choice. When WorkBuddy's UI or tool context
   already identifies built-in `ImageGen`, `VideoGen`, or a manually configured
   media model as the selected execution path, use that host-selected model and
   do not reroute it through `puretokens-image`. This includes a manually
   configured Pure Tokens model: do not create a duplicate MCP task for the
   same request. A model name in ordinary chat text is not by itself proof of a
   WorkBuddy UI selection; resolve it under the shared catalog-first policy.
2. For a bare image/video request or an explicit Pure Tokens media request
   without that host-selected execution context, discover the
   `puretokens-image` MCP tools with `ToolSearch` and use the matching Pure
   Tokens tool when it is callable. If MCP is unavailable, use the shared
   Direct Cloud branch only when WorkBuddy actually has its required HTTPS,
   injected-credential, and local-delivery capabilities.
3. Follow the shared workflow below without creating a second model-selection
   or retry policy. Its live catalog, exact-model, single-submission, polling,
   delivery, and failure rules are authoritative.
4. Do not expose credentials, Router tokens, local authorization URLs, or
   upstream URLs.
</system_reminder>

## WorkBuddy execution requirement

ToolSearch only discovers the deferred MCP tools; it neither invokes the MCP
nor starts a media-model request.

In WorkBuddy, invoke every discovered Pure Tokens MCP tool through
`DeferExecuteTool`, never by merely naming it in text. Use these exact deferred
tool names: `mcp__puretokens-image__puretokens_list_media_models`,
`mcp__puretokens-image__puretokens_generate_image`,
`mcp__puretokens-image__puretokens_image_result`,
`mcp__puretokens-image__puretokens_generate_video`, and
`mcp__puretokens-image__puretokens_video_result`.

The required order is: ToolSearch -> DeferExecuteTool(list_media_models) ->
choose the exact catalog model -> DeferExecuteTool(generate). If that generate
call already returns native media content, stop there. Otherwise, only a
returned task id may be followed by DeferExecuteTool(result) until the native
result arrives. `gpt-image-2` returns its native MCP image in the generate
call; after that image is present, never call `puretokens_image_result`.

Do not use `show_widget`, `ImageGen`, `VideoGen`, `WebSearch`, SVG, HTML, or a
text-only response as a fallback for a Pure Tokens media request. A Pure Tokens
image/video is successful only after the corresponding `DeferExecuteTool`
result contains the MCP-returned exact model and its native result or local
delivery metadata.

A discovered tool, a tool name in an answer, a task-less response, or a
rendered widget is never proof that a model was called. If deferred invocation
is unavailable or fails, report that state and do not fabricate a visual result
or claim Pure Tokens generated it.
