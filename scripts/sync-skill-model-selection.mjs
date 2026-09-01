import { readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { repositoryRoot } from "./skill-registry.mjs";

const catalogPath = path.join(repositoryRoot, "references", "media-model-catalog.json");
const targets = [
  { capability: "image", skill: "puretokens-image" },
  { capability: "video", skill: "puretokens-video" }
];

function parseMode() {
  const write = process.argv.includes("--write");
  const check = process.argv.includes("--check");
  if (write === check) throw new Error("use exactly one of --write or --check");
  return { write, check };
}

export function buildSelection(catalog, capability) {
  if (catalog?.schemaVersion !== 2 || !Array.isArray(catalog.models)) {
    throw new Error("media-model-catalog.json must be a schemaVersion=2 model catalog");
  }
  return {
    $schema: "https://puretokensx.com/schemas/model-selection.schema.json",
    schemaVersion: 1,
    capability,
    catalogUpdatedAt: catalog.updatedAt,
    catalogCapturedAt: catalog.serviceCatalog?.capturedAt,
    models: catalog.models
      .filter((model) => model.capabilities?.includes(capability))
      .map((model) => ({
        id: model.id,
        aliases: model.aliases || [],
        ...(model.parameterSchema ? { parameterSchema: model.parameterSchema } : {})
      }))
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
    const next = serialized(buildSelection(catalog, target.capability));
    const output = path.join(repositoryRoot, "skills", target.skill, "references", "model-selection.json");
    let current = "";
    try {
      current = await readFile(output, "utf8");
    } catch (error) {
      if (error?.code !== "ENOENT") throw error;
    }
    if (current === next) continue;
    changed = true;
    if (write) await writeFile(output, next);
    else process.stderr.write(`${path.relative(repositoryRoot, output)}: model selection reference is out of sync; run npm run docs:sync-media-models\n`);
  }
  if (check && changed) process.exitCode = 1;
}

if (process.argv[1] === fileURLToPath(import.meta.url)) await main();
