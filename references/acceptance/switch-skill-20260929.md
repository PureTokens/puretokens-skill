# Switch × Skills 脱敏离线证据 — 2026-09-29

性质：开发工作区的源码／夹具证据，**不是实机通过，不是发布证明**。
Skills 0.18.7 候选（HEAD 4e65eed＋工作区）；Switch 0.4.0 开发工作区
（HEAD 332959e＋工作区）。所有配置／凭据和 API 响应均为合成样例。
未读取真实客户端认证文件、应用真实配置、提交计费任务或部署。

| 检查 | 结果 | 范围 |
| --- | --- | --- |
| VS Code Switch writer | 7 passed | Rust adapter夹具，含实际 writer 对照共享JSON、macOS/Windows目录和保留用户配置 |
| Switch 专用模型与分组 | 26 passed / 4 files | 固定 Node 24.18.0，model-capabilities、runtime-group-snapshot、useClientApplyActions、ConfigApplySkillGuide |
| Switch TraeWork/矩阵 | 23 passed | 合成原生服务／故障恢复与矩阵一致性；不表示被禁用的正式配置入口可用 |
| Web Key 模型名单 | 1 passed | 完整所选并集、组外排除、旧模型移除、同Key刷新、保存未确认失败、无自动重试 |
| Web 既有 Key 回归 | 2 passed | VS Code 受管 Key 更新不旋转，网关忽略保存时拒绝成功 |
| Skills VS Code与媒体/音频/Jev 聚焦测试 | passed | 实际 dispatcher＋合成宿主记录／HTTP；200/401/403、精确专用路径、目录刷新，附件模拟 |
| 六平台原生执行器 | passed | Go 1.26.0，源码标识及六份产物哈希校验通过 |
| `npm run check` | passed | 仓库／生成文档／Schema／Go全测与vet；Node 109 passed、0 failed、1 skipped |
| 文档／跨仓库一致性 | passed | 20客户端×6能力×5阶段完整；本地链接、共享fixture字节、三个仓库diff空白检查 |

Node 跳过项是本机缺少 PowerShell 的下载诊断测试，不是Windows运行通过；
Windows两个架构是交叉构建，双平台路径是合成夹具，仍没有Windows真实客户端交付证据。
第一次构建期间补充了路由歧义保护，输入变更守卫正确拒绝混合构建；
随后针对最终源码完整重建六个平台，并通过最终工程门禁。

构建源码 SHA-256：
`65084cb372369d9928c1b05dd0c3cfed644a303379838785e7ed534973cb2f57`。
共享合成fixture SHA-256：
`05f5917b3ec0bcdf640cf5793295f15aa9418f15610773c6d6bd893478643d05`。
完整开发日志位于本机 `/tmp/pt-switch-skill-check-20260929.log`；
版本绑定产物摘要见 `runtime/executor/build-proof.json`，未发布。

首次 Switch 前端检查用默认 Node 26.9.0通过但有引擎警告；随后使用项目
指定 Node 24.18.0重新执行，同样26项通过。该警告不当作固定工具链通过。
新测试编写时曾因夹具未建立目录、目录样例缺少 capabilities、视频带参数
触发未安装 profile 的目录读取而失败；只修正测试夹具及不依赖安装根的
基础提交用例，没有放宽生产校验或把预查当调用成功。

共享配置样例：

- Skills `test/fixtures/switch-vscode-connection.json`
- Switch `tests/fixtures/client-configs/vscode/skill-connection.json`

内容包含明确标注的合成 Key；真实配置从未复制入库。读取器不修改文件，
不在错误结果中输出 Key、匹配指纹或比较结果。多组即使同Key仍不合并；
便携profile、动态认证和无法解释的覆盖安全停止。

公开线上只读证据见 [目录快照](switch-skill-public-catalog-20260929.json)。
GitHub latest返回0.18.6，published_at为2026-09-19T10:58:55Z。
公开稳定 `release-manifest.json` 同时确认schema2、0.18.6与
`4e65eed34bc8e889f8577c006950d377dfabdd24`，不能把工作区候选当安装默认。
匿名目录可见不等于实际测试Key分组可见，公开价格不是可执行报价。

剩余：每个矩阵行①实际加载／执行、②当前会话与Switch记录对应、
③真实分组名单／拒绝、④生产专用请求、⑤真实附件或文本交付均须单独实测。
新候选 `host-acceptance.json` 保持原pending／unavailable，无新增passed。
详见 [矩阵](../switch-skill-compatibility.md)与[实机费用方案](../switch-skill-test-plan.md)。
