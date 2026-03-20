package dto

import "time"

// CalculateTargetRequest 计算目标请求
type CalculateTargetRequest struct {
	Gender       string `json:"gender" binding:"required,oneof=男 女" example:"男"`
	Age          int    `json:"age" binding:"required,min=0,max=120" example:"30"`
	Crowd        string `json:"crowd" binding:"required" example:"普通人群"`
	NutrientName string `json:"nutrient_name" binding:"required" example:"维生素C"`
	Scenario     string `json:"scenario" binding:"required,oneof=A B C" example:"A"`
	UserID       string `json:"user_id,omitempty" example:"user_123"` // 可选用户标识，用于关联历史记录
}

// CalculateTargetResponse 计算目标响应
type CalculateTargetResponse struct {
	Code      int        `json:"code"`
	Message   string     `json:"message"`
	Data      TargetData `json:"data"`
	VersionID string     `json:"version_id,omitempty"` // 计算结果的版本ID（用于后续查询和对比）
}

// TargetData 目标数据
type TargetData struct {
	NutrientName     string  `json:"nutrient_name"`
	OriginalMean     float64 `json:"original_mean"`     // 原始平均摄入量
	OriginalCV       float64 `json:"original_cv"`       // 原始变异系数
	TargetMedian     float64 `json:"target_median"`     // 目标中位数
	TargetP95        float64 `json:"target_p95"`        // 目标P95值
	UL               float64 `json:"ul"`                // UL值
	ExceedUL         bool    `json:"exceed_ul"`         // 是否超过UL
	Warning          string  `json:"warning"`           // 警告信息
	Unit             string  `json:"unit"`              // 单位
	AdjustmentFactor float64 `json:"adjustment_factor"` // 调整因子
}

// CalculationResult Redis中存储的完整计算结果
type CalculationResult struct {
	VersionID   string                 `json:"version_id"`   // 唯一版本ID
	Request     CalculateTargetRequest `json:"request"`      // 原始请求参数
	Result      TargetData             `json:"result"`       // 计算结果
	CreatedAt   time.Time              `json:"created_at"`   // 计算时间
	ExpireAt    time.Time              `json:"expire_at"`    // 过期时间
}

// HistoryItem 历史记录列表项
type HistoryItem struct {
	VersionID    string    `json:"version_id"`
	NutrientName string    `json:"nutrient_name"`
	Gender       string    `json:"gender"`
	Age          int       `json:"age"`
	Crowd        string    `json:"crowd"`
	Scenario     string    `json:"scenario"`
	CreatedAt    time.Time `json:"created_at"`
	TargetMedian float64   `json:"target_median"`
	Unit         string    `json:"unit"`
}

// HistoryListResponse 历史记录列表响应
type HistoryListResponse struct {
	Code     int           `json:"code"`
	Message  string        `json:"message"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
	Data     []HistoryItem `json:"data"`
}

// HistoryDetailResponse 历史记录详情响应
type HistoryDetailResponse struct {
	Code     int                `json:"code"`
	Message  string             `json:"message"`
	Data     CalculationResult  `json:"data"`
}

// CompareRequest 对比请求
type CompareRequest struct {
	VersionID1 string `json:"version_id1" binding:"required" example:"calc_abc123"`
	VersionID2 string `json:"version_id2" binding:"required" example:"calc_def456"`
}

// CompareDiffItem 差异项
type CompareDiffItem struct {
	Field       string      `json:"field"`        // 字段名
	Description string      `json:"description"`  // 字段描述
	Value1      interface{} `json:"value1"`       // 版本1的值
	Value2      interface{} `json:"value2"`       // 版本2的值
	Diff        float64     `json:"diff"`         // 差异值（数值型）
	DiffPercent string      `json:"diff_percent"` // 差异百分比（数值型）
	ChangeType  string      `json:"change_type"`  // 变化类型：increase/decrease/no_change/different
}

// CompareResponse 对比响应
type CompareResponse struct {
	Code       int             `json:"code"`
	Message    string          `json:"message"`
	Version1   CalculationResult `json:"version1"`
	Version2   CalculationResult `json:"version2"`
	DiffSummary CompareSummary  `json:"diff_summary"`
	Diffs      []CompareDiffItem `json:"diffs"`
}

// CompareSummary 对比摘要
type CompareSummary struct {
	TotalFields     int `json:"total_fields"`      // 总字段数
	ChangedFields   int `json:"changed_fields"`    // 有变化的字段数
	IncreasedFields int `json:"increased_fields"`  // 增加的字段数
	DecreasedFields int `json:"decreased_fields"`  // 减少的字段数
}

// 通用响应
type CommonResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

