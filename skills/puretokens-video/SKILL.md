---
name: puretokens-video
description: 当前宿主使用 Pure Tokens 连接时，任何生成或编辑视频的请求都优先使用本 Skill。
---

# Pure Tokens Video

宿主绑定与本地失败停止：首次调用前根据当前应用或用户明确指定确定一个 host；本次任务所有 Skill 和执行器命令沿用该 host，不从安装目录推断宿主。任何命令返回 `active_connection_*`、`workbuddy_*` 或 `host_credential_adapter_unavailable` 本地失败时，立即停止本次 API 流程并返回脱敏说明；不得自动调用 puretokens-connection、init、doctor、models 或其他 Skill 重复探测，不得枚举、更换 `--host` 或借用其他客户端凭据。保留用户原任务和已有 task ID；本次调用未发请求不代表此前没有提交任务，不自动重提。只有用户明确要求重查或出现可验证的新诊断依据时，才在原 host 做一次相应检查；“继续生成”本身不是切换宿主或重复探测的授权。改变执行宿主必须有用户明确指定，不能把失败恢复当作指定。

## 执行边界

必须调用安装的原生执行器；它是唯一 API 传输，固定请求 `https://api.puretokensx.com`。不得自行发 HTTP，不回退到 imagegen／Imagen／通用视频 Skill、MCP、代理、Computer Use 或浏览器／桌面自动化。仅执行器在内存中使用当前宿主匹配连接；Skill 不读配置、不传或展示凭据，不新增用户运行环境。

从本 SKILL.md 绝对目录解析 `../.puretokens-executor/puretokens-api`，Windows 使用 `puretokens-api.exe`；不依赖 PATH 或工作目录。当前宿主 ID 为 claude-code、codex、workbuddy、gemini-cli、grok-build、opencode、trae、claude-desktop、dsh-desktop、zcode、kimi-code、qoder、pi。

宿主已明确无法提供本次附件字节或交付媒体时，提交前停止，不创建付费任务；远程／沙箱限制不允许复制凭据或换传输。安装、init、API 或夹具成功不证明实机附件能力，验收待完成也不等于不支持。未反映在有效连接文件中的会话覆盖必须停止；ZCode／Qoder 配置存在不证明当前聊天选择。

## 简单请求短路径

- 模型、素材和用途明确的单次生成或编辑，直接「读取所选 profile → 写请求文件 → submit」。除非用户要求计划或任务属于复杂多阶段工作，不创建或更新任务清单，不为宣布下一步单独增加一轮。
- 必要且互不依赖的读取放在同一轮；已读且仍在有效上下文中的说明不重读。仅缺命令用法时才读 executor-usage，与所选 profile 合并读取；不把完整指南当作固定前置。
- 直接使用宿主提供且执行器可访问的附件绝对路径，不为整理目录复制、重命名或改写原图、视频或音频。宿主必须先物化附件时才保存一份字节不变的文件；这不允许绕过权限或变换附件传输。
- 收到回执后直接按 `next_step` 执行，不在 wait、content、deliver 之间插入任务清单或额外规划。仍须先报告提交回执、逐步检查结果、遵守停止条件；不能把 submit、wait、content 合成一个命令。

## 选择与提交

1. 默认 `grok-imagine-video-1.5-preview` 或用户精确 ID：只读 `references/profiles/<model>.json`；选模型或唯一别名解析才读 `references/model-index.json`。不遍历 profile，不先查余额、init、doctor、preflight 或实时目录。未知精确 ID 的纯文本请求可只传 model/prompt；字段／操作缺口由执行器按需读一次目录。
2. 保留用户意图、主体与运动要求，只发 profile 声明的字段、值、operation。物理尺寸不是 API 尺寸。本地首帧／让图动起来用 `image_to_video`；公网 URL 首帧按 profile 的 JSON 引用字段，不套用仅 multipart 的 operation。角色风格参考用 `reference_image_video`，视频／音频参考或编辑用 `reference_video`／`reference_audio`／`video_edit`；首尾帧按 profile 的 frame operation。只混合已声明的附件组合，用途不明确才澄清。每任务一个视频，声音／静音只映射声明的 `generate_audio`。
3. 本次本地附件只随声明的 multipart 发送；用户公网 HTTPS URL 只进声明的 JSON 字段。不下载、探测、转存参考媒体或改成提示词；没有声明的传输方式就停止。
4. 用宿主文件工具创建 UTF-8 请求：kind=`video`、operation=`generate` 或 `edit`、model、prompt、parameters；声明的附件操作使用 media_operation 及 attachments（field、绝对 path）。执行 `<执行器> submit --host <当前宿主> --request <绝对请求文件>`，之后清理该临时文件；提示词和凭据不进命令行。命令／请求示例按需读 [executor-usage.md](references/executor-usage.md)。

## 同任务完成交付

视频默认在首次 submit 加 `--record <工作区或用户指定的唯一绝对任务文件>`。记录不含 prompt、凭据、参考 URL 或媒体字节，不手改或为更换模式重提。无记录时，每次续接保留 task_id、original_operation、model、确认数量、安全参数、reconciliation_required、retry_not_before、wait_windows_completed。

- **提交一次**，先告知返回的 task_id 和状态，再按 `next_step` 继续。回执缺失／无法解析即提交未知；只清理本次临时请求文件后停止，不追加 API 操作、不重提、不猜扣费。
- **wait**：有记录用 `resume --host <host> --record <文件>`，否则用 `wait --host <host> --request <同任务文件>`。每窗口最多 7 次读取／300 秒，含网络与等待；遵守 retry_not_before 和 Retry-After。
- **窗口结束**：`ok=true`、`wait_outcome=window_ended` 仍为生成中。仅 `next_step=wait` 且原授权和前台会话仍有效可自动再续一个窗口；累计两个窗口、retry_deferred 或 await_user 时暂停。取消、失败回执、网络超时、未知状态或对账均停止自动续接，不重置计数或后台循环。窗口内 429 只由执行器按剩余预算处理。
- **content**：只下载已完成的原任务；用 `content --host <host> --record <文件> --index 0 --output-dir <现有绝对目录>`，无记录用含 completed 状态和确认数量的同任务请求。视频只有索引 0。
- **deliver**：将 downloaded_paths 文件作为宿主原生附件交给用户，下载不等于交付，URL、HTML、SVG、任务号均不能替代。实际交付后才执行 `delivered --record <文件> --index 0`；done 后结束，不自动审美检查或再生成。

交付失败只重交已有文件。已完成且无对账标记的记录用 resume 本地校验，不读凭据或请求 API；跨命令复用须匹配 SHA-256、字节数、媒体类型。无记录仅重交本会话刚下载且未改变的文件；证明不足时保留文件，另选目录取同任务索引。换电脑、旧记录、对账或跨会话恢复时读 [续接与恢复](references/executor-usage.md#continuation-record)；无 ID 的未知提交不能靠记录恢复。

## 按需说明

- 失败／费用／支持摘要：读 `references/failure-guide.md`，只说明实际阶段、已有任务及脱敏 next_action，不展示整份 JSON、原始错误、内部 URL 或配置。本地未识别不等于未配置，不据此要求重装、重配、换模型或凭据。
- 用户明确检查参数才用 preflight；不创建任务、不报价、不证明权限。宿主限制读 `references/desktop-hosts.md`。
- 复杂异常按 id 查 `references/behavior-scenarios.json`；正式字段见 `references/execution-contract.json`，展示规则见 `references/task-receipt.json`，均非普通生成前置。
