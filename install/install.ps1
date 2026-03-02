param(
  [string]$Repo = "wingoo/landrop",
  [string]$Version = "latest",
  [string]$InstallDir = "$env:LOCALAPPDATA\Programs\landrop"
)

$ErrorActionPreference = "Stop"

$asset = "landrop_windows_amd64.zip"
if ($Version -eq "latest") {
  $assetUrl = "https://github.com/$Repo/releases/latest/download/$asset"
  $checksumsUrl = "https://github.com/$Repo/releases/latest/download/checksums.txt"
} else {
  $assetUrl = "https://github.com/$Repo/releases/download/$Version/$asset"
  $checksumsUrl = "https://github.com/$Repo/releases/download/$Version/checksums.txt"
}

Write-Host "Resolving release: $Repo ($Version)"

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
$tmpZip = Join-Path $env:TEMP "landrop_windows_amd64.zip"
$tmpChecksums = Join-Path $env:TEMP ("landrop_checksums_" + [guid]::NewGuid().ToString("N") + ".txt")
$tmpDir = Join-Path $env:TEMP ("landrop_" + [guid]::NewGuid().ToString("N"))

Invoke-WebRequest -UseBasicParsing -Uri $assetUrl -OutFile $tmpZip
Invoke-WebRequest -UseBasicParsing -Uri $checksumsUrl -OutFile $tmpChecksums

$checksums = Get-Content -Path $tmpChecksums
$entry = $checksums | Where-Object { $_ -match "\s$asset$" } | Select-Object -First 1
if (-not $entry) {
  throw "Could not find checksum entry for $asset"
}
$expectedSha = ($entry -split '\s+')[0].ToLowerInvariant()
$actualSha = (Get-FileHash -Algorithm SHA256 -Path $tmpZip).Hash.ToLowerInvariant()
if ($actualSha -ne $expectedSha) {
  throw "Checksum verification failed for $asset"
}

Expand-Archive -Path $tmpZip -DestinationPath $tmpDir -Force

$src = Join-Path $tmpDir "landrop.exe"
if (-not (Test-Path $src)) {
  throw "landrop.exe not found in archive"
}

$dst = Join-Path $InstallDir "landrop.exe"
Copy-Item -Force $src $dst

Remove-Item -Force $tmpZip
Remove-Item -Force $tmpChecksums
Remove-Item -Recurse -Force $tmpDir

Write-Host "Installed: $dst"
Write-Host "If needed, add to PATH: $InstallDir"
Write-Host "Run: landrop.exe --help"
