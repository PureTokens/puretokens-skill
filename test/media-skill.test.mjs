import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import path from "node:path";
import test from "node:test";
import { repositoryRoot } from "../scripts/skill-registry.mjs";

const skillRoot = path.join(repositoryRoot, "skills", "puretokens_media");

async function readSkill() {
  return readFile(path.join(skillRoot, "SKILL.md"), "utf8");
}

test("media Skill publishes the complete MCP tool contract", async () => {
  const [skillText, manifestText, directCloudContract] = await Promise.all([
    readSkill(),
    readFile(path.join(skillRoot, "skill.json"), "utf8"),
    readFile(path.join(skillRoot, "references", "direct-cloud-contract.md"), "utf8")
  ]);
  const manifest = JSON.parse(manifestText);
  assert.equal(manifest.mcp.server, "puretokens-image");
  assert.deepEqual(manifest.mcp.tools, [
    "puretokens_list_media_models",
    "puretokens_generate_image",
    "puretokens_image_result",
    "puretokens_generate_video",
    "puretokens_video_result"
  ]);
  assert.equal(manifest.rules.stableRequestIdRequired, true);
  assert.equal(manifest.rules.reuseRequestIdOnHostRetry, true);
  assert.equal(manifest.rules.neverAutoResubmitAfterError, true);
  assert.equal(manifest.rules.neverAutoSwitchModelAfterError, true);
  assert.equal(manifest.rules.successRequiresNativeMediaResult, true);
  assert.equal(manifest.rules.completedMediaRequiresLocalFileDelivery, true);
  assert.equal(manifest.rules.completedImageRequiresNativeImageContent, true);
  assert.equal(manifest.rules.completedMediaReportsExactModel, true);
  assert.equal(manifest.rules.completedMediaReportsLocalDelivery, true);
  assert.equal(manifest.rules.completedVideoUsesBoundedNativeResource, true);
  assert.equal(manifest.rules.usesCuratedNaturalLanguageAliasRegistry, true);
  assert.equal(manifest.rules.usesDeterministicMediaDefaults, true);
  assert.equal(manifest.rules.defaultResultCount, 1);
  assert.equal(manifest.rules.explicitCountRequiredForMultipleResults, true);
  assert.equal(manifest.rules.supportsDirectCloudWithoutDesktop, true);
  assert.equal(manifest.rules.directCloudRequiresHostInjectedCredentials, true);
  assert.equal(manifest.rules.directCloudUsesExplicitGatewayEndpointCapabilities, true);
  assert.equal(manifest.rules.directCloudImagesAlwaysAsync, true);
  assert.equal(manifest.rules.directCloudDeliversSynchronousImageBytes, true);
  assert.equal(manifest.rules.directCloudDeliversAsyncContentBytes, true);
  assert.equal(manifest.naturalLanguageAliases, "skills/puretokens_media/references/natural-language-aliases.json");
  assert.equal(manifest.directCloudContract, "skills/puretokens_media/references/direct-cloud-contract.md");
  assert.equal(manifest.workBuddyAdapter, "skills/puretokens_media/adapters/workbuddy-execution.md");
  assert.match(skillText, /第一步必须调用[：:][\s\S]*puretokens_list_media_models/);
  assert.match(skillText, /稳定的 `request_id`/);
  assert.match(skillText, /同一个 `task_id`/);
  assert.match(skillText, /MCP 通道以 `structuredContent\.model` 为事实来源/);
  assert.match(skillText, /`type == resource`/);
  assert.match(skillText, /Direct Cloud 通道/);
  assert.match(skillText, /不需要 Pure Tokens Desktop、Router、额外 CLI 或 MCP/);
  assert.match(directCloudContract, /PURETOKENS_API_KEY/);
  assert.match(directCloudContract, /PURETOKENS_API_BASE_URL/);
  assert.match(directCloudContract, /always set `async: true`/);
  assert.match(directCloudContract, /synchronous `data\[\]\.b64_json`/);
  assert.match(directCloudContract, /synchronous[\s\S]*`data\[\]\.url`/);
  assert.match(directCloudContract, /asynchronous task/);
  assert.match(skillText, /GET \/v1\/media\/models/);
  assert.match(skillText, /默认只请求 `n=1` 个结果/);
  assert.match(skillText, /Direct Cloud 图片提交必须始终传 `async: true`/);
  assert.match(skillText, /同步 `data\[\]\.b64_json`、同步 `data\[\]\.url` 和异步任务/);
  assert.match(skillText, /视频始终按异步任务处理/);
  assert.match(skillText, /Direct Cloud 通道由宿主的 Direct Cloud 执行层完成/);
  assert.match(skillText, /缺少“可调用 MCP 工具”或“HTTPS 执行能力与已注入的 Direct Cloud 凭据”/);
  assert.doesNotMatch(skillText, /GET \/v1\/models/);
  assert.doesNotMatch(skillText, /PURETOKENS_API_KEY/);
});

test("media Skill does not advertise tools absent from the media MCP", async () => {
  const skillText = await readSkill();
  const manifest = JSON.parse(await readFile(path.join(skillRoot, "skill.json"), "utf8"));
  assert.doesNotMatch(skillText, /puretokens_get_balance|puretokens_get_model_price/);
  assert.deepEqual(manifest.mcp.tools, [
    "puretokens_list_media_models",
    "puretokens_generate_image",
    "puretokens_image_result",
    "puretokens_generate_video",
    "puretokens_video_result"
  ]);
});

test("natural language aliases resolve only to explicit catalog model IDs", async () => {
  const aliases = JSON.parse(await readFile(
    path.join(skillRoot, "references", "natural-language-aliases.json"),
    "utf8"
  ));
  const image2 = aliases.aliases.find((entry) => entry.phrases.includes("image2"));
  assert.deepEqual(image2, {
    phrases: ["image2", "image 2", "gpt image 2", "openai image 2"],
    capability: "image",
    modelIds: ["gpt-image-2"]
  });
  const aliasesByPhrase = new Map(aliases.aliases.flatMap((entry) => entry.phrases.map((phrase) => [phrase, entry])));
  assert.deepEqual(aliasesByPhrase.get("nano banana pro").modelIds, ["gemini-3.0-pro-image"]);
  assert.deepEqual(aliasesByPhrase.get("nano banana 2").modelIds, ["gemini-3.1-flash-image"]);
  assert.deepEqual(aliasesByPhrase.get("nano banana").modelIds, [
    "gemini-3.0-pro-image",
    "gemini-3.1-flash-image"
  ]);
  assert.deepEqual(aliases.defaults, {
    image: {
      modelId: "gpt-image-2",
      when: "The user asks to generate an image without naming a model."
    },
    video: {
      modelId: "grok-imagine-video-1.5",
      when: "The user asks to generate a video without naming a model."
    }
  });
  assert.match(aliases.normalization, /complete registered phrases/i);
});

test("media Skill behavior scenarios cover ambiguity, empty catalog, unavailable MCP, failure, and timeout", async () => {
  const [skillText, scenariosText] = await Promise.all([
    readSkill(),
    readFile(path.join(skillRoot, "references", "behavior-scenarios.json"), "utf8")
  ]);
  const scenarios = JSON.parse(scenariosText).scenarios;
  assert.deepEqual(scenarios.map((scenario) => scenario.id), [
    "default-single-result",
    "explicit-multiple-results",
    "direct-cloud-image-delivery",
    "ambiguous-model",
    "no-media-model",
    "mcp-unavailable",
    "task-failure",
    "task-timeout",
    "content-delivery-failure"
  ]);
  assert.deepEqual(scenarios.slice(3, 4).map((scenario) => scenario.firstTool), [
    "puretokens_list_media_models",
  ]);
  assert.match(skillText, /模型不存在或匹配多个/);
  assert.match(skillText, /目录为空/);
  assert.match(skillText, /MCP 不可用/);
  assert.match(skillText, /safeToResubmit=false/);
  assert.match(skillText, /轮询超时/);
  assert.deepEqual(scenarios.find((scenario) => scenario.id === "mcp-unavailable")?.expected, "use_direct_cloud_when_available_or_report_missing_execution_capability");
  assert.equal(scenarios.find((scenario) => scenario.id === "direct-cloud-image-delivery")?.expected, "force_async_true_and_write_actual_image_bytes_before_reporting_success");
  assert.match(skillText, /不得自动换模型/);
  assert.match(skillText, /不得自动重新提交/);
  assert.match(skillText, /同步图片结果、`\/content` 或本机写入失败/);
});

test("the shared media source has target-specific Claude and WorkBuddy deliveries", async () => {
  const [skillText, manifestText] = await Promise.all([
    readSkill(),
    readFile(path.join(skillRoot, "skill.json"), "utf8")
  ]);
  const manifest = JSON.parse(manifestText);
  assert.equal(manifest.distribution.claudeDesktop.format, "zip");
  assert.equal(manifest.distribution.claudeDesktop.archiveRoot, "puretokens_media");
  assert.equal(manifest.distribution.claudeDesktop.enableAfterImport, true);
  assert.equal(manifest.distribution.workbuddy.generatedSkillName, "puretokens_workbuddy_router");
  assert.equal(manifest.distribution.workbuddy.alwaysApply, true);
  const workBuddyAdapter = await readFile(path.join(skillRoot, "adapters", "workbuddy-execution.md"), "utf8");
  assert.match(workBuddyAdapter, /DeferExecuteTool/);
  assert.equal(manifest.distribution.codex.managedByDesktop, true);
  assert.equal(manifest.distribution.codex.managedSkillDirectory, "~/.codex/skills/puretokens_media");
  assert.equal(manifest.distribution.codex.requiresPluginFeature, false);
  assert.equal(manifest.distribution.codex.supportsDirectCloud, true);
  assert.equal(manifest.distribution.codex.managedMcpWhenDesktopAvailable, true);
  assert.match(skillText, /跨客户端共用的媒体行为源/);
  assert.match(skillText, /不得自行写入、替换或删除任何客户端的 Skill、MCP、Router 或 Secret 配置/);
});
