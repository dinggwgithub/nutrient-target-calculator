$Platform = "linux"
$OutputDir = "./bin"
$Version = ""

$ProjectName = "target-calculator-api"
$MainPackage = "."

# Generate version
if ([string]::IsNullOrEmpty($Version)) {
    $Version = "v2.0.0-" + (Get-Date -Format "yyyyMMdd")
}

$GitCommit = git rev-parse --short HEAD 2>$null
if ($LASTEXITCODE -ne 0) {
    $GitCommit = "unknown"
}

# Create output directory
if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
}

Write-Host "========================================"
Write-Host "Project: $ProjectName"
Write-Host "Version: $Version"
Write-Host "Git Commit: $GitCommit"
Write-Host "========================================"

# Build function
function Build-Target {
    param(
        [string]$GOOS,
        [string]$GOARCH,
        [string]$OutputName
    )
    
    Write-Host "`nBuilding: $OutputName ($GOOS/$GOARCH)"
    
    $env:GOOS = $GOOS
    $env:GOARCH = $GOARCH
    $env:CGO_ENABLED = "0"
    
    $LDFLAGS = "-s -w -X main.Version=$Version -X main.GitCommit=$GitCommit"
    $OutputPath = Join-Path $OutputDir $OutputName
    
    go build -ldflags $LDFLAGS -o $OutputPath $MainPackage
    
    if ($LASTEXITCODE -eq 0) {
        $fileSize = [math]::Round((Get-Item $OutputPath).Length / 1MB, 2)
        Write-Host "Build success: $OutputPath ($fileSize MB)"
    } else {
        Write-Host "Build failed!"
        exit 1
    }
}

# Build based on platform
Build-Target -GOOS "linux" -GOARCH "amd64" -OutputName "$ProjectName-$Version-linux-amd64"

# Reset env
$env:GOOS = $null
$env:GOARCH = $null
$env:CGO_ENABLED = $null

Write-Host "`n========================================"
Write-Host "Build completed!"
Write-Host "Output directory: $OutputDir"
Write-Host "========================================"

# List outputs
Write-Host "`nOutput files:"
Get-ChildItem $OutputDir | ForEach-Object {
    $size = [math]::Round($_.Length / 1MB, 2)
    Write-Host "  - $($_.Name) ($size MB)"
}
