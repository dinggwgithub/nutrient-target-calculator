package dto

import "time"

type HistoryRecord struct {
	VersionID     string    `json:"version_id"`
	Gender        string    `json:"gender"`
	Age           int       `json:"age"`
	Crowd         string    `json:"crowd"`
	NutrientName  string    `json:"nutrient_name"`
	Scenario      string    `json:"scenario"`
	TargetData    TargetData `json:"target_data"`
	CreatedAt     time.Time `json:"created_at"`
}

type HistoryListResponse struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    HistoryListData  `json:"data"`
}

type HistoryListData struct {
	Total    int             `json:"total"`
	Records  []HistoryRecord `json:"records"`
}

type HistoryQueryRequest struct {
	Gender       string `form:"gender" binding:"omitempty,oneof=男 女"`
	Age          int    `form:"age" binding:"omitempty,min=0,max=120"`
	Crowd        string `form:"crowd"`
	NutrientName string `form:"nutrient_name"`
	Limit        int    `form:"limit" binding:"omitempty,min=1,max=100"`
}

type CompareRequest struct {
	VersionID1 string `json:"version_id1" binding:"required"`
	VersionID2 string `json:"version_id2" binding:"required"`
}

type CompareResponse struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Data    CompareData  `json:"data"`
}

type CompareData struct {
	Record1   HistoryRecord     `json:"record1"`
	Record2   HistoryRecord     `json:"record2"`
	Diff      TargetDiff        `json:"diff"`
}

type TargetDiff struct {
	TargetMedianDiff     float64 `json:"target_median_diff"`
	TargetMedianDiffPct  float64 `json:"target_median_diff_pct"`
	TargetP95Diff        float64 `json:"target_p95_diff"`
	TargetP95DiffPct     float64 `json:"target_p95_diff_pct"`
	OriginalMeanDiff     float64 `json:"original_mean_diff"`
	OriginalMeanDiffPct  float64 `json:"original_mean_diff_pct"`
	ExceedULChanged      bool    `json:"exceed_ul_changed"`
	WarningChanged       bool    `json:"warning_changed"`
}
