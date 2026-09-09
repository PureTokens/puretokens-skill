// Maintainer build only. Users receive one native executable, never a Go toolchain.
import { createHash } from "node:crypto";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import { buildRoot as root, executorInputs, checkToolchain, compileExecutor } from "./executor-build-proof.mjs";
const directory=path.join(root,"runtime/executor");
const inputs=await executorInputs(root);
const {version}=JSON.parse(await readFile(path.join(root,"package.json")));
checkToolchain(inputs.config,directory);
const artifacts={};
await mkdir(path.join(directory,"bin"),{recursive:true});
for(const platform of inputs.config.platforms){
  const relative=`bin/puretokens-api-${platform}${platform.startsWith("windows-")?".exe":""}`;
  compileExecutor(root,inputs,version,platform,path.join(directory,relative));
  artifacts[platform]={path:relative,sha256:createHash("sha256").update(await readFile(path.join(directory,relative))).digest("hex")};
  console.log(`Built ${platform}`);
}
if((await executorInputs(root)).sourceSha256!==inputs.sourceSha256)throw new Error("Inputs changed during build; rebuild before packaging");
await writeFile(path.join(directory,"manifest.json"),JSON.stringify({schemaVersion:1,name:"puretokens-api-executor",version,artifacts},null,2)+"\n");
await writeFile(path.join(directory,"build-proof.json"),JSON.stringify({schemaVersion:1,version,toolchain:inputs.config.toolchain,sourceSha256:inputs.sourceSha256,inputs:inputs.inputs,artifacts},null,2)+"\n");
