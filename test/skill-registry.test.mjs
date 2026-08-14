import assert from "node:assert/strict";
import test from "node:test";
import { collectSkillRecords, validateRepository } from "../scripts/skill-registry.mjs";

test("the skill registry has the shared media skill and WorkBuddy router", async () => {
  const { registry, records } = await collectSkillRecords();
  assert.equal(registry.schemaVersion, 1);
  assert.deepEqual(registry.skills.map((skill) => skill.name), [
    "puretokens_media",
    "puretokens_workbuddy_router"
  ]);
  assert.equal(records.length, 2);
  const media = records.find((record) => record.manifest.name === "puretokens_media");
  const workbuddyRouter = records.find((record) => record.manifest.name === "puretokens_workbuddy_router");
  assert.equal(registry.skills[0].version, media.manifest.version);
  assert.equal(media.frontmatter.name, "puretokens_media");
  assert.equal(workbuddyRouter.frontmatter.name, "puretokens_workbuddy_router");
  assert.equal(workbuddyRouter.frontmatter.alwaysApply, "true");
  assert.deepEqual(workbuddyRouter.manifest.supportedClients, ["workbuddy"]);
});

test("the repository passes its safety and registry validation", async () => {
  assert.deepEqual(await validateRepository(), []);
});
