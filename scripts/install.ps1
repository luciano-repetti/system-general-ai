# system-general-ai bootstrap — Windows (PowerShell 5.1+)
#
# Usage:
#   irm https://raw.githubusercontent.com/luciano-repetti/system-general-ai/main/scripts/install.ps1 | iex
#
# This script:
#   1. Detects architecture
#   2. Downloads the matching binary from GitHub Releases
#   3. Installs to %LOCALAPPDATA%\system-general-ai\bin\system-general-ai.exe (+ sgai.exe alias)
#   4. Adds the bin dir to the User PATH if missing
#   5. Runs `system-general-ai install` with default settings

$ErrorActionPreference = 'Stop'

$Repo      = 'luciano-repetti/system-general-ai'
$BinName   = 'system-general-ai'
$TargetDir = Join-Path $env:LOCALAPPDATA 'system-general-ai\bin'

function Write-Info { param($Msg) Write-Host ":: $Msg" -ForegroundColor Cyan }
function Write-Ok   { param($Msg) Write-Host "✓  $Msg" -ForegroundColor Green }
function Fail       { param($Msg) Write-Host "error: $Msg" -ForegroundColor Red; exit 1 }

function Get-Arch {
    if ([Environment]::Is64BitOperatingSystem) {
        if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') { return 'arm64' }
        return 'amd64'
    }
    Fail 'Unsupported architecture (32-bit Windows is not supported).'
}

function Get-LatestTag {
    $api = "https://api.github.com/repos/$Repo/releases/latest"
    try {
        $resp = Invoke-RestMethod -Uri $api -Headers @{ 'User-Agent' = 'system-general-ai-installer' }
        return $resp.tag_name
    } catch {
        Fail "Could not fetch latest release from $api : $_"
    }
}

function Ensure-Path {
    param($Dir)
    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    if ($userPath -split ';' -contains $Dir) { return }
    $newPath = if ($userPath) { "$userPath;$Dir" } else { $Dir }
    [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
    $env:Path = "$env:Path;$Dir"
    Write-Info "Added $Dir to User PATH. New shells will see it automatically."
}

# 1. Resolve target
$arch     = Get-Arch
$tag      = Get-LatestTag
$platform = "windows_$arch"
$archive  = "${BinName}_${tag}_${platform}.zip"
$url      = "https://github.com/$Repo/releases/download/$tag/$archive"

Write-Info "Installing $BinName $tag for $platform"

# 2. Download + extract
New-Item -ItemType Directory -Path $TargetDir -Force | Out-Null
$tmp = New-Item -ItemType Directory -Path (Join-Path $env:TEMP "sgai-$([guid]::NewGuid().ToString('N'))") -Force
try {
    $archivePath = Join-Path $tmp $archive
    Invoke-WebRequest -Uri $url -OutFile $archivePath -UseBasicParsing
    Expand-Archive -Path $archivePath -DestinationPath $tmp -Force

    $extracted = Join-Path $tmp "$BinName.exe"
    if (-not (Test-Path $extracted)) { Fail "$BinName.exe not found in archive." }

    $dest = Join-Path $TargetDir "$BinName.exe"
    Copy-Item $extracted $dest -Force
    Copy-Item $extracted (Join-Path $TargetDir 'sgai.exe') -Force
    Write-Ok "Binary installed at $dest (alias: sgai.exe)"
} finally {
    Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}

# 3. Ensure PATH
Ensure-Path -Dir $TargetDir

# 4. Run zero-config install
Write-Info 'Running zero-config install ...'
& (Join-Path $TargetDir "$BinName.exe") install
if ($LASTEXITCODE -ne 0) { Fail 'system-general-ai install failed.' }

Write-Ok 'Done. Restart the selected AI tool(s) to apply.'

