# Image rules

当前只支持文生图，不支持参考图、上传图片、局部重绘、去物体或其他图片编辑。遇到编辑请求，说明可用文字描述生成新的图片，不得访问用户媒体或假装编辑成功。

## 数量与尺寸

- 默认只请求 `1` 张。只有用户明确说“张、幅、个、results”等数量时才传 `n`；`n` 只能是 `1` 至 `6` 的整数。当前模型或通道不能接受该数量时停止，绝不能拆成多次付费提交。
- `size` 只支持 `1024x1024`、`1536x1024`、`1024x1536`；`image_size` 只支持 `1K`、`2K`、`4K`。
- `200cm × 230cm`、米、毫米或英寸是物理尺寸，既不是数量，也不能传给 `size`。停止并说明 API 不能精确保证物理成品尺寸，列出上述像素画布和 `image_size` 选项；不得猜 DPI、换算、裁切或选择最接近值。
- 未支持的像素画布或 `image_size` 同样停止并列出可选值；不得静默近似或仅改写提示词后声称已满足。

## 图片模型与通道

先保留用户在宿主 UI 或工具上下文主动选择的、能证明精确 `image` 模型和真实图片交付的 Pure Tokens 原生执行器；普通聊天模型配置、模型文字或组件不是执行器。

在 Codex 或 CC Switch 中，未指定模型或明确要求 `gpt-image-2` / `image2` 时，使用当前 Pure Tokens 连接认证单次调用 `POST https://api.puretokensx.com/v1/images/generations`，固定 `model: "gpt-image-2"`。不读目录、不调用 MCP、Direct Cloud 或上游地址，也不轮询或失败回退。

明确选择其他图片模型时，只有宿主已将当前 Pure Tokens 连接提供为可调用、已认证的 HTTPS Images API 执行器，并能返回或交付真实原生图片字节，才读取同一连接的 `GET https://api.puretokensx.com/v1/media/models`，确认精确 `image` 模型后单次提交同一 Images API。仅保存 API Key 或只能读目录不是执行能力。

其他宿主依次使用：已选原生执行器；可调用的 `puretokens-image` MCP（先 `puretokens_list_media_models`，只选认证目录的精确 `image` ID，再以精确 `model` 和稳定 `request_id` 一次 `puretokens_generate_image`）；或已由宿主 Secret/环境机制注入 Pure Tokens 凭据、具备 HTTPS 和真实本机交付能力的 Direct Cloud。

若 CC Switch/Codex 当前连接不能直连用户指定模型，任何备用提交前必须说明不能直连的精确模型、已验证的同模型备用通道，并明确不会换模型；没有同模型通道则停止。Connection Images API 失败后同样停止，不能改走 MCP 或 Direct Cloud。

`gpt-image-2` 的 MCP 生成调用返回原生图片后即完成，绝不再调用 `puretokens_image_result`。其他 MCP 图片模型只对同一返回 `task_id` 调用结果工具。Direct Cloud 图片始终传 `async: true`，但只有获得真实图片字节并实际交付时才成功。下载、保存或结果读取失败时不能声称成功；用户只能要求继续读取同一任务，或明确发起新的任务。

## 面向用户的特殊情况

- 模型名称、别名或供应商匹配到多个候选时，列出当前认证目录中的精确图片模型 ID 和已返回的显示信息，请用户选择；目录为空、模型不存在或没有 `image` 能力时，说明当前连接没有可用的图片模型。不得猜测或改用相近模型。
- MCP 不可用时，只能在宿主已经具备 HTTPS、已注入 Pure Tokens 凭据和真实本机交付能力的情况下使用 Direct Cloud；否则明确说明缺少可调用的图片执行通道。只有确认是 Desktop 受管 MCP 时，才引导用户选择包含该模型的分组、点击“验证并应用”、重启客户端并新建会话；自管连接只应检查自己的凭据和模型范围。
- 生成、结果读取或交付报错（包括 `safeToResubmit=false`）时，如实说明当前任务失败或状态未知；不得自动换模型或重新提交。轮询超时或仍在处理中时，引导用户稍后继续查询同一 `task_id`，而不是新建任务。
