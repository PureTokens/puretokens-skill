# Pure Tokens Skills 当前开发状态

更新日期：2026-09-10。此文件只维护当前任务与未解决事项；历史实现、发布记录及审查证据见 Git、CHANGELOG 和 references/audits，不将历史状态作为当前事实。

## 目标与授权

用户要求补齐 Kimi Code、Qoder，并全面检查六个 Skill 的残留文档、契约与流程，实施清理优化。只修改本仓库；相邻 Switch 可只读核实，不修改或部署。用户已授权提交推送本轮修改；未授权真实客户端配置/安装或付费任务。

基线为 main 的 a7ea449（媒体链路与模型能力更新，已推送）。当前产品版本0.18.0；本轮修改仍在工作区，必须保留新增文件。宿主默认安装支持12项，11个凭据适配器；Trae仍无已验证适配器。

## 有效决定与实现

- 六个独立Skill入口与单次原生执行器不变；固定API请求、精确连接读取、无凭据持久化或传输回退。安装不增加用户运行时依赖。
- 默认或精确模型只读选中profile；索引仅选择/别名解析时读，实时目录按需。普通生成不运行init、doctor、余额或preflight。
- 异步submit只发一次并立即回执；同任务status/wait/content独立。图片默认5秒首次查询、3秒间隔、40次/120秒窗口；视频7次/300秒窗口。最多两个自动前台窗口，预算跨续接保存。错误、对账和无法容纳的重试停止自动续等。
- 下载时计算摘要；记录中的文件证明匹配后可本地恢复交付，实际附件handoff成功后才标记delivered。逐索引下载交付，不后台预取或重提。
- Kimi Code按默认模型关联连接解析；Qoder按真实端点选择唯一连接，不按名称识别、不比较多个凭据。安装、doctor和凭据适配共享相同目录优先级；Kimi doctor还检查已声明共享Skill目录。
- 新宿主依据：Switch 0ccc470两客户端适配器、已安装Kimi Code程序静态skillRoots/data-home代码与Qoder官方文档，详见credential-adapters.md。旧Python Kimi CLI不是新Kimi Code的路径依据。未读取真实连接。
- 本轮清理：宿主排查段移至按需desktop-hosts，纠正init不依赖配置选择的过度概括；校验器从host-support推导适配器清单；共享严格JSON解析器去除ZCode专属命名；未知宿主路径解析明确拒绝；清除本文件旧候选/旧计时等冲突叙述。历史迁移资产保留，不能当垃圾删除。

## 验证与剩余事项

新宿主阶段npm run check通过（Go、go vet、60项Node），六平台独立重建校验通过；Windows脚本覆盖已增加，未Windows实机运行。清理阶段也已通过npm run check（Go、go vet、60项Node）、executor:verify与git diff --check；六平台执行器、两份迁移归档及六份草稿平台包已重建。release:validate再次在同一模型目录新鲜度门禁停止。

- release:validate此前在模型目录采集超过7天处停止；快照2026-09-03，不伪造刷新。权威目录更新后同步并重跑门禁，不覆盖已确认的补充模型资料。
- 真实宿主API、原生附件交付及端到端耗时仍待指定环境验收；host-acceptance.json是唯一验收结果。夹具、静态源码和安装成功不能代替实机。
- 远程WorkBuddy诊断及Windows安装体验的实机结果仍未确认。
- 精确报价、服务端幂等和未知任务找回未实现；不从任务记录或本地校验推导保证。
- 发布前重建六平台执行器、两份迁移归档与平台候选；脏工作区候选sourceCommit=null不是正式发布资源。提交时包含新源码/测试/说明。

## 维护入口

AGENTS.md与CONTRIBUTING.md定义长期约束；skills/为安装指令及profile，schemas/为字段契约，runtime/executor/为行为实现，references/host-support.json为宿主声明，references/host-acceptance.json为实机证据。工程门禁npm run check；发布门禁npm run release:validate；executor:verify独立重建比较六平台产物。
