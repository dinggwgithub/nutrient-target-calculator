package controller

import (
	"net/http"
	"strconv"
	"target-calculator-from-db/dto"
	"target-calculator-from-db/service"

	"github.com/gin-gonic/gin"
)

// GetHistoryList 获取用户历史记录列表
// @Summary 获取用户历史记录列表
// @Description 获取指定用户的计算历史记录列表，按时间倒序排列
// @Tags 历史记录
// @Accept json
// @Produce json
// @Param user_id query string true "用户ID" example:"user123"
// @Param limit query int false "返回数量限制，默认10" example:"10"
// @Success 200 {object} dto.HistoryListResponse "成功返回历史记录列表"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /target/history [get]
func GetHistoryList(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "user_id 参数不能为空",
			"data":    nil,
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	data, err := service.GetUserHistoryList(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取历史记录失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, dto.HistoryListResponse{
		Code:    200,
		Message: "获取成功",
		Data:    *data,
	})
}

// GetHistoryDetail 获取历史记录详情
// @Summary 获取历史记录详情
// @Description 根据 Redis Key 获取单条历史记录的完整详情
// @Tags 历史记录
// @Accept json
// @Produce json
// @Param key query string true "Redis Key" example:"calc:user123:1710900000:abc123"
// @Success 200 {object} dto.HistoryDetailResponse "成功返回历史记录详情"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "记录不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /target/history/detail [get]
func GetHistoryDetail(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "key 参数不能为空",
			"data":    nil,
		})
		return
	}

	record, err := service.GetHistoryDetail(key)
	if err != nil {
		if err.Error() == "calculation result not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "记录不存在或已过期",
				"data":    nil,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取详情失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, dto.HistoryDetailResponse{
		Code:    200,
		Message: "获取成功",
		Data:    *record,
	})
}

// CompareHistory 对比两个历史版本
// @Summary 对比两个历史版本
// @Description 对比两个历史计算记录的差异，返回详细的字段对比信息
// @Tags 历史记录
// @Accept json
// @Produce json
// @Param request body dto.CompareRequest true "对比请求参数"
// @Success 200 {object} dto.CompareResponse "成功返回对比结果"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "记录不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /target/history/compare [post]
func CompareHistory(c *gin.Context) {
	var req dto.CompareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数验证失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	data, err := service.CompareTwoRecords(req.Key1, req.Key2)
	if err != nil {
		if err.Error() == "calculation result not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "记录不存在或已过期",
				"data":    nil,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "对比失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, dto.CompareResponse{
		Code:    200,
		Message: "对比成功",
		Data:    *data,
	})
}

// DeleteHistory 删除历史记录
// @Summary 删除历史记录
// @Description 根据 Redis Key 删除指定的历史记录
// @Tags 历史记录
// @Accept json
// @Produce json
// @Param request body dto.DeleteHistoryRequest true "删除请求参数"
// @Success 200 {object} dto.DeleteHistoryResponse "删除成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /target/history [delete]
func DeleteHistory(c *gin.Context) {
	var req dto.DeleteHistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数验证失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	if err := service.DeleteHistoryRecord(req.Key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, dto.DeleteHistoryResponse{
		Code:    200,
		Message: "删除成功",
	})
}
