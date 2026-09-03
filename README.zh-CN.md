<p align="center">
  <img src="./assets/brand/puretokens-skill-hero.png" alt="Pure Tokens 官方 Skills" width="100%" />
</p>

# Pure Tokens Skills

本仓库提供六个独立 Skill：

| Skill | 用途 |
| --- | --- |
| `puretokens-balance` | 通过固定的 Pure Tokens API 读取当前余额快照。 |
| `puretokens-connection` | 检查固定 Pure Tokens API 的身份声明，不展示连接配置。 |
| `puretokens-models` | 查询固定 API 的认证模型目录，并说明已声明的模型能力、参数和媒体操作。 |
| `puretokens-image` | 通过固定的 Pure Tokens Images API 生图，并按 profile 支持图片编辑。 |
| `puretokens-video` | 通过固定的 Pure Tokens Videos API 生视频，并按 profile 支持图生视频、参考图/视频/音频视频和视频编辑。 |
| `puretokens-update` | 安装或安全升级本机官方 Pure Tokens Skills。 |

按需安装到受支持宿主已声明的全局 Skill 目录。用户不需要安装 Node、npm、Git 或任何依赖：

```bash
# macOS 或 Linux：将下方 target 替换为对应宿主目录。
installer_dir="$(mktemp -d)"
curl --fail --location --proto '=https' --tlsv1.2 \
  https://raw.githubusercontent.com/PureTokens/puretokens-skill/main/runtime/puretokens-skill-install.sh \
  --output "$installer_dir/puretokens-skill-install.sh"
sh "$installer_dir/puretokens-skill-install.sh" sync --target "$HOME/.agents/skills"
rm -rf "$installer_dir"
```

```powershell
# Windows PowerShell：将下方 target 替换为对应宿主目录。
$installerDir = Join-Path ([System.IO.Path]::GetTempPath()) ("puretokens-skill-" + [Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $installerDir | Out-Null
Invoke-WebRequest https://raw.githubusercontent.com/PureTokens/puretokens-skill/main/runtime/puretokens-skill-install.ps1 -OutFile "$installerDir\puretokens-skill-install.ps1"
& "$installerDir\puretokens-skill-install.ps1" sync -Target "$env:USERPROFILE\.agents\skills"
Remove-Item -LiteralPath $installerDir -Recurse -Force
```

Claude Code 使用 `~/.claude/skills`，Codex 使用 `~/.agents/skills`，WorkBuddy 使用 `~/.workbuddy/skills`，Gemini CLI 使用 `~/.gemini/skills`，Grok Build 使用 `~/.grok/skills`，OpenCode 使用 `~/.config/opencode/skills`，Trae 使用 `~/.trae/skills`。

## 让 Agent 协助安装

客户端下载页会提取下方固定标题下的第一个代码块。必须保持该标题和唯一的 `text` 代码块稳定；本节不得放入其他代码块。

### 直接复制给具备本机终端的 Agent

```text
请从 https://github.com/PureTokens/puretokens-skill 安装或更新官方 Pure Tokens Skills。
```

## 宿主支持

CC Switch 是连接配置工具，不是 Skill 宿主。Skill 自身始终调用固定的公开 Pure Tokens API origin。受管直连运行器只会为本次固定请求在内存中狭义匹配已配置的 Pure Tokens 模型凭据；绝不展示、复制、保存或索取该凭据。

| 宿主 | 当前专项 Skill 交付方式 | 直连 API 执行 | 用户操作 |
| --- | --- | --- | --- |
| Claude Code | 手动安装源文件 | 已验证的受管运行器 | 将所需 Skill 安装到 `~/.claude/skills`。 |
| Codex | 手动安装源文件 | 已验证的受管运行器 | 将所需 Skill 安装到 `~/.agents/skills`。 |
| WorkBuddy | 手动安装源文件 | 已验证的受管运行器 | 将所需 Skill 安装到 `~/.workbuddy/skills`。 |
| Gemini CLI | 手动安装源文件 | 已验证的受管运行器 | 将所需 Skill 安装到 `~/.gemini/skills`。 |
| Grok Build | 手动安装源文件 | 已验证的受管运行器 | 将所需 Skill 安装到 `~/.grok/skills`。 |
| OpenCode | 手动安装源文件 | 已验证的受管运行器 | 将所需 Skill 安装到 `~/.config/opencode/skills`。 |
| Trae | 手动安装源文件 | 手动凭据配置 | 将所需 Skill 安装到 `~/.trae/skills`。 |

唯一事实来源是 `references/host-support.json`。CLI 刻意不会猜测宿主目录。

## 直连 API 契约

面向 API 的 Skill 都调用固定的公开 API origin：`https://api.puretokensx.com`。请求使用完整 URL，不使用用户选择的 Base URL 或备用 endpoint。安装在 Skill 同级目录的受管 `.puretokens-runtime/puretokens-direct-api.mjs` 会在 Claude Code、Codex、WorkBuddy、Gemini CLI、Grok Build 和 OpenCode 中，只从各宿主已记录的固定配置位置狭义匹配一个 Pure Tokens 凭据。通常配置应指向 `https://api.puretokensx.com/v1`（或规定的仅 origin 形式）；WorkBuddy 也允许在相同固定 origin 下保存无 query、无 fragment 的 `/v1/...` 单模型资源 URL。它只在内存中保留该凭据，并且只用于允许的固定 Pure Tokens API 路径认证。它绝不打印、复制、保存或索取 Key；不接受任意 URL，不使用 MCP、代理或 sidecar。`puretokens-update` 只处理本机更新，不调用 API endpoint。

直连契约要求已认证的完整 URL HTTPS 请求、JSON 任务响应、同任务状态查询和原生媒体字节交付。Claude Code、Codex、WorkBuddy、Gemini CLI、Grok Build 和 OpenCode 通过受管本地运行器完成认证绑定。WorkBuddy 的 POST 请求体使用一个有界、规范的 Base64 参数，而不是标准输入，因为它的 Bash 执行链路不能可靠关闭 stdin；运行器只在内存中解码。Trae 是手动配置例外：当前没有已批准的本地凭据读取契约，Skill 会安全停止，而不会读取或猜测其用户状态。

`puretokens-connection` 只调用一次 `GET https://api.puretokensx.com/v1`。只有固定端点明确返回 `status: "ok"`、`name: "Pure Tokens API"` 与 `base_url: "/v1"` 时，才确认该固定 API 的标识。受管运行器可在内部为该请求使用凭据，但 Skill 不会展示或报告用户实际配置的 Base URL 或凭据；这只是端点公开声明检查，不是密码学防伪证明。

`puretokens-models` 只调用一次 `GET https://api.puretokensx.com/v1/media/models`。它将认证目录转换为用户可读的信息：精确模型 ID、实际返回的能力、已声明可选参数和媒体操作。用户可以问图生视频、参考媒体、时长、画幅或分辨率等明确技术要求有哪些兼容模型；Skill 只会根据实时 profile 实际声明的字段和值筛选，不会提交媒体任务、重试目录请求、回退到 README 静态目录，或编造质量、价格、速度和可用性排序。

`puretokens-image` 的普通文生图会直接调用 `POST https://api.puretokensx.com/v1/images/generations` 并传 `async: true`，不会先读取模型目录。已安装的版本化选择资料负责默认 `gpt-image-2` 的普通生图和已声明的本地附件图片编辑、唯一别名和已知图片参数。只有用户明确查询当前模型、要求安装资料未声明的参数或媒体操作，或模型/参数/capability 被 API 拒绝后需要解释时，才读取一次实时目录。目录读取失败绝不能阻止原本有效的核心文生图，也绝不会触发自动重提。

`puretokens-video` 的普通文生视频会直接调用 `POST https://api.puretokensx.com/v1/videos`，不会先读取模型目录。已安装选择资料负责默认 `grok-imagine-video-1.5-preview`、唯一别名、已知参数和媒体操作。只有明确查询当前模型、为满足安装资料未声明的参数或媒体操作，或模型/参数/capability 被 API 拒绝后需要解释时，才读取一次实时目录。目录读取失败绝不能阻止原本有效的核心文生视频，也绝不会触发自动重提。Skill 只轮询并交付同一任务的原生字节。

完整直连 API 契约见 `references/direct-api-execution-contract.json`。若任务被接受前的直连请求失败，Skill 只报告实际返回的失败，不会猜测用户的 Base URL、认证或路由原因；不会切换执行路径，也不识别或分支处理其他中转服务。

## 模型访问分组

`GET https://api.puretokensx.com/v1/media/models` 只返回当前受管 Key 已覆盖的模型。如果在允许的一次诊断后，用户指定的精确模型（例如 `minimax-h3`）仍未返回，Skill 会说明当前连接未返回该模型，不会提交或静默切换模型。若用户预期可以使用它，Skill 会引导用户在 Pure Tokens 客户端配置中勾选包含该精确模型的分组，创建或选择覆盖所选分组的受管 Key，执行“验证并应用”，然后新开当前宿主会话并明确重试。

Skill 绝不猜测分组名称，也不会声称某模型属于哪个分组；除非认证 API 明确返回模型到分组的映射，认证媒体目录本身不能提供这项结论。

## 余额

`puretokens-balance` 只执行一次 `GET https://api.puretokensx.com/api/product/desktop/account/balance`，在可用时使用已配置凭据的直连运行器。它只报告接口返回的字段。若直连请求失败，会报告实际返回的结果并引导用户到 Pure Tokens 客户端余额入口；绝不会猜余额、尝试其他路径或索取凭据。

## Skill 升级

`puretokens-update` 专门处理用户明确提出的安装、更新或同步本机官方 Skills 的请求。macOS/Linux 使用已安装的原生 Shell 安装器，Windows 使用已安装的 PowerShell 安装器。安装器通过 HTTPS 下载官方 `main` 源码，在改动目标目录前对完整安装载荷做静态校验，然后执行同步。用户无需安装 Node、npm、包管理器、Git 或依赖。当前官方 Skill 全部安装或升级成功后，它会删除已验证的旧受管 Skill 目录，以及同名的旧隐藏备份，安装缺失官方 Skill，只升级受管且同名的当前 Skill 目录；只要遇到当前或旧的非受管同名目录，或非受管运行器目录，整个同步会在改动前停止。仅当目标是 Codex 目录时，它会通过官方 Codex 插件接口移除精确匹配且已安装的旧 `puretokens-media` 插件；无法移除时会明确提示用户到 Codex Plugins 卸载或联系工作区管理员。同步成功会明确输出当前官方 Skill 版本。它绝不读取连接设置或凭据，也绝不会在媒体任务中自行运行。

## 图片尺寸和数量

一个生图请求绝不会拆成多次付费提交。数量和所有尺寸控制都是模型级的：`n`、`size`、`image_size`、`aspect_ratio`、`width`、`height` 只有已安装的精确模型资料，或为安装资料缺失的明确需求按需读取一次的目录，声明字段和值时才会发送。未声明 `n` 时，Skill 不会自行补充。

## 图片请求体验

`puretokens-image` 会在提交前区分文生图、将图片作为视觉参考、以及编辑现有图片。用户提供 URL 或附件却没有说明角色时，它只会询问“这是参考图还是待编辑图片？”，不会猜测。当前本地参考图会使用所选模型已声明的 multipart `image_edit` operation，同时在提示词中保留“参考”语义；不需要独立上传或转存。`n` 只表示同一完整简报的变体，绝不表示多个不同资产：海报、头像、横幅等多项请求会请用户先选定本次要生成的第一项，而不是创建多笔付费任务。

Skill 会将自然语言整理为简洁的图片简报，同时保留已说明的用途、主体、场景、风格、构图、逐字文本和限制；不会擅自添加品牌、文案、主体或改变操作。完成时只有宿主实际确认附件或本地路径，才会报告它；否则只说明原生图片字节已通过当前宿主交付，不编造预览或下载位置。

`200cm × 230cm` 这类物理尺寸无法精确保证，也绝不会传给 `n`、`size` 或其他 API 字段。Skill 会说明限制，并列出已安装资料或按需目录查询声明的像素或语义尺寸选项。若模型声明 `width` 与 `height` 必须成对，用户只给一个尺寸时会在提交前停止；模型声明多个尺寸表达的优先顺序时，只会提交用户已给选项中优先级最高的一种。

请求 `n` 张图时，交付会在任务成功后从同一任务严格按顺序读取零基索引 `0` 到 `n-1`，一次只读取一个索引。只有每个请求索引都拿到原生字节才算成功。部分结果会明确列出已交付和缺失索引，并且只允许继续读取该任务缺失的内容。Skill 不会预取或重复下载已交付内容；每一项原生结果交付给运行环境后才读取下一项。

## 模型参数资料与任务回执

普通图片和视频任务使用已安装的版本化选择资料，不会因实时目录预检增加一次请求。任何可选字段都必须在该资料中存在且值兼容。只有用户明确查询当前能力、要求安装资料缺失的参数或媒体操作时，Skill 才按需读取一次目录；模型/参数/capability 被拒绝后，也只允许为解释读取一次，绝不会自动重试。普通文生视频必须提供提示词；只有已安装或按需资料明确声明精确单参考例外时才可省略。

媒体输入由已安装资料控制；只有用户明确要求安装资料未声明或当前资料不兼容的媒体操作时，才按需读取一次 profile。用户明确提供的公网 HTTPS URL、file ID 或 voice ID 只能写入资料声明的精确字段和允许 transport；Skill 不会下载、探测、检查可访问性、转存或改写它。公网 URL 图片编辑还必须有精确声明的 JSON `image_edit` operation；仅有参考图字段不等于支持编辑。当前本地图片的参考或编辑会直接使用已声明的 multipart `image_edit` operation，并严格遵守路径、字段和数量限制；即使传输 operation 名为 `image_edit`，参考图仍保持参考语义。用户当前请求附带的原生媒体则只能走已声明的 `multipart_file` operation，并由已验证运行器随这一条 Images 或 Videos API 请求发送。运行器只接受当前请求明确指定的常规媒体文件，并限制附件数量与大小，不会调用独立上传 API。多个原生附件类型需要明确的组合 operation；多个公网 URL/ID 字段则不能与已声明的互斥或模式限制冲突。若本地参考图没有已声明的兼容 `image_edit` multipart operation，或运行器无法交付其字节，Skill 会在提交前停止，请用户提供该图的公网 HTTPS URL，或在新的明确请求中忽略附件并改为纯文生图；绝不会查看图片内容或把它转写为提示词。视频中用户明确说“首帧/第一帧”时，会选择模型已声明的 `image_to_video` operation 及精确字段；只附图片但未说明首帧或参考图角色时，会在计费前询问。Skill 不会伪造 URL 或 file ID，也不会把媒体请求静默改为纯文生。

媒体 Skill 在提交、继续查询、核验、完成和失败时统一返回回执：已返回的精确模型 ID、已返回的任务 ID、当前状态、请求操作、请求数量、尺寸/参数、完成时的已交付数量和下一步。失败时只有 API 明确返回公开机器码才会在 `api_error_code` 中原样显示，否则写“未返回”；同时会返回失败阶段、API 返回的 HTTP 状态（未返回时写“未返回”）和安全错误信息。HTTP 429 且存在有效 `Retry-After` 时会给出等待秒数。回执绝不暴露原始响应正文、上游标识、内部 URL、堆栈、请求数据、凭据或用户媒体。回执中的 `task_id` 会优先从所选模型 lifecycle 声明的顶层 ID 字段规范化；未声明 lifecycle 时只接受顶层 `task_id` 或 `id`，绝不从 URL 或嵌套响应数据猜测。任务元数据未返回时会明确写“未返回”。

如果提交 `POST` 已开始，但运行器最终输出为空、截断、格式异常或无法使用，提交结果就是未知。Skill 会输出一次失败回执后结束当前回复，不会继续调用工具、轮询或围绕同一提交反复推理；不会为了拿到任务 ID 再发一次 `POST` 或创建替代任务，也不会声称任务未创建、未扣费或已退款。没有任务 ID 时，后续只说“继续”或“再试”会要求用户明确确认要新建一次付费任务；只有确认后的下一轮才能创建一个新任务。

## 异步轮询

只有提交返回规范化 `task_id` 后才开始轮询；自动轮询只在本次提交所在的用户回合，或用户明确继续查询同一 `task_id` 的用户回合内运行。每次状态响应先检查 `reconciliation_required: true`：无论生命周期状态是什么，都将其视为终止性的运营状态，停止常规轮询、保留同一任务 ID，不提交替代任务、不推断退款。否则优先使用所选模型声明的 lifecycle 状态；未声明时只把 `pending`、`queued`、`running`、`in_progress` 视为处理中。状态缺失或未识别时会如实报告并停止自动轮询。同一任务最多一个状态请求在途，绝不会创建后台计时器、队列或工作器。状态查询收到 HTTP 429 时，只有有效正数 `Retry-After` 未超出本轮剩余预算才会等待并继续查询**同一任务**；否则停止本轮且绝不重提。正常处理中，生图依次等待 `3、6、12、24、30、30` 秒，最多读取同一任务状态 6 次；生视频依次等待 `5、10、20、40、60、60` 秒，最多读取 7 次。每个有界轮询窗口最多持续：生图 120 秒，生视频 300 秒。非 429 的 5xx、传输错误或超时时立即停止本轮。如仍在处理中会连同任务 ID 如实报告；用户明确继续查询时，才会为**同一任务**开启一个新的有界轮询窗口。Skill 不会把到期或读取错误当失败，也不会提交替代任务。

媒体字节不会缓存到 Skill 状态、提示词或日志中。只有任务终态成功后才读取内容，每个任务最多一个内容读取在途；运行环境交付原生字节后才会进行下一次读取。运行器只接受原生图片、视频、音频或 octet-stream 内容响应，并限制可交付的总字节；不支持或超限时会失败，绝不改用 URL 或重提任务。如果运行环境无法在不创建无界后台工作、重复读取或缓存副本的前提下交付，Skill 会报告同任务交付不可用，不会用 URL 代替或重提任务。已确认交付即结束当前回复：Skill 不会自动检查成片、搜索历史或工作区、调用其他 Skill 或创建新的媒体任务。

## 使用示例

- 连接：`这个 Pure Tokens Skill 能确认它调用的 API 吗？` Skill 只检查固定端点的 `GET https://api.puretokensx.com/v1` 声明，不会展示配置。
- 模型：`查看 Pure Tokens 当前可用的视频模型、它们已声明的时长和画幅选项，以及哪些支持图生视频。` 此查询只读，不会提交任务。
- 生图：`使用 gpt-image-2 生成一张 2K、16:9 的雪后黎明小镇插画。`
- 其他图片模型：`用 nano banana pro 生成一张简洁的产品海报。` Skill 会解析唯一的已安装别名后直接提交。
- 参考图 URL：`使用 gpt-image-2，并以此公网参考图 URL 生图：https://example.com/reference.png`。Skill 使用已安装资料声明的字段和 transport；若安装资料没有匹配能力，才读取一次目录。
- 图片编辑：`用 gpt-image-2 编辑我附上的图片：把阴天改成晴朗的日落。` Skill 只会通过 gpt-image-2 已声明的 multipart 图片编辑 operation 发送当前附件；只有该 operation 缺失时才读取一次目录。
- 本地参考图：`使用 gpt-image-2，以我附上的图片作为构图参考，生成一个安静的冬日版本。` Skill 会使用已声明的 multipart 图片编辑传输，但保持“参考图生图”语义；不会上传或转存该文件。
- 生视频：`用 grok 1.5 video 生成一段六秒钟的电影感海上日出。`
- 图生视频：`用 grok 1.5 video 把这个公网图片 URL 制作成六秒视频：https://example.com/reference.png`。Skill 使用已安装资料声明的 URL 字段和 transport；只有对应操作缺失时才读取一次目录。
- 首帧视频：`用 wan3.0-video，将我附上的图片作为首帧。` Skill 会选择已声明的 `image_to_video` operation 和 `first_frame_image` 字段。
- 参考视频：`用 seedance-2.5 根据我附上的视频制作一段六秒视频。` Skill 使用已安装资料声明的 `reference_video` operation；只有资料缺失时才读取一次目录。
- 参考音频：`用 minimax h3 根据我附上的音频生成一段视频。` Skill 使用已安装资料声明的 `reference_audio` operation；只有资料缺失时才读取一次目录。
- 视频编辑：`编辑我附上的视频：将白天改为夜景。` 只有已安装资料或按需读取的 profile 声明 `video_edit` 时，才会向 `https://api.puretokensx.com/v1/videos/edits` 提交。
- 继续已有任务：`继续查询任务 <task_id>。` Skill 只读取该任务，绝不会自动提交替代任务。
- 升级：`升级我的 Pure Tokens Skills。` 更新 Skill 会校验官方 `main`，并安全同步本机 Skill 目录。

## 模型发现

用户想知道 Pure Tokens 实际可用哪些模型、哪些模型支持某项媒体操作，或哪些模型接受某个已声明参数时，应使用 `puretokens-models`。它读取认证后的 `GET https://api.puretokensx.com/v1/media/models`，只展示精确模型 ID、能力、可选参数资料和 `input_schema.operations`，不补全缺失字段。兼容模型清单只是技术匹配：只根据已声明 capability、字段/值和 operation 筛选，不做主观质量或价格推荐。

README 仅用于发现能力。模型能力只来自基础模型目录明确声明的图片/视频能力，绝不通过名称推断。每个安装后的图片/视频 Skill 都携带从同一目录生成的、按能力拆分的 `references/model-selection.json`；别名只有唯一对应一个精确模型 ID 时才可使用。

<!-- media-model-catalog:start -->
## 媒体模型清单

已与基础模型目录同步：2026-09-03T04:00:37.535Z。

这份清单由 Pure Tokens 基础模型目录中的明确图片/视频能力生成。实际执行时，精确模型和所需能力仍必须出现在认证后的 GET https://api.puretokensx.com/v1/media/models 响应中。

README 只从基础目录中带有明确图片/视频能力的模型生成，不通过模型名称推断。已安装快照用于普通生成的模型选择和已知参数；实时目录只在明确查询、安装资料缺口或提交被拒后的诊断时按需读取。发布前从受控基础目录刷新，并运行 `npm run release:validate`；当快照超过七天时发布校验会失败。

### 图片模型

| 模型 ID | 提供方 | 也可以这样说 | 适合 | 示例 |
| --- | --- | --- | --- | --- |
| `gpt-image-2` | OpenAI | `image2` | 图片生成 | `用 gpt-image-2 生成一张图片。` |
| `grok-imagine-image` | xAI | `grok image` | 图片生成 | `用 grok-imagine-image 生成一张图片。` |
| `grok-imagine-image-2.0` | xAI | `grok image 2.0` | 图片生成 | `用 grok-imagine-image-2.0 生成一张图片。` |
| `grok-imagine-image-quality` | xAI | 仅精确 ID | 图片生成 | `用 grok-imagine-image-quality 生成一张图片。` |
| `nano-banana-2` | Google | `nano banana 2` | 图片生成 | `用 nano-banana-2 生成一张图片。` |
| `nano-banana-2-lite` | Google | 仅精确 ID | 图片生成 | `用 nano-banana-2-lite 生成一张图片。` |
| `nano-banana-pro` | Google | `nano banana pro` | 图片生成 | `用 nano-banana-pro 生成一张图片。` |
| `seedream-5.0-pro` | ByteDance | 仅精确 ID | 图片生成 | `用 seedream-5.0-pro 生成一张图片。` |

### 视频模型

| 模型 ID | 提供方 | 也可以这样说 | 适合 | 示例 |
| --- | --- | --- | --- | --- |
| `grok-imagine-video` | xAI | `grok video` | 视频生成 | `用 grok-imagine-video 生成一条视频。` |
| `grok-imagine-video-1.5` | xAI | 仅精确 ID | 视频生成 | `用 grok-imagine-video-1.5 生成一条短视频。` |
| `grok-imagine-video-1.5-preview` | xAI | `grok 1.5 video` | 视频生成 | `用 grok-imagine-video-1.5-preview 生成一条视频。` |
| `minimax-h3` | MiniMax | `minimax h3` | 视频生成 | `用 minimax-h3 生成一条视频。` |
| `seedance-2.0` | ByteDance | 仅精确 ID | 视频生成 | `用 seedance-2.0 生成一条视频。` |
| `seedance-2.0-fast` | ByteDance | 仅精确 ID | 视频生成 | `用 seedance-2.0-fast 生成一条视频。` |
| `seedance-2.0-mini` | ByteDance | 仅精确 ID | 视频生成 | `用 seedance-2.0-mini 生成一条视频。` |
| `seedance-2.5` | ByteDance | 仅精确 ID | 视频生成 | `用 seedance-2.5 生成一条视频。` |
| `wan3.0-video` | Qwen | `wan3 video`, `wan 3 video` | 视频生成 | `用 wan3.0-video 生成一条短视频。` |
| `wan3.0-video-prime` | Qwen | `wan3 video prime`, `wan 3 video prime` | 视频生成 | `用 wan3.0-video-prime 生成一条短视频。` |

<!-- media-model-catalog:end -->

## 更新

拉取最新仓库后，分别升级已经安装的 Skill：

```bash
node bin/puretokens-skill.js upgrade puretokens-image --target ~/.agents/skills
```

发布前运行：

```bash
npm run docs:sync-media-models-from-service
npm run release:validate
```
