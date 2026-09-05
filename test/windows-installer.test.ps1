$ErrorActionPreference = 'Stop'
$repository = Split-Path -Parent $PSScriptRoot
$root = Join-Path ([IO.Path]::GetTempPath()) ('pt-installer-test-' + [Guid]::NewGuid().ToString('N'))
$target = Join-Path $root 'skills'
$installer = Join-Path $repository 'runtime/puretokens-skill-install.ps1'
try {
  New-Item -ItemType Directory $root | Out-Null
  & $installer sync -Target $target
  $binary = Join-Path $target '.puretokens-executor/puretokens-api.exe'
  $expected = (Get-Content (Join-Path $repository 'package.json') -Raw | ConvertFrom-Json).version
  if ((& $binary --version) -ne $expected) { throw 'Executor version mismatch' }
  & $installer sync -Target $target
  $lock = [IO.File]::Open((Join-Path $target '.puretokens-install.lock'), 'OpenOrCreate', 'ReadWrite', 'None')
  $rejected = $false
  try { & $installer sync -Target $target } catch { $rejected = $true } finally { $lock.Dispose() }
  if (-not $rejected) { throw 'Concurrent installation was not rejected' }
  $image = Join-Path $target 'puretokens-image'
  $stage = Join-Path $target '.puretokens-skill-stage-fixture'
  New-Item -ItemType Directory (Join-Path $stage 'backup') -Force | Out-Null
  New-Item -ItemType File (Join-Path $stage 'transaction-v1') | Out-Null
  '[{"name":"puretokens-image","action":"replace"}]' | Set-Content (Join-Path $stage 'plan.json')
  Move-Item $image (Join-Path $stage 'backup/puretokens-image')
  & $installer sync -Target $target
  if (Test-Path $stage) { throw 'Interrupted transaction not recovered' }
  if (-not (Test-Path (Join-Path $image 'SKILL.md'))) { throw 'Skill missing after recovery' }
} finally { if (Test-Path $root) { Remove-Item $root -Recurse -Force } }
