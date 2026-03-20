# 营养素目标计算器 (target-calculator-from-db)

基于数据库摄入数据和DRIs计算营养素目标摄入量的Go+Gin API服务。

## 功能特性

- 根据人群特征（性别、年龄、人群类型）查询营养素摄入分布数据
- 支持三种计算场景（A/B/C）：
  - **场景A**：高目标，达到RNI/AI以保证97.5%的人群达到EAR
  - **场景B**：中等目标，保证50%的人群达到EAR
  - **场景C**：保守目标，基于原始分布的轻微调整
- 计算目标中位数和P95百分位数
- 检查P95是否超过可耐受最高摄入量(UL)并发出警告
- 完整的Swagger API文档支持

## 技术栈

- **Go 1.21+**
- **Gin Web Framework** - HTTP Web框架
- **GORM** - ORM数据库操作
- **MySQL** - 数据存储
- **Swaggo** - API文档生成

## 项目结构

```
nutrient-target-calculator/
├── main.go                 # 应用入口
├── config/
│   └── database.go         # 数据库配置
├── models/
│   └── nutrient.go         # 数据模型
├── dto/
│   └── request.go          # 请求响应DTO
├── service/
│   └── calculator.go       # 业务逻辑服务
├── controller/
│   └── target.go           # HTTP控制器
├── routes/
│   └── routes.go           # 路由配置
├── docs/                   # Swagger文档（自动生成）
├── test/
│   └── request_test.http   # API测试请求
└── go.mod
```

## 数据库表

系统依赖以下数据表：

1. **nutrient_average** - 营养素平均摄入量表
2. **nutrient_variation** - 营养素变异系数表
3. **dris_references** - DRIs参考值表

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

- **API端点**: `http://localhost:8080/api/target/calculate`
- **Swagger文档**: `http://localhost:8080/swagger/index.html`
- **健康检查**: `http://localhost:8080/api/health`

## API接口说明

### POST /api/target/calculate

计算营养素目标摄入量。

**请求参数**:

```json
{
    "gender": "男",           // 性别：男/女
    "age": 30,               // 年龄：0-120
    "crowd": "普通人群",      // 人群类型
    "nutrient_name": "维生素C", // 营养素名称
    "scenario": "A"          // 计算场景：A/B/C
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
    }
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
