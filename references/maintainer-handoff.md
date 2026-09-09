# Pure Tokens Skills 开发交接

更新日期：2026-09-09。本文是此项目唯一的开发交接状态；后续按关键变化覆盖更新，不追加逐轮对话。长期产品约束以根目录 `AGENTS.md` 和实际执行契约为准，本文不改变它们。

## 当前状态校正（2026-09-09）

- 产品版本为 0.18.0。产品代码、安装脚本及归档已推送 main：3b60371 为累计更新，5d7b22b 为 WorkBuddy 诊断与清理收尾，d1aefbb 为 Windows 原生启动及安装流程收敛。
- 用户明确要求将当前全部修改提交，包含本文件和两份历史审查材料；不再将它们排除在版本管理之外。
- 下方带日期的开发过程保留为历史证据。其中“未提交／未推送”、旧 HEAD、候选状态及当时验收数量不代表当前状态；以本节和实际 Git 状态为准。
- 最近一次产品提交前 npm run check（Go、go vet、59 项 Node）、npm run release:validate 和 git diff --check 通过。Windows 专项及远程 WorkBuddy 实机效果仍未确认。
- 待办：远程 WorkBuddy 连接无匹配的真实原因；新版官方入口的安装耗时、输出和额外窗口验收。不得用本地夹具替代实机结果。

## 目标与来源

用户要求由当前会话接手 Pure Tokens Skill 开发与升级，并在删除旧会话后仍能继续维护。旧会话名称为「skill 开发」，ID 为 `01a0201f-42d8-71e3-b313-363ed7ab1b00`，工作历史覆盖 2026-08-21 至 2026-09-06（北京时间）。已梳理其用户请求时间线、关键纠正、近期交付和当前源码；本记录只保存有效决定、证据位置及未完成事项，不保存完整对话。

旧会话最初位于相邻 Switch 仓库，真正的 Skill 工作目录为 `/Users/renlimin/code/puretokens-skill`。旧会话引用更早会话 `01a01f0e-85a6-7d01-98f2-a6fc216d10a4`，本次没有独立重读该更早会话；当前约束和待办已由后续历史及本仓库核对。

## 当前基线与授权范围

- 当前分支 `main`；本地 HEAD `1cc3082`，提交标题 `fix: correct balance and harden skill workflows`，对应 0.17.0。
- 旧会话 2026-09-05 最后一次推送记录确认上述提交已到远端 main。这是历史证据，本次未重新查询远端发布或部署状态。
- 工作区现在是 **0.18.0 本地候选**（包含此前未发布的 0.17.1 修复及两个新增客户端），含大量已修改文件及尚未跟踪的新源码、测试、验收文档。它们属于旧会话交付，必须保留；不能当作临时文件清除。版本号、二进制及迁移包已更新，但这些变化尚未提交。
- 用户 2026-09-05 明确要求“你只能改 skill，需要改别的项目要告知我”。可读取相关项目来核实接口；不能把 Skill 修复擅自扩大成 Web／Switch 修改和部署。
- 用户多次说明截图来自另一台远程 Mac 或 Windows 测试机，不能据此改本机真实宿主安装或连接。开发默认使用隔离夹具、本地 HTTP 服务；真实宿主安装、认证和付费媒体验收按指定环境及授权执行。
- 此次交接不包含提交、推送、发布、真实安装更新或删除旧会话操作。旧推送指令已在对应版本完成，不视为对当前候选的发布指令。

## 有效产品决定

1. 保持六个独立入口：`puretokens-image`、`puretokens-video`、`puretokens-balance`、`puretokens-models`、`puretokens-connection`、`puretokens-update`。旧聚合 media/router 不是新功能入口。
2. 用户只配置好 Pure Tokens 连接并安装 Skill，就应能使用已支持的媒体能力；可以通过第三方 CC Switch 或手工配置，不要求 Pure Tokens Desktop、MCP、Node/npm/Python/Go、浏览器自动化、代理或常驻服务。
3. 最终实现采用随安装包分发的单次原生 Go 执行器。执行器只读取当前宿主已声明的有效连接记录，在内存中使用匹配凭据；不能扫描用户目录、索取或输出凭据。早期“绝不读任何凭据”已被用户后续明确许可的精确适配器方案替代。
4. 产品只请求固定 Pure Tokens API；余额使用既有 console 用量和公开换算接口。个人全局生图规则只是历史参考，不能复制其中的其他服务地址或认证路径到产品。
5. 普通生成只读小型模型索引和选中 profile，不自动查余额、init、doctor、preflight 或实时目录。明确发现模型、请求字段／operation 缺口、拒绝诊断才按需查询目录。
6. 模型列表来自受控基础目录，不能从供应商名称、渠道路由或聊天模型列表推测数量和能力。参数、数量、参考图、编辑、音频等按具体 profile 声明；物理尺寸不能直接作为 API size。当前图片默认 `gpt-image-2`，视频默认 `grok-imagine-video-1.5-preview`。
7. 异步提交一次，立即回传任务 ID；status/wait/content 始终使用原任务。未知提交、超时、下载失败不能自动重提。下载不是交付；多图按零起点逐份下载、实际交付后再取下一份。
8. 本地附件按声明 multipart 随该操作提交；公开参考 URL 仅走声明的 JSON 字段。早期独立上传／转存设想已废弃，不能恢复成新前置步骤。
9. 用户提示简短、可操作，不能输出整份内部响应或把本地诊断码伪装成 API 返回码。余额失败不能推断所有媒体认证失效，金额缺失不能填零或猜测。
10. 安装升级保护完整受管文件清单，不覆盖同名未知目录或用户新增、修改内容。README 中指定标题下第一个 `text` 安装提示块由客户端下载页提取，修改时必须保留提取契约；提示应便于用户直接复制。

## 当前任务：Windows 安装流程收敛（2026-09-09）

用户要求优化安装期间大量黑色终端。已将 Windows sync 内的原生执行器调用统一到 Invoke-NativeExecutor：UseShellExecute=false、CreateNoWindow=true、UTF-8 双管道并发读取、关闭 stdin、保留 ExitCode；不使用 shell、提权或备用启动。参数处理支持空格及末尾反斜杠，拒绝引号／换行。

- 覆盖归属验证、管理清单、安装完整性、旧运行器检查和 init；Codex 专属旧插件迁移不在 WorkBuddy 路径上。
- 首次安装 README 中英文及更新 Skill 明确一次官方入口，失败即报告并停止，禁止现场 probe／shim／Python 补丁和换启动方式；必需审批仍遵守。原生等待上限本地校验 30 秒、init 195 秒、输出结束 5 秒；fetch 下载目录清理异常不覆盖安装结果。
- 不减少校验次数；解决本安装器创建额外控制台窗口的来源，不能控制 WorkBuddy 最外层命令窗口，未实测黑窗消失。
- 新增 Windows 5.1／7 启动器回归：成功输出、失败退出码、带空格和末尾反斜杠的目录参数；本机无 PowerShell，尚未运行该专项。
- 更新 Skill 安装指令避免逐目录手工重复命令，更新来源摘要和安装归档／候选包。尚未提交推送。
- 上轮 WorkBuddy 分类诊断及收尾修复已推送 5d7b22b。远程连接无匹配的具体根因仍待实机诊断。

## 已完成：全链路修复已完成，实机验收受连接条件阻塞（2026-09-08）

用户已明确要求按审查建议实施。已完成六项代码/门禁修复，保留 0.18.0 候选全部既有工作；未提交／推送／发布。

- 附件句柄、大小、时间及摘要校验，发送中完整性校验；同大小/同时间的变化也拒绝，保持 unknown 提交语义。
- 显式记录绑定下载 SHA-256、字节数、媒体类型；复用及 delivered 均核对。旧记录可续查原任务，无证明文件换目录下载，保留旧文件。
- 固定提交目录选择器不再复用旧 marker；Shell/PowerShell 增加旧入口配新宿主的隔离回归。
- 模型参数验证先于记录持久化；保守字符串投影禁止非法自由文本。
- 固定 Go 1.26.0 构建配置，源码/依赖/构建输入摘要、内嵌身份、六平台产物证明；release:validate 实际重建逐字节比较。当前平台实际二进制 build-info/offline preflight 进入 Go 测试。
- Linux 进入验收结构；五个 Linux 客户端目标 pending，其他尚未验证目标 unavailable 并说明原因。100 个汇总检查 pending，9 类详细用例逐宿主另计；不能声称真实可用。
- 验证完成：npm run check 通过（Go、go vet、57 项 Node）；npm run release:validate 通过，包括六平台可复现构建比较；git diff --check 通过。六种执行器、两份迁移归档、六份平台候选已重建。Windows PowerShell 新增用例未在本机运行，CI 入口已保留。
- 实际当前 Codex init 在网络前安全停止：active_connection_not_puretokens，api_request_executed=false；未创建付费任务。未读取/展示连接详情或借用其他客户端；真实媒体/附件验收保持 pending。本机现有 image/video Skill 元数据为 0.13.19，未发现新原生执行器安装，未修改实际安装。
- 下一个必要条件：用户在当前宿主选择可用的 Pure Tokens 连接后继续；先用原生执行器复核，再按已有授权完成明确的真实宿主安装、图片/编辑/续接/附件及视频验收，不走替代传输。当前候选未发布，sourceCommit=null。
- 修复前复现与问题对应关系：references/audits/2026-09-08-chain-review.md。回归断言正确行为的测试已经加入，不运行旧的“断言缺陷存在”探针。

## 已完成：ZCode 适配（2026-09-08）

用户要求落实 Switch 新增 ZCode 的 Skill 支持，保留当前 0.18.0 候选已有工作。

- 已接入六个 Skill、Shell／PowerShell 安装更新入口、凭据分派、init／doctor、宿主矩阵与验收记录。共十个安装宿主，九个凭据适配器。
- 依据 Switch ZCode adapter 和本机 ZCode 应用程序代码，连接来自 `.zcode/v2/config.json`；明确绝对 `ZCODE_DATA_BASE_DIR` 是 `.zcode` 的父目录。安装／doctor 使用同一目录规则；不读取真实连接或旧版存储。
- 只选择唯一启用且实际端点匹配的连接，要求支持的内联 bearer 格式；重复 JSON 键、多匹配、缺失凭据、引用凭据和畸形记录安全停止。会话模型由 ZCode 运行时提供；旧 root model 字段不能证明当前会话选择，已废弃该选择方案，init 明确只验证 API 能力。
- Go 夹具覆盖有效连接、选择歧义、重复键、错误端点、禁用、凭据缺失及引用、过深结构、宿主分派、路径覆盖、doctor 和禁止旧存储回退。Node 隔离测试覆盖六 Skill 安装、重复更新、空格目录和相对目录拒绝。Windows 入口用例已加入脚本，未在 Windows 实机运行。
- 回归中发现 Apple Silicon Node 与转译 Shell 架构不一致导致测试包缺失产物；下载夹具现与 Shell 安装器使用相同架构选择。宿主断言更新为十个。
- 六种执行器、两份迁移归档和六份平台候选已重建；`npm run check` 全部通过（Go 测试、go vet、53 项 Node 测试）；`npm run release:validate` 和 `git diff --check` 通过。80 项真实宿主检查仍 pending；远程工作区同步不能代替执行器和连接，原生附件交付未验收。
- 当前未提交、推送、发布或更新真实客户端。平台候选 sourceCommit=null，不能作为正式发布资源。

## 已完成：新增桌面客户端（2026-09-07）

用户要求支持 Switch 新增的两个客户端。已核对 Switch 注册表，新增的是 **Claude Desktop**（claude-desktop）和 **DSH Desktop**（dsh-desktop）。本次只修改 Skill 仓库，沿用既有跨仓库和真实宿主操作边界。

- 六个 Skill 的宿主清单、安装更新选择、凭据分派和诊断均已纳入两个客户端，版本统一为 0.18.0；保留全部 0.17.1 候选修复。
- Claude Desktop：读取本机 Desktop 3P 模式和当前 appliedId 指向的 UUID 配置，验证 gateway/static/bearer 及固定 Pure Tokens 端点；不扫描其他配置或借用 Claude Code 凭据。Skill 安装到本地 Code 会话的 ~/.claude/skills，与 Claude Code 共享，支持显式 CLAUDE_CONFIG_DIR。云端、SSH、WSL、Cowork 隔离会话不等于本机，无法访问本机执行器／连接即停止。
- DSH：从当前本地 Harness 的 settings.yaml 选择默认 provider/model，再按 apiKeyEnv 读取同目录 .credentials.yaml version 1 的一个 refs 条目；先验证端点，不回退到进程环境变量 Key。仅支持明确 DSH_HOME 或平台默认应用目录。YAML 解析编译进执行器，拒绝重复键、别名、合并键、多文档及过深结构。
- DSH 默认安装位置：macOS 应用数据下 dsh-desktop/harness/skills，Windows APPDATA 下对应目录；显式 DSH_HOME 可覆盖。doctor 检查本地和共享 Agents 的已声明 Skill 根，不扫描任意工程目录。
- 证据：Switch 的 crates/client-adapters/claude-desktop/src/lib.rs、dsh-desktop/src/lib.rs；官方源码／文档证据和精确字段记录在 references/credential-adapters.md。安装后的按需说明为各 Skill 的 references/desktop-hosts.md。
- 验证：npm run check 全部通过，含 Go 测试、go vet、52 项 Node 测试；两个新宿主使用合成记录和隔离安装，未读取真实连接或提交付费任务。六个平台执行器、两份迁移归档和六份草稿平台包已重建。Windows PowerShell 5.1／7 新增路径测试已入脚本，但本机未执行 Windows 真机。真实宿主检查共 72 项，全部仍 pending。
- 候选未提交／推送／发布／安装到真实宿主。平台包 sourceCommit 为 null（工作区未提交），不能被当作正式固定提交资源。下一步为指定宿主真实验收及用户要求的候选交付。

## 当前候选已实现

具体变更见 `CHANGELOG.zh-CN.md` 的 0.17.1 和工作区 diff。

| 项目 | 当前实现与证据 |
| --- | --- |
| 图片轮询 | `runtime/executor/main.go`、`polling_test.go`：新 pending 图片接受回执后 20 秒首次查询，此后每 3 秒；每窗口最多 40 次／120 秒。复用 `retry_not_before` 绝对时间，交接耗时扣除、续接不重置。实际 API Retry-After 优先，本地等待不伪装成 API header。视频保持最多 7 次／300 秒及原退避节奏。 |
| 余额格式 | `balance.go`、`balance_test.go`：仅余额请求内存补齐缺失的前缀，保留媒体认证原样；同时脱敏两种形式，区分余额认证拒绝与公开单位信息失败。历史上已做合成凭据的 Web 余额处理器联测，不能等同于远程用户验收。 |
| 升级归属保护 | `installation_guard.go`、两套原生安装器、`installation-history.json`：校验完整目录及文件哈希；新增、改动、缺失、符号链接会阻止覆盖和危险恢复。 |
| 历史迁移 | 从仓库历史生成、去重的目录快照用于识别没有归属标记的原版旧安装；旧会话记录 256 份通过测试。这不是 256 个发布版本，也不来自用户文件扫描。 |
| 媒体验证 | `media_containers.go`、`media_validation.go`、`reliability_test.go`：加强 GIF/WebM 结构及截断检查，坏文件不得保存为成功或复用。容器验证不等于真实视频播放验收。 |
| 错误公开化 | `public_errors.go`：受控分类提示，仅保留实际返回且识别为公开分类的错误码，不透传任意服务端文字。 |
| 对账续接 | `task_records.go`：用户明确 resume 对账记录时查询原任务一次，按响应刷新标记，不创建替代任务。 |
| 验收与维护 | 修复来源信息工具函数；加入真实宿主证据校验、Windows 文件保护回归及 fuzz targets。六个平台执行器和旧版迁移归档已在旧会话重建。 |

0.17.0 已包括可选任务记录、同任务恢复和实际交付标记、显式 preflight、按需 doctor、模型筛选、下载期限、更新检查、固定提交下载和平台资源选择。不要因这些功能出现在历史建议中再次实现。

## 尚未解决与下一步

1. **当前候选交付**：0.18.0 仍在本地。后续提交时必须包括必要的未跟踪实现、夹具与安装归属历史；先核对完整 diff 和产物一致性。发布平台资源是单独动作，脏工作区生成的包可能 `sourceCommit: null`，不能宣称已固定到发布提交。
2. **端到端耗时**：用户实测反馈上游生成约 40 秒、拿图需要 4–5 分钟。只修复了轮询；没有证据证明整体延迟已解决。下一轮应记录用户发起、POST 发出、任务接受、发现完成、下载完成、附件可见等时间，区分模型调度、服务端、网络与宿主交付开销。20 秒初始等待是待实测验证的参数，不是最优结论。
3. **宿主验收**：`references/host-acceptance.json` 现有 100 项汇总检查 pending，详细实机结果 0 条。先确定实际测试宿主、系统、安装版本及执行器哈希，再按 `host-acceptance-guide.md` 验证余额、图片、上传图编辑、生成图再编辑、多图、视频、跨会话续接和原生附件交付。本机过去被发现装有 0.13.19 是旧会话观察，本次未检查，不能当作当前事实。
4. **后续性能建议未实现**：阶段计时、按宿主提供已验证的简短交付步骤、减少确定性步骤之间的模型调度。旧会话曾建议合并等待与下载，但未实现；当前契约仍要求独立命令，不能把建议当作已批准的契约变更。60–90 秒只是当时提出的目标，不是承诺。
5. **Web 编辑 400**：旧会话 2026-09-06 核实 Web 生产 `7793cb81` 使用生成路径校验编辑参考图，develop 修复为 `f26f819e`，本地旧逻辑可复现拒绝。这是历史状态；后续必须重新核对部署，不能现在断言生产仍未修。相关文件位于相邻 Web 仓库的 `patches/puretokenplus-new-api/patches/139-image-edit-route-aware-staging.patch`。Skill 余额补丁不能解决此服务端问题。
6. **服务端依赖**：精确报价／支付前费用保证、服务端幂等、无任务 ID 的未知提交找回尚未实现。不得从本地校验或任务记录推导这些保证。
7. **宿主能力限制**：九个凭据适配器有 fixture 覆盖；Trae 仍无已验证适配器，安装成功不等于能调用 API。宿主忽略 Skill 优先级时，仓库声明本身不能强制改变工具选择。

## 维护入口与相称验证

- `AGENTS.md`：产品和安全边界；`CONTRIBUTING.md`：工程、构建、发布顺序。
- `skills/`：实际安装指令、逐模型 profile、执行示例及场景；`schemas/`：请求／回执契约。
- `runtime/executor/`：原生执行器、生命周期、适配器、媒体校验与测试。
- `runtime/puretokens-skill-fetch.sh`、`runtime/puretokens-skill-fetch.ps1`：原生获取；同目录 `puretokens-skill-install.sh`、`puretokens-skill-install.ps1`：原生同步。
- `references/media-model-catalog.json`：当前受控快照为 2026-09-03，8 个图片、10 个视频模型；这是安装快照，不是当前线上授权清单。由基础目录同步脚本更新，再生成 Skill profile 和 README。发布新鲜度上限 7 天。
- `runtime/executor/manifest.json`：平台产物及校验；`scripts/build-installation-history.mjs`：从受审查的仓库历史构建迁移清单。
- `dist/puretokens-skill-install*.zip` 及 `scripts/legacy-bootstrap/`：有意保留的旧更新器迁移桥，不能当作废弃垃圾删除。

源码改变后执行 `npm run executor:build`，必要时重建迁移归档和平台包；同时更新相关指令、契约、schema、版本、来源哈希和双语 changelog。工程门槛是 `npm run check`；发布再跑 `npm run release:validate`。平台包发布需匹配提交及校验值。Windows 的 PowerShell 5.1 与 7 分别验证；本机夹具不替代 Windows 真机。

2026-09-07 初次交接重新执行 `npm run check` 完整通过：契约与生成资料一致性、宿主证据一致性、Go 测试、go vet、51 项 Node 测试全部通过，0 失败。`npm run release:validate` 和 `git diff --check` 也通过。56 项真实宿主检查仍 pending；本次未运行 Windows 真机验收或额外 fuzz campaign。未发起付费任务、未读取真实连接凭据、未改真实宿主安装、未修改相邻项目。
