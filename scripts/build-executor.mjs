// Maintainer build only. Users receive one native executable, never a Go toolchain.
import { execFileSync } from "node:child_process";
import { createHash } from "node:crypto";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const directory = path.join(root, "runtime", "executor");
const { version } = JSON.parse(await readFile(path.join(root, "package.json"), "utf8"));
const artifacts = {};
await mkdir(path.join(directory, "bin"), { recursive: true });
for (const os of ["darwin", "linux", "windows"]) {
  for (const arch of ["amd64", "arm64"]) {
    const platform = `${os}-${arch}`;
    const relative = `bin/puretokens-api-${platform}${os === "windows" ? ".exe" : ""}`;
    execFileSync("go", ["build", "-trimpath", "-buildvcs=false", "-ldflags", `-s -w -X main.executorVersion=${version}`, "-o", path.join(directory, relative), "."], {
      cwd: directory, env: { ...process.env, CGO_ENABLED: "0", GOOS: os, GOARCH: arch }, stdio: "inherit"
    });
    artifacts[platform] = { path: relative, sha256: createHash("sha256").update(await readFile(path.join(directory, relative))).digest("hex") };
    console.log(`Built ${platform}`);
  }
}
await writeFile(path.join(directory, "manifest.json"), `${JSON.stringify({ schemaVersion: 1, name: "puretokens-api-executor", version, artifacts }, null, 2)}\n`);
