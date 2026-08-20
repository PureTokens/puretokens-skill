import { createHash } from "node:crypto";
import { readFile } from "node:fs/promises";
import path from "node:path";
import { repositoryRoot } from "./skill-registry.mjs";

export const mediaSkillName = "puretokens_media";

export const mediaSkillSourceFiles = [
  "SKILL.md",
  "skill.json",
  "agents/openai.yaml",
  "adapters/workbuddy-execution.md",
  "references/behavior-scenarios.json",
  "references/direct-cloud-contract.md",
  "references/model-catalog-contract.md",
  "references/natural-language-aliases.json"
];

export async function getMediaSkillProvenance() {
  const sourceRoot = path.join(repositoryRoot, "skills", mediaSkillName);
  const files = await Promise.all(mediaSkillSourceFiles.map(async (relativePath) => ({
    relativePath,
    contents: await readFile(path.join(sourceRoot, relativePath))
  })));
  const manifest = JSON.parse(files.find((file) => file.relativePath === "skill.json").contents.toString("utf8"));
  const hash = createHash("sha256");
  for (const file of files) {
    hash.update(file.relativePath);
    hash.update("\0");
    hash.update(file.contents);
    hash.update("\0");
  }
  return {
    name: manifest.name,
    version: manifest.version,
    sourceSha256: hash.digest("hex")
  };
}
