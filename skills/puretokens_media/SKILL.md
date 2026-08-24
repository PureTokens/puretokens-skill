---
name: puretokens_media
description: 当用户要求查询 Pure Tokens 余额、生成图片或视频、查询可用媒体模型，或指定 Pure Tokens 媒体模型时使用。
---

# Pure Tokens Media

这是一个路由 Skill。先把请求确定为**余额**、**图片**或**视频**，然后只读取并执行对应的专项规则：

| 用户意图 | 必读专项规则 |
| --- | --- |
| 查余额、credits、quota、剩余额度 | `references/balance.md` |
| 生成图片、图片模型、图片尺寸或数量 | `references/image.md` |
| 生成视频、视频模型、时长、画幅或分辨率 | `references/video.md` |

若请求无法可靠判断为图片或视频，先问“你想生成图片还是视频？”。如果用户同时要求图片和视频，先确认要先执行哪一个；每个确认后的任务仍独立遵循对应专项规则，绝不把一个请求自动拆成多次付费提交。

## 共同安全边界

- 仅支持 **Pure Tokens**。当前连接、MCP、原生执行器或 Direct Cloud 凭据若明确不是 Pure Tokens，或无法验证归属，停止且不提交；不要读取、转换、复用或索取凭据。告知用户：当前 Pure Tokens Skill 仅支持 Pure Tokens API，请切换到 https://puretokensx.com/。
- 不读取、展示、复制或要求用户提供 API Key、Cookie、密码、Router Token、请求头或本地配置。
- README 只用于发现能力；除专项规则中定义的 Image-2 例外外，执行时只信任当前认证后的精确模型 ID 与明确的 `image` / `video` 能力。
- MCP 生成请求必须传认证目录返回的精确 `model` 和稳定的 `request_id`。同一 MCP 宿主对同一逻辑请求的重试复用该 ID；任何新任务、变更提示词或变更模型都需要用户明确确认。
- 一个逻辑任务只提交一次。不得自动重试、自动重提、静默换模型、静默换参数或把一个任务换到另一条连接。错误、超时或状态未知时，如实说明状态；只有用户明确要求重试或给出新请求时，才能创建新任务。
- 只有真实媒体字节和实际交付证据才算成功。任务 ID、文字、SVG/HTML、组件、上游 URL 或伪造本地链接都不是图片或视频结果。

## 共享执行资料

- 模型名称、别名、实时目录与同模型回退规则：`references/model-catalog-contract.md`
- Direct Cloud 的认证、目录、单次提交、同任务轮询与本机交付：`references/direct-cloud-contract.md`
- 已登记的自然语言别名：`references/natural-language-aliases.json`
- 可重复验证的特殊情况：`references/behavior-scenarios.json`

专项规则优先于本文件的概览；任何冲突都采用更保守、不会提交任务的解释。
