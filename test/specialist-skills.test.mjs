import assert from "node:assert/strict";
import { execFile } from "node:child_process";
import { mkdtemp, readFile, rm } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { promisify } from "node:util";
import { checkMediaCatalogFreshness } from "../scripts/check-media-catalog-freshness.mjs";
import { collectSkillRecords, repositoryRoot, validateRepository } from "../scripts/skill-registry.mjs";
import { sourceRows, buildPublishedCatalog } from "../scripts/sync-media-model-catalog-from-service.mjs";
import { buildSelection } from "../scripts/sync-skill-model-selection.mjs";

const names = ["puretokens_balance", "puretokens_connection", "puretokens_models", "puretokens_image", "puretokens_video"];

test("registry exposes exactly the five specialist Skills", async () => {
  const { registry, records } = await collectSkillRecords();
  assert.deepEqual(registry.skills.map((skill) => skill.name), names);
  assert.equal(records.length, 5);
  assert.deepEqual([...new Set(records.map((record) => record.manifest.version))], ["0.9.0"]);
  assert.deepEqual(await validateRepository(), []);
});

test("image and video policies use the current API without MCP fallback", async () => {
  const [balance, connection, models, image, video] = await Promise.all(names.map((name) => readFile(path.join(repositoryRoot, "skills", name, "SKILL.md"), "utf8")));
  assert.match(connection, /GET \/v1/);
  assert.match(connection, /不读取、扫描、展示或索取真实 Base URL、凭据、provider 标签或宿主配置/);
  assert.match(models, /GET \/v1\/media\/models/);
  assert.match(models, /不得调用 Images\/Videos 提交、任务状态、内容、余额或其他路径/);
  assert.match(models, /`input_schema\.constraints`/);
  assert.match(image, /POST \/v1\/images\/generations/);
  assert.match(image, /`200cm × 230cm`/);
  assert.match(video, /POST \/v1\/videos/);
  assert.match(video, /`grok-imagine-video-1.5-preview`/);
  const policy = balance + connection + models + image + video;
  assert.match(policy, /不检查或判断 Base URL、provider 标签、服务归属或凭据/);
  assert.doesNotMatch(policy, /MCP|Direct Cloud|krill|当前连接不是 Pure Tokens|无法确认归属|puretokensx\.com/i);
});

test("bilingual READMEs include safe copyable Agent installation prompts", async () => {
  const [english, chinese] = await Promise.all([
    readFile(path.join(repositoryRoot, "README.md"), "utf8"),
    readFile(path.join(repositoryRoot, "README.zh-CN.md"), "utf8")
  ]);
  assert.match(english, /^<p align="center">\n  <img src="\.\/assets\/brand\/puretokens-skill-hero\.png" alt="Pure Tokens Official Skills" width="100%" \/>\n<\/p>/);
  assert.match(english, /## Agent-assisted installation/);
  assert.match(english, /\$env:USERPROFILE\\\.agents\\skills/);
  assert.match(english, /Never overwrite an existing destination unless it is the managed Skill with that same name/);
  assert.match(english, /Do not read, display, copy, modify, or ask for any API Key, Base URL, authentication file, model configuration, or MCP configuration/);
  assert.match(chinese, /## 让 Agent 协助安装/);
  assert.match(chinese, /^<p align="center">\n  <img src="\.\/assets\/brand\/puretokens-skill-hero\.png" alt="Pure Tokens 官方 Skills" width="100%" \/>\n<\/p>/);
  assert.match(chinese, /\$env:USERPROFILE\\\.agents\\skills/);
  assert.match(chinese, /若已有目录不是同名受管 Skill，绝不覆盖，报告冲突/);
  assert.match(chinese, /不得读取、展示、复制、修改或索取任何 API Key、Base URL、认证文件、模型配置或 MCP 配置/);
});

test("specialist manifests use the configured connection without provider inspection", async () => {
  for (const name of names) {
    const manifest = JSON.parse(await readFile(path.join(repositoryRoot, "skills", name, "skill.json"), "utf8"));
    assert.equal(manifest.rules.usesCurrentConfiguredConnection, true);
    assert.equal(manifest.rules.usesRelativeApiPathsOnly, true);
    assert.equal(manifest.rules.doesNotInspectConnectionIdentity, true);
  }
});

test("installed model selections are generated from the published media catalog", async () => {
  const catalog = JSON.parse(await readFile(path.join(repositoryRoot, "references", "media-model-catalog.json"), "utf8"));
  for (const [skill, capability] of [["puretokens_image", "image"], ["puretokens_video", "video"]]) {
    const selection = JSON.parse(await readFile(path.join(repositoryRoot, "skills", skill, "references", "model-selection.json"), "utf8"));
    const expected = catalog.models
      .filter((model) => model.capabilities.includes(capability))
      .map((model) => ({ id: model.id, aliases: model.aliases, ...(model.parameterSchema ? { parameterSchema: model.parameterSchema } : {}) }));
    assert.equal(selection.schemaVersion, 1);
    assert.equal(selection.$schema, "https://puretokensx.com/schemas/model-selection.schema.json");
    assert.equal(selection.capability, capability);
    assert.equal(selection.catalogCapturedAt, catalog.serviceCatalog.capturedAt);
    assert.deepEqual(selection.models, expected);
  }
});

test("installed contracts cover bounded requests, task recovery, and user-facing failure guidance", async () => {
  const requiredScenarios = {
    puretokens_balance: ["balance-account-session-unavailable", "balance-response"],
    puretokens_connection: ["connection-identity-confirmed", "connection-identity-unconfirmed", "connection-identity-unavailable", "connection-base-url-request", "connection-identity-assurance"],
    puretokens_models: ["models-catalog-unavailable", "models-catalog-empty", "models-exact-id-unavailable", "models-capability-filter-empty", "models-parameter-profile-absent", "models-operation-profile-absent", "models-requirement-ambiguous", "models-catalog-response"],
    puretokens_image: ["image-model-alias-ambiguous", "image-model-unavailable", "image-model-parameter-profile-unavailable", "image-model-parameter-unsupported", "image-execution-unavailable", "image-count-invalid", "image-pixel-size-invalid", "image-physical-size", "image-edit-profile-unavailable", "image-public-url-reference-profile-unavailable", "image-public-url-reference-invalid", "image-public-url-reference-ambiguous", "image-edit-attachment-unavailable", "image-edit-input-unsupported", "image-task-pending", "image-task-poll-delay", "image-task-poll-resource-bound", "image-task-polling-deadline", "image-task-explicit-continuation", "image-task-terminal-failure", "image-task-timeout-or-unknown", "image-content-delivery-failure", "image-content-index-missing", "image-content-resource-bound"],
    puretokens_video: ["video-model-alias-ambiguous", "video-model-unavailable", "video-model-parameter-profile-unavailable", "video-model-parameter-unsupported", "video-execution-unavailable", "video-parameter-unsupported", "video-resolution-mode-unsupported", "video-image-operation-profile-unavailable", "video-media-operation-profile-unavailable", "video-image-transport-unavailable", "video-public-url-reference-profile-unavailable", "video-public-url-reference-invalid", "video-public-url-reference-ambiguous", "video-prompt-requirement", "video-image-count-unsupported", "video-media-combination-unsupported", "video-input-media-unsupported", "video-task-pending", "video-task-poll-delay", "video-task-poll-resource-bound", "video-task-polling-deadline", "video-task-explicit-continuation", "video-task-terminal-failure", "video-task-timeout-or-unknown", "video-content-delivery-failure", "video-content-resource-bound"]
  };
  for (const name of names) {
    const root = path.join(repositoryRoot, "skills", name, "references");
    const contract = JSON.parse(await readFile(path.join(root, "execution-contract.json"), "utf8"));
    const scenarios = JSON.parse(await readFile(path.join(root, "behavior-scenarios.json"), "utf8"));
    assert.equal(contract.schemaVersion, 1);
    assert.equal(contract.$schema, "https://puretokensx.com/schemas/media-execution-contract.schema.json");
    assert.equal(scenarios.schemaVersion, 1);
    assert.equal(scenarios.$schema, "https://puretokensx.com/schemas/media-behavior-scenarios.schema.json");
    assert.equal(contract.result.neverAutoResubmit ?? false, ["puretokens_image", "puretokens_video"].includes(name));
    const ids = new Set(scenarios.scenarios.map((scenario) => scenario.id));
    for (const id of requiredScenarios[name]) assert.ok(ids.has(id), `${name} missing ${id}`);
  }
  const image = JSON.parse(await readFile(path.join(repositoryRoot, "skills", "puretokens_image", "references", "execution-contract.json"), "utf8"));
  assert.equal(image.operations.submit.conditionalBodyFields.all_user_optional_fields, "only_when_the_exact_authenticated_live_input_schema_declares_the_field_and_value");
  assert.equal(image.operations.submit.fixedBody.async, true);
  assert.equal(image.operations.catalog.path, "/v1/media/models");
  assert.deepEqual(image.operations.edit.allowedPaths, ["/v1/images/generations", "/v1/images/edits"]);
  assert.equal(image.operations.edit.pathSource, "authenticated_live_profile_operation_path");
  assert.equal(image.operations.edit.contentType, "multipart/form-data");
  assert.equal(image.operations.edit.requiresAuthenticatedLiveProfileOperation, "image_edit");
  assert.deepEqual(image.operations.edit.requiredBodyFields, ["model", "prompt", "image", "async"]);
  assert.equal(image.inputMediaValidation.publicUrlReferenceRequires, "authenticated_live_catalog_property_and_declared_reference_transport");
  assert.equal(image.inputMediaValidation.nativeAttachmentRequires, "authenticated_live_catalog_input_schema_operation");
  assert.equal(image.inputMediaValidation.imageEditOperation, "image_edit");
  assert.equal(image.inputMediaValidation.onlyCurrentUserExplicitMedia, true);
  assert.equal(image.inputMediaValidation.skillPassesUserProvidedPublicUrlsWithoutDownloadingOrRehosting, true);
  assert.equal(image.inputMediaValidation.skillNeverDownloadsOrRehostsMedia, true);
  assert.equal(image.inputMediaValidation.gatewayStagesCurrentRequestAttachment, true);
  assert.equal(image.inputMediaValidation.requiresHostNativeAttachmentByteDelivery, true);
  assert.equal(image.result.sameTaskOnly, true);
  assert.equal(image.result.neverAutoResubmit, true);
  assert.deepEqual(image.transport.requiredHostCapabilities, ["authenticated_relative_http", "json_task_response", "native_media_byte_delivery", "same_task_continuation"]);
  assert.equal(image.parameterValidation.everyNewSubmissionReadsAuthenticatedLiveCatalog, true);
  assert.equal(image.parameterValidation.everySelectedModelRequiresExactLiveCatalogIdAndImageCapability, true);
  assert.equal(image.parameterValidation.allOptionalParametersRequire, "authenticated_live_catalog_input_schema");
  assert.equal(image.contentRetrieval.indexBase, 0);
  assert.equal(image.contentRetrieval.allowedIndexes, "0..requestedCount-1");
  assert.equal(image.contentRetrieval.completeDeliveryRequiresEveryRequestedIndex, true);
  assert.equal(image.contentRetrieval.fetchOnlyAfterTerminalSuccess, true);
  assert.equal(image.contentRetrieval.sequentialByIndex, true);
  assert.equal(image.contentRetrieval.oneInFlightContentReadPerTask, true);
  assert.equal(image.contentRetrieval.neverPrefetchOrRefetchDeliveredContent, true);
  assert.equal(image.contentRetrieval.handoffBeforeNextContentRead, true);
  assert.deepEqual(image.polling.fallbackDelaysSeconds, [3, 6, 12, 24, 30]);
  assert.equal(image.polling.serverDelay, "honor_valid_positive_retry_after_before_fallback");
  assert.equal(image.polling.serverDelayMustFitRemainingAutomaticDeadline, true);
  assert.equal(image.polling.steadyDelaySeconds, 30);
  assert.equal(image.polling.automaticDeadlineSeconds, 120);
  assert.equal(image.polling.maxAutomaticStatusReads, 6);
  assert.equal(image.polling.oneInFlightStatusReadPerTask, true);
  assert.equal(image.polling.automaticPollingScope, "submission_or_explicit_same_task_continuation_turn_only_no_background_timer_or_queue");
  assert.equal(image.polling.afterStatusReadError, "stop_automatic_polling_and_require_explicit_same_task_continuation");
  assert.equal(image.polling.explicitContinuation, "new_bounded_same_task_polling_window_only_when_explicitly_requested");
  assert.equal(image.polling.afterDeadline, "report_pending_and_require_explicit_same_task_continuation");
  const video = JSON.parse(await readFile(path.join(repositoryRoot, "skills", "puretokens_video", "references", "execution-contract.json"), "utf8"));
  assert.equal(video.operations.catalog.path, "/v1/media/models");
  assert.equal(video.operations.submit.path, "/v1/videos");
  assert.deepEqual(video.operations.submit.requiredBodyFields, ["model"]);
  assert.equal(video.operations.edit.path, "/v1/videos/edits");
  assert.equal(video.operations.edit.requiresAuthenticatedLiveProfileOperation, "video_edit");
  assert.equal(video.result.sameTaskOnly, true);
  assert.equal(video.result.neverAutoResubmit, true);
  assert.equal(video.parameterValidation.allModelsOptionalParametersRequire, "authenticated_live_catalog_input_schema");
  assert.equal(video.parameterValidation.promptRequirementUsesAuthenticatedLivePropertiesAndConstraints, true);
  assert.equal(video.parameterValidation.resolutionUsesAuthenticatedLiveModeConstraint, true);
  assert.equal(video.inputMediaValidation.imageToVideoOperation, "image_to_video");
  assert.equal(video.inputMediaValidation.referenceImageVideoOperation, "reference_image_video");
  assert.equal(video.inputMediaValidation.referenceVideoOperation, "reference_video");
  assert.equal(video.inputMediaValidation.referenceAudioOperation, "reference_audio");
  assert.equal(video.inputMediaValidation.videoEditOperation, "video_edit");
  assert.equal(video.inputMediaValidation.publicUrlOrDeclaredIdReferenceRequires, "authenticated_live_catalog_property_and_declared_reference_transport");
  assert.deepEqual(video.inputMediaValidation.allowedTransports, ["multipart_file"]);
  assert.equal(video.inputMediaValidation.onlyCurrentUserExplicitMedia, true);
  assert.equal(video.inputMediaValidation.skillPassesUserProvidedPublicUrlsOrDeclaredIdsWithoutDownloadingOrRehosting, true);
  assert.equal(video.inputMediaValidation.skillNeverDownloadsOrRehostsMedia, true);
  assert.equal(video.inputMediaValidation.gatewayStagesCurrentRequestAttachment, true);
  assert.equal(video.inputMediaValidation.multipleNativeAttachmentTypesRequireOneDeclaredCombinedOperation, true);
  assert.equal(video.inputMediaValidation.multiplePublicUrlOrIdFieldsRequireNoDeclaredConflict, true);
  assert.equal(video.contentRetrieval.fetchOnlyAfterTerminalSuccess, true);
  assert.equal(video.contentRetrieval.oneInFlightContentReadPerTask, true);
  assert.equal(video.contentRetrieval.neverPrefetchOrRefetchDeliveredContent, true);
  assert.equal(video.contentRetrieval.handoffBeforeNextContentRead, true);
  assert.deepEqual(video.polling.fallbackDelaysSeconds, [5, 10, 20, 40, 60]);
  assert.equal(video.polling.serverDelay, "honor_valid_positive_retry_after_before_fallback");
  assert.equal(video.polling.serverDelayMustFitRemainingAutomaticDeadline, true);
  assert.equal(video.polling.steadyDelaySeconds, 60);
  assert.equal(video.polling.automaticDeadlineSeconds, 300);
  assert.equal(video.polling.maxAutomaticStatusReads, 7);
  assert.equal(video.polling.oneInFlightStatusReadPerTask, true);
  assert.equal(video.polling.automaticPollingScope, "submission_or_explicit_same_task_continuation_turn_only_no_background_timer_or_queue");
  assert.equal(video.polling.afterStatusReadError, "stop_automatic_polling_and_require_explicit_same_task_continuation");
  assert.equal(video.polling.explicitContinuation, "new_bounded_same_task_polling_window_only_when_explicitly_requested");
  assert.equal(video.polling.afterDeadline, "report_pending_and_require_explicit_same_task_continuation");

  const balance = JSON.parse(await readFile(path.join(repositoryRoot, "skills", "puretokens_balance", "references", "execution-contract.json"), "utf8"));
  assert.equal(balance.operations.read.path, "/api/product/desktop/account/balance");
  assert.equal(balance.operations.read.requiresExistingAuthenticatedAccountSession, true);
  assert.equal(balance.operations.read.responseSchema, "https://puretokensx.com/schemas/balance-snapshot.schema.json");

  const connection = JSON.parse(await readFile(path.join(repositoryRoot, "skills", "puretokens_connection", "references", "execution-contract.json"), "utf8"));
  assert.equal(connection.kind, "connection");
  assert.equal(connection.operations.identity.path, "/v1");
  assert.deepEqual(connection.transport.requiredHostCapabilities, ["authenticated_relative_http", "json_task_response"]);
  assert.deepEqual(connection.result, {
    identitySource: "api_declared_public_health",
    expectedStatus: "ok",
    expectedName: "Pure Tokens API",
    expectedBasePath: "/v1",
    neverRetry: true,
    doesNotExposeHostConfiguration: true,
    unconfirmedDoesNotProveOtherService: true
  });

  const models = JSON.parse(await readFile(path.join(repositoryRoot, "skills", "puretokens_models", "references", "execution-contract.json"), "utf8"));
  assert.equal(models.kind, "models");
  assert.equal(models.operations.catalog.path, "/v1/media/models");
  assert.deepEqual(models.transport.requiredHostCapabilities, ["authenticated_relative_http", "json_task_response"]);
  assert.equal(models.result.reportOnlyAuthenticatedCatalog, true);
  assert.equal(models.result.doesNotSubmitMediaTasks, true);
  assert.equal(models.result.neverRetry, true);
  assert.equal(models.result.noStaticCatalogFallback, true);
  assert.equal(models.result.compatibilityShortlistsRequireDeclaredCapabilityAndInputSchema, true);
});

test("host matrix is explicit and matches every specialist manifest", async () => {
  const hostSupport = JSON.parse(await readFile(path.join(repositoryRoot, "references", "host-support.json"), "utf8"));
  assert.equal(hostSupport.$schema, "https://puretokensx.com/schemas/host-support.schema.json");
  const supported = hostSupport.supported.map((host) => host.id).sort();
  assert.deepEqual(supported, ["claude-code", "claude-desktop", "codex", "gemini-cli"]);
  assert.deepEqual(Object.fromEntries(hostSupport.supported.filter((host) => host.delivery === "manual-source").map((host) => [host.id, host.globalSkillDirectory])), {
    codex: "~/.agents/skills",
    "claude-code": "~/.claude/skills",
    "gemini-cli": "~/.gemini/skills"
  });
  assert.equal(hostSupport.nativeExecutionContract, "references/host-native-execution-contract.json");
  for (const host of hostSupport.supported) {
    assert.deepEqual(host.nativeExecution, {
      authenticatedRelativeHttp: true,
      jsonTaskResponse: true,
      nativeMediaByteDelivery: true,
      sameTaskContinuation: true
    });
  }
  const nativeContract = JSON.parse(await readFile(path.join(repositoryRoot, "references", "host-native-execution-contract.json"), "utf8"));
  assert.deepEqual(nativeContract.requiredCapabilities, ["authenticated_relative_http", "json_task_response", "native_media_byte_delivery", "same_task_continuation"]);
  assert.deepEqual(nativeContract.acceptanceScenarios.map((scenario) => scenario.id), ["connection-identity-read", "catalog-read", "media-submit", "same-task-status", "native-media-delivery"]);
  assert.equal(nativeContract.acceptanceScenarios.find((scenario) => scenario.id === "connection-identity-read").request, "GET /v1");
  assert.match(nativeContract.acceptanceScenarios.find((scenario) => scenario.id === "media-submit").request, /\/v1\/videos\/edits/);
  assert.deepEqual(nativeContract.userMediaInput, {
    onlyCurrentUserExplicitMedia: true,
    publicUrlOrDeclaredIdRequiresLiveProfilePropertyAndTransport: true,
    requiresProfileDeclaredTransport: true,
    requiresHostToExposeExactAttachmentRepresentation: true,
    skillNeverDownloadsOrRehosts: true,
    pureTokensGatewayStagesMultipartAttachments: true,
    fallback: "When the current host cannot deliver the user's attachment as a profile-declared transport, stop before submission, explain the declared transport, and do not invent a URL, file ID, or upload path. When it can deliver multipart_file bytes, send them only with the one Images/Videos API request; the Pure Tokens gateway owns short-lived R2 staging and never returns the provider-facing URL. A public HTTPS URL, file ID, or voice ID explicitly supplied by the user may be sent only in the exact live-profile field whose declared transport permits it; the Skill never downloads, validates, or rehosts it."
  });
  assert.deepEqual(nativeContract.mediaDeliveryResourceBounds, {
    oneInFlightStatusReadPerTask: true,
    noBackgroundPollingOrQueue: true,
    readContentOnlyAfterTerminalSuccess: true,
    oneInFlightContentReadPerTask: true,
    sequentialImageContentIndexes: true,
    neverPrefetchOrRefetchDeliveredContent: true,
    handoffNativeBytesBeforeNextRead: true,
    doesNotPersistMediaBytesInSkillStateOrLogs: true,
    fallback: "If the host cannot hand off a native media response without creating unbounded background work, cached copies, or duplicated reads, stop before content delivery and report that same-task native delivery is unavailable. Do not substitute a URL, HTML, SVG, status text, or a re-submitted task."
  });
  for (const name of names) {
    const manifest = JSON.parse(await readFile(path.join(repositoryRoot, "skills", name, "skill.json"), "utf8"));
    assert.deepEqual([...manifest.supportedClients].sort(), supported);
    assert.deepEqual([...manifest.excludedClients].sort(), ["grok-build", "opencode", "trae", "workbuddy"]);
    assert.deepEqual(manifest.distribution.codex, {
      manualInstallationSupported: true,
      globalSkillDirectory: "~/.agents/skills",
      requiresPluginFeature: false
    });
    assert.deepEqual(manifest.distribution.claudeCode, {
      manualInstallationSupported: true,
      globalSkillDirectory: "~/.claude/skills"
    });
    assert.deepEqual(manifest.distribution.geminiCli, {
      manualInstallationSupported: true,
      globalSkillDirectory: "~/.gemini/skills"
    });
  }
  const imageManifest = JSON.parse(await readFile(path.join(repositoryRoot, "skills", "puretokens_image", "skill.json"), "utf8"));
  assert.equal(imageManifest.rules.profiledPublicUrlReferenceSupported, true);
  assert.equal(imageManifest.rules.profiledImageEditSupported, true);
  assert.equal(imageManifest.rules.usesResourceBoundedPolling, true);
  assert.equal(imageManifest.rules.doesNotCreateBackgroundMediaWork, true);
  assert.equal(imageManifest.rules.usesSequentialNativeContentDelivery, true);
  assert.equal(Object.hasOwn(imageManifest.rules, "textToImageOnly"), false);
  const videoManifest = JSON.parse(await readFile(path.join(repositoryRoot, "skills", "puretokens_video", "skill.json"), "utf8"));
  assert.equal(videoManifest.rules.profiledPublicUrlOrDeclaredIdReferenceSupported, true);
  assert.equal(videoManifest.rules.profiledImageToVideoSupported, true);
  assert.equal(videoManifest.rules.profiledReferenceImageVideoSupported, true);
  assert.equal(videoManifest.rules.profiledReferenceVideoSupported, true);
  assert.equal(videoManifest.rules.profiledReferenceAudioSupported, true);
  assert.equal(videoManifest.rules.profiledVideoEditSupported, true);
  assert.equal(videoManifest.rules.modeSpecificResolutionRequiresLiveConstraint, true);
  assert.equal(videoManifest.rules.usesResourceBoundedPolling, true);
  assert.equal(videoManifest.rules.doesNotCreateBackgroundMediaWork, true);
  assert.equal(videoManifest.rules.usesSequentialNativeContentDelivery, true);
  assert.equal(Object.hasOwn(videoManifest.rules, "textToVideoOnly"), false);
});

test("formal reference schemas ship with the package", async () => {
  for (const schema of [
    "media-execution-contract.schema.json",
    "media-behavior-scenarios.schema.json",
    "model-selection.schema.json",
    "host-support.schema.json",
    "host-native-execution-contract.schema.json",
    "catalog-freshness.schema.json",
    "balance-snapshot.schema.json",
    "task-receipt.schema.json"
  ]) {
    const document = JSON.parse(await readFile(path.join(repositoryRoot, "schemas", schema), "utf8"));
    assert.equal(document.$schema, "https://json-schema.org/draft/2020-12/schema");
    assert.match(document.$id, /^https:\/\/puretokensx\.com\/schemas\//);
  }
});

test("CLI installs each specialist Skill", async (t) => {
  const target = await mkdtemp(path.join(os.tmpdir(), "puretokens-skill-cli-"));
  t.after(() => rm(target, { recursive: true, force: true }));
  const run = promisify(execFile);
  for (const name of names) {
    await run(process.execPath, [path.join(repositoryRoot, "bin", "puretokens-skill.js"), "install", name, "--target", target], { cwd: repositoryRoot });
    assert.equal(JSON.parse(await readFile(path.join(target, name, "skill.json"), "utf8")).name, name);
    if (["puretokens_image", "puretokens_video"].includes(name)) {
      const manifest = JSON.parse(await readFile(path.join(target, name, "skill.json"), "utf8"));
      assert.equal(manifest.taskReceipt, "references/task-receipt.json");
      assert.equal(JSON.parse(await readFile(path.join(target, name, manifest.taskReceipt), "utf8")).kind, name.replace("puretokens_", ""));
    }
  }
});

test("task receipts always expose the same core task fields", async () => {
  const core = ["exact_model_id", "task_id", "returned_state", "requested_operation", "requested_count", "requested_size_or_parameters", "next_action"];
  for (const name of ["puretokens_image", "puretokens_video"]) {
    const receipt = JSON.parse(await readFile(path.join(repositoryRoot, "skills", name, "references", "task-receipt.json"), "utf8"));
    for (const phase of ["submission", "continuation", "completion", "failure"]) {
      for (const field of core) assert.ok(receipt[phase].requiredFields.includes(field), `${name} ${phase} missing ${field}`);
    }
    assert.ok(receipt.completion.requiredFields.includes("delivered_count"));
  }
});

test("catalog parameter profiles propagate into installed model selections", () => {
  const previous = {
    models: [{ id: "example-image", aliases: ["example"], goodFor: { en: "Image generation", zh: "图片生成" }, example: { en: "Example", zh: "示例" } }]
  };
  const rows = sourceRows({
    vendors: [{ id: 9, name: "Example vendor" }],
    data: [{ model_name: "example-image", vendor_id: 9, capabilities: ["image"], media_input_schema: { properties: { size: { enum: ["1024x1024"] } } } }]
  });
  const catalog = buildPublishedCatalog(previous, rows, { sourcePath: "/controlled/catalog", capturedAt: "2026-08-21T02:46:19.421Z" });
  const selection = buildSelection(catalog, "image");
  assert.deepEqual(selection.models, [{ id: "example-image", aliases: ["example"], parameterSchema: { properties: { size: { enum: ["1024x1024"] } } } }]);
});

test("catalog freshness gate is deterministic", async () => {
  const catalog = JSON.parse(await readFile(path.join(repositoryRoot, "references", "media-model-catalog.json"), "utf8"));
  const capturedAt = Date.parse(catalog.serviceCatalog.capturedAt);
  const passed = await checkMediaCatalogFreshness({ maxAgeDays: 7, now: capturedAt + 3 * 24 * 60 * 60 * 1000 });
  assert.equal(passed.maxAgeDays, 7);
  await assert.rejects(
    checkMediaCatalogFreshness({ maxAgeDays: 1, now: capturedAt + 3 * 24 * 60 * 60 * 1000 }),
    /days old/
  );
});
