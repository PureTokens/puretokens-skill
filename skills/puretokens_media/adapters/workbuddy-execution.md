<system_reminder>
## WorkBuddy media routing

This is the WorkBuddy delivery of the shared Pure Tokens Media Skill. Apply
these routing rules only when the user asks to generate or create an image or
video, asks for a media model, or explicitly mentions Pure Tokens media. Do
not affect normal text, coding, analysis, image understanding, or editing
requests.

1. Before WorkBuddy built-in `ImageGen` or `VideoGen`, discover the
   `puretokens-image` MCP tools with `ToolSearch` and use the matching Pure
   Tokens tool. Use the built-in tool only when the user explicitly asks for
   that built-in tool.
2. Follow the shared workflow below without creating a second model-selection
   or retry policy. Its live catalog, exact-model, single-submission, polling,
   delivery, and failure rules are authoritative.
3. Do not expose credentials, Router tokens, local authorization URLs, or
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
choose the exact catalog model -> DeferExecuteTool(generate) ->
DeferExecuteTool(result) until the native result arrives.

Do not use `show_widget`, `ImageGen`, `VideoGen`, `WebSearch`, SVG, HTML, or a
text-only response as a fallback for a Pure Tokens media request. A Pure Tokens
image/video is successful only after the corresponding `DeferExecuteTool`
result contains the MCP-returned exact model and its native result or local
delivery metadata.

A discovered tool, a tool name in an answer, a task-less response, or a
rendered widget is never proof that a model was called. If deferred invocation
is unavailable or fails, report that state and do not fabricate a visual result
or claim Pure Tokens generated it.
