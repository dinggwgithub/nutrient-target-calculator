package controller

import (
	"net/http"
	"strconv"
	"target-calculator-from-db/dto"
	"target-calculator-from-db/service"

	"github.com/gin-gonic/gin"
)

// CalculateTarget 计算营养素目标
// @Summary 计算营养素目标摄入量
// @Description 根据人群特征、营养素名称和场景，计算目标中位数并检查P95是否超UL，结果自动保存到Redis
// @Tags 营养素目标计算
// @Accept json
// @Produce json
// @Param request body dto.CalculateTargetRequest true "计算请求参数"
// @Success 200 {object} dto.CalculateTargetResponse "成功返回计算结果，包含version_id用于后续查询和对比"
// @Failure 400 {object} dto.CommonResponse "请求参数错误"
// @Failure 500 {object} dto.CommonResponse "服务器内部错误"
// @Router /target/calculate [post]
func CalculateTarget(c *gin.Context) {
	var req dto.CalculateTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.CommonResponse{
			Code:    400,
			Message: "参数验证失败: " + err.Error(),
			Data:    nil,
		})
		return
	}

	// 调用服务层计算
	result, err := service.CalculateNutrientTarget(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.CommonResponse{
			Code:    500,
			Message: "计算失败: " + err.Error(),
			Data:    nil,
		})
		return
	}

	// 自动保存结果到Redis
	versionID, err := service.SaveCalculationResult(req, result)
	if err != nil {
		// 保存失败不影响计算结果返回，只打日志
		c.Error(err)
	}

	// 返回成功响应，包含versionID
	c.JSON(http.StatusOK, dto.CalculateTargetResponse{
		Code:      200,
		Message:   "计算成功",
		Data:      *result,
		VersionID: versionID,
	})
}

// GetHistoryList 获取计算历史记录列表
// @Summary 获取计算历史记录列表
// @Description 分页获取计算历史记录，可通过user_id过滤特定用户的记录
// @Tags 历史记录管理
// @Accept json
// @Produce json
// @Param user_id query string false "用户ID，可选，不填则返回所有记录" example:"user_123"
// @Param page query int false "页码，默认1" example:"1"
// @Param page_size query int false "每页条数，默认10" example:"10"
// @Success 200 {object} dto.HistoryListResponse "成功返回历史记录列表"
// @Failure 500 {object} dto.CommonResponse "服务器内部错误"
// @Router /target/history [get]
func GetHistoryList(c *gin.Context) {
	userID := c.Query("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	items, total, err := service.GetHistoryList(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.CommonResponse{
			Code:    500,
			Message: "获取历史记录失败: " + err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, dto.HistoryListResponse{
		Code:     200,
		Message:  "获取成功",
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     items,
	})
}

// GetHistoryDetail 获取历史记录详情
// @Summary 获取历史记录详情
// @Description 根据version_id获取计算结果的详细信息
// @Tags 历史记录管理
// @Accept json
// @Produce json
// @Param version_id path string true "计算结果版本ID" example:"calc_abc123def456"
// @Success 200 {object} dto.HistoryDetailResponse "成功返回历史记录详情"
// @Failure 404 {object} dto.CommonResponse "记录不存在或已过期"
// @Failure 500 {object} dto.CommonResponse "服务器内部错误"
// @Router /target/history/{version_id} [get]
func GetHistoryDetail(c *gin.Context) {
	versionID := c.Param("version_id")
	if versionID == "" {
		c.JSON(http.StatusBadRequest, dto.CommonResponse{
			Code:    400,
			Message: "version_id不能为空",
			Data:    nil,
		})
		return
	}

	result, err := service.GetCalculationResult(versionID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.CommonResponse{
			Code:    404,
			Message: "记录不存在或已过期: " + err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, dto.HistoryDetailResponse{
		Code:    200,
		Message: "获取成功",
		Data:    *result,
	})
}

// CompareHistory 对比两个历史版本
// @Summary 对比两个历史版本差异
// @Description 对比两个计算版本的参数和结果差异，生成详细的对比报告
// @Tags 版本对比
// @Accept json
// @Produce json
// @Param request body dto.CompareRequest true "对比请求参数"
// @Success 200 {object} dto.CompareResponse "成功返回对比结果，包含详细的差异分析"
// @Failure 400 {object} dto.CommonResponse "请求参数错误"
// @Failure 404 {object} dto.CommonResponse "某个版本不存在或已过期"
// @Failure 500 {object} dto.CommonResponse "服务器内部错误"
// @Router /target/compare [post]
func CompareHistory(c *gin.Context) {
	var req dto.CompareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.CommonResponse{
			Code:    400,
			Message: "参数验证失败: " + err.Error(),
			Data:    nil,
		})
		return
	}

	result, err := service.CompareCalculations(req.VersionID1, req.VersionID2)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.CommonResponse{
			Code:    404,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
