import { mkdir, readFile, readdir, rm, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { repositoryRoot } from "./skill-registry.mjs";

const catalogPath = path.join(repositoryRoot, "references", "media-model-catalog.json");
const targets = [
  { capability: "image", skill: "puretokens-image", defaultModel: "gpt-image-2" },
  { capability: "video", skill: "puretokens-video", defaultModel: "grok-imagine-video-1.5-preview" }
];

function parseMode() {
  const write = process.argv.includes("--write");
  const check = process.argv.includes("--check");
  if (write === check) throw new Error("use exactly one of --write or --check");
  return { write, check };
}

export function buildModelIndex(catalog, capability, defaultModel) {
  if (catalog?.schemaVersion !== 2 || !Array.isArray(catalog.models)) {
    throw new Error("media-model-catalog.json must be a schemaVersion=2 model catalog");
  }
  return {
    $schema: "https://puretokensx.com/schemas/model-index.schema.json",
    schemaVersion: 1,
    capability,
    catalogUpdatedAt: catalog.updatedAt,
    catalogCapturedAt: catalog.serviceCatalog?.capturedAt,
    defaultModel,
    models: catalog.models
      .filter((model) => model.capabilities?.includes(capability))
      .map((model) => ({
        id: model.id,
        aliases: model.aliases || [],
        profile: `profiles/${model.id}.json`
      }))
  };
}

export function buildModelProfile(catalog, capability, model) {
  return {
    $schema: "https://puretokensx.com/schemas/model-profile.schema.json",
    schemaVersion: 1,
    capability,
    catalogUpdatedAt: catalog.updatedAt,
    catalogCapturedAt: catalog.serviceCatalog?.capturedAt,
    id: model.id,
    parameterSchema: model.parameterSchema || {}
  };
}

function serialized(value) {
  return `${JSON.stringify(value, null, 2)}\n`;
}

async function main() {
  const { write, check } = parseMode();
  const catalog = JSON.parse(await readFile(catalogPath, "utf8"));
  let changed = false;
  for (const target of targets) {
    const referencesRoot = path.join(repositoryRoot, "skills", target.skill, "references");
    const indexPath = path.join(referencesRoot, "model-index.json");
    const profilesRoot = path.join(referencesRoot, "profiles");
    const models = catalog.models.filter((model) => model.capabilities?.includes(target.capability));
    if (!models.some((model) => model.id === target.defaultModel)) {
      throw new Error(`${target.skill}: default model ${target.defaultModel} is absent from the controlled ${target.capability} catalog`);
    }
    const expected = new Map([
      [indexPath, serialized(buildModelIndex(catalog, target.capability, target.defaultModel))],
      ...models.map((model) => [path.join(profilesRoot, `${model.id}.json`), serialized(buildModelProfile(catalog, target.capability, model))])
    ]);
    for (const [output, next] of expected) {
      let current = "";
      try {
        current = await readFile(output, "utf8");
      } catch (error) {
        if (error?.code !== "ENOENT") throw error;
      }
      if (current === next) continue;
      changed = true;
      if (write) {
        await mkdir(path.dirname(output), { recursive: true });
        await writeFile(output, next);
      } else {
        process.stderr.write(`${path.relative(repositoryRoot, output)}: model index/profile reference is out of sync; run npm run docs:sync-media-models\n`);
      }
    }
    const legacyPath = path.join(referencesRoot, "model-selection.json");
    try {
      await readFile(legacyPath, "utf8");
      changed = true;
      if (write) await rm(legacyPath);
      else process.stderr.write(`${path.relative(repositoryRoot, legacyPath)}: retired monolithic selection must be removed; run npm run docs:sync-media-models\n`);
    } catch (error) {
      if (error?.code !== "ENOENT") throw error;
    }
    let profileEntries = [];
    try {
      profileEntries = await readdir(profilesRoot, { withFileTypes: true });
    } catch (error) {
      if (error?.code !== "ENOENT") throw error;
    }
    for (const entry of profileEntries) {
      if (!entry.isFile() || !entry.name.endsWith(".json")) continue;
      const output = path.join(profilesRoot, entry.name);
      if (expected.has(output)) continue;
      changed = true;
      if (write) await rm(output);
      else process.stderr.write(`${path.relative(repositoryRoot, output)}: retired model profile must be removed; run npm run docs:sync-media-models\n`);
    }
  }
  if (check && changed) process.exitCode = 1;
}

if (process.argv[1] === fileURLToPath(import.meta.url)) await main();
