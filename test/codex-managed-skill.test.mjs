import assert from "node:assert/strict";
import { mkdtemp, readFile, rm, stat } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { renderCodexManagedSkill, writeCodexManagedSkill } from "../scripts/render-codex-managed-skill.mjs";
import { getMediaSkillProvenance, mediaSkillSourceFiles } from "../scripts/media-skill-provenance.mjs";
import { repositoryRoot } from "../scripts/skill-registry.mjs";

const sourceRoot = path.join(repositoryRoot, "skills", "puretokens_media");

test("Codex managed Skill is generated from the shared media source", async (t) => {
  const temporaryRoot = await mkdtemp(path.join(os.tmpdir(), "puretokens-codex-skill-"));
  t.after(() => rm(temporaryRoot, { recursive: true, force: true }));

  const [sourceManifestText, provenance] = await Promise.all([
    readFile(path.join(sourceRoot, "skill.json"), "utf8"),
    getMediaSkillProvenance()
  ]);
  const destination = path.join(temporaryRoot, "puretokens_media");
  await writeCodexManagedSkill(destination);
  const sourceManifest = JSON.parse(sourceManifestText);
  const delivery = JSON.parse(await readFile(path.join(destination, "source-delivery.json"), "utf8"));

  assert.equal(delivery.delivery, "codex-managed-skill");
  assert.equal(delivery.managedBy, "Pure Tokens Desktop");
  assert.deepEqual(delivery.derivedFrom, provenance);
  for (const relativePath of mediaSkillSourceFiles) {
    const [source, rendered] = await Promise.all([
      readFile(path.join(sourceRoot, relativePath)),
      readFile(path.join(destination, relativePath))
    ]);
    assert.deepEqual(rendered, source, relativePath);
  }
  assert.equal(JSON.parse(await readFile(path.join(destination, "skill.json"), "utf8")).version, sourceManifest.version);
});

test("Codex managed Skill renderer has no Plugin or MCP payload", async () => {
  const { files } = await renderCodexManagedSkill();
  assert.equal(files.has(".mcp.json"), false);
  assert.equal(files.has(".codex-plugin/plugin.json"), false);
  assert.equal([...files.keys()].some((file) => file.startsWith("plugins/")), false);
});

test("the repository has no retired Codex Plugin delivery", async () => {
  await assert.rejects(stat(path.join(repositoryRoot, "plugins", "puretokens-media")));
  await assert.rejects(stat(path.join(repositoryRoot, ".agents", "plugins", "marketplace.json")));
});
