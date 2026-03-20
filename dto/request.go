package dto

// CalculateTargetRequest 计算目标请求
type CalculateTargetRequest struct {
	Gender       string `json:"gender" binding:"required,oneof=男 女" example:"男"`
	Age          int    `json:"age" binding:"required,min=0,max=120" example:"30"`
	Crowd        string `json:"crowd" binding:"required" example:"普通人群"`
	NutrientName string `json:"nutrient_name" binding:"required" example:"维生素C"`
	Scenario     string `json:"scenario" binding:"required,oneof=A B C" example:"A"`
}

// CalculateTargetResponse 计算目标响应
type CalculateTargetResponse struct {
	Code           int         `json:"code"`
	Message        string      `json:"message"`
	Data           TargetData  `json:"data"`
}

// TargetData 目标数据
type TargetData struct {
	NutrientName   string      `json:"nutrient_name"`
	OriginalMean   float64     `json:"original_mean"`   // 原始平均摄入量
	OriginalCV     float64     `json:"original_cv"`     // 原始变异系数
	TargetMedian   float64     `json:"target_median"`   // 目标中位数
	TargetP95      float64     `json:"target_p95"`      // 目标P95值
	UL             float64     `json:"ul"`              // UL值
	ExceedUL       bool        `json:"exceed_ul"`       // 是否超过UL
	Warning        string      `json:"warning"`         // 警告信息
	Unit           string      `json:"unit"`            // 单位
	AdjustmentFactor float64   `json:"adjustment_factor"` // 调整因子
}
