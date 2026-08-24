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

const names = ["puretokens_balance", "puretokens_image", "puretokens_video"];

test("registry exposes exactly the three specialist Skills", async () => {
  const { registry, records } = await collectSkillRecords();
  assert.deepEqual(registry.skills.map((skill) => skill.name), names);
  assert.equal(records.length, 3);
  assert.deepEqual([...new Set(records.map((record) => record.manifest.version))], ["0.8.1"]);
  assert.deepEqual(await validateRepository(), []);
});

test("image and video policies use the current API without MCP fallback", async () => {
  const [balance, image, video] = await Promise.all(names.map((name) => readFile(path.join(repositoryRoot, "skills", name, "SKILL.md"), "utf8")));
  assert.match(image, /POST \/v1\/images\/generations/);
  assert.match(image, /`200cm × 230cm`/);
  assert.match(video, /POST \/v1\/videos/);
  assert.match(video, /`grok-imagine-video-1.5-preview`/);
  const policy = balance + image + video;
  assert.match(policy, /不检查或判断 Base URL、provider 标签、服务归属或凭据/);
  assert.doesNotMatch(policy, /MCP|Direct Cloud|krill|当前连接不是 Pure Tokens|无法确认归属|puretokensx\.com/i);
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
    puretokens_image: ["image-model-alias-ambiguous", "image-model-unavailable", "image-model-parameter-profile-unavailable", "image-model-parameter-unsupported", "image-execution-unavailable", "image-count-invalid", "image-pixel-size-invalid", "image-physical-size", "image-edit-input", "image-task-pending", "image-task-poll-delay", "image-task-polling-deadline", "image-task-terminal-failure", "image-task-timeout-or-unknown", "image-content-delivery-failure", "image-content-index-missing"],
    puretokens_video: ["video-model-alias-ambiguous", "video-model-unavailable", "video-model-parameter-profile-unavailable", "video-model-parameter-unsupported", "video-execution-unavailable", "video-parameter-unsupported", "video-input-media", "video-task-pending", "video-task-poll-delay", "video-task-polling-deadline", "video-task-terminal-failure", "video-task-timeout-or-unknown", "video-content-delivery-failure"]
  };
  for (const name of names) {
    const root = path.join(repositoryRoot, "skills", name, "references");
    const contract = JSON.parse(await readFile(path.join(root, "execution-contract.json"), "utf8"));
    const scenarios = JSON.parse(await readFile(path.join(root, "behavior-scenarios.json"), "utf8"));
    assert.equal(contract.schemaVersion, 1);
    assert.equal(contract.$schema, "https://puretokensx.com/schemas/media-execution-contract.schema.json");
    assert.equal(scenarios.schemaVersion, 1);
    assert.equal(scenarios.$schema, "https://puretokensx.com/schemas/media-behavior-scenarios.schema.json");
    assert.equal(contract.result.neverAutoResubmit ?? false, name === "puretokens_balance" ? false : true);
    const ids = new Set(scenarios.scenarios.map((scenario) => scenario.id));
    for (const id of requiredScenarios[name]) assert.ok(ids.has(id), `${name} missing ${id}`);
  }
  const image = JSON.parse(await readFile(path.join(repositoryRoot, "skills", "puretokens_image", "references", "execution-contract.json"), "utf8"));
  assert.deepEqual(image.operations.submit.conditionalBodyFields.n.integerRange, [1, 6]);
  assert.deepEqual(image.operations.submit.conditionalBodyFields.size.allowed, ["1024x1024", "1536x1024", "1024x1536"]);
  assert.equal(image.operations.submit.fixedBody.async, true);
  assert.equal(image.result.sameTaskOnly, true);
  assert.equal(image.result.neverAutoResubmit, true);
  assert.deepEqual(image.transport.requiredHostCapabilities, ["authenticated_relative_http", "json_task_response", "native_media_byte_delivery", "same_task_continuation"]);
  assert.equal(image.parameterValidation.nonDefaultModelOptionalParametersRequire, "authenticated_live_catalog_input_schema");
  assert.equal(image.contentRetrieval.indexBase, 0);
  assert.equal(image.contentRetrieval.allowedIndexes, "0..requestedCount-1");
  assert.equal(image.contentRetrieval.completeDeliveryRequiresEveryRequestedIndex, true);
  assert.deepEqual(image.polling.fallbackDelaysSeconds, [2, 3, 5, 8]);
  assert.equal(image.polling.serverDelay, "honor_valid_retry_after_before_fallback");
  assert.equal(image.polling.steadyDelaySeconds, 15);
  assert.equal(image.polling.automaticDeadlineSeconds, 120);
  assert.equal(image.polling.afterDeadline, "report_pending_and_require_explicit_same_task_continuation");
  const video = JSON.parse(await readFile(path.join(repositoryRoot, "skills", "puretokens_video", "references", "execution-contract.json"), "utf8"));
  assert.equal(video.operations.catalog.path, "/v1/media/models");
  assert.equal(video.operations.submit.path, "/v1/videos");
  assert.equal(video.result.sameTaskOnly, true);
  assert.equal(video.result.neverAutoResubmit, true);
  assert.equal(video.parameterValidation.allModelsOptionalParametersRequire, "authenticated_live_catalog_input_schema");
  assert.deepEqual(video.polling.fallbackDelaysSeconds, [2, 3, 5, 8]);
  assert.equal(video.polling.serverDelay, "honor_valid_retry_after_before_fallback");
  assert.equal(video.polling.steadyDelaySeconds, 15);
  assert.equal(video.polling.automaticDeadlineSeconds, 300);
  assert.equal(video.polling.afterDeadline, "report_pending_and_require_explicit_same_task_continuation");

  const balance = JSON.parse(await readFile(path.join(repositoryRoot, "skills", "puretokens_balance", "references", "execution-contract.json"), "utf8"));
  assert.equal(balance.operations.read.path, "/api/product/desktop/account/balance");
  assert.equal(balance.operations.read.requiresExistingAuthenticatedAccountSession, true);
  assert.equal(balance.operations.read.responseSchema, "https://puretokensx.com/schemas/balance-snapshot.schema.json");
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
  assert.deepEqual(nativeContract.acceptanceScenarios.map((scenario) => scenario.id), ["catalog-read", "media-submit", "same-task-status", "native-media-delivery"]);
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
    if (name !== "puretokens_balance") {
      const manifest = JSON.parse(await readFile(path.join(target, name, "skill.json"), "utf8"));
      assert.equal(manifest.taskReceipt, "references/task-receipt.json");
      assert.equal(JSON.parse(await readFile(path.join(target, name, manifest.taskReceipt), "utf8")).kind, name.replace("puretokens_", ""));
    }
  }
});

test("task receipts always expose the same core task fields", async () => {
  const core = ["exact_model_id", "task_id", "returned_state", "requested_count", "requested_size_or_parameters", "next_action"];
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
    data: [{ model_name: "example-image", vendor_id: 9, capabilities: ["image"], input_schema: { properties: { size: { enum: ["1024x1024"] } } } }]
  });
  const catalog = buildPublishedCatalog(previous, rows, { sourcePath: "/controlled/catalog", capturedAt: "2026-08-21T02:46:19.421Z" });
  const selection = buildSelection(catalog, "image");
  assert.deepEqual(selection.models, [{ id: "example-image", aliases: ["example"], parameterSchema: { properties: { size: { enum: ["1024x1024"] } } } }]);
});

test("catalog freshness gate is deterministic", async () => {
  const passed = await checkMediaCatalogFreshness({ maxAgeDays: 7, now: Date.parse("2026-08-24T02:46:19.421Z") });
  assert.equal(passed.maxAgeDays, 7);
  await assert.rejects(
    checkMediaCatalogFreshness({ maxAgeDays: 1, now: Date.parse("2026-08-24T02:46:19.421Z") }),
    /days old/
  );
});
