$ErrorActionPreference = 'Stop'
$repository = Split-Path -Parent $PSScriptRoot
$root = Join-Path ([IO.Path]::GetTempPath()) ('pt-installer-test-' + [Guid]::NewGuid().ToString('N'))
$target = Join-Path $root 'skills'
$installer = Join-Path $repository 'runtime/puretokens-skill-install.ps1'
$savedEnvironment = @{}
foreach ($name in @('USERPROFILE', 'HOME', 'CODEX_HOME')) { $savedEnvironment[$name] = [Environment]::GetEnvironmentVariable($name, 'Process') }
try {
  New-Item -ItemType Directory $root | Out-Null
  $env:USERPROFILE = Join-Path $root 'home'
  $env:HOME = $env:USERPROFILE
  $env:CODEX_HOME = Join-Path $env:USERPROFILE '.codex'
  New-Item -ItemType Directory $env:USERPROFILE | Out-Null
  foreach ($engine in @('powershell.exe', 'pwsh')) {
    $command = Get-Command $engine -ErrorAction Stop
    $engineTarget = Join-Path $root $engine
    # Exercise the exact public -File entrance in both Windows PowerShell 5.1
    # and PowerShell 7. Bypass applies to this child process only.
    & $command.Source -NoProfile -ExecutionPolicy Bypass -File $installer sync -Target $engineTarget
    if ($LASTEXITCODE -ne 0) { throw "$engine advertised installer entry failed" }
    $capture = Join-Path $root "capture-$engine.ps1"
    $receipt = Join-Path $root "receipt-$engine.txt"
    @'
param($Installer, $Target, $Receipt)
& $Installer init -Target $Target | Set-Content -LiteralPath $Receipt -Encoding UTF8
'@ | Set-Content -LiteralPath $capture -Encoding ASCII
    & $command.Source -NoProfile -ExecutionPolicy Bypass -File $capture -Installer $installer -Target $engineTarget -Receipt $receipt
    $expectedHeading = '# Pure Tokens Skill ' + [char]0x4f7f + [char]0x7528 + [char]0x987b + [char]0x77e5
    if (-not (Get-Content -LiteralPath $receipt -Raw -Encoding UTF8).Contains($expectedHeading)) { throw "$engine did not preserve the UTF-8 usage guide" }
    & $command.Source -NoProfile -ExecutionPolicy Bypass -File (Join-Path $repository 'runtime/puretokens-skill-fetch.ps1') -?
    if ($LASTEXITCODE -ne 0) { throw "$engine download entry could not be parsed" }
  }
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
} finally {
  foreach ($name in $savedEnvironment.Keys) { [Environment]::SetEnvironmentVariable($name, $savedEnvironment[$name], 'Process') }
  if (Test-Path $root) { Remove-Item $root -Recurse -Force }
}
