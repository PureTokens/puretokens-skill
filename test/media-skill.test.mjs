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
  const [skillText, manifestText] = await Promise.all([
    readSkill(),
    readFile(path.join(skillRoot, "skill.json"), "utf8")
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
  assert.match(skillText, /第一步必须调用[：:][\s\S]*puretokens_list_media_models/);
  assert.match(skillText, /稳定的 `request_id`/);
  assert.match(skillText, /同一个 `task_id`/);
});

test("media Skill behavior scenarios cover ambiguity, empty catalog, unavailable MCP, failure, and timeout", async () => {
  const [skillText, scenariosText] = await Promise.all([
    readSkill(),
    readFile(path.join(skillRoot, "references", "behavior-scenarios.json"), "utf8")
  ]);
  const scenarios = JSON.parse(scenariosText).scenarios;
  assert.deepEqual(scenarios.map((scenario) => scenario.id), [
    "ambiguous-model",
    "no-media-model",
    "mcp-unavailable",
    "task-failure",
    "task-timeout"
  ]);
  assert.deepEqual(scenarios.slice(0, 3).map((scenario) => scenario.firstTool), [
    "puretokens_list_media_models",
    "puretokens_list_media_models",
    "puretokens_list_media_models"
  ]);
  assert.match(skillText, /模型不存在或匹配多个/);
  assert.match(skillText, /目录为空/);
  assert.match(skillText, /MCP 不可用/);
  assert.match(skillText, /safeToResubmit=false/);
  assert.match(skillText, /轮询超时/);
  assert.match(skillText, /不得自动换模型/);
  assert.match(skillText, /不得自动重新提交/);
});

test("Claude Desktop distribution remains an explicit upload and enablement flow", async () => {
  const [skillText, manifestText] = await Promise.all([
    readSkill(),
    readFile(path.join(skillRoot, "skill.json"), "utf8")
  ]);
  const manifest = JSON.parse(manifestText);
  assert.equal(manifest.distribution.claudeDesktop.format, "zip");
  assert.equal(manifest.distribution.claudeDesktop.archiveRoot, "puretokens_media");
  assert.equal(manifest.distribution.claudeDesktop.enableAfterImport, true);
  assert.match(skillText, /bundle puretokens_media --format claude-desktop/);
  assert.match(skillText, /上传 ZIP 并打开开关/);
  assert.match(skillText, /不能算 Claude Desktop 安装完成/);
});
