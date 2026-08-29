import { readFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { repositoryRoot } from "./skill-registry.mjs";

const policyPath = path.join(repositoryRoot, "references", "catalog-freshness.json");
const catalogPath = path.join(repositoryRoot, "references", "media-model-catalog.json");

function parseArgs() {
  const index = process.argv.indexOf("--max-age-days");
  const nowIndex = process.argv.indexOf("--now");
  const maxAgeDays = index < 0 ? undefined : Number(process.argv[index + 1]);
  if (index >= 0 && (!Number.isInteger(maxAgeDays) || maxAgeDays < 1 || maxAgeDays > 31)) {
    throw new Error("--max-age-days must be an integer from 1 through 31");
  }
  const now = nowIndex < 0 ? Date.now() : Date.parse(process.argv[nowIndex + 1] || "");
  if (Number.isNaN(now)) throw new Error("--now must be an ISO-8601 timestamp");
  return { maxAgeDays, now };
}

export async function checkMediaCatalogFreshness({ maxAgeDays: requestedMaxAgeDays, now } = {}) {
  const [policy, catalog] = await Promise.all([
    readFile(policyPath, "utf8").then(JSON.parse),
    readFile(catalogPath, "utf8").then(JSON.parse)
  ]);
  const maxAgeDays = requestedMaxAgeDays ?? policy?.maxAgeDays;
  if (policy?.schemaVersion !== 1 || !Number.isInteger(maxAgeDays)) {
    throw new Error("catalog freshness policy is invalid");
  }
  const capturedAt = Date.parse(catalog?.serviceCatalog?.capturedAt || "");
  if (Number.isNaN(capturedAt)) throw new Error("media model catalog has no valid capturedAt timestamp");
  const currentTime = now ?? Date.now();
  const ageMs = currentTime - capturedAt;
  if (ageMs < 0) throw new Error("media model catalog capturedAt is in the future");
  const maxAgeMs = maxAgeDays * 24 * 60 * 60 * 1000;
  if (ageMs > maxAgeMs) {
    throw new Error(`media model catalog is ${Math.floor(ageMs / 86400000)} days old; refresh it before release with npm run docs:sync-media-models-from-service`);
  }
  return { ageMs, maxAgeDays };
}

async function main() {
  const { maxAgeDays, now } = parseArgs();
  const result = await checkMediaCatalogFreshness({ maxAgeDays, now });
  process.stdout.write(`Media model catalog freshness passed (${Math.floor(result.ageMs / 3600000)} hours old; limit ${result.maxAgeDays} days).\n`);
}

if (process.argv[1] === fileURLToPath(import.meta.url)) await main();
