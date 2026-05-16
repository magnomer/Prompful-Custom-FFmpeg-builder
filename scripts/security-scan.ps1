$ErrorActionPreference = "Stop"

$repositoryRoot = Split-Path -Parent $PSScriptRoot
$violations = @()
$observedDescriptions = @{}

# This scanner documents the source boundaries used by this app.
# It is intentionally small and text-based: it looks for sensitive Go
# primitives and checks whether they appear in the parts of the app that
# currently own those responsibilities.
$rules = @(
    [PSCustomObject]@{
        Needle = "exec.Command"
        AllowedPrefixes = @("app.go", "internal/execution/")
        Description = "processes are started either by the app shell integration or by the managed execution layer"
    },
    [PSCustomObject]@{
        Needle = "os.StartProcess"
        AllowedPrefixes = @()
        Description = "low-level process startup is not part of this app"
    },
    [PSCustomObject]@{
        Needle = "syscall.Exec"
        AllowedPrefixes = @()
        Description = "process replacement is not part of this app"
    },
    [PSCustomObject]@{
        Needle = "http.NewRequest"
        AllowedPrefixes = @("internal/download/")
        Description = "network downloads are implemented by the download layer"
    },
    [PSCustomObject]@{
        Needle = "http.Client"
        AllowedPrefixes = @("internal/download/")
        Description = "HTTP clients are owned by the download layer"
    },
    [PSCustomObject]@{
        Needle = ".Do(request)"
        AllowedPrefixes = @("internal/download/")
        Description = "HTTP request execution is owned by the download layer"
    },
    [PSCustomObject]@{
        Needle = "net.Dial"
        AllowedPrefixes = @()
        Description = "raw network dialing is not part of this app"
    },
    [PSCustomObject]@{
        Needle = "os.RemoveAll"
        AllowedPrefixes = @("app.go", "internal/workspace/")
        Description = "recursive deletion is limited to workspace cleanup code paths"
    },
    [PSCustomObject]@{
        Needle = "os.Remove("
        AllowedPrefixes = @("app.go", "internal/download/", "internal/scripting/", "internal/workspace/")
        Description = "single-file deletion is used for workspace cleanup, download replacement, and generated script replacement"
    },
    [PSCustomObject]@{
        Needle = "os.Rename"
        AllowedPrefixes = @("app.go", "internal/download/", "internal/workspace/")
        Description = "file and directory moves are used for workspace setup and atomic download replacement"
    },
    [PSCustomObject]@{
        Needle = "os.WriteFile"
        AllowedPrefixes = @("app.go", "internal/audit/", "internal/scripting/")
        Description = "direct file writes are used for generated scripts, audit/report files, and app-owned result files"
    },
    [PSCustomObject]@{
        Needle = '"-lc"'
        AllowedPrefixes = @("app.go")
        Description = "shell-string execution is limited to app-owned MSYS2 maintenance commands"
    },
    [PSCustomObject]@{
        Needle = "unsafe."
        AllowedPrefixes = @()
        Description = "unsafe Go operations are not part of this app"
    }
)

function Convert-ToRepositoryPath($FullName) {
    $relativePath = [System.IO.Path]::GetRelativePath($repositoryRoot, $FullName)
    return $relativePath.Replace([System.IO.Path]::DirectorySeparatorChar, '/').Replace([System.IO.Path]::AltDirectorySeparatorChar, '/')
}

function Test-AllowedPath($Path, $AllowedPrefixes) {
    foreach ($allowedPrefix in $AllowedPrefixes) {
        if ($Path -eq $allowedPrefix -or $Path.StartsWith($allowedPrefix, [System.StringComparison]::Ordinal)) {
            return $true
        }
    }
    return $false
}

Get-ChildItem -Path $repositoryRoot -Recurse -Filter *.go | ForEach-Object {
    $relativePath = Convert-ToRepositoryPath $_.FullName

    if ($relativePath.StartsWith("scripts/", [System.StringComparison]::Ordinal)) { return }
    if ($relativePath.EndsWith("_test.go", [System.StringComparison]::Ordinal)) { return }

    $text = Get-Content $_.FullName -Raw

    foreach ($rule in $rules) {
        if ($text.Contains($rule.Needle)) {
            $observedDescriptions[$rule.Description] = $true

            if (-not (Test-AllowedPath $relativePath $rule.AllowedPrefixes)) {
                $script:violations += "${relativePath}: $($rule.Description)"
            }
        }
    }
}

if ($violations.Count -gt 0) {
    Write-Host "app source boundary scan found mismatches:" -ForegroundColor Red
    $violations | ForEach-Object { Write-Host " - $_" -ForegroundColor Red }
    exit 1
}

Write-Host "app source boundary scan passed" -ForegroundColor Green
Write-Host "observed app behavior:"
foreach ($rule in $rules) {
    if ($observedDescriptions.ContainsKey($rule.Description)) {
        Write-Host " - $($rule.Description)"
    }
}
