import assert from "node:assert/strict";
import { execFile } from "node:child_process";
import { mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { promisify } from "node:util";
import test from "node:test";
import { repositoryRoot } from "../scripts/skill-registry.mjs";
import { getMediaSkillProvenance, mediaSkillSourceFiles } from "../scripts/media-skill-provenance.mjs";

const execFileAsync = promisify(execFile);
const cli = path.join(repositoryRoot, "bin", "puretokens-skill.js");
const sourceRoot = path.join(repositoryRoot, "skills", "puretokens_media");

async function runCli(args) {
  return execFileAsync(process.execPath, [cli, ...args], { cwd: repositoryRoot });
}

test("Claude Desktop bundle includes the skill root and required files", async (t) => {
  const temporaryRoot = await mkdtemp(path.join(os.tmpdir(), "puretokens-skill-bundle-"));
  t.after(() => rm(temporaryRoot, { recursive: true, force: true }));
  const output = path.join(temporaryRoot, "puretokens_media.zip");
  await runCli(["bundle", "puretokens_media", "--format", "claude-desktop", "--out", output]);
  const entries = readStoredZipEntries(await readFile(output));
  assert.deepEqual([...entries.keys()].sort(), [
    "puretokens_media/SKILL.md",
    "puretokens_media/adapters/workbuddy-execution.md",
    "puretokens_media/references/behavior-scenarios.json",
    "puretokens_media/references/direct-cloud-contract.md",
    "puretokens_media/references/model-catalog-contract.md",
    "puretokens_media/references/natural-language-aliases.json",
    "puretokens_media/skill.json",
    "puretokens_media/source-delivery.json"
  ]);
  for (const relativePath of mediaSkillSourceFiles.filter((file) => file !== "agents/openai.yaml")) {
    assert.deepEqual(entries.get(`puretokens_media/${relativePath}`), await readFile(path.join(sourceRoot, relativePath)), relativePath);
  }
  const provenance = await getMediaSkillProvenance();
  assert.deepEqual(JSON.parse(entries.get("puretokens_media/source-delivery.json").toString("utf8")), {
    delivery: "claude-desktop",
    derivedFrom: provenance
  });
});

test("install, upgrade, and explicit uninstall only manage the named Skill directory", async (t) => {
  const temporaryRoot = await mkdtemp(path.join(os.tmpdir(), "puretokens-skill-cli-"));
  t.after(() => rm(temporaryRoot, { recursive: true, force: true }));
  const skillDirectory = path.join(temporaryRoot, "puretokens_media");
  await runCli(["install", "puretokens_media", "--target", temporaryRoot]);
  const manifest = JSON.parse(await readFile(path.join(skillDirectory, "skill.json"), "utf8"));
  assert.equal(manifest.version, (await getMediaSkillProvenance()).version);
  await writeFile(path.join(skillDirectory, "SKILL.md"), "local modification\n");
  await runCli(["upgrade", "puretokens_media", "--target", temporaryRoot]);
  const upgraded = await readFile(path.join(skillDirectory, "SKILL.md"), "utf8");
  assert.match(upgraded, /Pure Tokens 媒体编排 Skill/);
  await runCli(["uninstall", "puretokens_media", "--target", temporaryRoot, "--yes"]);
  await assert.rejects(readFile(path.join(skillDirectory, "SKILL.md")));
});

function readStoredZipEntries(bundle) {
  const entries = new Map();
  let offset = 0;
  while (offset + 30 <= bundle.length && bundle.readUInt32LE(offset) === 0x04034b50) {
    const compression = bundle.readUInt16LE(offset + 8);
    const compressedSize = bundle.readUInt32LE(offset + 18);
    const nameLength = bundle.readUInt16LE(offset + 26);
    const extraLength = bundle.readUInt16LE(offset + 28);
    assert.equal(compression, 0, "bundle entries must be stored without compression");
    const nameStart = offset + 30;
    const dataStart = nameStart + nameLength + extraLength;
    const dataEnd = dataStart + compressedSize;
    assert.ok(dataEnd <= bundle.length, "bundle entry must fit in the archive");
    entries.set(bundle.subarray(nameStart, nameStart + nameLength).toString("utf8"), bundle.subarray(dataStart, dataEnd));
    offset = dataEnd;
  }
  assert.ok(entries.size > 0, "bundle must contain local ZIP entries");
  return entries;
}

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
