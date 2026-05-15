$ErrorActionPreference = "Stop"

$repositoryRoot = Split-Path -Parent $PSScriptRoot
$violations = @()

function Add-Violation($Path, $Message) {
    $script:violations += "${Path}: ${Message}"
}

Get-ChildItem -Path $repositoryRoot -Recurse -Include *.go | ForEach-Object {
    $relativePath = Resolve-Path -Relative $_.FullName
    if ($relativePath -match "scripts[\/]") { return }
    $text = Get-Content $_.FullName -Raw

    if ($text -match "exec\.Command" -and $relativePath -notmatch "internal[\\/]execution[\\/]execution\.go") {
        Add-Violation $relativePath "exec.Command usage is allowed only in internal/execution."
    }
    if (($text -match "http\.Get" -or $text -match "http\.DefaultClient" -or $text -match "\.Do\(request\)") -and $relativePath -notmatch "internal[\\/]download[\\/]download\.go") {
        Add-Violation $relativePath "HTTP download primitives are allowed only in internal/download."
    }
    if ($text -match "os\.RemoveAll" -and $relativePath -notmatch "internal[\\/]workspace") {
        Add-Violation $relativePath "os.RemoveAll is allowed only in controlled workspace code."
    }
    if ($text -match '"-lc"') {
        Add-Violation $relativePath "bash -lc shell-string execution is forbidden; use approved script files."
    }
}

if ($violations.Count -gt 0) {
    Write-Host "Security scan failed:" -ForegroundColor Red
    $violations | ForEach-Object { Write-Host " - $_" -ForegroundColor Red }
    exit 1
}

Write-Host "Security scan passed." -ForegroundColor Green
