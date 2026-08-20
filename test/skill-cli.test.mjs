import assert from "node:assert/strict";
import { execFile } from "node:child_process";
import { mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { promisify } from "node:util";
import test from "node:test";
import { repositoryRoot } from "../scripts/skill-registry.mjs";
import { getMediaSkillProvenance } from "../scripts/media-skill-provenance.mjs";

const execFileAsync = promisify(execFile);
const cli = path.join(repositoryRoot, "bin", "puretokens-skill.js");

async function runCli(args) {
  return execFileAsync(process.execPath, [cli, ...args], { cwd: repositoryRoot });
}

test("Claude Desktop bundle includes the skill root and required files", async (t) => {
  const temporaryRoot = await mkdtemp(path.join(os.tmpdir(), "puretokens-skill-bundle-"));
  t.after(() => rm(temporaryRoot, { recursive: true, force: true }));
  const output = path.join(temporaryRoot, "puretokens_media.zip");
  await runCli(["bundle", "puretokens_media", "--format", "claude-desktop", "--out", output]);
  const bundle = await readFile(output);
  assert.equal(bundle.readUInt32LE(0), 0x04034b50);
  for (const file of ["SKILL.md", "skill.json", "source-delivery.json", "adapters/workbuddy-execution.md", "references/model-catalog-contract.md", "references/direct-cloud-contract.md", "references/behavior-scenarios.json", "references/natural-language-aliases.json"]) {
    assert.ok(bundle.includes(Buffer.from(`puretokens_media/${file}`)));
  }
  const provenance = await getMediaSkillProvenance();
  assert.ok(bundle.includes(Buffer.from(provenance.sourceSha256)));
});

test("install, upgrade, and explicit uninstall only manage the named Skill directory", async (t) => {
  const temporaryRoot = await mkdtemp(path.join(os.tmpdir(), "puretokens-skill-cli-"));
  t.after(() => rm(temporaryRoot, { recursive: true, force: true }));
  const skillDirectory = path.join(temporaryRoot, "puretokens_media");
  await runCli(["install", "puretokens_media", "--target", temporaryRoot]);
  const manifest = JSON.parse(await readFile(path.join(skillDirectory, "skill.json"), "utf8"));
  assert.equal(manifest.version, "0.4.1");
  await writeFile(path.join(skillDirectory, "SKILL.md"), "local modification\n");
  await runCli(["upgrade", "puretokens_media", "--target", temporaryRoot]);
  const upgraded = await readFile(path.join(skillDirectory, "SKILL.md"), "utf8");
  assert.match(upgraded, /Pure Tokens 媒体编排 Skill/);
  await runCli(["uninstall", "puretokens_media", "--target", temporaryRoot, "--yes"]);
  await assert.rejects(readFile(path.join(skillDirectory, "SKILL.md")));
});

test("agent installation instructions fail closed in ordinary ChatGPT chats", async () => {
  const [english, chinese] = await Promise.all([
    readFile(path.join(repositoryRoot, "README.md"), "utf8"),
    readFile(path.join(repositoryRoot, "README.zh-CN.md"), "utf8")
  ]);

  assert.match(english, /normal ChatGPT conversation/);
  assert.match(english, /Do not identify it as Codex merely because/);
  assert.match(chinese, /普通 ChatGPT 对话/);
  assert.match(chinese, /不能只因模型或运行时标签显示 Codex/);
});
