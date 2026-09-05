param([switch]$Child)
$ErrorActionPreference = "Stop"
if (-not $Child) {
  foreach ($engine in @("powershell.exe", "pwsh")) {
    & $engine -NoProfile -ExecutionPolicy Bypass -File $PSCommandPath -Child
    if ($LASTEXITCODE -ne 0) { throw "$engine distribution fixtures failed" }
  }
  return
}
$repository = Split-Path -Parent $PSScriptRoot
$root = Join-Path ([IO.Path]::GetTempPath()) ("pt-download-test-" + [Guid]::NewGuid().ToString("N"))
$savedEnvironment = @{}
foreach ($name in @("USERPROFILE", "HOME", "CODEX_HOME", "TMP", "TEMP")) { $savedEnvironment[$name] = [Environment]::GetEnvironmentVariable($name, "Process") }
try {
  New-Item -ItemType Directory -Path $root | Out-Null
  $env:USERPROFILE = Join-Path $root "home"
  $env:HOME = $env:USERPROFILE
  $env:CODEX_HOME = Join-Path $env:USERPROFILE ".codex"
  $env:TMP = Join-Path $root "tmp"
  $env:TEMP = $env:TMP
  New-Item -ItemType Directory -Path $env:USERPROFILE, $env:TMP | Out-Null
  $global:fixtureCommit = "a" * 40
  $global:fixtureVersion = (Get-Content -LiteralPath (Join-Path $repository "package.json") -Raw -Encoding UTF8 | ConvertFrom-Json).version
  $platform = if ([Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture.ToString() -eq "Arm64") { "windows-arm64" } else { "windows-amd64" }
  $source = Join-Path $root "puretokens-skill"
  New-Item -ItemType Directory -Path (Join-Path $source "runtime/executor/bin") -Force | Out-Null
  foreach ($file in @("README.md", "package.json")) { Copy-Item -LiteralPath (Join-Path $repository $file) -Destination $source }
  Copy-Item -LiteralPath (Join-Path $repository "skills") -Destination $source -Recurse
  foreach ($file in @("puretokens-skill-install.sh", "puretokens-skill-install.ps1", "puretokens-skill-fetch.sh", "puretokens-skill-fetch.ps1")) {
    Copy-Item -LiteralPath (Join-Path (Join-Path $repository "runtime") $file) -Destination (Join-Path $source "runtime")
  }
  Copy-Item -LiteralPath (Join-Path $repository "runtime/executor/manifest.json") -Destination (Join-Path $source "runtime/executor")
  Copy-Item -LiteralPath (Join-Path $repository "runtime/executor/bin/puretokens-api-$platform.exe") -Destination (Join-Path $source "runtime/executor/bin")
  $global:fixtureSource = $source
  $global:fixtureRoot = $root
  Add-Type -AssemblyName System.IO.Compression.FileSystem
  [IO.Compression.ZipFile]::CreateFromDirectory($source, (Join-Path $root "platform.zip"), [IO.Compression.CompressionLevel]::Fastest, $true)
  $pinnedSource = Join-Path $root "puretokens-skill-$global:fixtureCommit"
  Copy-Item -LiteralPath $source -Destination $pinnedSource -Recurse
  [IO.Compression.ZipFile]::CreateFromDirectory($pinnedSource, (Join-Path $root "source.zip"), [IO.Compression.CompressionLevel]::Fastest, $true)
  $release = @{
    version = $global:fixtureVersion
    sourceCommit = $global:fixtureCommit
    files = @{ $platform = @{ filename = "puretokens-skill-$global:fixtureVersion-$platform.zip"; sha256 = (Get-FileHash -LiteralPath (Join-Path $root "platform.zip") -Algorithm SHA256).Hash.ToLowerInvariant() } }
  }
  $release | ConvertTo-Json -Depth 6 | Set-Content -LiteralPath (Join-Path $root "release.json") -Encoding UTF8
  function global:Invoke-WebRequest {
    param($Uri, $OutFile, $TimeoutSec, $Headers, $UserAgent, [switch]$UseBasicParsing, [switch]$PassThru)
    $global:fixtureRequests.Add([string]$Uri)
    switch -Regex ($Uri) {
      '/commits/main$' { @{ sha = $global:fixtureCommit } | ConvertTo-Json | Set-Content -LiteralPath $OutFile -Encoding UTF8 }
      "^https://raw.githubusercontent.com/PureTokens/puretokens-skill/$global:fixtureCommit/package.json$" { Copy-Item -LiteralPath (Join-Path $global:fixtureSource "package.json") -Destination $OutFile }
      "^https://raw.githubusercontent.com/PureTokens/puretokens-skill/$global:fixtureCommit/runtime/puretokens-skill-install.ps1$" { Copy-Item -LiteralPath (Join-Path $global:fixtureSource "runtime/puretokens-skill-install.ps1") -Destination $OutFile }
      '/release-manifest.json$' {
        if ($global:fixtureMode -eq "source") { return [PSCustomObject]@{ StatusCode = 404 } }
        Copy-Item -LiteralPath (Join-Path $global:fixtureRoot "release.json") -Destination $OutFile
      }
      "^https://codeload.github.com/PureTokens/puretokens-skill/zip/$global:fixtureCommit$" {
        if ($global:fixtureConcurrentSource) {
          & (Join-Path $global:fixtureConcurrentSource "runtime/puretokens-skill-install.ps1") sync -Target $global:fixtureConcurrentTarget -Source $global:fixtureConcurrentSource | Out-Null
        }
        Copy-Item -LiteralPath (Join-Path $global:fixtureRoot "source.zip") -Destination $OutFile
      }
      '/puretokens-skill-[0-9.]+-windows-(amd64|arm64).zip$' { Copy-Item -LiteralPath (Join-Path $global:fixtureRoot "platform.zip") -Destination $OutFile }
      default { throw "Unexpected fixture request: $Uri" }
    }
    return [PSCustomObject]@{ StatusCode = 200 }
  }
  $fetch = Join-Path $source "runtime/puretokens-skill-fetch.ps1"
  foreach ($mode in @("source", "platform")) {
    $global:fixtureMode = $mode
    $global:fixtureRequests = New-Object 'System.Collections.Generic.List[string]'
    $target = Join-Path $root "target-$mode"
    & $fetch check-update -Target $target
    if ($global:fixtureRequests.Count -ne 2 -or (Test-Path -LiteralPath $target)) { throw "check-update changed local installation or fetched more than metadata" }
    $global:fixtureRequests.Clear()
    & $fetch install -Target $target
    if (-not (Test-Path -LiteralPath (Join-Path $target ".puretokens-executor/puretokens-api.exe"))) { throw "download did not install the platform executor" }
    $sourceRequests = @($global:fixtureRequests | Where-Object { $_ -like "https://codeload.github.com/*" })
    if (($mode -eq "source" -and $sourceRequests.Count -ne 1) -or ($mode -eq "platform" -and $sourceRequests.Count -ne 0)) { throw "incorrect distribution selection" }
    if (@(Get-ChildItem -LiteralPath $env:TMP -Force).Count -ne 0) { throw "private download staging was not cleaned" }
  }
  $bootstrap = Join-Path $root "legacy-bootstrap"
  New-Item -ItemType Directory -Path $bootstrap | Out-Null
  Copy-Item -LiteralPath $fetch -Destination $bootstrap
  "exit 79" | Set-Content -LiteralPath (Join-Path $bootstrap "puretokens-skill-install.ps1") -Encoding ASCII
  $global:fixtureRequests.Clear()
  & (Join-Path $bootstrap "puretokens-skill-fetch.ps1") check-update -Target (Join-Path $root "legacy-target")
  if (@($global:fixtureRequests | Where-Object { $_ -like "*/runtime/puretokens-skill-install.ps1" }).Count -ne 1) { throw "legacy sibling selector was not replaced by the pinned selector" }

  $newer = Join-Path $root "newer-source"
  Copy-Item -LiteralPath $source -Destination $newer -Recurse
  @{ version = "99.0.0" } | ConvertTo-Json | Set-Content -LiteralPath (Join-Path $newer "package.json") -Encoding UTF8
  $global:fixtureConcurrentSource = $newer
  $global:fixtureConcurrentTarget = Join-Path $root "target-source"
  $global:fixtureMode = "source"
  $rejected = $false
  try { & $fetch update -Target $global:fixtureConcurrentTarget } catch { $rejected = $_.Exception.Message -like "*downgrade was stopped under the update lock*" }
  if (-not $rejected) { throw "an older download overwrote the newer installation" }
  $currentRuntime = Get-Content -LiteralPath (Join-Path $global:fixtureConcurrentTarget ".puretokens-executor/runtime.json") -Raw -Encoding UTF8 | ConvertFrom-Json
  if ($currentRuntime.version -ne "99.0.0") { throw "newer version was not preserved" }
  $global:fixtureConcurrentSource = $null
  $global:fixtureMode = "platform"
  $release.files[$platform].sha256 = "0" * 64
  $release | ConvertTo-Json -Depth 6 | Set-Content -LiteralPath (Join-Path $root "release.json") -Encoding UTF8
  $rejected = $false
  try { & $fetch install -Target (Join-Path $root "bad-checksum") } catch { $rejected = $_.Exception.Message -like "*checksum mismatch*" }
  if (-not $rejected -or (Test-Path -LiteralPath (Join-Path $root "bad-checksum"))) { throw "checksum mismatch was not rejected before mutation" }
} finally {
  Remove-Item function:global:Invoke-WebRequest -ErrorAction SilentlyContinue
  foreach ($name in $savedEnvironment.Keys) { [Environment]::SetEnvironmentVariable($name, $savedEnvironment[$name], "Process") }
  if (Test-Path -LiteralPath $root) { Remove-Item -LiteralPath $root -Recurse -Force }
}
