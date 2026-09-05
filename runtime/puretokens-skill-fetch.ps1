# Run with powershell.exe -NoProfile -ExecutionPolicy Bypass -File <this-file>.
# This process-scoped option does not change machine-wide execution policy.
[CmdletBinding()]
param(
  [Parameter(Position = 0, Mandatory = $true)]
  [ValidateSet("check-update", "install", "update")]
  [string]$Command,
  [Alias("Host")]
  [ValidateSet("claude-code", "codex", "workbuddy", "gemini-cli", "grok-build", "opencode", "trae")]
  [string]$HostId,
  [string]$Target
)
$ErrorActionPreference = "Stop"
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
function Fail([string]$Message) { throw "Pure Tokens Skill download: $Message" }
function Read-Json([string]$File) { return (Get-Content -LiteralPath $File -Raw -Encoding UTF8 | ConvertFrom-Json) }
function Get-OfficialFile([string]$Url, [string]$Destination) {
  try {
    $result = Invoke-WebRequest -UseBasicParsing -Uri $Url -OutFile $Destination -PassThru -TimeoutSec 180 -Headers @{ Accept = "application/vnd.github+json" } -UserAgent "puretokens-skill-installer"
    return [int]$result.StatusCode
  } catch {
    if ($null -ne $_.Exception.Response -and [int]$_.Exception.Response.StatusCode -eq 404) { return 404 }
    Fail "the official source download failed; installed files were not changed"
  }
}
$locationOptions = @{}
if ($HostId) { $locationOptions.HostId = $HostId }
if ($Target) { $locationOptions.Target = $Target }
if ($locationOptions.Count -eq 0) { Fail "-Host or -Target is required" }
$downloadRoot = Join-Path ([IO.Path]::GetTempPath()) ("puretokens-download-" + [Guid]::NewGuid().ToString("N"))
try {
  New-Item -ItemType Directory -Path $downloadRoot | Out-Null
  $commitFile = Join-Path $downloadRoot "commit.json"
  if ((Get-OfficialFile "https://api.github.com/repos/PureTokens/puretokens-skill/commits/main" $commitFile) -ne 200) { Fail "the official main revision could not be resolved" }
  $sourceCommit = (Read-Json $commitFile).sha
  if ($sourceCommit -cnotmatch '^[0-9a-f]{40}$') { Fail "the official revision is invalid" }
  $packageFile = Join-Path $downloadRoot "package.json"
  if ((Get-OfficialFile "https://raw.githubusercontent.com/PureTokens/puretokens-skill/$sourceCommit/package.json" $packageFile) -ne 200) { Fail "the pinned official version could not be read" }
  $availableVersion = (Read-Json $packageFile).version
  if ($availableVersion -notmatch '^\d+\.\d+\.\d+$') { Fail "the official version is invalid" }
  $locationInstaller = Join-Path $PSScriptRoot "puretokens-skill-install.ps1"
  if (-not (Test-Path -LiteralPath $locationInstaller -PathType Leaf) -or -not (Select-String -LiteralPath $locationInstaller -SimpleMatch -Pattern "# puretokens-locate-v1" -Quiet)) {
    $locationInstaller = Join-Path $downloadRoot "puretokens-skill-install.ps1"
    if ((Get-OfficialFile "https://raw.githubusercontent.com/PureTokens/puretokens-skill/$sourceCommit/runtime/puretokens-skill-install.ps1" $locationInstaller) -ne 200) { Fail "the pinned directory selector is unavailable" }
  }
  $Target = & $locationInstaller locate @locationOptions
  $installedVersion = "not_installed"
  $installedFile = Join-Path $Target ".puretokens-executor/runtime.json"
  if (Test-Path -LiteralPath $installedFile -PathType Leaf) {
    try {
      $installed = Read-Json $installedFile
      if ($installed.name -eq "puretokens-api-executor") {
        $installedVersion = if ($installed.version -match '^\d+\.\d+\.\d+$') { $installed.version } else { "unverified" }
      }
    } catch { $installedVersion = "unverified" }
  }
  $state = if ($installedVersion -eq $availableVersion) { "current" } else { "version_differs" }
  Write-Output "Pure Tokens Skills update check: installed=$installedVersion available=$availableVersion status=$state source_commit=$sourceCommit"
  if ($Command -eq "check-update") { return }
  if ($installedVersion -match '^\d+\.\d+\.\d+$' -and [version]$installedVersion -gt [version]$availableVersion) { Fail "the installed version is newer than official main; automatic downgrade was stopped" }
  $architecture = [Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture.ToString().ToLowerInvariant()
  $platform = switch ($architecture) { "x64" { "windows-amd64" }; "arm64" { "windows-arm64" }; default { Fail "this operating system and CPU has no platform executor" } }
  $releaseOrigin = "https://github.com/PureTokens/puretokens-skill/releases/download/v$availableVersion"
  $releaseFile = Join-Path $downloadRoot "release.json"
  $archive = Join-Path $downloadRoot "source.zip"
  $releaseStatus = Get-OfficialFile "$releaseOrigin/release-manifest.json" $releaseFile
  $distribution = "source"
  if ($releaseStatus -eq 200) {
    $release = Read-Json $releaseFile
    if ($release.version -eq $availableVersion -and $release.sourceCommit -ceq $sourceCommit) {
      $artifact = $release.files.PSObject.Properties[$platform].Value
      if ($null -eq $artifact -or $artifact.filename -cne "puretokens-skill-$availableVersion-$platform.zip" -or $artifact.sha256 -cnotmatch '^[0-9a-f]{64}$') { Fail "published platform metadata is invalid" }
      if ((Get-OfficialFile "$releaseOrigin/$($artifact.filename)" $archive) -ne 200) { Fail "the published platform archive is unavailable" }
      if ((Get-FileHash -LiteralPath $archive -Algorithm SHA256).Hash.ToLowerInvariant() -cne $artifact.sha256) { Fail "published platform archive checksum mismatch" }
      $distribution = "platform"
      $archiveRoot = "puretokens-skill"
    }
  } elseif ($releaseStatus -ne 404) { Fail "published platform metadata could not be checked" }
  if ($distribution -eq "source") {
    Write-Output "Matching published platform assets are unavailable; retrieving the pinned official source archive."
    if ((Get-OfficialFile "https://codeload.github.com/PureTokens/puretokens-skill/zip/$sourceCommit" $archive) -ne 200) { Fail "the pinned official source archive is unavailable" }
    $archiveRoot = "puretokens-skill-$sourceCommit"
  }
  Add-Type -AssemblyName System.IO.Compression.FileSystem
  $zip = [IO.Compression.ZipFile]::OpenRead($archive)
  try {
    if ($zip.Entries.Count -eq 0) { Fail "the archive is empty" }
    foreach ($entry in $zip.Entries) {
      if (-not $entry.FullName.StartsWith("$archiveRoot/", [StringComparison]::Ordinal) -or $entry.FullName -match '\\|(^|/)\.\.?(/|$)' -or (($entry.ExternalAttributes -shr 16) -band 0xF000) -eq 0xA000) { Fail "the archive contains an unsafe path or symlink" }
    }
  } finally { $zip.Dispose() }
  $unpacked = Join-Path $downloadRoot "unpacked"
  [IO.Compression.ZipFile]::ExtractToDirectory($archive, $unpacked)
  $sourceRoot = Join-Path $unpacked $archiveRoot
  if ((Read-Json (Join-Path $sourceRoot "package.json")).version -ne $availableVersion) { Fail "the archive version does not match the pinned source" }
  $syncOptions = @{ Source = $sourceRoot; Target = $Target }
  if ($HostId) { $syncOptions.HostId = $HostId }
  & (Join-Path $sourceRoot "runtime/puretokens-skill-install.ps1") sync @syncOptions
} finally {
  if (Test-Path -LiteralPath $downloadRoot) { Remove-Item -LiteralPath $downloadRoot -Recurse -Force }
}
