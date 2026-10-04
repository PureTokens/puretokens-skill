---
name: puretokens-audio
description: 当前宿主使用 Pure Tokens 连接，用户要求文字配音、朗读成音频、录音转文字、生成音效和声音场景、生成纯音乐或歌曲时使用。只写配音稿／歌词、解释用法、克隆音色和实时麦克风对话不触发本 Skill 的付费请求。
---

# Pure Tokens Audio

宿主绑定与本地失败停止：首次调用前根据当前应用或用户明确指定确定一个 host；本次任务所有 Skill 和执行器命令沿用该 host，不从安装目录推断宿主。任何命令返回 `active_connection_*`、`workbuddy_*` 或 `host_credential_adapter_unavailable` 本地失败时，立即停止本次 API 流程并返回脱敏说明；不得自动调用 puretokens-connection、init、doctor、models 或其他 Skill 重复探测，不得枚举、更换 `--host` 或借用其他客户端凭据。保留用户原任务和已有 task ID；本次调用未发请求不代表此前没有提交任务，不自动重提。只有用户明确要求重查或出现可验证的新诊断依据时，才在原 host 做一次相应检查；“继续生成”本身不是切换宿主或重复探测的授权。改变执行宿主必须有用户明确指定，不能把失败恢复当作指定。

## 需求与操作

<!-- generated:operations -->
| 用户需求 | 命令与请求 | 返回方式 | 恢复与交付 |
| --- | --- | --- | --- |
| 文字配音 | `audio`；`operation=speech` | 同步；一次请求最多90秒 | 仅已有文件 `audio-verify` 后重新附加；无任务查询 |
| 录音转写 | `audio`；`operation=transcribe` | 同步；一次请求最多90秒 | 无任务查询或结果找回；不自动重发 |
| 音效／声音场景 | `audio`；`operation=generate` | 同步；一次请求最多90秒 | 仅已有文件 `audio-verify` 后重新附加；无任务查询 |
| 纯音乐／歌曲 | `submit`；`kind=music` | 异步；先回执，再独立等待／下载 | 原任务 `resume`；仅实际附件交付后 `delivered` |

文件路径／下载回执不等于交付；附件须实际附加，文本须展示校验后的结果。未知提交不证明未处理或未扣费，均不自动重发。
<!-- /generated:operations -->

克隆声音、音频翻译、实时对话尚未接入，不用其他操作替代。
配音最多1000个 Unicode 字符；转写为8 MiB内的一份 MP3/WAV/OGG；
声音描述最多500字符。完整限制只读所选 profile。

固定入口分别为 `POST https://api.puretokensx.com/v1/audio/speech`、
`POST https://api.puretokensx.com/v1/audio/transcriptions`、
`POST https://api.puretokensx.com/v1/audio/generate`，仅由原生执行器请求。

仅问用法、静态支持范围或音色时读本地说明，不运行 init、doctor、models 或付费请求。只写／润色配音稿也不生成音频。要求当前账户可见模型才交给 puretokens-models 的 audio 查询；普通生成不查目录或余额。

## 音乐路径

音乐先读 [音乐指南](references/music-usage.md) 和对应模型 profile，使用
`submit/status/wait/resume/content/delivered` 与明确音乐任务记录；提交一次后
立即回执，再按原编号续接。音乐不使用下面的同步 audio/audio-verify 命令。
用户只要求写／润色歌词时不生成；纯音乐与人声有歧义才澄清。
当前能力需要配套网关 API Key 音乐接口已部署；404等失败停止，不改用网页接口。

## 配音／转写／声音场景的最短路径

1. 确定当前 host 和明确操作。默认配音 `stepaudio-2.5-tts`、转写 `stepaudio-2.5-asr`、声音生成 `stepaudio-3-gen-preview`，默认或精确模型只读对应 `references/profiles/<精确模型>.json`；需要选模型才读 [小型模型索引](references/model-index.json)。指定其他模型不猜映射或自动替换。
2. 配音保留用户原文、标点和换行；音色未指定可用已审核的 `cixingnansheng`，格式默认 MP3。用户要求具体风格／音色时仅映射已声明值，有歧义再澄清。语速范围 0.5–2；`instruction` 仅用于对应 profile 明确支持的模型。不要为达到情绪效果擅自修改朗读原文或插入控制标签。转写需要宿主提供本次明确附件的本地字节路径，无法获得就停止，不上传、转换、重托管或用摘要代替。
3. 仅缺命令或请求格式时读 [请求示例](references/executor-usage.md)，与必要 profile 合并读取；已读且仍有效的资料不重读。创建权限受限、绝对路径的 UTF-8 临时 JSON 文件。配音／声音生成提供现有、可写的用户输出目录；转写不提供输出目录。正文、附件路径不放在命令行。执行后清理本次请求文件；音频结果是用户文件，不作为隐式缓存删除。
4. 从本 SKILL.md 绝对目录解析 `../.puretokens-executor/puretokens-api`（Windows 为 `puretokens-api.exe`）。运行 `<执行器> audio --host <当前宿主> --request <绝对请求文件>`。执行器验证后只提交一次，同步等待最多 90 秒。无 task_id、wait、resume、content 或自动重试。
5. `next_step=deliver`：用宿主原生附件能力交付 `artifact.path` 的实际音频文件，只有附件交付成功才说已交付。路径、链接、SHA-256 和文字回执不等于音频。`next_step=done`：展示 `result.text`；这是模型转写，可能有识别或规范化差异，不保证逐字准确，也不擅自概括、翻译。其他回执按停止条件处理。

必须使用安装的原生执行器和当前宿主连接；不读／展示宿主配置，不另索 Key，不新增 SDK、Node、Python、MCP、代理、后台服务或浏览器传输，也不调用其他音频／媒体 Skill 回退。当前宿主 ID 为 claude-code、codex、workbuddy、gemini-cli、grok-build、opencode、trae、claude-desktop、dsh-desktop、deepseek-harness、minimax-code、zcode、kimi-code、qoder、pi、hermes、evox、vscode、octop。

## 配音／转写／声音场景的交付与停止条件

- 生成音频在收到响应时保存到明确输出目录；回执含路径、格式、字节数、SHA-256。它只证明下载完成，不能证明宿主已经附加或用户已经播放。
- 已下载但交付失败：保留原 `artifact`，写成 `{"artifact":<原对象>}` 的临时文件，运行 `<执行器> audio-verify --host <原宿主> --request <绝对文件>`。该命令只校验本地文件，不读取凭据或联网；成功后重新附加同一文件。文件丢失／变化则停止，不重新生成。用户要求保存进度时可将回执作为明确工作区文件保存，不记录输入正文或凭据。
- `not_submitted` 才表示本次未提交。`unknown` 包含网络中断、超时或无法校验的响应，不能推断未处理／未扣费。`accepted` 但文件保存失败也不能自动重发。没有同步结果查询恢复、服务端幂等或退款保证；用户基于“超时没扣费”要求重做时，先纠正误解，再依其知情后的决定执行。
- 超长文本、过大文件或不支持格式：说明限制和所需选择，不静默截断、改写、转换或拆成更多收费请求。有限批量先说明次数，逐项串行并交付；遇失败停止，用户明确继续仅处理未尝试项，未知项不自动重做。
- 本地 profile 不是当前权限或价格承诺。未列出的模型、音色、实时／克隆接口不试探；`stepaudio-3-tts` 尚无本 Skill 审核音色，不自动继承其他模型音色。网关把部分 ASR 的 SSE 聚合成 JSON，不要求宿主自行解析供应商流。
- 失败只展示受控类别、阶段及下一步；不展示原始响应、请求体、内部地址或上传录音。不把转写或音频结果写进支持摘要。诊断细节按需读 `references/execution-contract.json` 与 `references/behavior-scenarios.json`。

## 按需说明

用户需要协助表达配音风格、声音场景或音乐要求时读 [场景指南](references/prompt-guide.md)；
明确的正文／歌词不先套模板。视频配独立音频等多步骤需求才读
[组合流程](references/workflows.md)。解释回执时读 [回执说明](references/receipt-guide.md)。
简单请求不先建计划、不增加 init、余额、目录或更新检查。
