# 2026-10-04 客户端／模型增量交付

> 以下是本轮发布准备前的证据快照。维护者随后于 2026-10-04 批准 0.18.7 带版本验收缺口发布，并要求不在面向用户的发布、安装或更新流程增加缺口提示。该决定仅记录在维护侧 `host-acceptance.json` 的版本例外中；166 项 pending、空实机证据和各项功能边界不变，工程／CI／可重现构建／上传产物校验继续执行。

版本：未发布0.18.7；稳定版核对基线0.18.6。用户批准本地修复，未授权本轮发布、真实配置读写或计费请求。没有执行这些操作。

## 修改

- Skills：H3首尾帧明确使用 `image_to_video`，保留首尾顺序与两个附件；正常声明内选择不读目录。同步当前29个媒体profile，移出目录未列出的seedream-5.0-pro；保留历史任务身份，默认图片gpt-image-2.5-flare／视频minimax_h3不变。
- Skills：新增MiniMax Code Desktop 3.1.0独立宿主、安装选择与凭证适配。凭证跟随Switch保存的defaultModel；多协议提供方、模型ID斜杠、区域偏好目录和拒绝边界有隔离夹具。原生执行器四类专用调用和目录有→无→有、401／403测试均只用合成HTTP。多个不同Key时只允许明确默认所选连接；不比较、不回退。未知会话覆盖仍不支持。
- Skills：20客户端×6能力×5阶段矩阵同步；移除当前目录中的Orca／Herdr，加入MiniMax；修正Trae配置preflight描述，但仍不增加Skill凭证或CDP传输。
- Switch：按宿主、能力和核对发布版本展示稳定入口／源码候选／无入口。只有已发布组合保留复制执行示例；音频／Jev、新宿主、VS Code多模型适配与向量／重排不会被稳定版安装提示误报可用。聊天模型准入不变。
- Switch：MiniMax实际writer输出与Skills共享合成样本在macOS／Windows布局逐项对照。更新产品方案、兼容矩阵、双语变更说明和交接契约。
- Web：仅同步product-client-api-contract文档；下载页提示词未修改，未改网关或部署。

## 验证

`npm run check`通过：Go全测／vet，Node111通过、1项因无PowerShell跳过。`npm run release:validate`通过：候选新鲜度、源码身份与六平台可重现构建一致。它不是稳定实机验收门禁通过。

Switch前端16项、类型检查、定向ESLint、MiniMax适配器12项及定向Rust格式检查通过。Web分组完整模型并集／刷新／拒绝权限夹具1项通过。三仓库修改范围空白检查通过。20客户端成员、顺序、20行矩阵与跨仓库共享样本已核对。

多结果机制原测试依赖移出的线上profile；现改为明确的fixture-multi-image合成profile，保留多结果交付、尺寸配对、类型与计数回归。当前15个图片profile均声明n=1，不以合成夹具扩展线上能力。

## 限制

MiniMax仅本地默认Desktop会话，CLI、不可确认的会话覆盖或路由覆盖不支持。独立安装器不读取桌面偏好配置；偏好文件存在而没有明确运行目录环境时，须使用已确认的绝对Skill目标，不能猜默认目录。执行器的凭证路径解析单独遵守声明的偏好文件和安全校验。

真实宿主加载、API和附件交付均未新增通过记录；候选仍有166个宿主／系统摘要待验收。Windows本机PowerShell执行未验收，跨编译不替代运行。音频／音乐／Jev仍仅源码／夹具，向量和重排仍未实现。真实计费测试须先给当前分组报价方案并另行授权；不自动重提。

详细源摘要、合成样本哈希与构建身份见同目录 `client-model-sync-20261004.json`。本地日志位于 `/tmp/pt-client-model-{build,check,release}-20261004.log`、`/tmp/pt-switch-skill-{tests,typecheck,lint}-20261004.log`、`/tmp/pt-web-skill-permissions-20261004.log`。
