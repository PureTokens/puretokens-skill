# 2026-10-03 客户端与模型同步验收

范围：Skills 0.18.7未发布候选；保留原有未提交工作。未读用户凭据、未写真实客户端配置、未发起计费请求、未提交／推送／发布。

- 官方DeepSeek Harness新增独立宿主、Shell／PowerShell安装位置、凭据读取、诊断、安装manifest和支持契约。默认本地Desktop范围；其他组合／覆盖按契约停止。
- Switch/Web客户端目录21项；Skills候选18个安装目标，17个凭据夹具适配器。Trae无适配器，Paseo无已验证宿主路径，Orca／Herdr依赖其实际内部Agent；不虚增支持。详见[五阶段矩阵](../switch-skill-compatibility.md)。
- 媒体profile共30个（16图片、14视频），新增12个、移出5个。按用户指定默认图片为gpt-image-2.5-flare、视频为minimax_h3；GPT Image新增Flare／Sunburst及1K／2K／4K；当前H3精确ID minimax_h3。旧任务身份不改写。
- 修复声明的https_url输入、标量／数组数量校验、Minimax首尾帧互斥及video_urls引用校验。JSON URL专用操作不伪装成本地附件支持；Veo缺少附件操作声明时停止。
- Step音频／Jev在公开目录可见，接口权限和输出仍未实机验收。Qwen音频撤销、实时、向量／重排未新增入口。

验证结果：

| 检查 | 结果 |
| --- | --- |
| npm run check | 通过：Go测试／vet；Node110通过、0失败、1跳过 |
| npm run release:validate | 通过：目录新鲜度、候选版本、六平台可复现构建 |
| Switch官方Harness writer | 9通过；共用fixture校验macOS／Windows写入形状 |
| Switch兼容矩阵 | 5通过 |
| Web专用模型权限并集／刷新 | 1通过 |
| PowerShell运行 | 本机无PowerShell，未验证；不能据此宣称Windows通过 |
| 实机／计费／原生附件 | 未执行，维持pending |

执行器源摘要：`2f9c377f18844759ceeb5c7a64f8c72f329892ceba4538dc2576afaac2a88ee1`。公开目录源摘要与精确变化见[脱敏快照](client-model-sync-20261003.json)。
初次同步日志：`/tmp/pt-client-model-{build,check,release-validate}-20261003.log`。

默认模型后续修改已验证：图片 `gpt-image-2.5-flare`、视频 `minimax_h3`（用户所说的minmax-h3，经线上精确ID核对）。入口／manifest／契约／索引及示例一致；校验器拒绝默认值漂移，用户指定和已有任务不改写。02:40 UTC的公开目录复核两者完整schema不变；全目录行数变为113，另一个型号seedream-5.0-pro未出现，原全量同步快照仍保留，此后续只改默认值。重新通过工程检查（110通过、1个PowerShell跳过）和发布校验（含六平台可复现构建），执行器生产源码未变。日志：`/tmp/pt-default-model-check-20261003.log`、`/tmp/pt-default-model-release-validate-20261003.log`。仍未发计费请求或发布。

修改仓库：puretokens-skill（运行时代码、安装、模型、测试、文档及二进制）；Pure_Tokens_Switch（共用writer夹具和测试、兼容矩阵及交接文档；不改变用户配置写入行为）；PureTokenPlus_web（交接文档；不改变网关或线上设置）。
