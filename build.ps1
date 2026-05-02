[CmdletBinding()]
param(
    [Alias("OutputName")]
    [string]$Output = "ws7.exe",
    [string]$OutputDir,
    [ValidateSet("Debug", "Release")]
    [string]$Configuration = "Debug",
    [switch]$Clean,
    [switch]$Incremental,
    [switch]$SkipTests,
    [switch]$NoConsole,
    [switch]$Console,
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
    if ($NoConsole -and $Console) {
        throw "Choose only one console mode override: use either -NoConsole or -Console."
    }

    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        throw "Go was not found in PATH. Install Go and try again."
    }

    if (-not (Test-Path "go.mod")) {
        throw "go.mod not found. Run this script from the repository root."
    }

    $targetPackage = "./cmd/ws7"

    $rawOutput = if ([string]::IsNullOrWhiteSpace($Output)) { "ws7.exe" } else { $Output.Trim() }
    $hasOutputDir = -not [string]::IsNullOrWhiteSpace($OutputDir)

    if ($hasOutputDir) {
        $rawOutputName = [System.IO.Path]::GetFileName($rawOutput)
        if ([string]::IsNullOrWhiteSpace($rawOutputName) -or $rawOutputName -ne $rawOutput) {
            throw "When -OutputDir is used, -Output must be a file name only (example: -Output ws7.exe -OutputDir .\\bin)."
        }

        $resolvedOutputDir = [System.IO.Path]::GetFullPath((Join-Path $scriptRoot $OutputDir.Trim()))
        $outputPath = Join-Path $resolvedOutputDir $rawOutputName
        Write-Host "Output path resolved from file '$rawOutputName' + directory '$resolvedOutputDir'."
    }
    else {
        $candidateOutputPath = [System.IO.Path]::GetFullPath((Join-Path $scriptRoot $rawOutput))
        $looksLikeDirectoryInput = $rawOutput -eq "." -or $rawOutput -eq ".." -or $rawOutput.EndsWith("\") -or $rawOutput.EndsWith("/")
        $candidateIsDirectory = (Test-Path -LiteralPath $candidateOutputPath -PathType Container)

        if ($looksLikeDirectoryInput -or $candidateIsDirectory -or $candidateOutputPath -eq $scriptRoot) {
            $outputPath = Join-Path $candidateOutputPath "ws7.exe"
            Write-Host "Output path resolved to directory '$candidateOutputPath'; using '$outputPath'."
        }
        else {
            $outputPath = $candidateOutputPath
        }
    }

    if (Test-Path -LiteralPath $outputPath -PathType Container) {
        throw "Output path '$outputPath' resolves to a directory. Provide a file name, e.g. -Output ws7.exe or -Output .\\bin\\ws7.exe"
    }

    $outputDir = Split-Path -Parent $outputPath
    if (-not [string]::IsNullOrWhiteSpace($outputDir)) {
        New-Item -ItemType Directory -Path $outputDir -Force | Out-Null
    }

    $doCleanBuild = $Clean -or (-not $Incremental)
    if ($doCleanBuild -and (Test-Path -LiteralPath $outputPath)) {
        try {
            $existingArtifact = Get-Item -LiteralPath $outputPath -Force
            if ($existingArtifact.PSIsContainer) {
                throw "Refusing to remove directory '$outputPath'."
            }
            Remove-Item -LiteralPath $outputPath -Force
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
    $hideConsole = ($Configuration -eq "Release")
    if ($NoConsole) {
        $hideConsole = $true
    }
    if ($Console) {
        $hideConsole = $false
    }
    $consoleMode = if ($hideConsole) { "NoConsole (windowsgui)" } else { "Console" }

    Write-Host "=== WS7 Build ==="
    Write-Host "Time:          $timestamp"
    Write-Host "Configuration: $Configuration"
    Write-Host "Output:        $outputPath"
    Write-Host "Mode:          $(if ($doCleanBuild) { 'Clean rebuild' } else { 'Incremental build' })"
    Write-Host "Go:            $goVersion"
    Write-Host "WS7 version:   $repoVersion"
    Write-Host "Build (hex):   $buildHex"
    Write-Host "Git commit:    $gitCommit"
    Write-Host "Console mode:  $consoleMode"

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
    if ($hideConsole) {
        $ldflags += " -H=windowsgui"
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
