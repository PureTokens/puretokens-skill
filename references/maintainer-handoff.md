# Pure Tokens Skills 当前开发状态

更新日期：2026-09-12。任务标识：stable-distribution-performance-2026-09-12。此文件仅维护当前任务；历史实现和发布记录见 Git 与 CHANGELOG。

## 目标与授权

当前任务：用户已批准稳定发布分发、同版完整性核验后快速返回、init 总预算 20 秒、指令短路径与性能验收优化，并明确同意修改安装跟随 main 的原约定。此前“推送发布吧”已授权提交、推送及符合门禁的发布。Codex 本机首批验收后，2026-09-12 用户最新明确指定“用 gemini cli 或者 open code 验收吧。图片随便测。视频测试 xai 或者 h3 就行了。”已授权这两个现有客户端及图片测试、xAI／minimax-h3 视频测试，不再索要重复授权。本轮选择 Gemini CLI，先做明确的只读 API 验收，再逐项测试少量图片和一条低分辨率 xAI 短视频；不批量提交、不失败重提、不改连接、不借用其他客户端凭据，不伪造证据或绕过门禁。实施基线 HEAD 为 a1b72527e0f63de68e098a2bd9b0566ae7cd5fe2。

保留既有 Trae 验收分类修复、实机记录和维护文档修改，不回退用户已有工作。当前候选版本为 0.18.2；0.18.1 分发保持不可变，其验收表已完整归档至 `references/acceptance/0.18.1-host-acceptance.json`。

最新指示：用户在 OpenCode 诊断后明确要求“那你去修复”。正在实施并验证 OpenCode 专项修复；Gemini 保持暂停。此前图片与 xAI／minimax-h3 视频测试授权保留；只有本次有新实现依据的同宿主 init 通过后才继续 API／媒体，失败则停止，不换宿主或凭据。没有新版本实机证据前，不能沿用 0.18.1 通过状态。

验收目标：默认仅取已发布稳定清单指定的平台 ZIP，无整库下载回退；同版且清单完整时不下载 ZIP、不写入、不 init；有改动或缺失则停止。init 两次只读请求共用 20 秒预算。六平台重建及包检查、工程与发布门禁通过；实机缺口如实保留，不以夹具标记通过。

## 有效决定与实现

- OpenCode 0.18.2 修复（当前有效）：支持有界 JSONC 多层合并、XDG 目录、显式配置和项目／系统管理层，以及端点验证后的原生 API-key 存储。明确 model 不回退；缺省 model 只接受唯一已启用的匹配服务，不据此声称当前聊天身份。动态引用、插件、远程／会话未表示的覆盖及不支持的认证安全停止。不读取真实配置原文或历史库，不改用户连接。精确读取范围已写入 credential-adapters.md，六份 desktop-hosts.md 同步。
- 新增回归覆盖选择歧义、JSONC、合并优先级、原生凭据优先级、错误覆盖不回退、禁用过滤、重复键、体积／深度、插件、MDM、非普通文件和动态字段名。动态字段名漏洞经独立复核确认，五个用例实际先红后绿；修复前原始文本和解析后键／值都检查替换标记，避免继承旧凭据。独立安全复核无剩余明确阻断项；仅为代码复核，不是实机验收。
- 当前验证：OpenCode 聚焦测试通过，最终 10 秒有界模糊测试 139598 次通过；只应用本次运行时修改的隔离 HEAD 4acf06d 工作区 Go 全测／vet 通过。六平台执行器已重建；npm run check 通过（Go／vet、Node 80 通过，1 项本机无 PowerShell 跳过）；release:validate 可重现构建通过；release:package 六平台草稿包校验通过，sourceCommit=null，不可发布。0.18.2 活跃验收已重置为 118 pending、38 unavailable、0 passed，旧证据未改版本。
- 实机结果／材料：`/tmp/pt-opencode-0182.I7Qp8i/install.mjs` 从校验过的本平台 ZIP 更新 OpenCode 默认 Skill 目录，915 ms 完成；七份清单通过，14 个无关兄弟项元数据未变，安装后 executor 为 0.18.2。自动 init 观察一次，仍停止于 `active_connection_selection_unconfirmed`；未创建媒体任务。详见 `references/acceptance/0.18.2-opencode-macos-arm64.md`。隔离工作区 `/tmp/pt-opencode-isolated.GkB5sw/repo`。未提交、推送或发布。下方 0.18.1／Gemini 段落均为历史证据，不替代本段当前状态。
- 进一步的原生脱敏探针显示：当前 OpenCode XDG 配置根下存在一个配置文件，但合并后的文档没有 `provider` 声明；因此 0.18.2 按契约返回 selection-unconfirmed，而不是从 auth store 或 provider 标签猜 endpoint。这个结果解释了本机 init 仍未通过：运行时修复已生效，但当前本地有效配置不在支持的“声明 endpoint + 原生 api auth”范围内。未输出路径、provider ID、凭据或配置内容。
- OpenCode 专项诊断（2026-09-12）：本机版本仍为 1.18.23；官方 v1.18.23 固定到 `ef2880f379129aa048be9e9353e30aa168d42c17`。原生执行器哈希与候选一致，同宿主一次显式 init 复查在 20 ms 以退出码 1 返回 `active_connection_selection_unconfirmed`，API 请求／身份确认／凭据验证均 false。这不是新 OpenCode 会话执行证据。根因边界：本仓库 `credentialFromOpenCodeFile` 要求单份全局 JSON 的根 model 含 provider；官方 defaultModel 允许 model 缺省，再看有效最近模型及可用模型，实际会话还受请求／Agent／会话选择和多层 JSON/JSONC 合并影响。六个合成用例复现现有拒绝边界及不可借用未选连接的限制，overlay 测试通过；第一次探针 JSON 缺括号已改为结构化序列化，没有改产品代码。公开两个终端执行路径没有提供可直接依赖的当前 provider/model 环境契约。没有读取真实配置、最近模型、会话库或认证存储，不断言用户实际用了哪个来源。修复应补有效选择证明，或明确按本项目既有独立服务适配器方式验证唯一已配置服务连接的较窄能力；不能删除保护后取第一个 provider、要求用户填默认 model 掩盖缺口或猜环境变量。调查源文件、overlay、脱敏重查在 `/tmp/pt-opencode-diagnosis.6vBdzU`；已补入现有 OpenCode 验收附件，验收状态不改。产品代码、0.18.1 分发字节及用户连接未改动；若实施运行时修复，必须处理不可变版本规则和新增配置读取范围。
- Gemini 退出码 41 专项（2026-09-12）：0.57.0 的实际重查在 965 ms 退出，工具调用 0、执行器回执为空；仅对内存中的 stderr 做允许列表分类，命中 `client_auth_method_unresolved`（客户端未解析出认证方式），不是 API 拒绝。交互启动实际停在认证选择界面，没有输入或确认信任、登录、凭据。排查期间安装包版本从 0.57.0 变为 0.59.0，未由本任务执行升级；重新观察 0.59.0，启动前后版本一致，精确检测到认证选择提示和 API Key 选项，没有目录信任提示。探针结束时需终止启动包装进程，不将其退出状态当作对话成功。0.57.0 源码的网关类型校验遗漏不是本次命中的错误，已废弃此根因假设；不能从本地诊断断言用户未配置或 Key 无效。当前需要用户在本项目的 Gemini 认证界面确认现有 Pure Tokens 连接并先恢复普通对话，不要求提供配置或凭据、不改用 Google 登录替代 Pure Tokens 测试。完成后继续同宿主 init 和已授权图片／xAI 或 minimax-h3 视频验收。本轮未修改产品或客户端连接配置、未创建媒体任务。脱敏材料：`/tmp/pt-real-host-acceptance.ZqrAie/gemini-init-recheck-summary.json`、`gemini-startup-ui-current-summary.json`。
- 当前实机验收环境：macOS 26.6.2（25G83）、arm64、zsh 5.9／系统 sh bash 3.2.57、Codex CLI 0.153.4。候选包为 CI 34676764666 的 0.18.1、提交 4acf06d，已重新通过 release:verify --publishable。Codex 默认目录首次真实安装完成（814 ms），7 份受管清单及执行器哈希通过，53 个无关兄弟项的元数据未变。新 Codex 进程发现 6 个启用的正确路径；临时新会话实际读取帮助并完成回复，但 CLI 进程未在 60 秒内退出，已终止，不能宣称命令退出／端到端延迟通过。
- 同一 macOS 的显式项目目录（含空格、中文）完成 0.18.0 到 0.18.1 升级（559 ms），无关测试文件保留，实际 Codex 在升级前后发现对应版本的 6 个项目 Skill；read-only verify-installed 无文件改动。项目升级未传 host，init 明确延后，不重复检查已失败连接。探针最初误把全局／项目双作用域和 /tmp 的物理别名当作失败；修正测试观察器后通过，没有改产品安装器。临时探针位于 /tmp/pt-real-host-acceptance.ZqrAie。
- 先前 Codex 自动 init 返回 active_connection_not_puretokens，在本地停止，固定 API 请求为零；该 Codex 流程不再重查。用户现已明确指定 Gemini CLI／OpenCode 作为新验收目标，并授权媒体范围。Gemini CLI 0.57.0 的真实非交互启动于约 1.1 秒以退出码 41 停止（认证类退出），零终端工具调用、没有执行器回执，不能记为 Pure Tokens API 失败或已执行 init。没有重试或更改 Gemini 配置。OpenCode 1.18.23 已完成独立候选安装：macOS ARM64 默认目录同步 0.18.1 成功（1636 ms），七份受管清单和执行器哈希通过，14 个无关兄弟项元数据未变；新 `opencode debug skill` 进程退出 0，列出全部六项 Pure Tokens Skill，其中 `puretokens-models` 由更高优先级共享 `.agents/skills` 目录解析且字节一致。自动 init 随后在本地停止于 `active_connection_selection_unconfirmed`，没有执行固定 API 请求或创建媒体任务。由于没有启动新对话调用已加载 Skill，OpenCode installation 仍保持 pending；API、图片／视频、恢复和原生附件交付均未执行。临时安全探针与脱敏汇总保存在 /tmp/pt-real-host-acceptance.ZqrAie，不保存原始宿主日志或配置。
- 真实证据已写入 references/acceptance/0.18.1-codex-macos-arm64.md 并关联 host-acceptance.json：仅 Codex/macOS installation 改 passed，真实证据 1 条（installation-verified），校验确认 117 pending、38 unavailable、1 passed。适用范围仅所记录 CLI／OS／架构，不接受 Codex 桌面 UI、其他系统或架构。公开 latest 安装链路、本地 CLI 退出、真实 API、媒体交付和其他客户端仍保留缺口。新增证据后 npm run check 通过（Go 全测／vet、Node 80 通过、1 项无 PowerShell 跳过），npm run release:validate 通过（0.18.1 immutable、六平台可重现构建一致），git diff --check 通过；--stable 仍正确拒绝其余待验收内容。此轮改动未提交、推送或发布；不新增平台支持声明。
- 验收分类修复：校验器读取 host-support 的完整适配器状态，不再只接收宿主 ID。注册表中 credentialAdapter=pending 的宿主，credentialFixtures 及除 installation 外的摘要和详细用例必须 unavailable；安装证据独立验收。缺失／未知／重复支持元数据和夹具状态不一致会拒绝。全量检查 13 个客户端，仅 Trae 存在此误分类：macOS、Windows 各 3 项 pending 改 unavailable 并说明原因，安装保持 pending，Linux 和其余客户端不变。分类修复完成时为 118 pending、38 unavailable、0 passed、真实证据 0 条；随后实机进展见上方。
- 本轮仅修改维护侧校验器、回归测试、验收表和维护文档；分发内容及 0.18.1 版本不变。没有增加 Trae 适配器，没有把“不支持”当作验收通过，没有放宽正式发布门禁。
- 当前优化已完成实现及本地验收：schema-v2 稳定清单固定版本、源码提交、平台包与选择器及执行器 SHA-256；包内仅带当前系统脚本。显式源码 sync 保留，普通用户不再跟随 main 或自动回退源码。同版路径先核验选中执行器，再由发布选择器 verify-installed 验证七份清单及选中版本／平台身份，不下载 ZIP、不写入、不 init。
- init 已用 executeInitContext 共用 20 秒总预算，Windows 进程保护上限改为 25 秒。更新 Skill、契约、场景、两份 README 和 AGENTS 已同步，首条安装提示词保持原样。
- 新 validate-platform-releases.mjs 对六包、完整目录、源码字节、SHA-256、脚本和平台选择做校验，--publishable 额外核验提交。publish.yml 只在手工指定已有版本标签时运行：三系统测试 -> 实机验收 --stable -> 六包检查 -> 草稿上传 -> 下载复核 -> 发布。未实际触发发布。--stable 要求现有已声明可用目标验收完成，不将空或全 unavailable 记录视为通过；开发门禁仍允许 pending。
- 阶段：代码已提交并成功推送，当前 HEAD 为 4acf06d406bbe009a5fffcc2343e03a4ec7443a4，本地与远端分支均为 codex/stable-distribution-0.18.1。用户已补充 workflow 权限，权限阻塞已解除。CI 34676764666 全部通过；正式发布仍缺真实客户端证据。
- 发布预检确认：远端 main 仍在实施基线，GitHub Release 列表为空，无 v0.18 系列远端标签。新版入口在发布包缺失时停止，因此只推候选分支，不更新线上 main；不能先切换安装入口再留用户等待首个稳定包。
- 保持六个独立 Skill、单次原生执行器、固定 API、同宿主凭据边界与一次提交约束；不新增用户运行时、代理或传输回退。普通生成不加 init、doctor 或目录预查。
- Pi：先匹配唯一端点，再允许读取同根精确 auth.json，按同一 provider ID 关联；支持的内联认证优先，空值、未知格式或动态值停止，不回退。只读这一额外声明文件，不扫描其他存储或任意环境。
- Pi 内置环境认证的 39 个已审 ID 阻止低优先级 models 回退；其余普通自定义名称仍可使用纯内联值。CLI、扩展、会话覆盖和未来客户端变化不在已验证范围。
- 原始上游依据为 Pi 71dca871bc80b6bc97be37f0ca3189399d651fff，env-api-keys.ts 的 SHA-256 为 6876b915ffe1342f8be69ef1f75721c8a265a332a1bbbf5afe36b6f80b332552。已废弃网页工具返回的不同列表；root curl 与独立复核的原始文件一致，36 个映射加 3 个特殊 ID，无缺项或多余项。
- Pi 的安装、凭据和 doctor 目录规则一致；媒体正文和两份 README 补齐 Pi。PowerShell 使用 OSArchitecture，新增 SysWOW64 实际 x86 入口测试。
- 已完整交付的记录恢复为 done；异构系统路径只用于验证任务文件身份，字节交付仍要求本机绝对路径和完整性证明。换机下载仅用显式输出目录重绑定原任务，不重提。
- 采用分发版本不可变策略，从 0.18.1 开始。release:validate 检查同版本逐提交历史及工作区内容，候选升版不等于可发布。该本地门禁不替代远端分支保护。
- 实机证据新增架构、Shell 和执行模式约束；WSL 必须对应 Linux。安装、API、完整媒体交付等级从同一证据派生，不将远程结果归入本地摘要。

## 最终验证与剩余事项

- npm run check 通过：Go 全部测试与 vet 通过；Node 76 通过、1 项因本机没有 PowerShell 跳过。新增 init 测试覆盖两请求同一截止时间、无重试，以及本地 HTTP 服务器在两个阶段分别阻塞响应体。
- npm run release:validate 通过：Skill 摘要、模型新鲜度、0.18.1 candidate 版本门禁及六平台可重现构建全部通过。npm run release:package 通过：六个当前平台专用归档、四个脚本和清单的完整内容及校验和一致。
- 实际 darwin-arm64 ZIP 为 3,052,437 字节。隔离 HOME／目标目录完成首次安装，核验六个 Skill 和执行器共七份受管清单；重复 install 和 update 各两次清单／选择器读取，零 ZIP、零安装文件写入、零 init。仅克隆临时合成 sourceCommit 供下载夹具路由，正式候选清单仍为 null；不是线上下载或实机客户端验收。探针位于 /tmp/pt-stable-artifact-smoke.1faEDJ/smoke.mjs，隔离安装目录已清理。
- 三项独立发现均已修复并复核关闭：Gemini 物理路径别名／尾斜杠；同版核验与并发更新交错导致错误版本回执；发布工作流 GITHUB_SHA 未绑定检出标签。前两项新增回归在隔离恢复旧函数后均失败、修复后通过；发布探针确认覆盖错误的 dispatch SHA 后按实际检出提交打包。
- compat_install、compat_credentials、compat_task_chain 复核完成，无剩余明确发现；早期 PowerShell 静态结论已由下述最终 Windows CI 补充。工作流 YAML 解析及 git diff --check 通过。
- 分类修复前 node scripts/validate-host-acceptance.mjs --stable 按预期拒绝：当时 124 个摘要项 pending、真实证据 0 条。实现阶段的 --publishable 曾正确拒绝 sourceCommit=null；提交后重跑 release:package 和 release:verify -- --publishable 均通过，六包 sourceCommit 与 12605c10bfad7cd9f348c2aadffdef4255e94c30 匹配。版本门禁为 immutable。源码证明通过仍不代表实机验收通过。
- 本轮发布预检重跑 npm run check、npm run release:validate 均通过。远端 main 的既有 CI 34518651764 在 Windows 附件替换测试失败，安装器测试未执行。根因：测试忽略删除已打开文件的失败，随后截短原文件，却要求原句柄成功传输。仅修改测试，检查操作结果；删除被阻止时验证原字节不变，关闭后再真正替换并验证旧请求拒绝新路径。Go 全测、vet、格式检查和独立复核通过；Windows 动态验证尚未执行，不放松产品附件校验。
- 首次推送因缺少 workflow scope 被拒。用户完成官方授权后，本轮确认 PureTokens 账户新增权限，再次推送成功，远端候选 SHA 与本地一致，main 未变化。没有改用其他凭据／协议，也没有删除工作流。
- CI 34675993724 三系统 Go／vet、Windows 三入口安装生命周期以及仓库门禁均通过，但 PowerShell 5.1 下载夹具的 ZIP 带反斜杠成员，被生产路径保护正确拒绝。测试提交 de8aa22 显式生成与正式包一致的正斜杠成员，并新增校验和正确仍拒绝反斜杠的坏包用例。下一轮 CI 34676482726 暴露 PS5.1 未自动加载 ZipArchiveMode 程序集，4acf06d 在夹具显式加载 System.IO.Compression；没有放宽生产路径检查。独立静态复核及本地分发／包测试 19 通过、1 项无 PS 跳过。正式分发字节未变，0.18.1 immutable 门禁通过。
- 最终 CI 34676764666：repository、macOS、Linux、Windows 全部 success，提交 SHA 为 4acf06d406bbe009a5fffcc2343e03a4ec7443a4。包括 Windows 原生 Go／vet、PowerShell 5.1、PowerShell 7 和 SysWOW64 x86 的安装生命周期／稳定下载夹具。此为隔离 CI，不是实际客户端 API／附件验收，也不代表 Windows ARM64 已运行。
- 已下载该 CI 的 platform-release-candidates 到 /tmp/pt-ci-release-34676764666.iZZMeV，并运行 npm run release:verify -- --publishable --directory <该目录>：六个平台 ZIP、四个脚本、清单完整内容、校验和及当前已提交源码身份全部通过。远端候选 SHA 再次核对一致，main 仍为 a1b7252，GitHub Release 列表仍为空，无 v0.18 系列标签。
- 前序推送交付完成；分类修复及新增实机证据尚未提交、推送或发布。正式发布仍待真实客户端验收，不能以 CI 通过替代。未创建标签、Release 或切换 main 安装入口。此本地交接记录刻意不随发布代码提交。下一步是在用户明确指定的适用实机测试环境补齐 API、参考附件、结果交付和跨会话证据，再按正式门禁发布；Trae 仅安排已声明平台的安装验收。实际 Codex 默认目录已安装候选 0.18.1；没有修改连接配置、借用其他宿主凭据或发起付费媒体任务。
- 分类修复验证：旧校验器放过实际错误记录的反例已复现。新增 4 组回归覆盖三系统无适配器摘要、所有详细 API 用例、夹具与支持元数据不一致、改名宿主无硬编码及安装-only 和完整媒体宿主混合发布。npm run check 通过（Go 全测／vet，Node 80 通过、1 项本机无 PowerShell 跳过）；npm run release:validate 通过，0.18.1 immutable 且六平台可重现构建一致。--stable 仍按预期拒绝缺少证据，独立契约复核无发现。隔离工作区 /tmp/pt-acceptance-classification.GdMHKf/repo 仅应用本轮 5 个文件的补丁，不带此交接记录；隔离 npm run check 和 npm run release:validate 均通过，结果与原工作区一致。逐项比较基线确认仅 6 个摘要状态由 pending 改为 unavailable，其他客户端和证据未变；git diff --check 通过。

## 已完成兼容基线证据（不是本次最终门禁）

- Pi 认证、完整交付恢复、跨系统路径与非法安装根目录已跑过失败复现及修复后的聚焦测试。新增相对文件真实字节反例、WSL 模式约束与版本历史回归。
- 最终 Pi 安全复核通过：独立原始源码集合比较及 go test -run '^TestPi' -count=1 均通过，无剩余明确发现；安装与任务恢复专项复核也已完成。
- Pi 拒绝列表校正后已重新执行 npm run executor:build、npm run check、npm run release:validate，全部通过。Go/go vet 通过，Node 67 通过、1 项因无 PowerShell 跳过；六平台源码、内嵌标识、产物哈希及可重现构建一致。版本门禁正确标记 0.18.1 为 candidate。
- npm run release:package 已重新生成六个平台草稿包，sourceCommit=null，不可作为已发布资源使用。逐包核验归档校验和、完整文件集合、各文件与当前源码的字节一致性、六个 Skill 版本以及恰好一个平台执行器，无缓存或杂项文件。
- 从实际 darwin-arm64 ZIP 在隔离临时目录完成首次安装与重复更新；0.18.1 版本、执行权限、六个 Skill 和执行器的七份受管清单均通过。未传宿主，init 明确延后；临时夹具已删除，未读取真实凭据或请求 API。
- 该兼容基线时 host-acceptance.json 为 124 个摘要项 pending、实机证据 0 条（已由本轮分类修复更新为 118 pending）。Windows 原生执行、ARM64/模拟环境、真实 API 与附件交付未验收；当时没有读取真实配置、安装到用户目录或发起付费请求。
- 官方模型目录沿用此前 2026-09-12T02:04:41.893Z 的 23 模型快照；保留分别审查的兼容补充，本轮不再重复刷新模型。
- 上述兼容基线已包含在本次最终门禁中，结果见“最终验证与剩余事项”。不得将自动化测试替代待验收的实机结果。

## 维护入口

AGENTS.md 与 CONTRIBUTING.md 为长期约束；runtime/executor/ 为实现；skills/ 为安装指令；host-support.json 为宿主声明；host-acceptance.json 为实机结果。工程门禁 npm run check；发布门禁 npm run release:validate；平台包 npm run release:package。临时上游与回归探针位于 /tmp/puretokens-pi-security-review.2p9PpF，永久回归已进入仓库。
