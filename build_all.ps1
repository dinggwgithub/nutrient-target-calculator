# 营养素目标计算器API - 多平台构建脚本
# 支持: linux/amd64, linux/arm64, windows/amd64

$ErrorActionPreference = "Stop"

# 项目配置
$ProjectName = "target-calculator-api"
$OutputDir = "./bin"
$Version = "v2.0.0"

# 检测Git commit
$GitCommit = git rev-parse --short HEAD 2>$null
if ($LASTEXITCODE -ne 0) {
    $GitCommit = "unknown"
}

# 确保输出目录存在
if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "项目名称: $ProjectName" -ForegroundColor Cyan
Write-Host "版本号: $Version" -ForegroundColor Cyan
Write-Host "Git Commit: $GitCommit" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# 定义要构建的平台
$platforms = @(
    @{ GOOS = "linux"; GOARCH = "amd64"; Ext = "" },
    @{ GOOS = "linux"; GOARCH = "arm64"; Ext = "" },
    @{ GOOS = "windows"; GOARCH = "amd64"; Ext = ".exe" }
)

foreach ($p in $platforms) {
    $env:GOOS = $p.GOOS
    $env:GOARCH = $p.GOARCH
    $env:CGO_ENABLED = "0"
    
    $OutputName = "$ProjectName-$Version-$($p.GOOS)-$($p.GOARCH)$($p.Ext)"
    $OutputPath = Join-Path $OutputDir $OutputName
    
    Write-Host "`n正在构建: $OutputName" -ForegroundColor Green
    
    go build -ldflags "-s -w" -o $OutputPath .
    
    if ($LASTEXITCODE -eq 0) {
        $fileSize = [math]::Round((Get-Item $OutputPath).Length / 1MB, 2)
        Write-Host "构建成功: $OutputPath ($fileSize MB)" -ForegroundColor Green
    } else {
        Write-Host "构建失败: $OutputName" -ForegroundColor Red
        exit 1
    }
}

# 重置环境变量
$env:GOOS = $null
$env:GOARCH = $null
$env:CGO_ENABLED = $null

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "所有平台构建完成!" -ForegroundColor Green
Write-Host "输出目录: $OutputDir" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# 列出构建产物
Write-Host "`n构建产物列表:" -ForegroundColor Cyan
Get-ChildItem $OutputDir | ForEach-Object {
    $size = [math]::Round($_.Length / 1MB, 2)
    Write-Host "  - $($_.Name) ($size MB)" -ForegroundColor Gray
}
