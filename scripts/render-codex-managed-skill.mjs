import { mkdir, readFile, rm, writeFile } from "node:fs/promises";
import path from "node:path";
import process from "node:process";
import { repositoryRoot } from "./skill-registry.mjs";
import { getSkillProvenance, getSkillSourceFiles, mediaSkillName } from "./media-skill-provenance.mjs";

const managedSkillRoot = path.join(repositoryRoot, "dist", "codex", mediaSkillName);
export const codexManagedSkillNames = [mediaSkillName];

export async function renderCodexManagedSkill(skillName = mediaSkillName) {
  const sourceRoot = path.join(repositoryRoot, "skills", skillName);
  const sourceFiles = await getSkillSourceFiles(skillName);
  const derivedFrom = await getSkillProvenance(skillName);
  const files = new Map();
  for (const relativePath of sourceFiles) {
    files.set(relativePath, await readFile(path.join(sourceRoot, relativePath)));
  }
  files.set(
    "source-delivery.json",
    Buffer.from(`${JSON.stringify({ delivery: "codex-managed-skill", managedBy: "Pure Tokens Desktop", derivedFrom }, null, 2)}\n`)
  );
  return { files, derivedFrom };
}

export async function writeCodexManagedSkill(outputRoot = managedSkillRoot) {
  const destination = path.resolve(outputRoot);
  const { files, derivedFrom } = await renderCodexManagedSkill();
  await rm(destination, { recursive: true, force: true });
  for (const [relativePath, contents] of files) {
    const target = path.join(destination, relativePath);
    await mkdir(path.dirname(target), { recursive: true });
    await writeFile(target, contents);
  }
  return derivedFrom;
}

export async function renderCodexManagedSkills(skillNames = codexManagedSkillNames) {
  return new Map(await Promise.all(skillNames.map(async (skillName) => [
    skillName,
    await renderCodexManagedSkill(skillName)
  ])));
}

if (import.meta.url === `file://${process.argv[1]}`) {
  const outputIndex = process.argv.indexOf("--out");
  const outputRoot = outputIndex >= 0 ? process.argv[outputIndex + 1] : undefined;
  if (outputIndex >= 0 && (!outputRoot || outputRoot.startsWith("--"))) {
    throw new Error("--out requires a directory");
  }
  const provenance = await writeCodexManagedSkill(outputRoot);
  process.stdout.write(`Rendered ${mediaSkillName} ${provenance.version} to ${path.resolve(outputRoot || managedSkillRoot)}\n`);
}
