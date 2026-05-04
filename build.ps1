[CmdletBinding()]
param(
    [Alias("OutputName")]
    [string]$Output,
    [string]$OutputDir,
    [ValidateSet("both", "wsmsx", "wsdev")]
    [string]$Target = "both",
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

    $targetCatalog = @{
        wsmsx = @{ Package = "./cmd/ws7";  DefaultOutput = "wsmsx.exe" }
        wsdev = @{ Package = "./cmd/wsdev"; DefaultOutput = "wsdev.exe" }
    }

    $selectedTargets = switch ($Target) {
        "wsmsx" { @("wsmsx") }
        "wsdev" { @("wsdev") }
        default { @("wsmsx", "wsdev") }
    }
    if ($selectedTargets -is [string]) {
        $selectedTargets = @($selectedTargets)
    }

    $hasOutput = -not [string]::IsNullOrWhiteSpace($Output)

    # Backward compatibility: old usage like `.\build.ps1 . -Run` should still
    # build and run wsmsx only, even though default target is now both.
    if ($selectedTargets.Count -ne 1 -and ($Run -or $hasOutput)) {
        Write-Warning "Default multi-target build narrowed to 'wsmsx' because -Run or -Output was supplied. Use -Target both explicitly without -Run/-Output for dual build."
        $selectedTargets = @("wsmsx")
    }

    foreach ($t in $selectedTargets) {
        $pkg = $targetCatalog[$t].Package
        $pkgPath = Join-Path $scriptRoot ($pkg -replace '^\./', '' -replace '/', '\\')
        if (-not (Test-Path -LiteralPath $pkgPath -PathType Container)) {
            throw "Target '$t' package path not found: $pkg"
        }
    }

    if ($Run -and $selectedTargets.Count -ne 1) {
        throw "-Run only works with a single target. Use -Target wsmsx or -Target wsdev."
    }

    $hasOutputDir = -not [string]::IsNullOrWhiteSpace($OutputDir)
    $outputPathMap = @{}
    $resolvedOutputDir = ""

    if ($selectedTargets.Count -eq 1) {
        $singleTarget = $selectedTargets[0]
        $rawOutput = if ($hasOutput) { $Output.Trim() } else { $targetCatalog[$singleTarget].DefaultOutput }

        if ($hasOutputDir) {
            $rawOutputName = [System.IO.Path]::GetFileName($rawOutput)
            if ([string]::IsNullOrWhiteSpace($rawOutputName) -or $rawOutputName -ne $rawOutput) {
                throw "When -OutputDir is used, -Output must be a file name only (example: -Output wsmsx.exe -OutputDir .\\bin)."
            }

            $resolvedOutputDir = [System.IO.Path]::GetFullPath((Join-Path $scriptRoot $OutputDir.Trim()))
            $outputPathMap[$singleTarget] = Join-Path $resolvedOutputDir $rawOutputName
            Write-Host "Output path resolved from file '$rawOutputName' + directory '$resolvedOutputDir'."
        }
        else {
            $candidateOutputPath = [System.IO.Path]::GetFullPath((Join-Path $scriptRoot $rawOutput))
            $looksLikeDirectoryInput = $rawOutput -eq "." -or $rawOutput -eq ".." -or $rawOutput.EndsWith("\\") -or $rawOutput.EndsWith("/")
            $candidateIsDirectory = (Test-Path -LiteralPath $candidateOutputPath -PathType Container)

            if ($looksLikeDirectoryInput -or $candidateIsDirectory -or $candidateOutputPath -eq $scriptRoot) {
                $outputPathMap[$singleTarget] = Join-Path $candidateOutputPath $targetCatalog[$singleTarget].DefaultOutput
                Write-Host "Output path resolved to directory '$candidateOutputPath'; using '$($outputPathMap[$singleTarget])'."
            }
            else {
                $outputPathMap[$singleTarget] = $candidateOutputPath
            }
        }
    }
    else {
        if ($hasOutput) {
            throw "When building multiple targets (-Target both), do not use -Output. Use -OutputDir to choose a folder."
        }
        $resolvedOutputDir = if ($hasOutputDir) {
            [System.IO.Path]::GetFullPath((Join-Path $scriptRoot $OutputDir.Trim()))
        } else {
            $scriptRoot
        }
        foreach ($t in $selectedTargets) {
            $outputPathMap[$t] = Join-Path $resolvedOutputDir $targetCatalog[$t].DefaultOutput
        }
    }

    foreach ($targetName in $selectedTargets) {
        $outputPath = $outputPathMap[$targetName]
        if (Test-Path -LiteralPath $outputPath -PathType Container) {
            throw "Output path '$outputPath' resolves to a directory."
        }
        $targetOutputDir = Split-Path -Parent $outputPath
        if (-not [string]::IsNullOrWhiteSpace($targetOutputDir)) {
            New-Item -ItemType Directory -Path $targetOutputDir -Force | Out-Null
        }
    }

    $doCleanBuild = $Clean -or (-not $Incremental)
    if ($doCleanBuild) {
        foreach ($targetName in $selectedTargets) {
            $outputPath = $outputPathMap[$targetName]
            if (-not (Test-Path -LiteralPath $outputPath)) {
                continue
            }
            try {
                $existingArtifact = Get-Item -LiteralPath $outputPath -Force
                if ($existingArtifact.PSIsContainer) {
                    throw "Refusing to remove directory '$outputPath'."
                }
                Remove-Item -LiteralPath $outputPath -Force
                Write-Host "Removed previous artifact: $outputPath"
            }
            catch {
                throw "Unable to remove previous artifact '$outputPath'. Close WS7/WSDEV or any process using the file and try again. $($_.Exception.Message)"
            }
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
    Write-Host "Targets:       $($selectedTargets -join ', ')"
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

    $builtArtifacts = @()
    foreach ($targetName in $selectedTargets) {
        $targetPackage = $targetCatalog[$targetName].Package
        $outputPath = $outputPathMap[$targetName]
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

        Write-Host "Building [$targetName] $targetPackage -> $outputPath"
        & go @buildArgs
        if ($LASTEXITCODE -ne 0) {
            throw "go build failed for target '$targetName'."
        }

        if (-not (Test-Path $outputPath)) {
            throw "Build finished but output was not created: $outputPath"
        }

        $artifact = Get-Item $outputPath
        $builtArtifacts += $artifact
        Write-Host "Build OK: $($artifact.FullName) ($($artifact.Length) bytes)"
    }

    if ($OpenOutputFolder) {
        Write-Host "Opening output folder..."
        $folder = if ($selectedTargets.Count -eq 1) {
            Split-Path -Parent $builtArtifacts[0].FullName
        } else {
            $resolvedOutputDir
        }
        & explorer.exe $folder
    }

    if ($Run) {
        $artifact = $builtArtifacts[0]
        Write-Host "Running $($artifact.Name)..."
        & $artifact.FullName
    }
}
finally {
    Pop-Location
}
