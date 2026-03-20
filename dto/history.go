package dto

// CalculationRecord 计算记录（存储在 Redis 中的完整记录）
type CalculationRecord struct {
	ID         string                 `json:"id"`                                    // 记录ID (UUID)
	UserID     string                 `json:"user_id"`                               // 用户ID
	Timestamp  int64                  `json:"timestamp"`                             // 计算时间戳
	CreatedAt  string                 `json:"created_at"`                            // 格式化时间
	Request    CalculateTargetRequest `json:"request"`                               // 请求参数
	Result     TargetData             `json:"result"`                                // 计算结果
	TTLSeconds int64                  `json:"ttl_seconds,omitempty" example:"86400"` // 剩余过期时间（秒）
}

// HistoryListRequest 查询历史记录请求
type HistoryListRequest struct {
	UserID string `json:"user_id" form:"user_id" binding:"required" example:"user123"` // 用户ID
	Limit  int    `json:"limit" form:"limit" example:"10"`                             // 返回数量限制，默认10
}

// HistoryListResponse 历史记录列表响应
type HistoryListResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    HistoryListData `json:"data"`
}

// HistoryListData 历史记录列表数据
type HistoryListData struct {
	Total   int                 `json:"total"`   // 总记录数
	Records []HistoryRecordItem `json:"records"` // 记录列表
}

// HistoryRecordItem 历史记录列表项（简化信息）
type HistoryRecordItem struct {
	ID           string  `json:"id"`            // 记录ID
	Key          string  `json:"key"`           // Redis Key
	Timestamp    int64   `json:"timestamp"`     // 计算时间戳
	CreatedAt    string  `json:"created_at"`    // 格式化时间
	Nutrient     string  `json:"nutrient"`      // 营养素名称
	Scenario     string  `json:"scenario"`      // 场景
	TargetMedian float64 `json:"target_median"` // 目标中位数
	Unit         string  `json:"unit"`          // 单位
}

// HistoryDetailRequest 查询历史记录详情请求
type HistoryDetailRequest struct {
	Key string `json:"key" form:"key" binding:"required" example:"calc:user123:1710900000:abc123"` // Redis Key
}

// HistoryDetailResponse 历史记录详情响应
type HistoryDetailResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    CalculationRecord `json:"data"`
}

// CompareRequest 对比两个历史版本请求
type CompareRequest struct {
	Key1 string `json:"key1" binding:"required" example:"calc:user123:1710900000:abc123"` // 第一个版本Key
	Key2 string `json:"key2" binding:"required" example:"calc:user123:1710900100:def456"` // 第二个版本Key
}

// CompareResponse 对比响应
type CompareResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    CompareData `json:"data"`
}

// CompareData 对比数据
type CompareData struct {
	Record1     CalculationRecord `json:"record1"`     // 第一个记录
	Record2     CalculationRecord `json:"record2"`     // 第二个记录
	Differences []DifferenceItem  `json:"differences"` // 差异项列表
	Summary     CompareSummary    `json:"summary"`     // 对比摘要
}

// DifferenceItem 单个字段差异项
type DifferenceItem struct {
	Field         string      `json:"field"`                    // 字段名
	FieldName     string      `json:"field_name"`               // 字段中文名
	OldValue      interface{} `json:"old_value"`                // 旧值
	NewValue      interface{} `json:"new_value"`                // 新值
	ChangeType    string      `json:"change_type"`              // 变化类型: increased, decreased, unchanged
	ChangePercent float64     `json:"change_percent,omitempty"` // 变化百分比（数值型）
}

// CompareSummary 对比摘要
type CompareSummary struct {
	TotalFields    int    `json:"total_fields"`    // 总字段数
	ChangedFields  int    `json:"changed_fields"`  // 变化字段数
	IncreasedCount int    `json:"increased_count"` // 增加数量
	DecreasedCount int    `json:"decreased_count"` // 减少数量
	MainChange     string `json:"main_change"`     // 主要变化描述
}

// DeleteHistoryRequest 删除历史记录请求
type DeleteHistoryRequest struct {
	Key string `json:"key" binding:"required" example:"calc:user123:1710900000:abc123"` // Redis Key
}

// DeleteHistoryResponse 删除历史记录响应
type DeleteHistoryResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
