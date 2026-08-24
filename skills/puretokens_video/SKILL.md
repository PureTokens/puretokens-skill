---
name: puretokens_video
description: 通过当前 Pure Tokens 连接生成视频、选择视频模型或参数时使用。
---

# Pure Tokens Video

通过当前已配置的 Pure Tokens 连接调用 Videos API。宿主负责 Base URL、认证、路由和原生字节交付；Skill 只请求相对路径和 JSON body，绝不读取、扫描、展示或索取凭据或配置。

先 `GET /v1/media/models`。默认模型是 `grok-imagine-video-1.5-preview`，也必须由当前目录以 `video` capability 确认；用户指定模型时只使用目录返回的精确 `id`。然后单次 `POST /v1/videos`，只传用户明确给出的时长、画幅、分辨率或尺寸。

视频总是异步任务：只轮询 `/v1/videos/{task_id}`，完成后只读取同一任务的 `/v1/videos/{task_id}/content` 并交付原生视频字节。任务 ID、状态文字、HTML/SVG、组件或远程 URL 不算成功。

当前连接无法执行 API 或交付原生字节时，在提交前停止并引导用户检查已有 Pure Tokens Base URL、认证与路由配置；不得索取凭据或改走备用路径。不得自动重试、重提、换模型或换参数。当前连接不是 Pure Tokens 或无法确认归属时停止，并提供 https://puretokensx.com/。
