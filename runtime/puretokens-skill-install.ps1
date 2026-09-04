[CmdletBinding()]
param(
  [Parameter(Position = 0, Mandatory = $true)]
  [ValidateSet("check", "sync")]
  [string]$Command,
  [Parameter(Mandatory = $false)]
  [string]$Target,
  [Parameter(Mandatory = $false)]
  [ValidateSet("claude-code", "codex", "workbuddy", "gemini-cli", "grok-build", "opencode", "trae")]
  [Alias("Host")]
  [string]$HostId,
  [Parameter(Mandatory = $false)]
  [string]$Source
)

$ErrorActionPreference = "Stop"
$currentSkills = @("puretokens-balance", "puretokens-connection", "puretokens-models", "puretokens-image", "puretokens-video", "puretokens-update")
$retiredSkills = @("puretokens_media", "puretokens_balance", "puretokens_connection", "puretokens_models", "puretokens_image", "puretokens_video", "puretokens_update", "puretokens_get_balance", "puretokens_get_model_price", "puretokens_workbuddy_router")

function Fail([string]$Message) { throw "Pure Tokens Skill installer: $Message" }

function Test-ManagedSkill([string]$Directory, [string]$Name) {
  $manifest = Join-Path $Directory "skill.json"
  if (-not (Test-Path -LiteralPath (Join-Path $Directory "SKILL.md") -PathType Leaf) -or -not (Test-Path -LiteralPath $manifest -PathType Leaf)) { return $false }
  try { return ((Get-Content -LiteralPath $manifest -Raw | ConvertFrom-Json).name -eq $Name) } catch { return $false }
}

function Test-LegacyNodeRuntime([string]$Directory) {
  $manifest = Join-Path $Directory "runtime.json"
  if (-not (Test-Path -LiteralPath (Join-Path $Directory "puretokens-direct-api.mjs") -PathType Leaf) -or -not (Test-Path -LiteralPath $manifest -PathType Leaf)) { return $false }
  try { return ((Get-Content -LiteralPath $manifest -Raw | ConvertFrom-Json).name -eq "puretokens-direct-api-runtime") } catch { return $false }
}

function Get-SourceVersion([string]$SourceRoot) {
  try {
    $version = (Get-Content -LiteralPath (Join-Path $SourceRoot "package.json") -Raw | ConvertFrom-Json).version
    if ($version -notmatch '^\d+\.\d+\.\d+$') { Fail "official source has an invalid Skill version" }
    return $version
  } catch { Fail "official source has an invalid Skill version" }
}

function Test-OfficialSource([string]$SourceRoot) {
  if (-not (Test-Path -LiteralPath (Join-Path $SourceRoot "README.md") -PathType Leaf)) { Fail "official source is missing README.md" }
  if (-not (Test-Path -LiteralPath (Join-Path $SourceRoot "package.json") -PathType Leaf)) { Fail "official source is missing package.json" }
  $null = Get-SourceVersion $SourceRoot
  if (-not (Test-Path -LiteralPath (Join-Path $SourceRoot "runtime\puretokens-skill-install.ps1") -PathType Leaf)) { Fail "official source is missing the Windows installer" }
  if (-not (Test-Path -LiteralPath (Join-Path $SourceRoot "runtime\puretokens-skill-install.sh") -PathType Leaf)) { Fail "official source is missing the macOS/Linux installer" }
  foreach ($name in $currentSkills) {
    if (-not (Test-ManagedSkill (Join-Path (Join-Path $SourceRoot "skills") $name) $name)) { Fail "official source has an invalid Skill: $name" }
  }
}

function Get-TargetForHost([string]$RequestedHost) {
  if ([string]::IsNullOrWhiteSpace($env:USERPROFILE)) { Fail "cannot resolve a host Skill directory because USERPROFILE is unavailable" }
  switch ($RequestedHost) {
    "claude-code" { return (Join-Path $env:USERPROFILE ".claude\skills") }
    "codex" { return (Join-Path $env:USERPROFILE ".agents\skills") }
    "workbuddy" { return (Join-Path $env:USERPROFILE ".workbuddy\skills") }
    "gemini-cli" { return (Join-Path $env:USERPROFILE ".gemini\skills") }
    "grok-build" { return (Join-Path $env:USERPROFILE ".grok\skills") }
    "opencode" { return (Join-Path $env:USERPROFILE ".config\opencode\skills") }
    "trae" { return (Join-Path $env:USERPROFILE ".trae\skills") }
    default { Fail "unsupported host: $RequestedHost" }
  }
}

function Restore-Target([string]$TargetRoot, [string]$StageRoot, [string[]]$Replaced, [string[]]$Created) {
  foreach ($name in $Replaced) {
    $destination = Join-Path $TargetRoot $name
    $backup = Join-Path (Join-Path $StageRoot "backup") $name
    if (-not (Test-Path -LiteralPath $backup)) { continue }
    if (Test-Path -LiteralPath $destination) { Remove-Item -LiteralPath $destination -Recurse -Force }
    Move-Item -LiteralPath $backup -Destination $destination -Force
  }
  foreach ($name in $Created) {
    $destination = Join-Path $TargetRoot $name
    if (Test-Path -LiteralPath $destination) { Remove-Item -LiteralPath $destination -Recurse -Force }
  }
}

function Remove-LegacyCodexPlugin([string]$TargetRoot) {
  $codexTarget = [System.IO.Path]::GetFullPath((Join-Path $env:USERPROFILE ".agents\skills")).TrimEnd('\\')
  if (-not [string]::Equals($TargetRoot.TrimEnd('\\'), $codexTarget, [System.StringComparison]::OrdinalIgnoreCase)) { return }
  $codex = Get-Command codex -ErrorAction SilentlyContinue
  if ($null -eq $codex) { Fail "cannot verify removal of legacy Codex plugin puretokens-media because the Codex CLI is unavailable; remove it in Codex Plugins, then run this installer again" }
  try {
    $pluginOutput = & $codex.Source plugin list --json 2>$null
    if ($LASTEXITCODE -ne 0) { throw "plugin list failed" }
    $plugins = ($pluginOutput -join "`n") | ConvertFrom-Json
    $legacyPlugin = @($plugins.installed | Where-Object { $_.name -eq "puretokens-media" })
  } catch { Fail "cannot inspect Codex Plugins; remove legacy puretokens-media in Codex Plugins, then run this installer again" }
  if ($legacyPlugin.Count -eq 0) { return }
  foreach ($plugin in $legacyPlugin) {
    $selector = if (-not [string]::IsNullOrWhiteSpace($plugin.pluginId)) { $plugin.pluginId } elseif (-not [string]::IsNullOrWhiteSpace($plugin.marketplaceName)) { "$($plugin.name)@$($plugin.marketplaceName)" } else { $plugin.name }
    & $codex.Source plugin remove $selector --json *> $null
    if ($LASTEXITCODE -ne 0) { Fail "could not remove legacy Codex plugin $selector; remove it in Codex Plugins, then run this installer again" }
  }
  try {
    $pluginOutput = & $codex.Source plugin list --json 2>$null
    if ($LASTEXITCODE -ne 0) { throw "plugin list failed" }
    $plugins = ($pluginOutput -join "`n") | ConvertFrom-Json
    $legacyPlugin = @($plugins.installed | Where-Object { $_.name -eq "puretokens-media" })
  } catch { Fail "could not verify removal of legacy Codex plugin puretokens-media" }
  if ($legacyPlugin.Count -ne 0) { Fail "legacy Codex plugin puretokens-media is still installed; remove it in Codex Plugins, then run this installer again" }
  Write-Output "Removed and verified legacy Codex plugin puretokens-media. Fully restart Codex before testing the new Skills."
}

$stageRoot = $null
try {
  if (-not [string]::IsNullOrWhiteSpace($Target) -and -not [string]::IsNullOrWhiteSpace($HostId)) { Fail "use either -Host or -Target, not both" }
  if ([string]::IsNullOrWhiteSpace($Target) -and [string]::IsNullOrWhiteSpace($HostId)) { Fail "-Host or -Target is required" }
  if (-not [string]::IsNullOrWhiteSpace($HostId)) { $Target = Get-TargetForHost $HostId }
  if (-not [System.IO.Path]::IsPathRooted($Target)) { Fail "-Target must be an absolute Skill directory" }

  if (-not [string]::IsNullOrWhiteSpace($Source)) {
    if (-not [System.IO.Path]::IsPathRooted($Source)) { Fail "-Source must be an absolute official source directory" }
    if (-not (Test-Path -LiteralPath $Source -PathType Container)) { Fail "-Source does not exist: $Source" }
    $sourceRoot = (Resolve-Path -LiteralPath $Source).Path
  } else {
    $bundledSourceRoot = Split-Path -Parent $PSScriptRoot
    if ((Test-Path -LiteralPath (Join-Path $bundledSourceRoot "README.md") -PathType Leaf) -and (Test-Path -LiteralPath (Join-Path $bundledSourceRoot "package.json") -PathType Leaf) -and (Test-Path -LiteralPath (Join-Path $bundledSourceRoot "skills") -PathType Container)) {
      $sourceRoot = $bundledSourceRoot
    } else { Fail "run this source-only sync script from a fresh official Pure Tokens Skills main checkout or pass -Source <absolute-checkout-directory>" }
  }
  Test-OfficialSource $sourceRoot
  if ($Command -eq "check") { Write-Output "Official Pure Tokens Skill source passed native static validation."; exit 0 }
  $releaseVersion = Get-SourceVersion $sourceRoot

  New-Item -ItemType Directory -Path $Target -Force | Out-Null
  $targetRoot = (Resolve-Path -LiteralPath $Target).Path
  foreach ($name in $retiredSkills) {
    $destinations = @((Join-Path $targetRoot $name))
    $destinations += @(Get-ChildItem -LiteralPath $targetRoot -Force | Where-Object { $_.Name -like ("." + $name + ".retired-*") } | ForEach-Object { $_.FullName })
    foreach ($destination in $destinations) { if ((Test-Path -LiteralPath $destination) -and -not (Test-ManagedSkill $destination $name)) { Fail "unmanaged retired Skill conflicts: $destination" } }
  }
  foreach ($name in $currentSkills) {
    $destination = Join-Path $targetRoot $name
    if ((Test-Path -LiteralPath $destination) -and -not (Test-ManagedSkill $destination $name)) { Fail "unmanaged Skill conflicts: $destination" }
  }

  $stageRoot = Join-Path $targetRoot (".puretokens-skill-stage-" + [Guid]::NewGuid().ToString("N"))
  New-Item -ItemType Directory -Path (Join-Path $stageRoot "backup") -Force | Out-Null
  foreach ($name in $currentSkills) { Copy-Item -LiteralPath (Join-Path (Join-Path $sourceRoot "skills") $name) -Destination $stageRoot -Recurse }

  Remove-LegacyCodexPlugin $targetRoot
  $replaced = @()
  $created = @()
  try {
    foreach ($name in $currentSkills) {
      $destination = Join-Path $targetRoot $name
      if (Test-Path -LiteralPath $destination) { Move-Item -LiteralPath $destination -Destination (Join-Path (Join-Path $stageRoot "backup") $name); $replaced += $name } else { $created += $name }
      Move-Item -LiteralPath (Join-Path $stageRoot $name) -Destination $destination
    }
  } catch { Restore-Target $targetRoot $stageRoot $replaced $created; throw }
  foreach ($name in $retiredSkills) {
    $destinations = @((Join-Path $targetRoot $name))
    $destinations += @(Get-ChildItem -LiteralPath $targetRoot -Force | Where-Object { $_.Name -like ("." + $name + ".retired-*") } | ForEach-Object { $_.FullName })
    foreach ($destination in $destinations) { if (Test-Path -LiteralPath $destination) { Remove-Item -LiteralPath $destination -Recurse -Force; Write-Output "Removed retired managed $name from $destination" } }
  }
  $legacyRuntime = Join-Path $targetRoot ".puretokens-runtime"
  if ((Test-Path -LiteralPath $legacyRuntime) -and (Test-LegacyNodeRuntime $legacyRuntime)) { Remove-Item -LiteralPath $legacyRuntime -Recurse -Force; Write-Output "Removed retired managed Node runtime from $legacyRuntime" }
  Write-Output "Pure Tokens Skills $releaseVersion synchronized at $targetRoot"
} finally {
  if ($null -ne $stageRoot -and (Test-Path -LiteralPath $stageRoot)) { Remove-Item -LiteralPath $stageRoot -Recurse -Force -ErrorAction SilentlyContinue }
}
