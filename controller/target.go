package controller

import (
	"net/http"
	"target-calculator-from-db/dto"
	"target-calculator-from-db/service"

	"github.com/gin-gonic/gin"
)

// CalculateTarget 计算营养素目标
// @Summary 计算营养素目标摄入量
// @Description 根据人群特征、营养素名称和场景，计算目标中位数并检查P95是否超UL
// @Tags 营养素目标计算
// @Accept json
// @Produce json
// @Param request body dto.CalculateTargetRequest true "计算请求参数"
// @Success 200 {object} dto.CalculateTargetResponse "成功返回计算结果"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /target/calculate [post]
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

	// 调用服务层计算
	result, err := service.CalculateNutrientTarget(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "计算失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, dto.CalculateTargetResponse{
		Code:    200,
		Message: "计算成功",
		Data:    *result,
	})
}
