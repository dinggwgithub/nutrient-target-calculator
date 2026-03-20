package controller

import (
	"net/http"
	"target-calculator-from-db/dto"
	"target-calculator-from-db/service"

	"github.com/gin-gonic/gin"
)

// CalculateTarget 计算营养素目标
// @Summary 计算营养素目标摄入量
// @Description 根据人群特征、营养素名称和场景，计算目标中位数并检查P95是否超UL，结果自动保存到Redis历史记录（TTL=24小时）
// @Tags 营养素目标计算
// @Accept json
// @Produce json
// @Param request body dto.CalculateTargetRequest true "计算请求参数"
// @Success 200 {object} dto.CalculateTargetResponse "成功返回计算结果"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /target/calculate [post]
// @Example request
//
//	{
//	  "gender": "男",
//	  "age": 30,
//	  "crowd": "普通人群",
//	  "nutrient_name": "维生素C",
//	  "scenario": "A"
//	}
//
// @Example response
//
//	{
//	  "code": 200,
//	  "message": "计算成功",
//	  "data": {
//	    "version_id": "550e8400-e29b-41d4-a716-446655440000",
//	    "nutrient_name": "维生素C",
//	    "original_mean": 85.5,
//	    "original_cv": 0.25,
//	    "target_median": 100.0,
//	    "target_p95": 168.5,
//	    "ul": 2000.0,
//	    "exceed_ul": false,
//	    "warning": "",
//	    "unit": "mg/d",
//	    "adjustment_factor": 1.17,
//	    "created_at": "2024-01-15T10:30:00Z"
//	  }
//	}
func CalculateTarget(c *gin.Context) {
	var req dto.CalculateTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数验证失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	result, err := service.CalculateNutrientTarget(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "计算失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	historyRecord, err := service.SaveHistoryRecord(req, result)
	if err == nil && historyRecord != nil {
		result.VersionID = historyRecord.VersionID
		result.CreatedAt = historyRecord.CreatedAt
	}

	c.JSON(http.StatusOK, dto.CalculateTargetResponse{
		Code:    200,
		Message: "计算成功",
		Data:    *result,
	})
}
