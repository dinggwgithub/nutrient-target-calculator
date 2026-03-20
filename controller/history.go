package controller

import (
	"net/http"
	"target-calculator-from-db/dto"
	"target-calculator-from-db/service"

	"github.com/gin-gonic/gin"
)

// GetHistoryList 查询历史记录列表
// @Summary 查询历史计算记录列表
// @Description 根据条件查询历史计算记录，支持按性别、年龄、人群、营养素名称筛选
// @Tags 历史记录
// @Accept json
// @Produce json
// @Param gender query string false "性别" Enums(男, 女)
// @Param age query int false "年龄" minimum(0) maximum(120)
// @Param crowd query string false "人群类型"
// @Param nutrient_name query string false "营养素名称"
// @Param limit query int false "返回数量限制" default(20)
// @Success 200 {object} dto.HistoryListResponse "成功返回历史记录列表"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /target/history [get]
// @Example response
//
//	{
//	  "code": 200,
//	  "message": "查询成功",
//	  "data": {
//	    "total": 2,
//	    "records": [
//	      {
//	        "version_id": "550e8400-e29b-41d4-a716-446655440000",
//	        "gender": "男",
//	        "age": 30,
//	        "crowd": "普通人群",
//	        "nutrient_name": "维生素C",
//	        "scenario": "A",
//	        "target_data": {
//	          "version_id": "550e8400-e29b-41d4-a716-446655440000",
//	          "nutrient_name": "维生素C",
//	          "original_mean": 85.5,
//	          "original_cv": 0.25,
//	          "target_median": 100.0,
//	          "target_p95": 168.5,
//	          "ul": 2000.0,
//	          "exceed_ul": false,
//	          "warning": "",
//	          "unit": "mg/d",
//	          "adjustment_factor": 1.17,
//	          "created_at": "2024-01-15T10:30:00Z"
//	        },
//	        "created_at": "2024-01-15T10:30:00Z"
//	      }
//	    ]
//	  }
//	}
func GetHistoryList(c *gin.Context) {
	var req dto.HistoryQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数验证失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	result, err := service.GetHistoryList(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, dto.HistoryListResponse{
		Code:    200,
		Message: "查询成功",
		Data:    *result,
	})
}

// GetRecordByVID 根据版本ID查询单条历史记录
// @Summary 根据版本ID查询历史记录
// @Description 根据版本ID查询单条历史计算记录详情
// @Tags 历史记录
// @Accept json
// @Produce json
// @Param version_id path string true "版本ID"
// @Success 200 {object} map[string]interface{} "成功返回历史记录详情"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "记录不存在或已过期"
// @Router /target/history/{version_id} [get]
// @Example response
//
//	{
//	  "code": 200,
//	  "message": "查询成功",
//	  "data": {
//	    "version_id": "550e8400-e29b-41d4-a716-446655440000",
//	    "gender": "男",
//	    "age": 30,
//	    "crowd": "普通人群",
//	    "nutrient_name": "维生素C",
//	    "scenario": "A",
//	    "target_data": {
//	      "version_id": "550e8400-e29b-41d4-a716-446655440000",
//	      "nutrient_name": "维生素C",
//	      "original_mean": 85.5,
//	      "original_cv": 0.25,
//	      "target_median": 100.0,
//	      "target_p95": 168.5,
//	      "ul": 2000.0,
//	      "exceed_ul": false,
//	      "warning": "",
//	      "unit": "mg/d",
//	      "adjustment_factor": 1.17,
//	      "created_at": "2024-01-15T10:30:00Z"
//	    },
//	    "created_at": "2024-01-15T10:30:00Z"
//	  }
//	}
func GetRecordByVID(c *gin.Context) {
	versionID := c.Param("version_id")
	if versionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "version_id 不能为空",
			"data":    nil,
		})
		return
	}

	record, err := service.GetRecordByVersionID(versionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "记录不存在或已过期: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "查询成功",
		"data":    record,
	})
}

// CompareHistory 对比两个历史版本
// @Summary 对比两个历史计算版本
// @Description 对比两个历史计算记录的差异，包括目标中位数、P95值、原始均值等指标的变化
// @Tags 历史记录
// @Accept json
// @Produce json
// @Param request body dto.CompareRequest true "对比请求参数"
// @Success 200 {object} dto.CompareResponse "成功返回对比结果"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /target/compare [post]
// @Example request
//
//	{
//	  "version_id1": "550e8400-e29b-41d4-a716-446655440000",
//	  "version_id2": "660e8400-e29b-41d4-a716-446655440001"
//	}
//
// @Example response
//
//	{
//	  "code": 200,
//	  "message": "对比成功",
//	  "data": {
//	    "record1": {
//	      "version_id": "550e8400-e29b-41d4-a716-446655440000",
//	      "gender": "男",
//	      "age": 30,
//	      "crowd": "普通人群",
//	      "nutrient_name": "维生素C",
//	      "scenario": "A",
//	      "target_data": {
//	        "version_id": "550e8400-e29b-41d4-a716-446655440000",
//	        "nutrient_name": "维生素C",
//	        "original_mean": 85.5,
//	        "original_cv": 0.25,
//	        "target_median": 100.0,
//	        "target_p95": 168.5,
//	        "ul": 2000.0,
//	        "exceed_ul": false,
//	        "warning": "",
//	        "unit": "mg/d",
//	        "adjustment_factor": 1.17,
//	        "created_at": "2024-01-15T10:30:00Z"
//	      },
//	      "created_at": "2024-01-15T10:30:00Z"
//	    },
//	    "record2": {
//	      "version_id": "660e8400-e29b-41d4-a716-446655440001",
//	      "gender": "男",
//	      "age": 30,
//	      "crowd": "普通人群",
//	      "nutrient_name": "维生素C",
//	      "scenario": "B",
//	      "target_data": {
//	        "version_id": "660e8400-e29b-41d4-a716-446655440001",
//	        "nutrient_name": "维生素C",
//	        "original_mean": 85.5,
//	        "original_cv": 0.25,
//	        "target_median": 75.0,
//	        "target_p95": 126.3,
//	        "ul": 2000.0,
//	        "exceed_ul": false,
//	        "warning": "",
//	        "unit": "mg/d",
//	        "adjustment_factor": 0.88,
//	        "created_at": "2024-01-15T11:00:00Z"
//	      },
//	      "created_at": "2024-01-15T11:00:00Z"
//	    },
//	    "diff": {
//	      "target_median_diff": -25.0,
//	      "target_median_diff_pct": -25.0,
//	      "target_p95_diff": -42.2,
//	      "target_p95_diff_pct": -25.0,
//	      "original_mean_diff": 0.0,
//	      "original_mean_diff_pct": 0.0,
//	      "exceed_ul_changed": false,
//	      "warning_changed": false
//	    }
//	  }
//	}
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

	result, err := service.CompareRecords(req.VersionID1, req.VersionID2)
	if err != nil {
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
		Data:    *result,
	})
}
