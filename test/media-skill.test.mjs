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
  const files = [
    "SKILL.md",
    "references/balance.md",
    "references/image.md",
    "references/video.md",
    "references/direct-cloud-contract.md",
    "references/model-catalog-contract.md"
  ];
  return (await Promise.all(files.map((file) => readFile(path.join(skillRoot, file), "utf8")))).join("\n");
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
  assert.equal(manifest.rules.supportsHostExposedReadOnlyBalanceQuery, true);
  assert.equal(manifest.rules.balanceQueryDoesNotUseMediaMcp, true);
  assert.equal(manifest.rules.balanceQueryNeverEstimatesOrConvertsBalance, true);
  assert.equal(manifest.rules.balanceQueryRequiresHostExposedCapability, true);
  assert.equal(manifest.rules.balanceQueryFailureGuidesToCcSwitchUsageQuery, true);
  assert.equal(manifest.rules.supportsSkillDefinedImage2Api, true);
  assert.equal(manifest.rules.supportsSkillDefinedConnectionVideoApi, true);
  assert.equal(manifest.rules.supportsCatalogVerifiedConnectionImageApi, true);
  assert.equal(manifest.rules.connectionImageApiRequiresHostExecutionAndDelivery, true);
  assert.equal(manifest.rules.connectionImageFallbackIsDisclosedBeforeSubmission, true);
  assert.equal(manifest.rules.connectionVideoFallbackIsDisclosedBeforeSubmission, true);
  assert.equal(manifest.rules.specialCasesRequireUserFacingNextSteps, true);
  assert.equal(manifest.rules.unsupportedPixelCanvasesFailClosed, true);
  assert.equal(manifest.rules.imageCountIsOneThroughSix, true);
  assert.equal(manifest.rules.mediaEditingFailsClosedWithTextPromptGuidance, true);
  assert.equal(manifest.rules.supportsPureTokensOnly, true);
  assert.equal(manifest.rules.nonPureTokensConnectionFailsClosed, true);
  assert.equal(manifest.rules.physicalImageDimensionsFailClosed, true);
  assert.equal(manifest.rules.supportedImageCanvasesAreExplicit, true);
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
  assert.match(skillText, /这是一个路由 Skill/);
  assert.match(skillText, /`references\/balance\.md`/);
  assert.match(skillText, /`references\/image\.md`/);
  assert.match(skillText, /`references\/video\.md`/);
  assert.match(skillText, /稳定的 `request_id`/);
  assert.match(skillText, /精确 `model` 和稳定 `request_id`/);
  assert.match(skillText, /仅支持 \*\*Pure Tokens\*\*/);
  assert.match(skillText, /已认证、只读余额查询能力/);
  assert.match(skillText, /不调用五个媒体 MCP 工具/);
  assert.match(skillText, /供应商卡的“用量查询”/);
  assert.match(skillText, /构造余额 URL、请求头、请求体或响应格式/);
  assert.match(skillText, /https:\/\/puretokensx\.com\//);
  assert.match(skillText, /`200cm × 230cm`/);
  assert.match(skillText, /`1024x1024`、`1536x1024`、`1024x1536`/);
  assert.match(skillText, /`1K`、`2K`、`4K`/);
  assert.match(skillText, /不能传给 `size`/);
  assert.match(skillText, /POST https:\/\/api\.puretokensx\.com\/v1\/images\/generations/);
  assert.match(skillText, /POST https:\/\/api\.puretokensx\.com\/v1\/videos/);
  assert.match(skillText, /Codex 或 CC Switch/);
  assert.match(skillText, /可调用、已认证的 HTTPS Images API 执行器/);
  assert.match(skillText, /能返回或交付真实原生图片字节/);
  assert.match(skillText, /任何备用提交前必须说明/);
  assert.match(skillText, /没有同模型通道则停止/);
  assert.match(skillText, /`n` 只能是 `1` 至 `6` 的整数/);
  assert.match(skillText, /当前只支持文生图/);
  assert.match(skillText, /面向用户的特殊情况/);
  assert.match(skillText, /继续查询同一 `task_id`/);
  assert.match(skillText, /可调用、已认证的 HTTPS 视频执行器/);
  assert.match(skillText, /仅保存 API Key 或能读取目录不是视频执行能力/);
  assert.match(skillText, /Pure Tokens 原生执行器/);
  assert.match(skillText, /README 清单只用于发现/);
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
  assert.match(skillText, /默认只请求 `1` 张/);
  assert.match(skillText, /Direct Cloud 图片始终传 `async: true`/);
  assert.match(skillText, /真实本机交付能力/);
  assert.match(skillText, /synchronous `data\[\]\.b64_json`/);
  assert.match(skillText, /`gpt-image-2` 的 MCP 生成调用返回原生图片后即完成/);
  assert.match(skillText, /视频总是异步任务/);
  assert.match(skillText, /已由宿主 Secret\/环境机制注入 Pure Tokens 凭据/);
  assert.doesNotMatch(skillText, /GET \/v1\/models/);
  assert.doesNotMatch(skillText, /https?:\/\/api\.(?!puretokensx\.com)/);
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
    "host-exposed-balance-query",
    "cc-switch-balance-query-not-exposed-to-chat",
    "default-single-result",
    "explicit-multiple-results",
    "invalid-image-count",
    "direct-cloud-image-delivery",
    "direct-cloud-multiple-image-content",
    "ambiguous-model",
    "no-media-model",
    "new-live-catalog-model",
    "direct-cloud-no-media-model",
    "direct-cloud-missing-delivery-capability",
    "skill-defined-image2-api",
    "skill-defined-connection-image-api",
    "connection-image-api-not-executable",
    "skill-defined-connection-video-api",
    "connection-video-api-not-executable",
    "non-puretokens-provider",
    "physical-image-size-not-supported",
    "unsupported-image-canvas-or-resolution",
    "media-editing-not-supported",
    "catalog-selection-guidance",
    "workbuddy-explicit-host-model",
    "host-native-model-not-verified-for-media",
    "mcp-unavailable",
    "task-failure",
    "task-timeout",
    "content-delivery-failure"
  ]);
  assert.equal(scenarios.find((scenario) => scenario.id === "host-exposed-balance-query")?.expected, "call_the_host_exposed_balance_query_once_and_report_only_the_returned_balance_snapshot");
  assert.equal(scenarios.find((scenario) => scenario.id === "cc-switch-balance-query-not-exposed-to-chat")?.expected, "explain_that_the_current_puretokens_provider_card_can_show_balance_and_direct_the_user_to_its_usage_query_without_claiming_balance_is_unsupported");
  assert.equal(scenarios.find((scenario) => scenario.id === "ambiguous-model")?.firstTool, "puretokens_list_media_models");
  assert.match(skillText, /匹配到多个候选/);
  assert.match(skillText, /目录为空、模型不存在/);
  assert.match(skillText, /MCP 不可用时/);
  assert.match(skillText, /safeToResubmit=false/);
  assert.match(skillText, /轮询超时或仍在处理中/);
  assert.deepEqual(scenarios.find((scenario) => scenario.id === "mcp-unavailable")?.expected, "use_direct_cloud_when_available_or_report_missing_execution_capability");
  assert.equal(scenarios.find((scenario) => scenario.id === "direct-cloud-image-delivery")?.expected, "force_async_true_and_write_actual_image_bytes_before_reporting_success");
  assert.equal(scenarios.find((scenario) => scenario.id === "direct-cloud-multiple-image-content")?.expected, "retrieve_zero_based_content_indexes_in_declared_order");
  assert.equal(scenarios.find((scenario) => scenario.id === "direct-cloud-no-media-model")?.expected, "report_api_key_catalog_scope_without_desktop_group_instructions");
  assert.equal(scenarios.find((scenario) => scenario.id === "direct-cloud-missing-delivery-capability")?.expected, "report_missing_delivery_capability_without_submission");
  assert.equal(scenarios.find((scenario) => scenario.id === "invalid-image-count")?.expected, "stop_and_request_an_integer_count_from_one_through_six_without_split_submission");
  assert.equal(scenarios.find((scenario) => scenario.id === "skill-defined-image2-api")?.expected, "call_the_puretokens_connection_images_api_once_at_the_puretokens_user_endpoint_with_gpt_image_2_without_mcp_or_direct_cloud_fallback");
  assert.equal(scenarios.find((scenario) => scenario.id === "skill-defined-connection-image-api")?.expected, "read_the_active_connection_catalog_then_submit_one_exact_image_model_to_the_puretokens_user_images_endpoint_and_deliver_the_native_image_without_mcp_or_direct_cloud");
  assert.match(scenarios.find((scenario) => scenario.id === "connection-image-api-not-executable")?.expected || "", /before_any_fallback_submission_tell_the_user/);
  assert.deepEqual(scenarios.find((scenario) => scenario.id === "connection-image-api-not-executable")?.forbiddenTools, ["unverified_connection_image_submission", "silent_fallback", "automatic_model_switch", "credential_extraction", "automatic_resubmission"]);
  assert.equal(scenarios.find((scenario) => scenario.id === "skill-defined-connection-video-api")?.expected, "read_the_active_connection_catalog_then_submit_one_exact_video_model_to_the_puretokens_user_videos_endpoint_and_retrieve_the_same_task_content_without_mcp_or_direct_cloud");
  assert.match(scenarios.find((scenario) => scenario.id === "connection-video-api-not-executable")?.expected || "", /before_any_fallback_submission_tell_the_user/);
  assert.equal(scenarios.find((scenario) => scenario.id === "non-puretokens-provider")?.expected, "stop_without_submitting_and_tell_the_user_that_this_skill_supports_only_pure_tokens_with_https_puretokensx_com");
  assert.match(scenarios.find((scenario) => scenario.id === "physical-image-size-not-supported")?.expected || "", /report_the_exact_supported_pixel_canvases_1024x1024_1536x1024_1024x1536_and_optional_1K_2K_4K_resolution/);
  assert.equal(scenarios.find((scenario) => scenario.id === "unsupported-image-canvas-or-resolution")?.expected, "stop_and_list_supported_image_canvases_and_image_size_options_without_nearest_match");
  assert.equal(scenarios.find((scenario) => scenario.id === "media-editing-not-supported")?.expected, "explain_text_to_media_only_and_offer_text_prompt_generation_without_accessing_user_media");
  assert.equal(scenarios.find((scenario) => scenario.id === "catalog-selection-guidance")?.expected, "list_exact_candidates_and_give_route_specific_next_steps_without_auto_selection");
  assert.equal(scenarios.find((scenario) => scenario.id === "task-failure")?.expected, "report_safe_error_and_require_explicit_user_retry_without_fallback_or_resubmit");
  assert.equal(scenarios.find((scenario) => scenario.id === "task-timeout")?.expected, "report_pending_or_unknown_and_offer_same_task_check_without_resubmit");
  assert.equal(scenarios.find((scenario) => scenario.id === "content-delivery-failure")?.expected, "report_no_delivery_and_offer_same_task_result_check_or_explicit_new_task_without_success_claim");
  assert.equal(scenarios.find((scenario) => scenario.id === "new-live-catalog-model")?.expected, "use_the_exact_catalog_model_with_its_declared_capability_without_waiting_for_a_skill_release");
  assert.equal(scenarios.find((scenario) => scenario.id === "workbuddy-explicit-host-model")?.expected, "preserve_explicit_host_model_choice_without_mcp_reroute_or_duplicate_submission");
  assert.match(scenarios.find((scenario) => scenario.id === "host-native-model-not-verified-for-media")?.expected || "", /do_not_treat_host_model_configuration_as_media_execution/);
  assert.match(skillText, /不得自动换模型或重新提交/);
  assert.match(skillText, /下载、保存或结果读取失败时不能声称成功/);
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
  assert.match(skillText, /这是一个路由 Skill/);
  assert.match(skillText, /`references\/balance\.md`/);
  assert.match(skillText, /`references\/image\.md`/);
  assert.match(skillText, /`references\/video\.md`/);
});
