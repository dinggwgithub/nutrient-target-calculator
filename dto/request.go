package dto

import "time"

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
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    TargetData `json:"data"`
}

// TargetData 目标数据
type TargetData struct {
	VersionID        string    `json:"version_id"`
	NutrientName     string    `json:"nutrient_name"`
	OriginalMean     float64   `json:"original_mean"`
	OriginalCV       float64   `json:"original_cv"`
	TargetMedian     float64   `json:"target_median"`
	TargetP95        float64   `json:"target_p95"`
	UL               float64   `json:"ul"`
	ExceedUL         bool      `json:"exceed_ul"`
	Warning          string    `json:"warning"`
	Unit             string    `json:"unit"`
	AdjustmentFactor float64   `json:"adjustment_factor"`
	CreatedAt        time.Time `json:"created_at"`
}
