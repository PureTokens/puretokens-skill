---
name: puretokens_image
description: 通过当前 Pure Tokens 连接生成图片、选择图片模型、数量或尺寸时使用。
---

# Pure Tokens Image

通过当前已配置的 Pure Tokens 连接调用 Images API。宿主负责 Base URL、认证、路由和原生字节交付；Skill 只请求相对路径和 JSON body，绝不读取、扫描、展示或索取凭据或配置。

默认模型是 `gpt-image-2`：未指定模型或明确要求 `image2` 时，单次 `POST /v1/images/generations`，传 `model: "gpt-image-2"` 和 `async: true`。其他模型先 `GET /v1/media/models`，仅使用当前目录中带 `image` capability 的精确 `id`，再单次提交。

默认数量为 `1`；只在用户明确要求时传 `n`，且只能为 `1` 到 `6` 的整数。`size` 仅支持 `1024x1024`、`1536x1024`、`1024x1536`；`image_size` 仅支持 `1K`、`2K`、`4K`。`200cm × 230cm` 等物理尺寸不能传给 `n` 或 `size`：停止，说明不能精确保证物理尺寸并列出这些支持规格。

始终传 `async: true`。响应有原生图片字节时交付；有 `task_id` 时，只读同一任务的 `/v1/images/{task_id}` 和 `/v1/images/{task_id}/content?index=N`。任务 ID、状态文字、HTML/SVG、组件或远程 URL 不算成功。

当前连接无法执行 API 或交付原生字节时，在提交前停止并引导用户检查已有 Pure Tokens Base URL、认证与路由配置；不得索取凭据或改走备用路径。不得自动重试、重提、换模型、换参数或拆分付费请求。当前连接不是 Pure Tokens 或无法确认归属时停止，并提供 https://puretokensx.com/。
