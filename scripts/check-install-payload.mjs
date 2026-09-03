import { createHash } from "node:crypto";
import { execFile } from "node:child_process";
import { readdir, readFile } from "node:fs/promises";
import path from "node:path";
import { promisify } from "node:util";
import { repositoryRoot } from "./skill-registry.mjs";

const exec = promisify(execFile);
const archives = [
  path.join(repositoryRoot, "dist", "puretokens-skill-install.zip"),
  path.join(repositoryRoot, "dist", "puretokens-skill-install-payload.zip")
];
const payloadRoot = "puretokens-skill-main";
const expectedFiles = [
  "README.md",
  ...await filesUnder("runtime"),
  ...await filesUnder("skills")
].sort();

for (const archive of archives) {
  const { stdout } = await exec("unzip", ["-Z1", archive], { maxBuffer: 1024 * 1024 });
  const archivedFiles = stdout.split(/\r?\n/)
    .filter((entry) => entry.startsWith(`${payloadRoot}/`) && !entry.endsWith("/"))
    .map((entry) => entry.slice(`${payloadRoot}/`.length))
    .sort();

  if (JSON.stringify(archivedFiles) !== JSON.stringify(expectedFiles)) {
    throw new Error(`${path.basename(archive)} files do not exactly match README.md, runtime/, and skills/`);
  }

  for (const relativePath of expectedFiles) {
    const [source, payload] = await Promise.all([
      readFile(path.join(repositoryRoot, relativePath)),
      readArchiveFile(archive, `${payloadRoot}/${relativePath}`)
    ]);
    if (sha256(source) !== sha256(payload)) throw new Error(`${path.basename(archive)} differs from source: ${relativePath}`);
  }
}

async function filesUnder(relativeDirectory) {
  const directory = path.join(repositoryRoot, relativeDirectory);
  const entries = await readdir(directory, { withFileTypes: true });
  const files = [];
  for (const entry of entries) {
    const child = path.posix.join(relativeDirectory, entry.name);
    if (entry.isDirectory()) files.push(...await filesUnder(child));
    else if (entry.isFile()) files.push(child);
    else throw new Error(`install payload source contains unsupported entry: ${child}`);
  }
  return files;
}

async function readArchiveFile(archive, entry) {
  const { stdout } = await exec("unzip", ["-p", archive, entry], { encoding: "buffer", maxBuffer: 16 * 1024 * 1024 });
  return stdout;
}

function sha256(value) {
  return createHash("sha256").update(value).digest("hex");
}
