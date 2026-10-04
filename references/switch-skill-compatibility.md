# Switch × Skills 专用模型兼容矩阵

核对时间：2026-10-04（Asia/Shanghai）。Skills 最新正式发布为 **0.18.7**
（GitHub `published_at=2026-10-04T07:47:48Z`，非 draft/prerelease），分发提交
`aa28ac089a137459a284c68199a8abc21576ab44`。工程 CI `37186053573` 和正式发布
`37186386278` 均通过，六平台实际上传产物已核验。维护者批准的本版本验收例外
见 `host-acceptance.json`；166 项待验收与零实机证据保持不变。发布说明和安装／更新
流程不增加该维护记录的提示。

Switch 基线仍为 `332959e` 加当前工作区；其版本指引与 Web 合同已同步，
不代表这两个项目的新版本已经部署。音频／Jev指令随 0.18.7 分发，Switch暂不
开放其执行示例，等待专用接口和真实宿主交付验收。

目标链路：Switch 选择分组并应用 **聊天连接** → 用户单独安装 Skills →
当前客户端执行对应 Skill → 原生执行器从该客户端有效连接取得凭据 →
固定专用 API 按精确模型 ID 鉴权 → 当前客户端真实展示或附件交付。
专用模型不需要进入聊天模型列表；用户仍需有一个能执行工具的聊天模型。
只选专用分组而没有兼容聊天模型时，Switch 不应伪造聊天模型来完成应用。

## 每格的五项证据

每格写作 `发布层级 [①,②,③,④,⑤]`：

1. 当前客户端实际加载正确 Skill 并运行原生执行器。
2. 从 **Switch 实际写入的当前客户端连接**取得正确 Key，而非任意默认连接。
3. Key 的所选分组及多分组完整模型名单包含精确目标模型。
4. 使用精确模型 ID 和对应专用 API，收到可解释的专用结果。
5. 图片实际打开、音视频实际播放／原生附件交付、转写和 Jev 文本实际展示。

`发布` = **已发布支持**的实现，`源码` = **仅源码支持**，均不等于实机通过。
`夹` = **夹具通过**（仅合成配置／模拟 API）；`实` = **实机通过**；
`不` = **不支持**；`待` = 待实机验证，无通过声明。
③中的夹具只证明所选分组并集及刷新逻辑，**没有读取或验证任何用户真实 Key**。
①和⑤必须在同一客户端／系统／架构／会话模式上验收，不能拿安装器或文件校验替代。

当前 Switch／Web 源码公开目录为20个客户端；Skills 0.18.7 注册19个安装宿主、18个凭证夹具适配器。目录成员不代表线上已部署。

下表反映 0.18.7 分发实现；各行 ② 的精确路径及选择限制见
[凭证契约](credential-adapters.md)，系统支持范围见
[宿主验收记录](host-acceptance.json)。当前版本没有任何“实”格。
①均为“待”，虽然安装／执行器夹具已覆盖；⑤也均为“待”，模拟附件不是实际交付。

| 客户端（Switch ID） | 图片 | 视频 | 专用音频（含音乐） | Jev | 向量 | 重排 |
| --- | --- | --- | --- | --- | --- | --- |
| Claude Code (`claude-code`) | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 不 [不,夹,夹,不,不] | 不 [不,夹,夹,不,不] |
| Codex (`codex`) | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 不 [不,夹,夹,不,不] | 不 [不,夹,夹,不,不] |
| WorkBuddy (`workbuddy`) | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 不 [不,夹,夹,不,不] | 不 [不,夹,夹,不,不] |
| Gemini CLI (`gemini-cli`) | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 不 [不,夹,夹,不,不] | 不 [不,夹,夹,不,不] |
| Grok Build (`grok-build`) | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 不 [不,夹,夹,不,不] | 不 [不,夹,夹,不,不] |
| OpenCode (`opencode`) | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 不 [不,夹,夹,不,不] | 不 [不,夹,夹,不,不] |
| Claude Desktop 本地 Code (`claude-desktop`) | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 不 [不,夹,夹,不,不] | 不 [不,夹,夹,不,不] |
| DSH Desktop (`dsh-desktop`) | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 不 [不,夹,夹,不,不] | 不 [不,夹,夹,不,不] |
| 官方 DeepSeek Harness (`deepseek-harness`) | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 不 [不,夹,夹,不,不] | 不 [不,夹,夹,不,不] |
| MiniMax Code Desktop (`minimax-code`) | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 不 [不,夹,夹,不,不] | 不 [不,夹,夹,不,不] |
| ZCode (`zcode`) | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 不 [不,夹,夹,不,不] | 不 [不,夹,夹,不,不] |
| Kimi Code (`kimi-code`) | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 不 [不,夹,夹,不,不] | 不 [不,夹,夹,不,不] |
| Qoder (`qoder`) | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 不 [不,夹,夹,不,不] | 不 [不,夹,夹,不,不] |
| Pi (`pi`) | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 不 [不,夹,夹,不,不] | 不 [不,夹,夹,不,不] |
| Hermes (`hermes`) | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 不 [不,夹,夹,不,不] | 不 [不,夹,夹,不,不] |
| EvoX (`evox`) | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 不 [不,夹,夹,不,不] | 不 [不,夹,夹,不,不] |
| **VS Code 多模型** (`vscode`) | **发布** [待,夹,夹,夹,待] | **发布** [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 不 [不,夹,夹,不,不] | 不 [不,夹,夹,不,不] |
| Octop 本地工作区 (`octop`) | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 发布 [待,夹,夹,夹,待] | 不 [不,夹,夹,不,不] | 不 [不,夹,夹,不,不] |
| **TraeWork/SOLO** (`trae`) | 不 [待,不,待,不,不] | 不 [待,不,待,不,不] | 不 [待,不,待,不,不] | 不 [待,不,待,不,不] | 不 [不,不,待,不,不] | 不 [不,不,待,不,不] |
| Paseo (`paseo`) | 不 [待,不,待,不,不] | 不 [待,不,待,不,不] | 不 [待,不,待,不,不] | 不 [待,不,待,不,不] | 不 [不,不,待,不,不] | 不 [不,不,待,不,不] |

矩阵中②④是已有宿主适配器夹具和公共执行器夹具的**组合证据**，不是
每个宿主都执行过四类专用调用。VS Code与本次MiniMax Code测试贯穿实际命令分发、
Switch writer 对应文件、模拟权限拒绝和四类专用路径。向量／重排行的
②③只表示共享连接／分组机制已覆盖，不代表存在对应 Skill。
Cursor 当前不在 Skill 注册表及 Switch 当前可用配置入口内，六类均不支持。

## 官方 DeepSeek Harness 新增范围

Switch/Web当前源码有20个客户端；Skills 0.18.7 注册19个宿主，其中18个有凭据夹具，Trae仍没有适配器。官方Harness使用独立 `deepseek-harness`，仅默认本地Desktop组合；不是社区DSH。安装目录是DSH_HOME/skills或~/.dsh/skills，读取Desktop Cordis patch及其声明的refs。高优先级home patch、额外插件组合、会话覆盖、环境凭据覆盖或多匹配连接均不借用旧默认。0.18.7 已分发该新宿主实现，0.18.6 不含此适配。

共用writer夹具及实际dispatcher的图片200/401/403已覆盖；其他专用能力是共享执行器组合证据。①⑤仍待实机，不能宣传安装即可完成全部能力。来源及本次模型差异见[更新证据](acceptance/client-model-sync-20261003.json)。

## 分组权限、目录和已发布媒体

- Switch `config/catalog.rs` 的 `selected_group_model_ids` 保存完整有序并集；
  `compatible_client_model_ids` 仅决定聊天列表。专用模型名单不能用聊天子集覆盖。
- Switch `account/managed_keys.rs` 对相同分组的受管 Key 调用
  `update_managed_key_groups` 刷新模型名单，不旋转 Key。确认窗口过期时先刷新
  分组并拒绝不一致快照；配置事务检查 Key 覆盖所选分组。
- Web `product-client-resource-handlers.mjs` 从所选分组重算 `model_limits`，
  多分组写入 `auto_groups` 和限制，回读验证完整集合。单分组由组路由限制；
  不是“关闭名单限制就能用其他分组”。刷新必须重新应用／显式触发，不承诺后台自动刷新。
- 真实调用仍受 Key 状态、额度、渠道和网关专用路由约束。不能以目录可见、
  模型安装 profile、聊天可用或 init 成功证明专用 POST 有权限。普通生成
  不加目录预查；鉴权失败停止，不轮换 Key、不更换模型、不自动重提。
- 2026-10-04 06:04 UTC 匿名公开目录快照有112个模型，其中29个图片／视频模型（15图片、14视频）；本次移出seedream-5.0-pro；已审核的9个音频／音乐与3个Jev精确IDs也已可见。公开可见仍不证明某个Key有权调用。0.18.7 按用户指定默认图片为 `gpt-image-2.5-flare`、默认视频为 `minimax_h3`；这两个精确ID与输入schema已在本次快照复核一致。旧preview、旧连字符H3不再作为安装选项，不改写已有任务ID或自动替换用户精确选择。
- 图片和视频仍按异步任务协议：提交一次、先回执、再独立等待／取原任务；
  所选profile声明multipart时，图片编辑才发送当前附件原字节至 `/v1/images/edits`；仅声明JSON URL的型号不接受本地附件，不自动上传托管。
  下载成功只进入待宿主交付，只有真实交付后才标记 delivered。

## 专用接口与输出合同

专用调用固定在 `https://api.puretokensx.com`，不使用聊天条目的 URL 作请求目标：

| 能力 | 精确 ID 例子／权威来源 | 专用入口 | 交付 |
| --- | --- | --- | --- |
| 图片 | `gpt-image-2.5-flare`（默认）；已安装 image profile | `/v1/images/generations`、multipart `/v1/images/edits` | 同任务图片字节，再真实宿主附件 |
| 视频 | `minimax_h3`（默认）、`grok-imagine-video-1.5`；video profile | `/v1/videos`、声明的 `/v1/videos/edits` | 同任务 MP4/WebM，再实际播放／附件 |
| 配音／转写／声音 | `stepaudio-2.5-tts`、`stepaudio-2.5-asr`、`stepaudio-3-gen-preview`；[音频合同](audio-execution-contract.md) | `/v1/audio/speech`、`/v1/audio/transcriptions`、`/v1/audio/generate` | 同步音频字节或最终 text |
| 音乐 | `stepaudio-3-music-preview`；音频合同 | `/v1/audio/music/submit`，仅原任务 status/content | 异步 MP3/WAV，非同步音频命令 |
| Jev | `jev-latest`、`jev-1.13.0`、`jev-preview`；[Jev合同](jev-evaluation-contract.md) | `/typesafe/v1/systemone` | 同步 Choice/Score/Noul 类型结果 |

精确 ID 必须同时满足安装 profile 与真实 Key 分组／接口权限；
示例和源码允许列表不是对当前账户的推荐／授权。音频／Jev仍未实机通过。

## VS Code 修复范围与会话覆盖

0.18.6 在出现第二条 Pure Tokens 模型时停止。0.18.7 只在同一个
`customendpoint` 组内验证每条完整 URL、协议、唯一模型 ID 和内联 Bearer；
所有条目必须对应同一受支持服务和同一 Key。不同 Key、多匹配组、混合服务、
额外认证来源、动态值、重复 ID 和不支持的发现／路由覆盖都停止。
组名、Switch 标记不用于推断服务身份；不猜聊天框当前模型。
比较只在执行器内存中进行，回执没有 Key、摘要指纹或比较结果。

支持范围仍限 macOS／Windows **本地默认 profile**。便携模式已主动拒绝；
命名 profile、远程／WSL、宿主未反映在声明文件中的会话覆盖不支持。
无法从文件判断一个不可观察的会话覆盖，因此不能宣称“所有覆盖已检测”。
若宿主上下文与声明文件不一致，Skill 必须停止；覆盖隔离的真实验收仍待做。
OpenCode/Pi 的已声明覆盖顺序有离线回归，但也不代表所有宿主覆盖均通过。

## MiniMax Code 0.18.7 范围

静态桌面包3.1.0证据确认 `<runtime-data>/skills` 和 `config.yaml` 的默认选择。Switch可写三种协议提供方；执行器只取 `defaultModel` 指向的启用条目，不要求删除其他模型，不比较不同提供方Key。CLI／未知会话覆盖不支持；独立安装器遇到桌面偏好文件但没有运行目录环境时要求已确认的绝对目标，不猜默认路径。具体选择、文件安全和拒绝项见凭证契约。跨仓库同一合成writer夹具覆盖两系统，尚无实机验收。

## Trae 及新增客户端独立开发项

Trae 当前 Switch 槽位指 **TraeWork/SOLO**。当前源码 `traework_config/mod.rs::preflight` 已改为版本／宿主检查，配置通过宿主原生服务并要求明确重启确认；恢复预检仍不可用。旧“任何配置写入前都拒绝”的描述已作废。Switch内部使用CDP不构成Skills传输许可。宿主加密恢复记录、状态receipt及模型缓存不是Skill的当前会话凭据。

Skills仍没有Trae凭证适配器；成功配置不能算专用调用通过。没有新增数据库猜测、解密、浏览器／CDP、另存Key或借用其他客户端的路径。

后续必须先提供：宿主明确的当前连接只读
契约；Skill 加载目录和本地执行契约；macOS／Windows 的图片打开、音视频
播放和文本结果展示。取得这些依据后再开发 `trae` 适配器，随后分项实测；
旧 `~/.trae/skills` 可写不等于 TraeWork 能发现它。

Paseo 的 Switch 源码写独立 `agents.providers` 配置，不能假定被启动的 Codex
把它落盘到 `~/.codex`。须先核对其会话传递与 Skill 发现，当前无直接宿主支持。
Orca／Herdr已从当前Switch／Web公开目录移除，保留历史记录但不再列入当前矩阵；任何壳客户端都不能借用内部Agent身份宣称独立支持。

## 向量与重排：先定义任务和输出，尚未开发

向量候选任务：用户明确提供一组短文本，要求“输出这些文本的向量供检索使用”。
规划固定 `POST /v1/embeddings`，精确受审模型 ID，一次有界文本批次；
输出用户指定目录的 JSONL（输入序号、维度、数值）并交付真实附件，
聊天只显示条数／维度，不铺满向量、不建设隐式向量库或自动索引磁盘。

重排候选任务：用户提供一个查询与一组候选文本，要求“按相关性排序并保留原编号”。
规划固定 `POST /v1/rerank`，一次有界请求；输出原编号、排名和原始相关分，
需要时交付 JSON/CSV。分值不能冒充概率／准确率，不生成缺失事实或伪造聊天回复。

以上是任务设计，**未添加命令、Skill、接口白名单或模型支持声明**。
实现前需确认线上精确模型、真实请求／返回结构、计费、权限和有界输入；
不复用聊天 `/v1/chat/completions`，不把专用模型放进聊天列表。

## 脱敏证据与验收

- E1：`runtime/executor/{connection_diagnostics,client_file_credentials,desktop_credentials,opencode_credentials,pi_credentials,zcode_credentials,new_host_credentials}_test.go`：
  合成宿主配置、声明路径、选择与覆盖边界；不含用户凭据。
- E2：`test/fixtures/switch-vscode-connection.json` 与 Switch
  `tests/fixtures/client-configs/vscode/skill-connection.json` 字节一致。
  Switch `skill_connection_fixture_matches_actual_multi_model_writer` 在两个系统路径夹具
  中把实际 writer 输出与文件比较；Skill `TestSwitchVSCodeConnectionBoundary`
  覆盖三种聊天协议、同 Key、多 Key、多组、覆盖与不可变配置。
- E3：`TestSwitchVSCodeDedicatedPermissionsAndFreshCatalog` 运行实际 dispatcher；
  图片／视频／音频／Jev 的模拟 200、401、403，各仅一次专用 POST；
  目录有→无→有反映最新响应而不使用聊天模型列表。不是生产网关权限验收。
- E4：Web `services/bff/test/switch-skill-model-permissions.test.mjs`：
  完整专用模型并集刷新、移除旧名单、排除组外模型、保存未确认失败，
  不取 Key、不旋转 Key、不请求推理。Switch 客户端分组／确认快照测试提供上游证据。
- E5：`combined_workflow_test.go`、`chain_regression_test.go`、
  `audio_test.go`、`music_test.go`、`evaluation_test.go`：
  原字节附件、同任务续接、未知结果不重提、音频文件／转写及 Jev 类型。
  宿主附件交付在这些测试中模拟，⑤不因此改为实机通过。
- E7：[2026-10-04增量证据](acceptance/client-model-sync-20261004.json)：MiniMax静态桌面源码摘要、双仓库writer夹具及媒体目录变化。
- E6：[2026-10-03公开目录差异和宿主证据](acceptance/client-model-sync-20261003.json)：包含源摘要、精确增删ID和候选边界；9月29日快照仅保留作历史，不作为当前模型可用性或价格依据。
- 实际系统／客户端验收沿用 `host-acceptance.json`，不能复用不同版本、
  不同宿主或 Web/Pic 的成功结果。测试执行记录见
  [本轮离线证据](acceptance/switch-skill-20260929.md)。
- 下一阶段按[实机方案与成本](switch-skill-test-plan.md)单独授权。
  本轮没有读取真实客户端认证文件、修改真实配置或发起付费请求。
