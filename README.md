# 营养素目标计算器 (target-calculator-from-db)

基于数据库摄入数据和DRIs计算营养素目标摄入量的Go+Gin API服务。支持计算结果自动保存到Redis，提供历史记录查询和版本对比功能。

## 功能特性

- **营养素目标计算**：根据人群特征（性别、年龄、人群类型）查询营养素摄入分布数据
- **三种计算场景**（A/B/C）：
  - **场景A**：高目标，达到RNI/AI以保证97.5%的人群达到EAR
  - **场景B**：中等目标，保证50%的人群达到EAR
  - **场景C**：保守目标，基于原始分布的轻微调整
- **计算结果自动保存**：每次计算结果自动存储到Redis，TTL=1天
- **历史记录管理**：
  - 查询用户历史计算记录列表
  - 查看单条历史记录详情（含剩余过期时间）
  - 删除历史记录
- **版本对比功能**：对比两个历史版本的差异，显示字段变化详情
- **高性能**：所有历史数据存储在Redis，不落数据库

## 技术栈

- **Go 1.21+**
- **Gin Web Framework** - HTTP Web框架
- **GORM** - ORM数据库操作
- **MySQL** - 业务数据存储
- **Redis** - 历史记录缓存（TTL=1天）
- **Swaggo** - API文档生成

## 项目结构

```
target-calculator-from-db/
├── main.go                 # 应用入口
├── config/
│   ├── database.go         # 数据库配置
│   └── redis.go            # Redis配置
├── models/
│   └── nutrient.go         # 数据模型
├── dto/
│   ├── request.go          # 请求响应DTO
│   └── history.go          # 历史记录DTO
├── service/
│   ├── calculator.go       # 业务逻辑服务
│   └── history.go          # 历史记录服务
├── controller/
│   ├── target.go           # 计算控制器
│   └── history.go          # 历史记录控制器
├── routes/
│   └── routes.go           # 路由配置
├── docs/                   # Swagger文档（自动生成）
├── test/
│   └── request_test.http   # API测试请求
├── build.ps1               # Windows构建脚本
├── go.mod
└── README.md
```

## Redis 存储结构设计

### Key 命名规范

| Key 类型 | 格式 | 示例 |
|---------|------|------|
| 计算结果 | `calc:{user_id}:{timestamp}:{uuid}` | `calc:user123:1710900000:abc123` |
| 用户历史索引 | `history:{user_id}` | `history:user123` |

### 数据结构

**1. 计算结果（String，JSON格式）**
```json
{
  "id": "uuid",
  "user_id": "user123",
  "timestamp": 1710900000,
  "created_at": "2024-03-20 10:00:00",
  "request": { /* 原始请求参数 */ },
  "result": { /* 计算结果 */ },
  "ttl_seconds": 86400
}
```

**2. 用户历史索引（Sorted Set）**
- Key: `history:{user_id}`
- Score: 时间戳（用于排序）
- Member: 计算结果的 Key

### 过期策略

- **TTL**: 所有计算结果和用户历史索引统一设置 **24小时（1天）** 过期时间
- **自动清理**: Redis 自动删除过期数据，无需手动维护

## 快速开始

### 1. 环境要求

- Go 1.21+
- MySQL 8.0+
- Redis 6.0+

### 2. 安装依赖

```bash
go mod download
```

### 3. 配置数据库

编辑 `config/database.go` 和 `config/redis.go` 中的连接配置（默认已配置为本地开发环境）。

### 4. 生成Swagger文档

```bash
swag init
```

### 5. 编译运行

```bash
go build -o target-calculator.exe
./target-calculator.exe
```

或直接运行：

```bash
go run main.go
```

### 6. 访问API

服务启动后，访问以下地址：

- **Swagger文档**: `http://localhost:8080/swagger/index.html`
- **健康检查**: `http://localhost:8080/api/health`

## API接口说明

### 1. 计算营养素目标

**POST** `/api/target/calculate`

计算营养素目标摄入量，结果自动保存到Redis。

**请求参数**:
```json
{
    "gender": "男",           // 性别：男/女
    "age": 30,                // 年龄：0-120
    "crowd": "普通人群",       // 人群类型
    "nutrient_name": "维生素C", // 营养素名称
    "scenario": "A"           // 场景：A/B/C
}
```

**响应示例**:
```json
{
    "code": 200,
    "message": "计算成功",
    "data": {
        "nutrient_name": "维生素C",
        "original_mean": 85.5,
        "original_cv": 0.31,
        "target_median": 100.0,
        "target_p95": 156.8,
        "ul": 2000.0,
        "exceed_ul": false,
        "warning": "",
        "unit": "mg",
        "adjustment_factor": 1.17
    }
}
```

### 2. 获取历史记录列表

**GET** `/api/target/history?user_id={user_id}&limit={limit}`

获取指定用户的计算历史记录列表，按时间倒序排列。

**响应示例**:
```json
{
    "code": 200,
    "message": "获取成功",
    "data": {
        "total": 2,
        "records": [
            {
                "id": "abc123",
                "key": "calc:user123:1710900100:abc123",
                "timestamp": 1710900100,
                "created_at": "2024-03-20 10:01:40",
                "nutrient": "维生素C",
                "scenario": "B",
                "target_median": 92.75,
                "unit": "mg"
            },
            {
                "id": "def456",
                "key": "calc:user123:1710900000:def456",
                "timestamp": 1710900000,
                "created_at": "2024-03-20 10:00:00",
                "nutrient": "维生素C",
                "scenario": "A",
                "target_median": 100.0,
                "unit": "mg"
            }
        ]
    }
}
```

### 3. 获取历史记录详情

**GET** `/api/target/history/detail?key={key}`

根据 Redis Key 获取单条历史记录的完整详情，包含剩余过期时间。

**响应示例**:
```json
{
    "code": 200,
    "message": "获取成功",
    "data": {
        "id": "abc123",
        "user_id": "user123",
        "timestamp": 1710900000,
        "created_at": "2024-03-20 10:00:00",
        "ttl_seconds": 85000,
        "request": {
            "gender": "男",
            "age": 30,
            "crowd": "普通人群",
            "nutrient_name": "维生素C",
            "scenario": "A"
        },
        "result": {
            "nutrient_name": "维生素C",
            "original_mean": 85.5,
            "original_cv": 0.31,
            "target_median": 100.0,
            "target_p95": 156.8,
            "ul": 2000.0,
            "exceed_ul": false,
            "warning": "",
            "unit": "mg",
            "adjustment_factor": 1.17
        }
    }
}
```

### 4. 对比两个历史版本

**POST** `/api/target/history/compare`

对比两个历史计算记录的差异，返回详细的字段对比信息。

**请求参数**:
```json
{
    "key1": "calc:user123:1710900000:abc123",
    "key2": "calc:user123:1710900100:def456"
}
```

**响应示例**:
```json
{
    "code": 200,
    "message": "对比成功",
    "data": {
        "record1": { /* 第一个记录完整数据 */ },
        "record2": { /* 第二个记录完整数据 */ },
        "differences": [
            {
                "field": "request.scenario",
                "field_name": "场景",
                "old_value": "A",
                "new_value": "B",
                "change_type": "changed"
            },
            {
                "field": "result.target_median",
                "field_name": "目标中位数",
                "old_value": 100.0,
                "new_value": 92.75,
                "change_type": "decreased",
                "change_percent": -7.25
            },
            {
                "field": "result.target_p95",
                "field_name": "目标P95值",
                "old_value": 156.8,
                "new_value": 145.3,
                "change_type": "decreased",
                "change_percent": -7.34
            }
        ],
        "summary": {
            "total_fields": 11,
            "changed_fields": 3,
            "increased_count": 0,
            "decreased_count": 2,
            "main_change": "共有 3 个字段发生变化，2 个指标下降"
        }
    }
}
```

### 5. 删除历史记录

**DELETE** `/api/target/history`

根据 Redis Key 删除指定的历史记录。

**请求参数**:
```json
{
    "key": "calc:user123:1710900000:abc123"
}
```

**响应示例**:
```json
{
    "code": 200,
    "message": "删除成功"
}
```

## 构建生产环境

### Windows 下构建 Linux 版本

项目提供了 `build.ps1` PowerShell 脚本，支持在 Windows 环境下交叉编译 Linux 可执行文件。

#### 使用方法

```powershell
# 赋予执行权限（首次使用）
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# 执行构建脚本
.\build.ps1

# 或构建特定平台
.\build.ps1 -Platform linux/amd64
.\build.ps1 -Platform linux/arm64
.\build.ps1 -Platform windows/amd64
```

#### 支持的构建目标

| 平台 | 架构 | 输出文件名 |
|------|------|-----------|
| Linux | AMD64 | `target-calculator-linux-amd64` |
| Linux | ARM64 | `target-calculator-linux-arm64` |
| Windows | AMD64 | `target-calculator-windows-amd64.exe` |

#### 手动交叉编译

```bash
# Linux AMD64
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -ldflags="-s -w" -o target-calculator-linux-amd64

# Linux ARM64
$env:GOOS="linux"; $env:GOARCH="arm64"; go build -ldflags="-s -w" -o target-calculator-linux-arm64

# Windows AMD64（默认）
go build -ldflags="-s -w" -o target-calculator.exe
```

### 部署到 Linux 服务器

```bash
# 1. 上传可执行文件
scp target-calculator-linux-amd64 user@server:/opt/app/

# 2. 登录服务器设置权限
ssh user@server
chmod +x /opt/app/target-calculator-linux-amd64

# 3. 使用 systemd 管理服务（可选）
sudo systemctl start target-calculator
```

## 数据库表

系统依赖以下数据表：

1. **nutrient_average** - 营养素平均摄入量表
2. **nutrient_variation** - 营养素变异系数表
3. **dris_references** - DRIs参考值表

## 开发说明

### 添加新的营养素

在数据库中添加对应的 DRIs 参考值和摄入数据即可，无需修改代码。

### 修改 Redis TTL

编辑 `config/redis.go` 中的 `CalculationHistoryTTL` 常量：

```go
const (
    // 计算历史记录过期时间
    CalculationHistoryTTL = 24 * time.Hour  // 修改为需要的时长
)
```

### 自定义用户标识

当前版本使用请求头 `X-User-ID` 识别用户，未提供时默认为 `anonymous`。生产环境建议集成认证系统。

## 许可证

MIT License
