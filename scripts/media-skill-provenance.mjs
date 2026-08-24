import { createHash } from "node:crypto";
import { readFile, readdir } from "node:fs/promises";
import path from "node:path";
import { repositoryRoot } from "./skill-registry.mjs";

export const mediaSkillName = "puretokens_media";

export const mediaSkillSourceFiles = [
  "SKILL.md",
  "skill.json",
  "agents/openai.yaml",
  "adapters/workbuddy-execution.md",
  "references/behavior-scenarios.json",
  "references/balance.md",
  "references/direct-cloud-contract.md",
  "references/image.md",
  "references/model-catalog-contract.md",
  "references/natural-language-aliases.json",
  "references/video.md"
];

const sourceFilesBySkill = new Map([[mediaSkillName, mediaSkillSourceFiles]]);

export async function getSkillSourceFiles(skillName) {
  const explicitFiles = sourceFilesBySkill.get(skillName);
  if (explicitFiles) return explicitFiles;
  const files = [];
  await collectFiles(path.join(repositoryRoot, "skills", skillName), "", files);
  return files.sort();
}

export async function getSkillProvenance(skillName) {
  const sourceRoot = path.join(repositoryRoot, "skills", skillName);
  const sourceFiles = await getSkillSourceFiles(skillName);
  const files = await Promise.all(sourceFiles.map(async (relativePath) => ({
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

export async function getMediaSkillProvenance() {
  return getSkillProvenance(mediaSkillName);
}

async function collectFiles(root, relativePath, files) {
  for (const entry of await readdir(path.join(root, relativePath), { withFileTypes: true })) {
    const child = path.join(relativePath, entry.name);
    if (entry.isDirectory()) await collectFiles(root, child, files);
    else if (entry.isFile()) files.push(child);
    else throw new Error(`${root}: unsupported Skill source entry ${child}`);
  }
}
