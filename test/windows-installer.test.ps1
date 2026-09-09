$ErrorActionPreference = 'Stop'
$repository = Split-Path -Parent $PSScriptRoot
$root = Join-Path ([IO.Path]::GetTempPath()) ('pt-installer-test-' + [Guid]::NewGuid().ToString('N'))
$target = Join-Path $root 'skills'
$installer = Join-Path $repository 'runtime/puretokens-skill-install.ps1'
$savedEnvironment = @{}
foreach ($name in @('USERPROFILE', 'HOME', 'CODEX_HOME', 'APPDATA', 'DSH_HOME', 'CLAUDE_CONFIG_DIR', 'ZCODE_DATA_BASE_DIR')) { $savedEnvironment[$name] = [Environment]::GetEnvironmentVariable($name, 'Process') }
try {
  New-Item -ItemType Directory $root | Out-Null
  $env:USERPROFILE = Join-Path $root 'home'
  $env:HOME = $env:USERPROFILE
  $env:CODEX_HOME = Join-Path $env:USERPROFILE '.codex'
  New-Item -ItemType Directory $env:USERPROFILE | Out-Null
  foreach ($engine in @('powershell.exe', 'pwsh')) {
    $command = Get-Command $engine -ErrorAction Stop
    $env:APPDATA = Join-Path $root "Roaming with spaces"
    $env:DSH_HOME = ""
    $env:CLAUDE_CONFIG_DIR = ""
    $dshLocation = & $command.Source -NoProfile -ExecutionPolicy Bypass -File $installer locate -HostId dsh-desktop
    if ($LASTEXITCODE -ne 0 -or $dshLocation -ne (Join-Path $env:APPDATA "dsh-desktop\harness\skills")) { throw "$engine DSH default directory mismatch" }
    $claudeLocation = & $command.Source -NoProfile -ExecutionPolicy Bypass -File $installer locate -HostId claude-desktop
    if ($LASTEXITCODE -ne 0 -or $claudeLocation -ne (Join-Path $env:USERPROFILE ".claude\skills")) { throw "$engine Claude Desktop directory mismatch" }
    $env:DSH_HOME = Join-Path $root "Harness custom"
    $overrideLocation = & $command.Source -NoProfile -ExecutionPolicy Bypass -File $installer locate -HostId dsh-desktop
    if ($LASTEXITCODE -ne 0 -or $overrideLocation -ne (Join-Path $env:DSH_HOME "skills")) { throw "$engine DSH override mismatch" }
    $env:DSH_HOME = "relative"
    & $command.Source -NoProfile -ExecutionPolicy Bypass -File $installer locate -HostId dsh-desktop *> $null
    if ($LASTEXITCODE -eq 0) { throw "$engine accepted a relative Harness directory" }
    $env:DSH_HOME = ""
    $env:ZCODE_DATA_BASE_DIR = Join-Path $root "ZCode data"
    $zcodeLocation = & $command.Source -NoProfile -ExecutionPolicy Bypass -File $installer locate -HostId zcode
    if ($LASTEXITCODE -ne 0 -or $zcodeLocation -ne (Join-Path $env:ZCODE_DATA_BASE_DIR ".zcode\skills")) { throw "$engine ZCode override mismatch" }
    & $command.Source -NoProfile -ExecutionPolicy Bypass -File $installer sync -HostId zcode
    if ($LASTEXITCODE -ne 0) { throw "$engine ZCode installation failed" }
    $env:ZCODE_DATA_BASE_DIR = "relative"
    & $command.Source -NoProfile -ExecutionPolicy Bypass -File $installer locate -HostId zcode *> $null
    if ($LASTEXITCODE -eq 0) { throw "$engine accepted a relative ZCode directory" }
    $env:ZCODE_DATA_BASE_DIR = ""
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
    # The documented entry must preserve modifications under either engine.
    $personalFile = Join-Path $engineTarget "puretokens-image/personal.txt"
    "keep my work" | Set-Content -LiteralPath $personalFile -Encoding UTF8
    & $command.Source -NoProfile -ExecutionPolicy Bypass -File $installer sync -Target $engineTarget *> $null
    if ($LASTEXITCODE -eq 0 -or -not (Test-Path -LiteralPath $personalFile)) { throw "$engine overwrote personal files" }
    Remove-Item -LiteralPath $personalFile
    & $command.Source -NoProfile -ExecutionPolicy Bypass -File $installer sync -Target $engineTarget
    if ($LASTEXITCODE -ne 0) { throw "$engine could not update an unchanged managed installation" }
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
  $unmanaged = Join-Path $root "unmanaged/puretokens-image"
  New-Item -ItemType Directory -Path $unmanaged -Force | Out-Null
  "# Independent Skill" | Set-Content -LiteralPath (Join-Path $unmanaged "SKILL.md")
  '{"name":"puretokens-image"}' | Set-Content -LiteralPath (Join-Path $unmanaged "skill.json")
  $rejected = $false
  try { & $installer sync -Target (Split-Path -Parent $unmanaged) } catch { $rejected = $true }
  if (-not $rejected -or -not (Get-Content -LiteralPath (Join-Path $unmanaged "SKILL.md") -Raw).Contains("Independent")) { throw "unmanaged same-name Skill was replaced" }
} finally {
  foreach ($name in $savedEnvironment.Keys) { [Environment]::SetEnvironmentVariable($name, $savedEnvironment[$name], 'Process') }
  if (Test-Path $root) { Remove-Item $root -Recurse -Force }
}
