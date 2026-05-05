# nea-ai installer for Windows (PowerShell 5.1+).
#
# Usage:
#   iwr https://raw.githubusercontent.com/RDuuke/nea-ai/main/scripts/install.ps1 -UseBasicParsing | iex
#
# Optional environment variables:
#   $env:NEA_AI_VERSION   Pin a specific tag (e.g. v0.2.0). Defaults to latest release.
#   $env:NEA_AI_BIN_DIR   Install directory. Defaults to $env:LOCALAPPDATA\Programs\nea-ai.
#   $env:NEA_AI_NO_PATH   When set to 1, skip updating the user PATH.

$ErrorActionPreference = 'Stop'

$repo = 'RDuuke/nea-ai'
$binary = 'nea-ai'
$defaultBinDir = Join-Path $env:LOCALAPPDATA 'Programs\nea-ai'
$binDir = if ($env:NEA_AI_BIN_DIR) { $env:NEA_AI_BIN_DIR } else { $defaultBinDir }

function Write-Info($msg) {
  Write-Host "==> $msg"
}

function Resolve-Arch {
  $raw = $env:PROCESSOR_ARCHITECTURE
  switch ($raw) {
    'AMD64' { return 'amd64' }
    'ARM64' { return 'arm64' }
    default { throw "unsupported architecture: $raw" }
  }
}

function Resolve-Version {
  if ($env:NEA_AI_VERSION) {
    return $env:NEA_AI_VERSION
  }
  $url = "https://api.github.com/repos/$repo/releases/latest"
  $headers = @{ 'User-Agent' = 'nea-ai-install' }
  $resp = Invoke-RestMethod -Uri $url -UseBasicParsing -Headers $headers
  if (-not $resp.tag_name) { throw 'could not read tag_name from GitHub API' }
  return $resp.tag_name
}

function Verify-Checksum($archivePath, $sumsPath, $archiveName) {
  $expectedLine = Get-Content $sumsPath | Where-Object { $_ -match "\s$([regex]::Escape($archiveName))\s*$" }
  if (-not $expectedLine) { throw "no checksum entry for $archiveName" }
  $expected = ($expectedLine -split '\s+')[0].ToLower()
  $actual = (Get-FileHash -Algorithm SHA256 -Path $archivePath).Hash.ToLower()
  if ($expected -ne $actual) {
    throw "checksum mismatch for $archiveName (expected $expected, got $actual)"
  }
}

function Add-ToUserPath($dir) {
  if ($env:NEA_AI_NO_PATH -eq '1') {
    Write-Host ""
    Write-Host "$dir is not in your PATH and NEA_AI_NO_PATH=1 is set."
    Write-Host "Add it manually or restart your shell after updating PATH."
    return
  }
  $current = [Environment]::GetEnvironmentVariable('Path', 'User')
  if ($null -eq $current) { $current = '' }
  $entries = $current -split ';' | Where-Object { $_ -ne '' }
  if ($entries -contains $dir) { return }
  $newPath = if ($current.TrimEnd(';') -eq '') { $dir } else { "$($current.TrimEnd(';'));$dir" }
  [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
  $env:Path = "$env:Path;$dir"
  Write-Host ""
  Write-Host "Added $dir to user PATH. Open a new terminal for the change to take effect."
}

function Main {
  $arch = Resolve-Arch
  $tag = Resolve-Version
  $versionNoV = $tag.TrimStart('v')
  $archiveName = "${binary}_${versionNoV}_windows_${arch}.zip"
  $archiveUrl = "https://github.com/$repo/releases/download/$tag/$archiveName"
  $sumsUrl = "https://github.com/$repo/releases/download/$tag/checksums.txt"

  $tmp = Join-Path $env:TEMP "nea-ai-install-$([Guid]::NewGuid().ToString('N'))"
  New-Item -ItemType Directory -Force -Path $tmp | Out-Null

  try {
    Write-Info "downloading $archiveName"
    $archivePath = Join-Path $tmp $archiveName
    Invoke-WebRequest -Uri $archiveUrl -OutFile $archivePath -UseBasicParsing

    $sumsPath = Join-Path $tmp 'checksums.txt'
    Invoke-WebRequest -Uri $sumsUrl -OutFile $sumsPath -UseBasicParsing

    Write-Info "verifying sha256"
    Verify-Checksum -archivePath $archivePath -sumsPath $sumsPath -archiveName $archiveName

    Write-Info "extracting archive"
    Expand-Archive -Force -Path $archivePath -DestinationPath $tmp

    New-Item -ItemType Directory -Force -Path $binDir | Out-Null
    $exeSrc = Join-Path $tmp "$binary.exe"
    $exeDst = Join-Path $binDir "$binary.exe"
    Move-Item -Force -Path $exeSrc -Destination $exeDst

    Write-Info "installed $binary $tag -> $exeDst"
    Add-ToUserPath -dir $binDir
  }
  finally {
    if (Test-Path $tmp) {
      Remove-Item -Recurse -Force -Path $tmp
    }
  }
}

Main
