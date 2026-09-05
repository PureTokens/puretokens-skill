import assert from "node:assert/strict";
import { execFile as execFileCallback } from "node:child_process";
import { mkdir, mkdtemp, readFile, readdir, rm, writeFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { promisify } from "node:util";

import { collectSkillRecords, repositoryRoot, validateRepository } from "../scripts/skill-registry.mjs";

const execFile = promisify(execFileCallback);
const skillNames = [
  "puretokens-balance",
  "puretokens-connection",
  "puretokens-models",
  "puretokens-image",
  "puretokens-video",
  "puretokens-update"
];
const apiSkillNames = skillNames.filter((name) => name !== "puretokens-update");
const apiOrigin = "https://api.puretokensx.com";
const executorRequestStart = "invoke_the_checksum_verified_managed_native_executor_with_the_current_host_id_and_fixed_request";
const executorRequestActualFailureOnly = "report_only_an_actual_executor_network_credential_adapter_or_attachment_byte_failure_and_never_open_a_browser_desktop_ui_or_computer_use";

function currentExecutorPlatform() {
  const platform = process.platform === "darwin" ? "darwin" : process.platform === "linux" ? "linux" : "windows";
  const architecture = process.arch === "x64" ? "amd64" : process.arch;
  return `${platform}-${architecture}`;
}

test("repository validates its managed native executor contract", async () => {
  assert.deepEqual(await validateRepository(), []);
});

test("all specialist Skills are versioned together and use the expected source files", async () => {
  const packageManifest = JSON.parse(await readFile(path.join(repositoryRoot, "package.json"), "utf8"));
  const { registry, records } = await collectSkillRecords();
  assert.equal(packageManifest.version, "0.16.0");
  assert.deepEqual(registry.skills.map((entry) => entry.name), skillNames);
  assert.equal(records.length, skillNames.length);
  for (const record of records) {
    assert.equal(record.manifest.version, packageManifest.version);
    assert.equal(record.frontmatter.name, record.manifest.name);
  }
});

test("API Skills use the checksum-verified native executor without a user runtime", async () => {
  for (const name of apiSkillNames) {
    const base = path.join(repositoryRoot, "skills", name);
    const [skill, manifest, contract] = await Promise.all([
      readFile(path.join(base, "SKILL.md"), "utf8"),
      readFile(path.join(base, "skill.json"), "utf8").then(JSON.parse),
      readFile(path.join(base, "references", "execution-contract.json"), "utf8").then(JSON.parse)
    ]);
    assert.match(skill, /https:\/\/api\.puretokensx\.com/);
    assert.equal(manifest.rules.usesManagedNativeExecutor, true);
    assert.equal(manifest.rules.usesCurrentHostConfiguredCredential, true);
    assert.equal(manifest.rules.resolvesCurrentHostConnectionCredentialInMemoryForFixedRequest, true);
    assert.equal(manifest.rules.doesNotUseNodeRuntime, true);
    assert.equal(manifest.rules.doesNotUseComputerUseOrUiAutomation, true);
    assert.equal(manifest.rules.doesNotInvokeOtherMediaSkillsAsApiFallback, true);
    assert.equal(contract.transport.fixedApiOrigin, apiOrigin);
    assert.equal(contract.transport.usesFullApiUrls, true);
    assert.equal(contract.transport.usesManagedNativeExecutor, true);
    assert.equal(contract.transport.usesCurrentHostConfiguredCredential, true);
    assert.equal(contract.transport.credentialSourceNeverBecomesRequestTarget, true);
    assert.equal(contract.transport.neverExposesCredentialsOrHostConfiguration, true);
    assert.equal(contract.transport.doesNotUseNodeRuntime, true);
    assert.equal(contract.transport.doesNotUseMcpOrFallbackTransport, true);
    assert.equal(contract.transport.doesNotUseComputerUseOrUiAutomation, true);
    assert.equal(contract.transport.doesNotInvokeOtherMediaSkillsAsApiFallback, true);
    assert.equal(contract.transport.executionStart, executorRequestStart);
    assert.equal(contract.transport.whenManagedNativeExecutorUnavailable, executorRequestActualFailureOnly);
    assert.match(skill, /Computer Use/);
    assert.match(skill, /\.puretokens-executor\/puretokens-api/);
    assert.doesNotMatch(skill, /puretokens-direct-api\.mjs|\.puretokens-runtime|--json-base64|--multipart-base64|Authorization: Bearer/);
  }
  const manifest = JSON.parse(await readFile(path.join(repositoryRoot, "runtime", "executor", "manifest.json"), "utf8"));
  assert.equal(manifest.name, "puretokens-api-executor");
  assert.match(manifest.artifacts[currentExecutorPlatform()].sha256, /^[a-f0-9]{64}$/);
});

test("media routing prioritizes Pure Tokens specialists before generic media Skills", async () => {
  const directContract = JSON.parse(await readFile(path.join(repositoryRoot, "references", "direct-api-execution-contract.json"), "utf8"));
  assert.deepEqual(directContract.mediaRouting, {
    imageAndVideoRequestsPreferPureTokensSpecialists: true,
    imageSkill: "puretokens-image",
    videoSkill: "puretokens-video",
    priority: "when_current_host_connection_uses_puretokens_select_the_matching_puretokens_specialist_before_generic_imagegen_imagen_or_video_skills",
    afterSpecialistSelection: "never_fall_back_to_a_generic_media_skill",
    connectionContext: "managed_native_executor_resolves_one_matching_current_connection_credential_privately_for_the_fixed_request",
    metadataLimitation: "skill_metadata_expresses_routing_priority_but_cannot_override_a_host_that_ignores_installed_skill_selection"
  });
  for (const [name, genericSkill] of [["puretokens-image", "imagegen"], ["puretokens-video", "通用视频"]]) {
    const base = path.join(repositoryRoot, "skills", name);
    const [skill, manifest, agentMetadata, scenarios] = await Promise.all([
      readFile(path.join(base, "SKILL.md"), "utf8"),
      readFile(path.join(base, "skill.json"), "utf8").then(JSON.parse),
      readFile(path.join(base, "agents", "openai.yaml"), "utf8"),
      readFile(path.join(base, "references", "behavior-scenarios.json"), "utf8").then(JSON.parse)
    ]);
    assert.match(skill, /路由优先级/);
    assert.match(skill, new RegExp(genericSkill));
    assert.match(manifest.description, /Primary/);
    assert.match(agentMetadata, /Primary/);
    assert.equal(scenarios.scenarios[0].id, `${name.replace("puretokens-", "")}-specialist-routing-priority`);
  }
});

test("image and video start with the installed executor and never fall back", async () => {
  const image = await readFile(path.join(repositoryRoot, "skills", "puretokens-image", "SKILL.md"), "utf8");
  const video = await readFile(path.join(repositoryRoot, "skills", "puretokens-video", "SKILL.md"), "utf8");
  for (const [kind, skill] of [["image", image], ["video", video]]) {
    assert.match(skill, /唯一 API 传输/);
    assert.match(skill, /必须调用安装的原生执行器/);
    assert.match(skill, /不得自行发 HTTP/);
    assert.doesNotMatch(skill, /~\/.codex\/auth\.json|Authorization: Bearer/);
    const scenarios = JSON.parse(await readFile(path.join(repositoryRoot, "skills", `puretokens-${kind}`, "references", "behavior-scenarios.json"), "utf8"));
    assert.equal(scenarios.scenarios.some((scenario) => scenario.id === `${kind}-executor-request-start`), true);
    assert.equal(scenarios.scenarios.some((scenario) => scenario.id === `${kind}-executor-validation-failure`), true);
  }
  assert.match(image, /https:\/\/api\.puretokensx\.com\/v1\/images\/edits/);
  assert.match(image, /`image` 字段/);
});

test("media contracts retain fixed endpoints, asynchronous same-task handling, and native delivery", async () => {
  const image = JSON.parse(await readFile(path.join(repositoryRoot, "skills", "puretokens-image", "references", "execution-contract.json"), "utf8"));
  const video = JSON.parse(await readFile(path.join(repositoryRoot, "skills", "puretokens-video", "references", "execution-contract.json"), "utf8"));
  assert.equal(image.operations.submit.url, `${apiOrigin}/v1/images/generations`);
  assert.equal(image.operations.status.urlTemplate, `${apiOrigin}/v1/images/{task_id}`);
  assert.equal(image.operations.content.urlTemplate, `${apiOrigin}/v1/images/{task_id}/content?index={index}`);
  assert.equal(image.parameterValidation.defaultModel, "gpt-image-2");
  assert.equal(image.contentRetrieval.indexBase, 0);
  assert.equal(image.contentRetrieval.completeDeliveryRequiresEveryRequestedIndex, true);
  assert.equal(image.submissionFailure.hostExecutionPolicyBlockedBeforeRequestStarts, "report_validation_no_api_request_or_task_id_and_require_host_session_with_external_network_permission");
  assert.equal(video.operations.submit.url, `${apiOrigin}/v1/videos`);
  assert.equal(video.operations.status.urlTemplate, `${apiOrigin}/v1/videos/{task_id}`);
  assert.equal(video.operations.content.urlTemplate, `${apiOrigin}/v1/videos/{task_id}/content`);
  assert.equal(video.inputMediaValidation.nativeAttachmentBodyMode, "executor_multipart_only");
  assert.equal(video.submissionFailure.hostExecutionPolicyBlockedBeforeRequestStarts, "report_validation_no_api_request_or_task_id_and_require_host_session_with_external_network_permission");
  for (const contract of [image, video]) {
    assert.equal(contract.result.sameTaskOnly, true);
    assert.equal(contract.result.neverAutoResubmit, true);
    assert.equal(contract.transport.requiresNativeMediaByteDelivery, true);
  }
});

test("media Skills load a compact index and only the selected model profile", async () => {
  for (const [name, capability, defaultModel] of [
    ["puretokens-image", "image", "gpt-image-2"],
    ["puretokens-video", "video", "grok-imagine-video-1.5-preview"]
  ]) {
    const base = path.join(repositoryRoot, "skills", name);
    const [skill, manifest, index] = await Promise.all([
      readFile(path.join(base, "SKILL.md"), "utf8"),
      readFile(path.join(base, "skill.json"), "utf8").then(JSON.parse),
      readFile(path.join(base, "references", "model-index.json"), "utf8").then(JSON.parse)
    ]);
    assert.equal(manifest.modelIndex, "references/model-index.json");
    assert.equal(manifest.rules.usesLazyModelProfileLoading, true);
    assert.equal(manifest.rules.loadsBehaviorScenariosOnDemandOnly, true);
    assert.equal(index.capability, capability);
    assert.equal(index.defaultModel, defaultModel);
    assert.equal(index.models.some((model) => model.id === defaultModel), true);
    for (const model of index.models) {
      const profile = JSON.parse(await readFile(path.join(base, "references", model.profile), "utf8"));
      assert.equal(profile.id, model.id);
      assert.equal(profile.capability, capability);
      assert.equal(typeof profile.parameterSchema, "object");
    }
    assert.match(skill, /不要在每个请求加载全部/);
    assert.match(skill, /普通核心 POST 可只依据本 Skill、索引和选中 profile 执行/);
    assert.match(skill, /只有出现对应异常.*才读 `references\/behavior-scenarios\.json`/);
    await assert.rejects(readFile(path.join(base, "references", "model-selection.json"), "utf8"));
  }
});

test("legacy migration archive carries the current source and verified executor", async (t) => {
  const archive = path.join(repositoryRoot, "dist", "puretokens-skill-install.zip");
  const packageManifest = JSON.parse(await readFile(path.join(repositoryRoot, "package.json"), "utf8"));
  const { stdout: listing } = await execFile("unzip", ["-Z1", archive], { cwd: repositoryRoot });
  const entries = new Set(listing.trim().split("\n"));
  const prefix = "puretokens-skill-main/";
  for (const required of [
    `${prefix}README.md`,
    `${prefix}package.json`,
    `${prefix}runtime/puretokens-skill-install.sh`,
    `${prefix}runtime/puretokens-skill-install.ps1`,
    `${prefix}runtime/executor/manifest.json`,
    `${prefix}skills/index.json`,
    `${prefix}skills/puretokens-image/references/model-index.json`,
    `${prefix}skills/puretokens-video/references/model-index.json`
  ]) assert.equal(entries.has(required), true, `missing ${required}`);
  const executorManifest = JSON.parse(await readFile(path.join(repositoryRoot, "runtime", "executor", "manifest.json"), "utf8"));
  assert.equal(entries.has(`${prefix}runtime/executor/${executorManifest.artifacts[currentExecutorPlatform()].path}`), true);
  for (const retired of [
    `${prefix}runtime/runtime.json`,
    `${prefix}runtime/puretokens-direct-api.mjs`,
    `${prefix}skills/puretokens-image/references/model-selection.json`,
    `${prefix}skills/puretokens-video/references/model-selection.json`
  ]) assert.equal(entries.has(retired), false, `retired archive entry ${retired}`);
  const { stdout: index } = await execFile("unzip", ["-p", archive, `${prefix}skills/index.json`], { cwd: repositoryRoot });
  assert.equal(JSON.parse(index).skills.every((skill) => skill.version === packageManifest.version), true);

  const unpacked = await mkdtemp(path.join(os.tmpdir(), "puretokens-skill-legacy-archive-"));
  const target = path.join(unpacked, "target");
  t.after(() => rm(unpacked, { recursive: true, force: true }));
  await execFile("unzip", ["-q", archive, "-d", unpacked], { cwd: repositoryRoot });
  const source = path.join(unpacked, "puretokens-skill-main");
  const installer = path.join(source, "runtime", "puretokens-skill-install.sh");
  const { stdout } = await execFile("sh", [installer, "sync", "--target", target, "--source", source], { cwd: source });
  assert.match(stdout, /Pure Tokens Skills 0\.16\.0 synchronized with the native API executor at/);
  await assert.rejects(readFile(path.join(target, ".puretokens-runtime", "runtime.json"), "utf8"));
  const executor = path.join(target, ".puretokens-executor", process.platform === "win32" ? "puretokens-api.exe" : "puretokens-api");
  const { stdout: version } = await execFile(executor, ["--version"]);
  assert.equal(version.trim(), packageManifest.version);
});

test("published legacy updater bridge remains marker-only and carries current Skills", async () => {
  const archive = path.join(repositoryRoot, "dist", "puretokens-skill-install-payload.zip");
  const packageManifest = JSON.parse(await readFile(path.join(repositoryRoot, "package.json"), "utf8"));
  const prefix = "puretokens-skill-main/";
  const { stdout: listing } = await execFile("unzip", ["-Z1", archive], { cwd: repositoryRoot });
  const entries = new Set(listing.trim().split("\n"));
  for (const required of [
    `${prefix}runtime/runtime.json`,
    `${prefix}runtime/puretokens-direct-api.mjs`,
    `${prefix}skills/puretokens-image/SKILL.md`,
    `${prefix}skills/puretokens-video/SKILL.md`
  ]) assert.equal(entries.has(required), true, `missing ${required}`);
  assert.equal(entries.has(`${prefix}runtime/executor/manifest.json`), true);
  const [{ stdout: runtime }, { stdout: imageManifest }, { stdout: marker }] = await Promise.all([
    execFile("unzip", ["-p", archive, `${prefix}runtime/runtime.json`], { cwd: repositoryRoot }),
    execFile("unzip", ["-p", archive, `${prefix}skills/puretokens-image/skill.json`], { cwd: repositoryRoot }),
    execFile("unzip", ["-p", archive, `${prefix}runtime/puretokens-direct-api.mjs`], { cwd: repositoryRoot })
  ]);
  assert.equal(JSON.parse(runtime).legacyBootstrapOnly, true);
  assert.equal(JSON.parse(imageManifest).version, packageManifest.version);
  assert.doesNotMatch(marker, /api\.puretokensx\.com|fetch\(|Authorization/);
});

test("the public install prompt remains extractable in both README files", async () => {
  const expected = "Install or update the official Pure Tokens Skills from https://github.com/PureTokens/puretokens-skill.";
  for (const file of ["README.md", "README.zh-CN.md"]) {
    const readme = await readFile(path.join(repositoryRoot, file), "utf8");
    const match = readme.match(/### (?:Copy this to a terminal-capable local agent|复制给可在本机执行命令的 Agent)\n\n```text\n([^\n]+)\n```/);
    assert.equal(match?.[1], expected);
  }
});

test("host matrix lists seven executor hosts and the documented verified credential adapters", async () => {
  const support = JSON.parse(await readFile(path.join(repositoryRoot, "references", "host-support.json"), "utf8"));
  assert.deepEqual(support.supported.map((host) => host.id), ["claude-code", "codex", "workbuddy", "gemini-cli", "grok-build", "opencode", "trae"]);
  for (const host of support.supported) {
    assert.equal(host.delivery, "manual-source");
    assert.equal(host.directMediaExecution, "managed-native-executor");
  }
  assert.deepEqual(support.supported.filter((host) => host.credentialAdapter === "fixture-tested").map((host) => host.id), ["claude-code", "codex", "workbuddy", "gemini-cli", "grok-build", "opencode"]);
});

test("source installer synchronizes Skills and exactly one native executor", async (t) => {
  const target = await mkdtemp(path.join(os.tmpdir(), "puretokens-skill-executor-"));
  t.after(() => rm(target, { recursive: true, force: true }));
  const installer = path.join(repositoryRoot, "runtime", "puretokens-skill-install.sh");
  const { stdout } = await execFile("sh", [installer, "sync", "--target", target, "--source", repositoryRoot], { cwd: repositoryRoot });
  assert.match(stdout, /Pure Tokens Skills 0\.16\.0 synchronized with the native API executor at/);
  assert.match(stdout, /Pure Tokens Skill init: host ID was not supplied, so the connection check was deferred/);
  assert.match(stdout, /Pure Tokens Skill 使用须知/);
  assert.deepEqual((await readdir(target)).filter((entry) => skillNames.includes(entry)).sort(), [...skillNames].sort());
  await assert.rejects(readFile(path.join(target, ".puretokens-runtime", "runtime.json"), "utf8"));
  const executor = path.join(target, ".puretokens-executor", process.platform === "win32" ? "puretokens-api.exe" : "puretokens-api");
  const { stdout: version } = await execFile(executor, ["--version"]);
  assert.equal(version.trim(), "0.16.0");
});

test("source installer exposes an explicit init command and usage guide", async (t) => {
  const target = await mkdtemp(path.join(os.tmpdir(), "puretokens-skill-init-"));
  t.after(() => rm(target, { recursive: true, force: true }));
  const installer = path.join(repositoryRoot, "runtime", "puretokens-skill-install.sh");
  await execFile("sh", [installer, "sync", "--target", target, "--source", repositoryRoot], { cwd: repositoryRoot });
  const { stdout } = await execFile("sh", [installer, "init", "--target", target], { cwd: repositoryRoot });
  assert.match(stdout, /Pure Tokens Skill init: host ID was not supplied, so the connection check was deferred/);
  assert.match(stdout, /生成一张日落时分的山间湖泊图片/);
  assert.match(stdout, /更新 Pure Tokens Skills/);
});

test("source installer removes only the verified retired Node runtime", async (t) => {
  const target = await mkdtemp(path.join(os.tmpdir(), "puretokens-skill-legacy-runtime-"));
  t.after(() => rm(target, { recursive: true, force: true }));
  const verified = path.join(target, ".puretokens-runtime");
  await mkdir(verified);
  await writeFile(path.join(verified, "runtime.json"), JSON.stringify({ name: "puretokens-direct-api-runtime" }));
  await writeFile(path.join(verified, "puretokens-direct-api.mjs"), "retired runtime\n");
  const installer = path.join(repositoryRoot, "runtime", "puretokens-skill-install.sh");
  const { stdout } = await execFile("sh", [installer, "sync", "--target", target, "--source", repositoryRoot], { cwd: repositoryRoot });
  assert.match(stdout, /Removed retired managed Node runtime/);
  await assert.rejects(readFile(path.join(verified, "runtime.json"), "utf8"));

  const unknown = path.join(target, ".puretokens-runtime");
  await mkdir(unknown);
  await writeFile(path.join(unknown, "runtime.json"), JSON.stringify({ name: "someone-elses-runtime" }));
  await execFile("sh", [installer, "sync", "--target", target, "--source", repositoryRoot], { cwd: repositoryRoot });
  assert.equal(JSON.parse(await readFile(path.join(unknown, "runtime.json"), "utf8")).name, "someone-elses-runtime");
});

test("catalog freshness gate remains present for releases", async () => {
  const policy = JSON.parse(await readFile(path.join(repositoryRoot, "references", "catalog-freshness.json"), "utf8"));
  assert.equal(policy.maxAgeDays, 7);
  assert.equal(policy.releaseCommand, "npm run release:validate");
});
