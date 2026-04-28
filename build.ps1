[CmdletBinding()]
param(
    [string]$Output = "ws7.exe",
    [ValidateSet("Debug", "Release")]
    [string]$Configuration = "Debug",
    [switch]$Clean,
    [switch]$Incremental,
    [switch]$SkipTests,
    [switch]$Run,
    [switch]$OpenOutputFolder
)

$ErrorActionPreference = "Stop"

function Get-RepoVersion {
    $versionFile = Join-Path $scriptRoot "internal\version\version.go"
    if (-not (Test-Path $versionFile)) {
        return "unknown"
    }

    $line = Select-String -Path $versionFile -Pattern 'const\s+Version\s*=\s*"([^"]+)"' | Select-Object -First 1
    if ($null -eq $line) {
        return "unknown"
    }
    return $line.Matches[0].Groups[1].Value
}

function Get-GitCommit {
    if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
        return "n/a"
    }

    try {
        $commit = (& git --no-pager rev-parse --short HEAD 2>$null)
        if ([string]::IsNullOrWhiteSpace($commit)) {
            return "n/a"
        }
        return $commit.Trim()
    }
    catch {
        return "n/a"
    }
}

$scriptRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
Push-Location $scriptRoot

try {
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        throw "Go was not found in PATH. Install Go and try again."
    }

    if (-not (Test-Path "go.mod")) {
        throw "go.mod not found. Run this script from the repository root."
    }

    $targetPackage = "./cmd/ws7"
    $outputPath = [System.IO.Path]::GetFullPath((Join-Path $scriptRoot $Output))
    $outputDir = Split-Path -Parent $outputPath
    if (-not [string]::IsNullOrWhiteSpace($outputDir)) {
        New-Item -ItemType Directory -Path $outputDir -Force | Out-Null
    }

    $doCleanBuild = $Clean -or (-not $Incremental)
    if ($doCleanBuild -and (Test-Path $outputPath)) {
        try {
            Remove-Item -Force $outputPath
            Write-Host "Removed previous artifact: $outputPath"
        }
        catch {
            throw "Unable to remove previous artifact '$outputPath'. Close WS7 or any process using the file and try again. $($_.Exception.Message)"
        }
    }

    $goVersion = (& go version).Trim()
    $repoVersion = Get-RepoVersion
    $gitCommit = Get-GitCommit
    $timestamp = (Get-Date).ToString("yyyy-MM-dd HH:mm:ss")
    $unixSeconds = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
    $buildHex = "{0:x}" -f $unixSeconds

    Write-Host "=== WS7 Build ==="
    Write-Host "Time:          $timestamp"
    Write-Host "Configuration: $Configuration"
    Write-Host "Output:        $outputPath"
    Write-Host "Mode:          $(if ($doCleanBuild) { 'Clean rebuild' } else { 'Incremental build' })"
    Write-Host "Go:            $goVersion"
    Write-Host "WS7 version:   $repoVersion"
    Write-Host "Build (hex):   $buildHex"
    Write-Host "Git commit:    $gitCommit"

    if (-not $SkipTests) {
        Write-Host "Running tests..."
        & go test ./...
        if ($LASTEXITCODE -ne 0) {
            throw "Tests failed. Build aborted."
        }
    }

    $ldflags = "-X ws7/internal/version.BuildID=$buildHex"
    $buildArgs = @("build", "-o", $outputPath)
    if ($Configuration -eq "Release") {
        $ldflags += " -s -w"
        $buildArgs += @("-trimpath")
    }
    $buildArgs += @("-ldflags", $ldflags)
    $buildArgs += $targetPackage

    Write-Host "Building $targetPackage..."
    & go @buildArgs
    if ($LASTEXITCODE -ne 0) {
        throw "go build failed."
    }

    if (-not (Test-Path $outputPath)) {
        throw "Build finished but output was not created: $outputPath"
    }

    $artifact = Get-Item $outputPath
    Write-Host "Build OK: $($artifact.FullName) ($($artifact.Length) bytes)"

    if ($OpenOutputFolder) {
        Write-Host "Opening output folder..."
        & explorer.exe $outputDir
    }

    if ($Run) {
        Write-Host "Running $($artifact.Name)..."
        & $outputPath
    }
}
finally {
    Pop-Location
}

