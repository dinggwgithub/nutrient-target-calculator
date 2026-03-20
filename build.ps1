param(
    [string]$Os = "",
    [string]$Arch = "",
    [switch]$BuildAll = $false,
    [string]$OutputDir = "build",
    [string]$AppName = "target-calculator"
)

$ErrorActionPreference = "Stop"

function Build-App {
    param(
        [string]$TargetOs,
        [string]$TargetArch,
        [string]$OutputPath
    )

    Write-Host "Building for $TargetOs/$TargetArch..." -ForegroundColor Cyan

    $env:GOOS = $TargetOs
    $env:GOARCH = $TargetArch

    $buildTime = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $ldflags = "-s -w -X main.BuildTime=$buildTime"
    $ldflags = "-s -w"

    go build -ldflags $ldflags -o $OutputPath .

    if ($LASTEXITCODE -eq 0) {
        Write-Host "  Success: $OutputPath" -ForegroundColor Green
    } else {
        Write-Host "  Failed to build for $TargetOs/$TargetArch" -ForegroundColor Red
        exit 1
    }
}

function Main {
    Write-Host "========================================" -ForegroundColor Yellow
    Write-Host "  Target Calculator Build Script" -ForegroundColor Yellow
    Write-Host "========================================" -ForegroundColor Yellow
    Write-Host ""

    if (-not (Test-Path $OutputDir)) {
        New-Item -ItemType Directory -Path $OutputDir | Out-Null
        Write-Host "Created output directory: $OutputDir" -ForegroundColor Gray
    }

    if ($BuildAll) {
        Write-Host "Building all platforms..." -ForegroundColor Cyan
        Write-Host ""

        Build-App -TargetOs "linux" -TargetArch "amd64" -OutputPath "$OutputDir/$AppName-linux-amd64"
        Build-App -TargetOs "linux" -TargetArch "arm64" -OutputPath "$OutputDir/$AppName-linux-arm64"
        Build-App -TargetOs "windows" -TargetArch "amd64" -OutputPath "$OutputDir/$AppName-windows-amd64.exe"

        Write-Host ""
        Write-Host "All builds completed!" -ForegroundColor Green
        Write-Host ""
        Write-Host "Output files:" -ForegroundColor Yellow
        Get-ChildItem -Path $OutputDir -Filter "$AppName-*" | ForEach-Object {
            $size = $_.Length / 1MB
            Write-Host "  $($_.Name) ($('{0:N2}' -f $size) MB)"
        }
    }
    elseif ($Os -ne "" -and $Arch -ne "") {
        $extension = if ($Os -eq "windows") { ".exe" } else { "" }
        $outputPath = "$OutputDir/$AppName-$Os-$Arch$extension"

        Build-App -TargetOs $Os -TargetArch $Arch -OutputPath $outputPath

        Write-Host ""
        Write-Host "Build completed!" -ForegroundColor Green
        $size = (Get-Item $outputPath).Length / 1MB
        Write-Host "Output: $outputPath ($('{0:N2}' -f $size) MB)"
    }
    else {
        Write-Host "Building for current platform..." -ForegroundColor Cyan
        Write-Host ""

        $currentOs = $env:GOOS
        $currentArch = $env:GOARCH

        if (-not $currentOs) {
            $currentOs = if ($IsWindows -or $env:OS -match "Windows") { "windows" } else { "linux" }
        }
        if (-not $currentArch) {
            $currentArch = "amd64"
        }

        $extension = if ($currentOs -eq "windows") { ".exe" } else { "" }
        $outputPath = "$OutputDir/$AppName$extension"

        Build-App -TargetOs $currentOs -TargetArch $currentArch -OutputPath $outputPath

        Write-Host ""
        Write-Host "Build completed!" -ForegroundColor Green
        $size = (Get-Item $outputPath).Length / 1MB
        Write-Host "Output: $outputPath ($('{0:N2}' -f $size) MB)"
    }

    Write-Host ""
    Write-Host "Usage examples:" -ForegroundColor Yellow
    Write-Host "  .\build.ps1                    # Build for current platform"
    Write-Host "  .\build.ps1 -BuildAll          # Build for all platforms"
    Write-Host "  .\build.ps1 -Os linux -Arch amd64   # Build for Linux/amd64"
    Write-Host "  .\build.ps1 -Os linux -Arch arm64   # Build for Linux/arm64"
    Write-Host "  .\build.ps1 -Os windows -Arch amd64 # Build for Windows"
}

Main
