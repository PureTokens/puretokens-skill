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
    # Simulate a host deletion guard without bypassing it: successful sync and
    # init must remain observable, the stage is retained, and the lock released.
    $guardRunner = Join-Path $root "guard-$engine.ps1"
    @'
param($Installer, $Target)
$ErrorActionPreference = 'Stop'
function Remove-Item {
  [CmdletBinding(SupportsShouldProcess=$true)]
  param([string]$LiteralPath, [switch]$Recurse, [switch]$Force)
  if ($LiteralPath -like '*.puretokens-skill-stage-*') { throw 'synthetic host cleanup denied' }
  Microsoft.PowerShell.Management\Remove-Item @PSBoundParameters
}
& $Installer sync -Target $Target
'@ | Set-Content -LiteralPath $guardRunner -Encoding ASCII
    $guardTarget = Join-Path $root "guard-target-$engine"
    $guardOutput = @(& $command.Source -NoProfile -ExecutionPolicy Bypass -File $guardRunner -Installer $installer -Target $guardTarget)
    if ($LASTEXITCODE -ne 0) { throw "$engine cleanup denial failed the completed sync" }
    $guardText = $guardOutput -join "`n"
    if ($guardText -notmatch 'synchronized with the native API executor' -or $guardText -notmatch 'cleanup_status: pending' -or $guardText -notmatch 'connection check was deferred') { throw "$engine lost sync, cleanup or init status" }
    foreach ($name in @('puretokens-balance','puretokens-connection','puretokens-models','puretokens-image','puretokens-video','puretokens-update','.puretokens-executor')) {
      if (-not (Test-Path (Join-Path (Join-Path $guardTarget $name) '.puretokens-managed.json'))) { throw "$engine missing installed inventory" }
    }
    if (@(Get-ChildItem $guardTarget -Directory -Force -Filter '.puretokens-skill-stage-*').Count -ne 1) { throw "$engine did not retain denied cleanup stage" }
    $released = [IO.File]::Open((Join-Path $guardTarget '.puretokens-install.lock'), 'OpenOrCreate', 'ReadWrite', 'None')
    $released.Dispose()
    & $command.Source -NoProfile -ExecutionPolicy Bypass -File $installer sync -Target $guardTarget
    if ($LASTEXITCODE -ne 0 -or @(Get-ChildItem $guardTarget -Directory -Force -Filter '.puretokens-skill-stage-*').Count -ne 0) { throw "$engine could not recover retained completed stage" }
    $launchProbe = Join-Path $root "launcher-$engine.ps1"
    @'
param($Installer, $Target)
$ErrorActionPreference = 'Stop'
$tokens = $null; $parseErrors = $null
$ast = [System.Management.Automation.Language.Parser]::ParseFile($Installer, [ref]$tokens, [ref]$parseErrors)
if ($parseErrors.Count) { throw 'installer parse failed' }
$function = $ast.Find({ param($node) $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Invoke-NativeExecutor' }, $true)
Invoke-Expression $function.Extent.Text
function Fail([string]$Message) { throw $Message }
$binary = Join-Path $Target '.puretokens-executor/puretokens-api.exe'
$ok = Invoke-NativeExecutor $binary @('--version')
if ($ok.ExitCode -ne 0 -or $ok.Output.Trim() -notmatch '^\d+\.\d+\.\d+$') { throw 'launcher lost stdout or exit code' }
$bad = Invoke-NativeExecutor $binary @('unsupported-command')
if ($bad.ExitCode -eq 0) { throw 'launcher swallowed failure' }
$folder = (Join-Path $Target 'puretokens-image') + [IO.Path]::DirectorySeparatorChar
$inventory = Invoke-NativeExecutor $binary @('install-inventory', '--directory', $folder, '--name', 'puretokens-image')
if ($inventory.ExitCode -ne 0 -or ($inventory.Output | ConvertFrom-Json).name -ne 'puretokens-image') { throw 'launcher corrupted quoted path' }
'@ | Set-Content -LiteralPath $launchProbe -Encoding UTF8
    & $command.Source -NoProfile -ExecutionPolicy Bypass -File $launchProbe -Installer $installer -Target $engineTarget
    if ($LASTEXITCODE -ne 0) { throw "$engine native launcher regression" }
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
