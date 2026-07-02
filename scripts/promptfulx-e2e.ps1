#Requires -Version 5.1
<#
.SYNOPSIS
  Build and end-to-end exercise the PromptfulX CLI (setup -> plan -> build).

.DESCRIPTION
  Compiles promptfulx.exe to build\bin, then runs:
    1. setup  - prepares the MSYS2 toolchain in the workspace (downloads MSYS2)
    2. plan   - prints the resolved FFmpeg build plan
    3. build  - compiles FFmpeg (long: can take 1-2 hours)

  Each stage's exit code is checked. Setup failure stops the run. Use
  -SetupOnly to validate the download/extract/toolchain path without the
  hours-long FFmpeg compile.

.PARAMETER Workspace
  Build workspace directory (absolute). Created if missing.

.PARAMETER FfmpegVersion
  Supported FFmpeg release (e.g. 8.1.2). See: promptfulx list versions.

.PARAMETER Preset
  Preset to start from (e.g. minimal, full). See: promptfulx list presets.

.PARAMETER Jobs
  Parallel build jobs. Defaults to the processor count.

.PARAMETER SetupOnly
  Run setup (and plan) but skip the FFmpeg build.

.PARAMETER Msys2Url
  Optional override for the MSYS2 archive URL.

.EXAMPLE
  .\scripts\promptfulx-e2e.ps1 -Workspace D:\PromptfulWork -SetupOnly

.EXAMPLE
  .\scripts\promptfulx-e2e.ps1 -Workspace D:\PromptfulWork -FfmpegVersion 8.1.2 -Preset minimal
#>
[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [string]$Workspace,

    [string]$FfmpegVersion = "8.1.2",
    [string]$Preset = "minimal",
    [int]$Jobs = [Environment]::ProcessorCount,
    [switch]$SetupOnly,
    [string]$Msys2Url
)

$ErrorActionPreference = "Stop"

# Resolve repo root as the parent of this script's directory.
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = Split-Path -Parent $scriptDir
Set-Location $repoRoot

$exe = Join-Path $repoRoot "build\bin\promptfulx.exe"

function Write-Stage([string]$text) {
    Write-Host ""
    Write-Host "==== $text ====" -ForegroundColor Cyan
}

function Invoke-Stage([string]$name, [string[]]$cliArgs) {
    Write-Stage $name
    Write-Host "> promptfulx $($cliArgs -join ' ')" -ForegroundColor DarkGray
    $start = Get-Date
    # Out-Host keeps the exe's stdout on the console instead of leaking it into
    # this function's return value; $LASTEXITCODE still reflects the exe.
    & $exe @cliArgs | Out-Host
    $code = $LASTEXITCODE
    $elapsed = (Get-Date) - $start
    Write-Host ("[{0}] exit={1} in {2:mm\:ss}" -f $name, $code, $elapsed) -ForegroundColor Gray
    return $code
}

# --- Build the CLI -----------------------------------------------------------
Write-Stage "build promptfulx.exe"
New-Item -ItemType Directory -Force (Split-Path -Parent $exe) | Out-Null
& go build -o $exe ".\cmd\promptfulx"
if ($LASTEXITCODE -ne 0) {
    Write-Host "go build failed" -ForegroundColor Red
    exit 1
}
Write-Host "built: $exe" -ForegroundColor Green

# --- Ensure workspace --------------------------------------------------------
if (-not [System.IO.Path]::IsPathRooted($Workspace)) {
    Write-Host "Workspace must be an absolute path: $Workspace" -ForegroundColor Red
    exit 2
}
New-Item -ItemType Directory -Force $Workspace | Out-Null

# --- setup -------------------------------------------------------------------
$setupArgs = @("setup", "--workspace", $Workspace, "--yes")
if ($Msys2Url) { $setupArgs += @("--msys2-url", $Msys2Url) }
$setupCode = Invoke-Stage "setup" $setupArgs
if ($setupCode -ne 0) {
    Write-Host "setup failed (exit $setupCode); stopping." -ForegroundColor Red
    exit $setupCode
}

# --- plan --------------------------------------------------------------------
$planArgs = @("plan", "--ffmpeg-version", $FfmpegVersion, "--preset", $Preset,
    "--workspace", $Workspace, "--jobs", "$Jobs")
$null = Invoke-Stage "plan" $planArgs

if ($SetupOnly) {
    Write-Stage "done (SetupOnly)"
    Write-Host "Setup + plan verified. Skipped FFmpeg build." -ForegroundColor Green
    exit 0
}

# --- build -------------------------------------------------------------------
Write-Host ""
Write-Host "The FFmpeg build can take 1-2 hours." -ForegroundColor Yellow
$buildArgs = @("build", "--ffmpeg-version", $FfmpegVersion, "--preset", $Preset,
    "--workspace", $Workspace, "--jobs", "$Jobs", "--yes")
$buildCode = Invoke-Stage "build" $buildArgs

Write-Stage "summary"
Write-Host ("FFmpeg {0} / preset {1}" -f $FfmpegVersion, $Preset)
if ($buildCode -eq 0) {
    Write-Host "BUILD SUCCEEDED (exit 0)" -ForegroundColor Green
} else {
    Write-Host "BUILD FAILED (exit $buildCode)" -ForegroundColor Red
}
exit $buildCode
