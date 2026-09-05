// Maintainer-only packaging. A platform release includes six Skills and one
// executable; no source checkout, other architectures or migration archives.
import { execFileSync } from "node:child_process";
import { createHash } from "node:crypto";
import { cp, mkdir, mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import path from "node:path";
import os from "node:os";
import { fileURLToPath } from "node:url";
const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const manifest = JSON.parse(await readFile(path.join(root, "runtime/executor/manifest.json"), "utf8"));
const destination = path.join(root, "dist", "releases");
await mkdir(destination, { recursive: true });
const files = {};
for (const [platform, artifact] of Object.entries(manifest.artifacts)) {
  const temp = await mkdtemp(path.join(os.tmpdir(), "pt-release-"));
  try {
    const bundle = path.join(temp, "puretokens-skill");
    await mkdir(path.join(bundle, "runtime/executor/bin"), { recursive: true });
    await cp(path.join(root, "skills"), path.join(bundle, "skills"), { recursive: true });
    for (const file of ["README.md", "package.json", "runtime/puretokens-skill-install.sh", "runtime/puretokens-skill-install.ps1"]) await cp(path.join(root, file), path.join(bundle, file));
    const binary = await readFile(path.join(root, "runtime/executor", artifact.path));
    if (createHash("sha256").update(binary).digest("hex") !== artifact.sha256) throw new Error(`Binary checksum mismatch: ${platform}`);
    await cp(path.join(root, "runtime/executor", artifact.path), path.join(bundle, "runtime/executor", artifact.path));
    await writeFile(path.join(bundle, "runtime/executor/manifest.json"), JSON.stringify({ ...manifest, artifacts: { [platform]: artifact } }, null, 2));
    const name = `puretokens-skill-${manifest.version}-${platform}.zip`;
    execFileSync("zip", ["-qr", path.join(destination, name), "puretokens-skill"], { cwd: temp });
    const bytes = await readFile(path.join(destination, name));
    files[platform] = { filename: name, sha256: createHash("sha256").update(bytes).digest("hex"), bytes: bytes.length };
  } finally { await rm(temp, { recursive: true, force: true }); }
}
await writeFile(path.join(destination, "release-manifest.json"), `${JSON.stringify({ version: manifest.version, files }, null, 2)}\n`);
console.log(`Prepared ${Object.keys(files).length} platform-only releases; publish together with this version before changing the public install source.`);
