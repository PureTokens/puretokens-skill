[CmdletBinding()]
param(
  [Parameter(Position = 0, Mandatory = $true)]
  [ValidateSet("check", "sync")]
  [string]$Command,
  [Parameter(Mandatory = $true)]
  [string]$Target
)

$ErrorActionPreference = "Stop"
$archiveUrl = "https://github.com/PureTokens/puretokens-skill/archive/refs/heads/main.zip"
$currentSkills = @("puretokens-balance", "puretokens-connection", "puretokens-models", "puretokens-image", "puretokens-video", "puretokens-update")
$retiredSkills = @("puretokens_media", "puretokens_balance", "puretokens_connection", "puretokens_models", "puretokens_image", "puretokens_video", "puretokens_update", "puretokens_get_balance", "puretokens_get_model_price", "puretokens_workbuddy_router")

function Fail([string]$Message) {
  throw "Pure Tokens Skill installer: $Message"
}

function Test-ManagedSkill([string]$Directory, [string]$Name) {
  $manifest = Join-Path $Directory "skill.json"
  if (-not (Test-Path -LiteralPath (Join-Path $Directory "SKILL.md") -PathType Leaf) -or -not (Test-Path -LiteralPath $manifest -PathType Leaf)) { return $false }
  try { return ((Get-Content -LiteralPath $manifest -Raw | ConvertFrom-Json).name -eq $Name) } catch { return $false }
}

function Test-ManagedRuntime([string]$Directory) {
  $manifest = Join-Path $Directory "runtime.json"
  if (-not (Test-Path -LiteralPath (Join-Path $Directory "puretokens-direct-api.mjs") -PathType Leaf) -or -not (Test-Path -LiteralPath $manifest -PathType Leaf)) { return $false }
  try { return ((Get-Content -LiteralPath $manifest -Raw | ConvertFrom-Json).name -eq "puretokens-direct-api-runtime") } catch { return $false }
}

function Test-OfficialSource([string]$SourceRoot) {
  if (-not (Test-Path -LiteralPath (Join-Path $SourceRoot "README.md") -PathType Leaf)) { Fail "official source is missing README.md" }
  $runtime = Join-Path $SourceRoot "runtime"
  if (-not (Test-ManagedRuntime $runtime)) { Fail "official source has an invalid managed runtime" }
  if (-not (Test-Path -LiteralPath (Join-Path $runtime "puretokens-skill-install.ps1") -PathType Leaf)) { Fail "official source is missing the native installer" }
  if (-not (Test-Path -LiteralPath (Join-Path $runtime "puretokens-skill-install.sh") -PathType Leaf)) { Fail "official source is missing the macOS/Linux native installer" }
  foreach ($name in $currentSkills) {
    if (-not (Test-ManagedSkill (Join-Path (Join-Path $SourceRoot "skills") $name) $name)) { Fail "official source has an invalid Skill: $name" }
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

$workspace = Join-Path ([System.IO.Path]::GetTempPath()) ("puretokens-skill-" + [Guid]::NewGuid().ToString("N"))
try {
  New-Item -ItemType Directory -Path $workspace | Out-Null
  $archive = Join-Path $workspace "puretokens-skill-main.zip"
  Invoke-WebRequest -Uri $archiveUrl -OutFile $archive
  $unpacked = Join-Path $workspace "unpacked"
  Expand-Archive -LiteralPath $archive -DestinationPath $unpacked
  $sourceRoot = Join-Path $unpacked "puretokens-skill-main"
  if (-not (Test-Path -LiteralPath $sourceRoot -PathType Container)) { Fail "official source archive has an unexpected layout" }
  Test-OfficialSource $sourceRoot
  if ($Command -eq "check") {
    Write-Output "Official Pure Tokens Skill source passed native static validation."
    exit 0
  }

  if (-not [System.IO.Path]::IsPathRooted($Target)) { Fail "-Target must be an absolute Skill directory" }
  New-Item -ItemType Directory -Path $Target -Force | Out-Null
  $targetRoot = (Resolve-Path -LiteralPath $Target).Path
  foreach ($name in $retiredSkills) {
    $retiredDestinations = @((Join-Path $targetRoot $name))
    $retiredDestinations += @(Get-ChildItem -LiteralPath $targetRoot -Force | Where-Object { $_.Name -like ("." + $name + ".retired-*") } | ForEach-Object { $_.FullName })
    foreach ($destination in $retiredDestinations) {
      if ((Test-Path -LiteralPath $destination) -and -not (Test-ManagedSkill $destination $name)) { Fail "unmanaged retired Skill conflicts: $destination" }
    }
  }
  $runtimeDestination = Join-Path $targetRoot ".puretokens-runtime"
  if ((Test-Path -LiteralPath $runtimeDestination) -and -not (Test-ManagedRuntime $runtimeDestination)) { Fail "unmanaged Pure Tokens runtime conflicts: $runtimeDestination" }
  foreach ($name in $currentSkills) {
    $destination = Join-Path $targetRoot $name
    if ((Test-Path -LiteralPath $destination) -and -not (Test-ManagedSkill $destination $name)) { Fail "unmanaged Skill conflicts: $destination" }
  }

  $stageRoot = Join-Path $targetRoot (".puretokens-skill-stage-" + [Guid]::NewGuid().ToString("N"))
  New-Item -ItemType Directory -Path (Join-Path $stageRoot "backup") -Force | Out-Null
  Copy-Item -LiteralPath (Join-Path $sourceRoot "runtime") -Destination $stageRoot -Recurse
  Move-Item -LiteralPath (Join-Path $stageRoot "runtime") -Destination (Join-Path $stageRoot ".puretokens-runtime")
  foreach ($name in $currentSkills) { Copy-Item -LiteralPath (Join-Path (Join-Path $sourceRoot "skills") $name) -Destination $stageRoot -Recurse }
  $replaced = @()
  $created = @()
  try {
    if (Test-Path -LiteralPath $runtimeDestination) {
      Move-Item -LiteralPath $runtimeDestination -Destination (Join-Path (Join-Path $stageRoot "backup") ".puretokens-runtime")
      $replaced += ".puretokens-runtime"
    } else { $created += ".puretokens-runtime" }
    Move-Item -LiteralPath (Join-Path $stageRoot ".puretokens-runtime") -Destination $runtimeDestination
    foreach ($name in $currentSkills) {
      $destination = Join-Path $targetRoot $name
      if (Test-Path -LiteralPath $destination) {
        Move-Item -LiteralPath $destination -Destination (Join-Path (Join-Path $stageRoot "backup") $name)
        $replaced += $name
      } else { $created += $name }
      Move-Item -LiteralPath (Join-Path $stageRoot $name) -Destination $destination
    }
  } catch {
    Restore-Target $targetRoot $stageRoot $replaced $created
    throw
  }
  foreach ($name in $retiredSkills) {
    $retiredDestinations = @((Join-Path $targetRoot $name))
    $retiredDestinations += @(Get-ChildItem -LiteralPath $targetRoot -Force | Where-Object { $_.Name -like ("." + $name + ".retired-*") } | ForEach-Object { $_.FullName })
    foreach ($destination in $retiredDestinations) {
      if (-not (Test-Path -LiteralPath $destination)) { continue }
      Remove-Item -LiteralPath $destination -Recurse -Force
      Write-Output "Removed retired managed $name from $destination"
    }
  }
  Remove-Item -LiteralPath $stageRoot -Recurse -Force
  Write-Output "Installed or upgraded official Pure Tokens Skills at $targetRoot"
} finally {
  if (Test-Path -LiteralPath $workspace) { Remove-Item -LiteralPath $workspace -Recurse -Force }
}
