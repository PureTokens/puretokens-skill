import assert from "node:assert/strict";
import { execFile } from "node:child_process";
import { readFile } from "node:fs/promises";
import path from "node:path";
import test from "node:test";
import { promisify } from "node:util";
import { repositoryRoot } from "../scripts/skill-registry.mjs";

const skillRoot = path.join(repositoryRoot, "skills", "puretokens_media");
const execFileAsync = promisify(execFile);

async function readSkill() {
  return readFile(path.join(skillRoot, "SKILL.md"), "utf8");
}

test("media Skill publishes the complete MCP tool contract", async () => {
  const [skillText, manifestText, directCloudContract, packageText, registryText] = await Promise.all([
    readSkill(),
    readFile(path.join(skillRoot, "skill.json"), "utf8"),
    readFile(path.join(skillRoot, "references", "direct-cloud-contract.md"), "utf8"),
    readFile(path.join(repositoryRoot, "package.json"), "utf8"),
    readFile(path.join(repositoryRoot, "skills", "index.json"), "utf8")
  ]);
  const manifest = JSON.parse(manifestText);
  const packageManifest = JSON.parse(packageText);
  const registry = JSON.parse(registryText);
  assert.equal(packageManifest.version, manifest.version);
  assert.equal(registry.skills.find((skill) => skill.name === manifest.name)?.version, manifest.version);
  assert.equal(manifest.mcp.server, "puretokens-image");
  assert.deepEqual(manifest.mcp.tools, [
    "puretokens_list_media_models",
    "puretokens_generate_image",
    "puretokens_image_result",
    "puretokens_generate_video",
    "puretokens_video_result"
  ]);
  assert.equal(manifest.rules.stableRequestIdRequired, true);
  assert.equal(manifest.rules.liveCatalogDefinesSupportedModels, true);
  assert.equal(manifest.rules.registeredAliasesAreConvenienceOnly, true);
  assert.equal(manifest.rules.reuseRequestIdOnHostRetry, true);
  assert.equal(manifest.rules.neverAutoResubmitAfterError, true);
  assert.equal(manifest.rules.neverAutoSwitchModelAfterError, true);
  assert.equal(manifest.rules.successRequiresNativeMediaResult, true);
  assert.equal(manifest.rules.completedMediaRequiresLocalFileDelivery, true);
  assert.equal(manifest.rules.completedImageRequiresNativeImageContent, true);
  assert.equal(manifest.rules.image2ReturnsNativeImageWithoutResultPolling, true);
  assert.equal(manifest.rules.completedMediaReportsExactModel, true);
  assert.equal(manifest.rules.completedMediaReportsLocalDelivery, true);
  assert.equal(manifest.rules.completedVideoUsesBoundedNativeResource, true);
  assert.equal(manifest.rules.usesCuratedNaturalLanguageAliasRegistry, true);
  assert.equal(manifest.rules.usesDeterministicMediaDefaults, true);
  assert.equal(manifest.rules.defaultResultCount, 1);
  assert.equal(manifest.rules.explicitCountRequiredForMultipleResults, true);
  assert.equal(manifest.rules.supportsDirectCloudWithoutDesktop, true);
  assert.equal(manifest.rules.supportsSkillDefinedImage2Api, true);
  assert.equal(manifest.rules.supportsSkillDefinedConnectionVideoApi, true);
  assert.equal(manifest.rules.directCloudRequiresHostInjectedCredentials, true);
  assert.equal(manifest.rules.directCloudUsesExplicitGatewayEndpointCapabilities, true);
  assert.equal(manifest.rules.directCloudImagesAlwaysAsync, true);
  assert.equal(manifest.rules.directCloudRequiresDeliveryCapability, true);
  assert.equal(manifest.rules.directCloudUsesZeroBasedContentIndexForMultipleImages, true);
  assert.equal(manifest.rules.directCloudDeliversSynchronousImageBytes, true);
  assert.equal(manifest.rules.directCloudDeliversAsyncContentBytes, true);
  assert.equal(manifest.naturalLanguageAliases, "skills/puretokens_media/references/natural-language-aliases.json");
  assert.equal(manifest.directCloudContract, "skills/puretokens_media/references/direct-cloud-contract.md");
  assert.equal(manifest.workBuddyAdapter, "skills/puretokens_media/adapters/workbuddy-execution.md");
  assert.match(skillText, /第一步必须调用[：:][\s\S]*puretokens_list_media_models/);
  assert.match(skillText, /稳定的 `request_id`/);
  assert.match(skillText, /原始 `task_id`/);
  assert.match(skillText, /MCP 通道以 `structuredContent\.model` 为事实来源/);
  assert.match(skillText, /`type == resource`/);
  assert.match(skillText, /Direct Cloud 通道/);
  assert.match(skillText, /本 Skill 定义的 Pure Tokens Connection Images API/);
  assert.match(skillText, /Pure Tokens Connection Videos API/);
  assert.match(skillText, /POST https:\/\/api\.puretokensx\.com\/v1\/images\/generations/);
  assert.match(skillText, /POST https:\/\/api\.puretokensx\.com\/v1\/videos/);
  assert.match(skillText, /不依赖系统、开发者或 AGENTS 指令/);
  assert.match(skillText, /Codex 或 CC Switch/);
  assert.match(skillText, /不要求 `puretokens-image` MCP 或额外的 Direct Cloud 凭据/);
  assert.match(skillText, /可执行认证 HTTPS 视频请求并可交付实际视频字节/);
  assert.match(skillText, /普通聊天连接仅配置 API Key/);
  assert.match(skillText, /原生 Pure Tokens 媒体执行器/);
  assert.match(skillText, /不能把它当作图片\/视频执行器/);
  assert.match(skillText, /不需要等待 Skill 发版/);
  assert.match(skillText, /不需要 Pure Tokens Desktop、Router、额外 CLI 或 MCP/);
  assert.match(directCloudContract, /PURETOKENS_API_KEY/);
  assert.match(directCloudContract, /PURETOKENS_API_BASE_URL/);
  assert.match(directCloudContract, /always set `async: true`/);
  assert.match(directCloudContract, /synchronous `data\[\]\.b64_json`/);
  assert.match(directCloudContract, /synchronous[\s\S]*`data\[\]\.url`/);
  assert.match(directCloudContract, /asynchronous task/);
  assert.match(directCloudContract, /do not invent a request[\s\S]*JSON field/);
  assert.match(directCloudContract, /cannot provide that delivery path; report the missing[\s\S]*without submitting/);
  assert.match(directCloudContract, /content\?index=N/);
  assert.match(skillText, /GET \/v1\/media\/models/);
  assert.match(skillText, /默认只请求 `n=1` 个结果/);
  assert.match(skillText, /Direct Cloud 图片提交必须始终传 `async: true`/);
  assert.match(skillText, /真实媒体交付能力/);
  assert.match(skillText, /同步 `data\[\]\.b64_json`、同步 `data\[\]\.url` 和异步任务/);
  assert.match(skillText, /`gpt-image-2` 的 MCP 生成调用会直接返回原生图片内容/);
  assert.match(skillText, /视频始终按异步任务处理/);
  assert.match(skillText, /Direct Cloud 通道由宿主的 Direct Cloud 执行层完成/);
  assert.match(skillText, /已注入的 Direct Cloud 凭据、HTTPS 执行能力与真实媒体交付能力/);
  assert.doesNotMatch(skillText, /GET \/v1\/models/);
  assert.doesNotMatch(skillText, /Authorization: Bearer/);
  assert.doesNotMatch(skillText, /api\.krill-ai\.com/);
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
  assert.deepEqual(aliasesByPhrase.get("nano banana pro").modelIds, ["nano-banana-pro"]);
  assert.deepEqual(aliasesByPhrase.get("nano banana 2").modelIds, ["nano-banana-2"]);
  assert.deepEqual(aliasesByPhrase.get("nano banana").modelIds, [
    "nano-banana-pro",
    "nano-banana-2"
  ]);
  assert.deepEqual(aliasesByPhrase.get("qwen image 2.0 pro").modelIds, ["qwen-image-2.0-pro"]);
  assert.deepEqual(aliasesByPhrase.get("minimax h3 video").modelIds, ["minimax-h3"]);
  assert.deepEqual(aliases.defaults, {
    image: {
      modelId: "gpt-image-2",
      when: "The user asks to generate an image without naming a model."
    },
    video: {
      modelId: "grok-imagine-video-1.5-preview",
      when: "The user asks to generate a video without naming a model."
    }
  });
  assert.match(aliases.normalization, /complete registered phrases/i);
});

test("published media model catalog keeps both READMEs and registered aliases in sync", async () => {
  const [catalogText, aliasesText] = await Promise.all([
    readFile(path.join(skillRoot, "references", "published-model-catalog.json"), "utf8"),
    readFile(path.join(skillRoot, "references", "natural-language-aliases.json"), "utf8")
  ]);
  const catalog = JSON.parse(catalogText);
  const aliases = JSON.parse(aliasesText);
  const models = new Map(catalog.models.map((model) => [model.id, model]));

  assert.equal(catalog.schemaVersion, 2);
  assert.match(catalog.updatedAt, /^\d{4}-\d{2}-\d{2}$/);
  assert.equal(catalog.serviceCatalog.source, "Pure Tokens base model catalog");
  assert.equal(catalog.serviceCatalog.path, "/api/product/docs/model-catalog");
  assert.match(catalog.serviceCatalog.capturedAt, /^\d{4}-\d{2}-\d{2}T/);
  assert.equal(catalog.models.length, 17);
  for (const id of [
    "gpt-image-2",
    "nano-banana-pro",
    "seedream-5.0-pro",
    "grok-imagine-video-1.5-preview",
    "seedance-2.5"
  ]) assert.equal(models.has(id), true, `${id} must be published from the base catalog`);
  assert.equal(models.has("grok-imagine-video-1.5"), false);
  assert.equal(models.has("gmj-video-seedance-2.5-long"), false);

  for (const model of catalog.models) {
    assert.equal(model.name, model.id);
    assert.equal(typeof model.provider, "string");
    assert.equal(Number.isInteger(model.vendorId), true);
    assert.equal(Array.isArray(model.capabilities), true);
    assert.equal(model.capabilities.length > 0, true);
  }

  for (const { capability, modelIds } of aliases.aliases) {
    for (const modelId of modelIds) {
      assert.equal(models.get(modelId)?.capabilities.includes(capability), true, `${modelId} must retain its documented capability`);
    }
  }
  assert.equal(models.get(aliases.defaults.image.modelId)?.capabilities.includes("image"), true);
  assert.equal(models.get(aliases.defaults.video.modelId)?.capabilities.includes("video"), true);

  await execFileAsync(process.execPath, [
    path.join(repositoryRoot, "scripts", "sync-readme-media-catalog.mjs"),
    "--check"
  ], { cwd: repositoryRoot });
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
    "direct-cloud-multiple-image-content",
    "ambiguous-model",
    "no-media-model",
    "new-live-catalog-model",
    "direct-cloud-no-media-model",
    "direct-cloud-missing-delivery-capability",
    "skill-defined-image2-api",
    "skill-defined-connection-video-api",
    "connection-video-api-not-executable",
    "workbuddy-explicit-host-model",
    "host-native-model-not-verified-for-media",
    "mcp-unavailable",
    "task-failure",
    "task-timeout",
    "content-delivery-failure"
  ]);
  assert.equal(scenarios.find((scenario) => scenario.id === "ambiguous-model")?.firstTool, "puretokens_list_media_models");
  assert.match(skillText, /模型不存在或匹配多个/);
  assert.match(skillText, /目录为空/);
  assert.match(skillText, /MCP 不可用/);
  assert.match(skillText, /safeToResubmit=false/);
  assert.match(skillText, /轮询超时/);
  assert.deepEqual(scenarios.find((scenario) => scenario.id === "mcp-unavailable")?.expected, "use_direct_cloud_when_available_or_report_missing_execution_capability");
  assert.equal(scenarios.find((scenario) => scenario.id === "direct-cloud-image-delivery")?.expected, "force_async_true_and_write_actual_image_bytes_before_reporting_success");
  assert.equal(scenarios.find((scenario) => scenario.id === "direct-cloud-multiple-image-content")?.expected, "retrieve_zero_based_content_indexes_in_declared_order");
  assert.equal(scenarios.find((scenario) => scenario.id === "direct-cloud-no-media-model")?.expected, "report_api_key_catalog_scope_without_desktop_group_instructions");
  assert.equal(scenarios.find((scenario) => scenario.id === "direct-cloud-missing-delivery-capability")?.expected, "report_missing_delivery_capability_without_submission");
  assert.equal(scenarios.find((scenario) => scenario.id === "skill-defined-image2-api")?.expected, "call_the_puretokens_connection_images_api_once_at_the_puretokens_user_endpoint_with_gpt_image_2_without_mcp_or_direct_cloud_fallback");
  assert.equal(scenarios.find((scenario) => scenario.id === "skill-defined-connection-video-api")?.expected, "read_the_active_connection_catalog_then_submit_one_exact_video_model_to_the_puretokens_user_videos_endpoint_and_retrieve_the_same_task_content_without_mcp_or_direct_cloud");
  assert.match(scenarios.find((scenario) => scenario.id === "connection-video-api-not-executable")?.expected || "", /preserve_a_user_selected_native_executor_or_continue_to_mcp_or_direct_cloud/);
  assert.equal(scenarios.find((scenario) => scenario.id === "new-live-catalog-model")?.expected, "use_the_exact_catalog_model_with_its_declared_capability_without_waiting_for_a_skill_release");
  assert.equal(scenarios.find((scenario) => scenario.id === "workbuddy-explicit-host-model")?.expected, "preserve_explicit_host_model_choice_without_mcp_reroute_or_duplicate_submission");
  assert.match(scenarios.find((scenario) => scenario.id === "host-native-model-not-verified-for-media")?.expected || "", /do_not_treat_host_model_configuration_as_media_execution/);
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
  assert.equal(manifest.distribution.workbuddy.manualInstallationSupported, true);
  assert.equal(manifest.distribution.workbuddy.preservesExplicitHostModelChoice, true);
  const workBuddyAdapter = await readFile(path.join(skillRoot, "adapters", "workbuddy-execution.md"), "utf8");
  assert.match(workBuddyAdapter, /DeferExecuteTool/);
  assert.match(workBuddyAdapter, /never call `puretokens_image_result`/);
  assert.equal(manifest.distribution.codex.managedByDesktop, true);
  assert.equal(manifest.distribution.codex.manualInstallationSupported, true);
  assert.equal(manifest.distribution.codex.managedSkillDirectory, "~/.codex/skills/puretokens_media");
  assert.equal(manifest.distribution.codex.requiresPluginFeature, false);
  assert.equal(manifest.distribution.codex.supportsDirectCloud, true);
  assert.equal(manifest.distribution.codex.managedMcpWhenDesktopAvailable, true);
  assert.equal(manifest.rules.desktopManagedDeliveryIsOptional, true);
  assert.equal(manifest.rules.preservesExplicitHostModelChoice, true);
  assert.equal(manifest.rules.supportsVerifiedHostNativeMediaExecution, true);
  assert.equal(manifest.rules.hostNativeExecutionRequiresVerifiedMediaCapability, true);
  assert.equal(manifest.rules.hostNativeExecutionAvoidsDuplicateSubmission, true);
  assert.equal(manifest.supportedClients.includes("grok-build"), false);
  assert.match(skillText, /跨客户端共用的媒体行为源/);
  assert.match(skillText, /不得自行写入、替换或删除任何客户端的 Skill、MCP、Router 或 Secret 配置/);
});
