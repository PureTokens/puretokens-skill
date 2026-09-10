import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { compileSchema } from "./support/schema-validator.mjs";
const read = async file => JSON.parse(await readFile(new URL(`../${file}`, import.meta.url), "utf8"));
test("continuation contracts require integrity, original operation and waiting context", async () => {
 const schema=await read("schemas/media-execution-contract.schema.json");
 const recordSchema=schema.properties.continuationRecord;
 const validate=compileSchema(recordSchema,[]);
 for(const kind of ["image","video"]){
  const contract=await read(`skills/puretokens-${kind}/references/execution-contract.json`);
  const record=contract.continuationRecord;
  assert.equal(validate(record).length,0);
  for(const field of ["downloadProofFields","reuseRequiresMatchingProof","validatesBeforeCreation","legacyRecordWithoutProof"]){
   const changed=structuredClone(record);delete changed[field];assert.ok(validate(changed).length,field);
  }
  for(const field of ["original_operation","download_proofs","retry_not_before","reconciliation_required","wait_windows_completed"]){
   const changed=structuredClone(record);changed.preserves=changed.preserves.filter(x=>x!==field);assert.ok(validate(changed).length,field);
  }
 }
});
test("shared desktop guidance cannot drift between installed Skills", async () => {
 const names=["balance","connection","image","models","update","video"];
 const docs=await Promise.all(names.map(name=>readFile(new URL(`../skills/puretokens-${name}/references/desktop-hosts.md`,import.meta.url),"utf8")));
 for(const doc of docs)assert.equal(doc,docs[0]);
 assert.match(docs[0],/connection presence alone is not a routing signal/);
});

test("retired user install routes stay absent from the current distribution", async () => {
 const manifest=await read("package.json");
 assert.equal(manifest.bin,undefined);
 assert.equal(manifest.scripts["dist:build-legacy-migration-archive"],undefined);
 for(const file of ["bin/puretokens-skill.js","dist/puretokens-skill-install.zip","dist/puretokens-skill-install-payload.zip","runtime/executor/installation-history.json"]){
  await assert.rejects(readFile(new URL(`../${file}`,import.meta.url)),{code:"ENOENT"});
 }
 const shell=await readFile(new URL("../runtime/puretokens-skill-install.sh",import.meta.url),"utf8");
 const powershell=await readFile(new URL("../runtime/puretokens-skill-install.ps1",import.meta.url),"utf8");
 assert.doesNotMatch(shell,/retired_skills|legacy_node_runtime|codex plugin/);
 assert.doesNotMatch(powershell,/retiredSkills|Test-LegacyNodeRuntime|Remove-LegacyCodexPlugin/);
 assert.match(shell,/init_target "\$target_root" "\$host"/);
 assert.match(powershell,/Invoke-Init \$targetRoot \$HostId/);
});
