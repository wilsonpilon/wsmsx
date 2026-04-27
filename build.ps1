param(
    [string]$Output = "ws7.exe",
    [ValidateSet("Debug", "Release")]
    [string]$Configuration = "Debug",
    [switch]$Clean,
    [switch]$SkipTests
)

$ErrorActionPreference = "Stop"

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

    if ($Clean -and (Test-Path $Output)) {
        Remove-Item -Force $Output
        Write-Host "Removed previous artifact: $Output"
    }

    if (-not $SkipTests) {
        Write-Host "Running tests..."
        & go test ./...
        if ($LASTEXITCODE -ne 0) {
            throw "Tests failed. Build aborted."
        }
    }

    $buildArgs = @("build", "-o", $Output)

    if ($Configuration -eq "Release") {
        $buildArgs += @("-ldflags", "-s -w")
    }

    $buildArgs += $targetPackage

    Write-Host "Building $targetPackage -> $Output ($Configuration)..."
    & go @buildArgs
    if ($LASTEXITCODE -ne 0) {
        throw "go build failed."
    }

    if (-not (Test-Path $Output)) {
        throw "Build finished but output was not created: $Output"
    }

    $artifact = Get-Item $Output
    $artifactPath = (Resolve-Path $Output).Path
    Write-Host "Build OK: $artifactPath ($($artifact.Length) bytes)"
}
finally {
    Pop-Location
}

