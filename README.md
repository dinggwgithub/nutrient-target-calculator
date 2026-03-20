# 营养素目标计算器 (target-calculator-from-db)

基于数据库摄入数据和DRIs计算营养素目标摄入量的Go+Gin API服务。支持Redis缓存、历史记录查询和版本对比功能。

## 功能特性

### 核心功能
- 根据人群特征（性别、年龄、人群类型）查询营养素摄入分布数据
- 支持三种计算场景（A/B/C）：
  - **场景A**：高目标，达到RNI/AI以保证97.5%的人群达到EAR
  - **场景B**：中等目标，保证50%的人群达到EAR
  - **场景C**：保守目标，基于原始分布的轻微调整
- 计算目标中位数和P95百分位数
- 检查P95是否超过可耐受最高摄入量(UL)并发出警告

### 新增功能 (v2.0)
- ✅ **Redis自动存储** - 计算结果自动保存到Redis，TTL=1天
- ✅ **历史记录查询** - 支持分页查询计算历史，按用户过滤
- ✅ **版本对比** - 支持两个计算版本的参数和结果差异对比
- ✅ **高性能** - 所有热点数据存储在Redis，不落地数据库
- ✅ **多平台交叉编译** - 支持Windows下构建Linux/amd64、Linux/arm64版本
- ✅ **完整的Swagger API文档** - 所有接口包含示例

## 技术栈

- **Go 1.21+**
- **Gin Web Framework** - HTTP Web框架
- **GORM** - ORM数据库操作
- **MySQL** - 基础数据存储
- **Redis** - 计算结果缓存和历史记录
- **Swaggo** - API文档生成

## 项目结构

```
nutrient-target-calculator/
├── main.go                 # 应用入口
├── config/
│   ├── database.go         # 数据库配置
│   └── redis.go            # Redis配置
├── models/
│   └── nutrient.go         # 数据模型
├── dto/
│   └── request.go          # 请求响应DTO
├── service/
│   ├── calculator.go       # 计算业务逻辑
│   └── redis_service.go    # Redis服务逻辑
├── controller/
│   └── target.go           # HTTP控制器
├── routes/
│   └── routes.go           # 路由配置
├── docs/                   # Swagger文档（自动生成）
├── test/
│   └── request_test.http   # API测试请求
├── build.ps1               # Windows构建脚本
└── go.mod
```

## Redis存储设计

### Key设计

| Key类型 | Key模式 | 说明 | TTL |
|---------|---------|------|-----|
| String | `calc:result:{version_id}` | 存储完整的计算结果JSON | 24小时 |
| ZSet | `calc:history:{user_id}` | 用户的计算历史索引（按时间排序） | 25小时 |
| ZSet | `calc:index` | 全局计算历史索引 | 25小时 |

### 存储结构

```json
{
  "version_id": "calc_abc123def456",
  "request": {
    "gender": "男",
    "age": 30,
    "crowd": "普通人群",
    "nutrient_name": "维生素C",
    "scenario": "A",
    "user_id": "user_123"
  },
  "result": {
    "nutrient_name": "维生素C",
    "original_mean": 85.5,
    "target_median": 100.0,
    "...": "..."
  },
  "created_at": "2024-01-01T12:00:00Z",
  "expire_at": "2024-01-02T12:00:00Z"
}
```

## 数据库表

系统依赖以下数据表：

1. **nutrient_average** - 营养素平均摄入量表
2. **nutrient_variation** - 营养素变异系数表
3. **dris_references** - DRIs参考值表

## 环境依赖

| 依赖 | 版本要求 | 说明 |
|------|----------|------|
| Go | 1.21+ | 编译运行 |
| MySQL | 8.0+ | 基础数据存储 |
| Redis | 5.0+ | 计算结果缓存 |

## 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 生成Swagger文档

```bash
swag init
```

### 3. 编译运行

```bash
go build -o target-calculator.exe
./target-calculator.exe
```

或直接运行：

```bash
go run main.go
```

### 4. 访问API

服务启动后，访问以下地址：

- **计算接口**: `http://localhost:8080/api/target/calculate`
- **历史列表**: `http://localhost:8080/api/target/history`
- **Swagger文档**: `http://localhost:8080/swagger/index.html`
- **健康检查**: `http://localhost:8080/api/health`

## 交叉编译部署

### Windows下构建Linux版本

使用提供的 `build.ps1` 脚本：

```powershell
# 构建Linux/amd64版本（默认）
.\build.ps1 -Platform linux

# 构建Linux/arm64版本（适用于ARM服务器、树莓派等）
.\build.ps1 -Platform linux-arm64

# 构建所有平台版本
.\build.ps1 -Platform all

# 指定输出目录和版本号
.\build.ps1 -Platform linux -OutputDir ./release -Version v2.0.0
```

### 手动构建

```powershell
# Windows PowerShell环境下构建Linux/amd64
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
go build -ldflags "-s -w" -o target-calculator-linux-amd64 .

# 构建Linux/arm64
$env:GOARCH = "arm64"
go build -ldflags "-s -w" -o target-calculator-linux-arm64 .

# 重置环境变量
$env:GOOS = $null
$env:GOARCH = $null
$env:CGO_ENABLED = $null
```

### Linux部署

```bash
# 上传构建好的二进制文件
scp target-calculator-linux-amd64 user@server:/opt/

# 添加执行权限
chmod +x /opt/target-calculator-linux-amd64

# 创建配置（如有需要）
# 确保服务器已安装Redis和MySQL

# 启动服务
nohup /opt/target-calculator-linux-amd64 &> /var/log/target-calculator.log &
```

## API接口说明

### 1. POST /api/target/calculate

计算营养素目标摄入量，结果自动保存到Redis。

**请求参数**:

```json
{
    "gender": "男",           // 性别：男/女
    "age": 30,               // 年龄：0-120
    "crowd": "普通人群",      // 人群类型
    "nutrient_name": "维生素C", // 营养素名称
    "scenario": "A",         // 计算场景：A/B/C
    "user_id": "user_123"    // 可选：用户标识，用于关联历史记录
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
        "original_cv": 0.25,
        "target_median": 100.0,
        "target_p95": 168.5,
        "ul": 200.0,
        "exceed_ul": false,
        "warning": "",
        "unit": "mg/d",
        "adjustment_factor": 1.17
    },
    "version_id": "calc_abc123def456"  // 用于后续查询和对比
}
```

### 2. GET /api/target/history

获取计算历史记录列表（分页）。

**查询参数**:
- `user_id` (可选): 用户ID，过滤特定用户的记录
- `page` (可选，默认1): 页码
- `page_size` (可选，默认10，最大100): 每页条数

**响应示例**:
```json
{
    "code": 200,
    "message": "获取成功",
    "total": 156,
    "page": 1,
    "page_size": 10,
    "data": [
        {
            "version_id": "calc_abc123def456",
            "nutrient_name": "维生素C",
            "gender": "男",
            "age": 30,
            "crowd": "普通人群",
            "scenario": "A",
            "created_at": "2024-01-01T12:00:00Z",
            "target_median": 100.0,
            "unit": "mg/d"
        }
    ]
}
```

### 3. GET /api/target/history/{version_id}

获取单个历史记录详情。

**路径参数**:
- `version_id`: 计算结果版本ID

**响应示例**:
```json
{
    "code": 200,
    "message": "获取成功",
    "data": {
        "version_id": "calc_abc123def456",
        "request": {...},
        "result": {...},
        "created_at": "2024-01-01T12:00:00Z",
        "expire_at": "2024-01-02T12:00:00Z"
    }
}
```

### 4. POST /api/target/compare

对比两个历史版本的差异。

**请求参数**:
```json
{
    "version_id1": "calc_abc123def456",
    "version_id2": "calc_def789ghi012"
}
```

**响应示例**:
```json
{
    "code": 200,
    "message": "对比成功",
    "version1": {...},
    "version2": {...},
    "diff_summary": {
        "total_fields": 12,
        "changed_fields": 3,
        "increased_fields": 1,
        "decreased_fields": 2
    },
    "diffs": [
        {
            "field": "age",
            "description": "年龄",
            "value1": 30,
            "value2": 40,
            "diff": 10,
            "diff_percent": "33.33%",
            "change_type": "increase"
        },
        {
            "field": "target_median",
            "description": "目标中位数",
            "value1": 100.0,
            "value2": 120.0,
            "diff": 20.0,
            "diff_percent": "20.00%",
            "change_type": "increase"
        }
    ]
}
```

## 计算算法说明

### 摄入分布模型

假设营养素摄入量服从**对数正态分布**，基于以下参数计算：

- **均值(Mean)**: 从 `nutrient_average` 表获取
- **变异系数(CV)**: 从 `nutrient_variation` 表获取
- **标准差(SD)**: `SD = Mean * CV`

### 百分位数计算 (P95)

使用对数正态分布的95百分位数计算公式：

```
σ = √(ln(1 + CV²))
μ = ln(Mean) - σ² / 2
P95 = exp(μ + 1.645 * σ)
```

### 场景计算逻辑

| 场景 | 目标中位数计算 | 说明 |
|------|--------------|------|
| A | `targetMedian = RNI/AI` | 达到推荐摄入量，保证97.5%人群满足需求 |
| B | `targetMedian = (mean + EAR) / 2` | 中等目标，保证50%人群满足需求 |
| C | `targetMedian = max(mean, EAR*0.8)` | 保守目标，保证平均水平达到EAR的80% |

## 错误处理

API返回标准的HTTP状态码：

- `200` - 成功
- `400` - 请求参数错误
- `500` - 服务器内部错误（如数据库查询失败）

错误响应格式：

```json
{
    "code": 400,
    "message": "参数验证失败: ...",
    "data": null
}
```

## 开发说明

### 更新Swagger文档

修改API注释后，需要重新生成文档：

```bash
swag init
```

### 数据库配置

数据库连接配置在 `config/database.go` 中：

```go
dsn := "root:123456@tcp(127.0.0.1:3306)/recipe_system?charset=utf8mb4&parseTime=True&loc=Local"
```

## 许可证

MIT
