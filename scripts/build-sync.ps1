# Build the local checkout and sync generated assets into the selected AI tool.
#
# Usage:
#   .\scripts\build-sync.ps1
#   .\scripts\build-sync.ps1 -Target codex
#   .\scripts\build-sync.ps1 -Target claude,gemini,codex
#   .\scripts\build-sync.ps1 -SkipTests

[CmdletBinding()]
param(
    [string[]]$Target = @('codex'),
    [switch]$SkipTests
)

$ErrorActionPreference = 'Stop'

function Write-Info { param($Msg) Write-Host ":: $Msg" -ForegroundColor Cyan }
function Write-Ok   { param($Msg) Write-Host "OK $Msg" -ForegroundColor Green }
function Fail       { param($Msg) Write-Host "error: $Msg" -ForegroundColor Red; exit 1 }

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoRoot = Resolve-Path (Join-Path $ScriptDir '..')
Set-Location $RepoRoot

if (-not $SkipTests) {
    Write-Info 'Running tests ...'
    go test ./internal/...
    if ($LASTEXITCODE -ne 0) { Fail 'go test failed.' }
    Write-Ok 'Tests passed.'
}

Write-Info 'Building local sgai.exe ...'
go build -o sgai.exe ./cmd/system-general-ai
if ($LASTEXITCODE -ne 0) { Fail 'go build failed.' }
Write-Ok 'Built .\sgai.exe'

Write-Info "Syncing target(s): $($Target -join ', ') ..."
& .\sgai.exe sync @Target
if ($LASTEXITCODE -ne 0) { Fail 'sgai sync failed.' }

Write-Ok 'Done. Restart the selected AI tool(s) to apply.'
