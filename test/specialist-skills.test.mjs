import assert from "node:assert/strict";
import { execFile } from "node:child_process";
import { mkdtemp, readFile, rm } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { promisify } from "node:util";
import { writeCodexManagedSkills } from "../scripts/render-codex-managed-skill.mjs";
import { renderWorkBuddyMediaSkill } from "../scripts/render-workbuddy-media-skill.mjs";
import { collectSkillRecords, repositoryRoot, validateRepository } from "../scripts/skill-registry.mjs";

const names = ["puretokens_balance", "puretokens_image", "puretokens_video"];

test("registry exposes exactly the three specialist Skills", async () => {
  const { registry, records } = await collectSkillRecords();
  assert.deepEqual(registry.skills.map((skill) => skill.name), names);
  assert.equal(records.length, 3);
  assert.deepEqual(await validateRepository(), []);
});

test("image and video policies use the current API without MCP fallback", async () => {
  const [image, video] = await Promise.all(names.slice(1).map((name) => readFile(path.join(repositoryRoot, "skills", name, "SKILL.md"), "utf8")));
  assert.match(image, /POST \/v1\/images\/generations/);
  assert.match(image, /`200cm × 230cm`/);
  assert.match(video, /POST \/v1\/videos/);
  assert.match(video, /`grok-imagine-video-1.5-preview`/);
  assert.doesNotMatch(image + video, /MCP|Direct Cloud|krill/i);
});

test("Codex and WorkBuddy generated payloads contain no MCP route", async (t) => {
  const output = await mkdtemp(path.join(os.tmpdir(), "puretokens-codex-skills-"));
  t.after(() => rm(output, { recursive: true, force: true }));
  await writeCodexManagedSkills(output);
  for (const name of names) assert.equal(JSON.parse(await readFile(path.join(output, name, "skill.json"), "utf8")).mcp, undefined);
  const workbuddy = await renderWorkBuddyMediaSkill();
  assert.equal(workbuddy.manifest.mcp, undefined);
  assert.doesNotMatch(workbuddy.entry, /MCP/i);
});

test("CLI installs each specialist Skill", async (t) => {
  const target = await mkdtemp(path.join(os.tmpdir(), "puretokens-skill-cli-"));
  t.after(() => rm(target, { recursive: true, force: true }));
  const run = promisify(execFile);
  for (const name of names) {
    await run(process.execPath, [path.join(repositoryRoot, "bin", "puretokens-skill.js"), "install", name, "--target", target], { cwd: repositoryRoot });
    assert.equal(JSON.parse(await readFile(path.join(target, name, "skill.json"), "utf8")).name, name);
  }
});
