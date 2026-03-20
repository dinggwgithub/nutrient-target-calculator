# 营养素目标计算器 - Windows 构建脚本
# 支持多平台交叉编译：linux/amd64, linux/arm64, windows/amd64

param(
    [Parameter(Mandatory=$false)]
    [string]$Platform = "all",

    [Parameter(Mandatory=$false)]
    [string]$OutputDir = "./build",

    [Parameter(Mandatory=$false)]
    [switch]$Clean
)

# 设置错误处理
$ErrorActionPreference = "Stop"

# 颜色定义
$ColorInfo = "Cyan"
$ColorSuccess = "Green"
$ColorWarning = "Yellow"
$ColorError = "Red"

# 获取版本号（从 git tag 或时间戳）
function Get-Version {
    try {
        $tag = git describe --tags --abbrev=0 2>$null
        if ($tag) {
            return $tag
        }
    } catch {
        # 忽略错误
    }
    return (Get-Date -Format "yyyyMMdd")
}

# 获取 git commit hash
function Get-GitCommit {
    try {
        return (git rev-parse --short HEAD 2>$null)
    } catch {
        return "unknown"
    }
}

# 清理构建目录
function Clear-BuildDirectory {
    param([string]$Dir)
    if (Test-Path $Dir) {
        Write-Host "清理构建目录: $Dir" -ForegroundColor $ColorWarning
        Remove-Item -Path $Dir -Recurse -Force
    }
    New-Item -ItemType Directory -Path $Dir -Force | Out-Null
}

# 构建函数
function Build-Target {
    param(
        [string]$GOOS,
        [string]$GOARCH,
        [string]$OutputName
    )

    Write-Host "========================================" -ForegroundColor $ColorInfo
    Write-Host "开始构建: $GOOS/$GOARCH" -ForegroundColor $ColorInfo
    Write-Host "输出文件: $OutputName" -ForegroundColor $ColorInfo
    Write-Host "========================================" -ForegroundColor $ColorInfo

    # 设置环境变量
    $env:GOOS = $GOOS
    $env:GOARCH = $GOARCH
    $env:CGO_ENABLED = "0"

    # 构建参数
    $version = Get-Version
    $commit = Get-GitCommit
    $buildTime = (Get-Date -Format "yyyy-MM-dd HH:mm:ss")

    $ldflags = @(
        "-s -w",                                    # 去除符号表和调试信息，减小体积
        "-X main.Version=$version",                 # 注入版本号
        "-X main.GitCommit=$commit",                # 注入 Git commit
        "-X main.BuildTime=$buildTime"              # 注入构建时间
    ) -join " "

    $outputPath = Join-Path $OutputDir $OutputName

    try {
        # 执行构建
        go build -ldflags "$ldflags" -o $outputPath

        if ($LASTEXITCODE -eq 0) {
            # 获取文件大小
            $fileInfo = Get-Item $outputPath
            $sizeKB = [math]::Round($fileInfo.Length / 1KB, 2)

            Write-Host "✓ 构建成功!" -ForegroundColor $ColorSuccess
            Write-Host "  文件: $outputPath" -ForegroundColor $ColorSuccess
            Write-Host "  大小: $sizeKB KB" -ForegroundColor $ColorSuccess
            return $true
        } else {
            Write-Host "✗ 构建失败!" -ForegroundColor $ColorError
            return $false
        }
    } catch {
        Write-Host "✗ 构建异常: $_" -ForegroundColor $ColorError
        return $false
    }
}

# 主函数
function Main {
    Write-Host @"
╔══════════════════════════════════════════════════════════════╗
║         营养素目标计算器 - 多平台构建脚本                     ║
║         Nutrient Target Calculator Build Script              ║
╚══════════════════════════════════════════════════════════════╝
"@ -ForegroundColor $ColorInfo

    # 显示构建信息
    Write-Host "版本: $(Get-Version)" -ForegroundColor $ColorInfo
    Write-Host "Commit: $(Get-GitCommit)" -ForegroundColor $ColorInfo
    Write-Host "构建时间: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")" -ForegroundColor $ColorInfo
    Write-Host "输出目录: $OutputDir" -ForegroundColor $ColorInfo
    Write-Host ""

    # 清理
    if ($Clean) {
        Clear-BuildDirectory -Dir $OutputDir
    } else {
        if (-not (Test-Path $OutputDir)) {
            New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
        }
    }

    # 检查 Go 环境
    try {
        $goVersion = go version
        Write-Host "Go 版本: $goVersion" -ForegroundColor $ColorInfo
    } catch {
        Write-Host "错误: 未找到 Go 环境，请确保 Go 已安装并添加到 PATH" -ForegroundColor $ColorError
        exit 1
    }

    # 下载依赖
    Write-Host "下载依赖..." -ForegroundColor $ColorInfo
    go mod download
    if ($LASTEXITCODE -ne 0) {
        Write-Host "错误: 依赖下载失败" -ForegroundColor $ColorError
        exit 1
    }

    # 执行构建
    $buildResults = @()

    switch ($Platform.ToLower()) {
        "linux/amd64" {
            $result = Build-Target -GOOS "linux" -GOARCH "amd64" -OutputName "target-calculator-linux-amd64"
            $buildResults += [PSCustomObject]@{ Platform = "linux/amd64"; Success = $result }
        }
        "linux/arm64" {
            $result = Build-Target -GOOS "linux" -GOARCH "arm64" -OutputName "target-calculator-linux-arm64"
            $buildResults += [PSCustomObject]@{ Platform = "linux/arm64"; Success = $result }
        }
        "windows/amd64" {
            $result = Build-Target -GOOS "windows" -GOARCH "amd64" -OutputName "target-calculator-windows-amd64.exe"
            $buildResults += [PSCustomObject]@{ Platform = "windows/amd64"; Success = $result }
        }
        "all" {
            $result1 = Build-Target -GOOS "linux" -GOARCH "amd64" -OutputName "target-calculator-linux-amd64"
            $buildResults += [PSCustomObject]@{ Platform = "linux/amd64"; Success = $result1 }

            $result2 = Build-Target -GOOS "linux" -GOARCH "arm64" -OutputName "target-calculator-linux-arm64"
            $buildResults += [PSCustomObject]@{ Platform = "linux/arm64"; Success = $result2 }

            $result3 = Build-Target -GOOS "windows" -GOARCH "amd64" -OutputName "target-calculator-windows-amd64.exe"
            $buildResults += [PSCustomObject]@{ Platform = "windows/amd64"; Success = $result3 }
        }
        default {
            Write-Host "错误: 不支持的平台 '$Platform'" -ForegroundColor $ColorError
            Write-Host "支持的平台: linux/amd64, linux/arm64, windows/amd64, all" -ForegroundColor $ColorWarning
            exit 1
        }
    }

    # 构建结果汇总
    Write-Host ""
    Write-Host "========================================" -ForegroundColor $ColorInfo
    Write-Host "构建结果汇总" -ForegroundColor $ColorInfo
    Write-Host "========================================" -ForegroundColor $ColorInfo

    $successCount = 0
    $failCount = 0

    foreach ($result in $buildResults) {
        $status = if ($result.Success) { "✓ 成功" } else { "✗ 失败" }
        $color = if ($result.Success) { $ColorSuccess } else { $ColorError }
        Write-Host "$($result.Platform): $status" -ForegroundColor $color

        if ($result.Success) { $successCount++ } else { $failCount++ }
    }

    Write-Host ""
    Write-Host "成功: $successCount, 失败: $failCount" -ForegroundColor $ColorInfo

    # 列出输出文件
    Write-Host ""
    Write-Host "输出文件列表:" -ForegroundColor $ColorInfo
    Get-ChildItem $OutputDir | ForEach-Object {
        $size = [math]::Round($_.Length / 1KB, 2)
        Write-Host "  $($_.Name) (${size}KB)" -ForegroundColor $ColorSuccess
    }

    # 生成部署脚本
    $deployScript = @"
#!/bin/bash
# 营养素目标计算器部署脚本
# 生成时间: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")

APP_NAME="target-calculator"
VERSION="$(Get-Version)"

echo "开始部署 \$APP_NAME v\$VERSION"

# 检查系统架构
ARCH=\$(uname -m)
if [ "\$ARCH" = "x86_64" ]; then
    BINARY="target-calculator-linux-amd64"
elif [ "\$ARCH" = "aarch64" ]; then
    BINARY="target-calculator-linux-arm64"
else
    echo "不支持的架构: \$ARCH"
    exit 1
fi

echo "检测到架构: \$ARCH, 使用二进制: \$BINARY"

# 设置权限
chmod +x \$BINARY

# 启动服务（可根据需要修改为 systemd 方式）
echo "启动服务..."
./\$BINARY
"@

    $deployScriptPath = Join-Path $OutputDir "deploy.sh"
    $deployScript | Out-File -FilePath $deployScriptPath -Encoding UTF8
    Write-Host ""
    Write-Host "已生成部署脚本: $deployScriptPath" -ForegroundColor $ColorSuccess

    if ($failCount -gt 0) {
        exit 1
    }
}

# 执行主函数
Main
