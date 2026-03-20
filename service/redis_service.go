package service

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"target-calculator-from-db/config"
	"target-calculator-from-db/dto"
	"time"

	"github.com/redis/go-redis/v9"
)

// GenerateVersionID 生成唯一的版本ID
func GenerateVersionID(req dto.CalculateTargetRequest) string {
	// 基于请求参数和时间戳生成唯一ID
	data := fmt.Sprintf("%s:%d:%s:%s:%s:%d",
		req.Gender, req.Age, req.Crowd, req.NutrientName, req.Scenario, time.Now().UnixNano())
	hash := md5.Sum([]byte(data))
	return "calc_" + hex.EncodeToString(hash[:])[:12]
}

// SaveCalculationResult 保存计算结果到Redis
func SaveCalculationResult(req dto.CalculateTargetRequest, result *dto.TargetData) (string, error) {
	versionID := GenerateVersionID(req)
	now := time.Now()
	expireAt := now.Add(config.CalculationTTL)

	// 构建完整的计算结果
	calcResult := dto.CalculationResult{
		VersionID: versionID,
		Request:   req,
		Result:    *result,
		CreatedAt: now,
		ExpireAt:  expireAt,
	}

	// 序列化JSON
	jsonData, err := json.Marshal(calcResult)
	if err != nil {
		return "", fmt.Errorf("序列化计算结果失败: %v", err)
	}

	// 1. 存储计算结果详情（String类型）
	resultKey := config.KeyPrefixCalculation + versionID
	err = config.RedisClient.Set(config.Ctx, resultKey, jsonData, config.CalculationTTL).Err()
	if err != nil {
		return "", fmt.Errorf("保存计算结果到Redis失败: %v", err)
	}

	// 2. 维护用户历史记录索引（ZSet，按时间排序）
	if req.UserID != "" {
		userHistoryKey := config.KeyPrefixHistory + req.UserID
		// 使用时间戳作为score，value为versionID
		err = config.RedisClient.ZAdd(config.Ctx, userHistoryKey, redisZAddMember(float64(now.Unix()), versionID)).Err()
		if err != nil {
			return "", fmt.Errorf("更新用户历史索引失败: %v", err)
		}
		// 设置用户历史索引的TTL（比结果多保留1小时，避免孤立索引）
		config.RedisClient.Expire(config.Ctx, userHistoryKey, config.CalculationTTL+time.Hour)
	}

	// 3. 维护全局索引（可选，用于管理员查看所有计算）
	err = config.RedisClient.ZAdd(config.Ctx, config.KeyPrefixIndex, redisZAddMember(float64(now.Unix()), versionID)).Err()
	if err != nil {
		return "", fmt.Errorf("更新全局索引失败: %v", err)
	}
	config.RedisClient.Expire(config.Ctx, config.KeyPrefixIndex, config.CalculationTTL+time.Hour)

	return versionID, nil
}

// GetCalculationResult 根据versionID获取计算结果
func GetCalculationResult(versionID string) (*dto.CalculationResult, error) {
	resultKey := config.KeyPrefixCalculation + versionID

	// 从Redis获取数据
	jsonData, err := config.RedisClient.Get(config.Ctx, resultKey).Result()
	if err != nil {
		return nil, fmt.Errorf("从Redis获取计算结果失败: %v", err)
	}

	// 反序列化
	var result dto.CalculationResult
	err = json.Unmarshal([]byte(jsonData), &result)
	if err != nil {
		return nil, fmt.Errorf("反序列化计算结果失败: %v", err)
	}

	return &result, nil
}

// GetHistoryList 获取历史记录列表（支持分页）
func GetHistoryList(userID string, page, pageSize int) ([]dto.HistoryItem, int64, error) {
	var historyKey string
	if userID != "" {
		historyKey = config.KeyPrefixHistory + userID
	} else {
		// 如果没有userID，返回全局历史记录
		historyKey = config.KeyPrefixIndex
	}

	// 获取总数
	total, err := config.RedisClient.ZCard(config.Ctx, historyKey).Result()
	if err != nil {
		return nil, 0, fmt.Errorf("获取历史记录总数失败: %v", err)
	}

	if total == 0 {
		return []dto.HistoryItem{}, 0, nil
	}

	// 分页获取versionID列表（按时间倒序）
	start := int64((page - 1) * pageSize)
	end := start + int64(pageSize) - 1

	versionIDs, err := config.RedisClient.ZRevRange(config.Ctx, historyKey, start, end).Result()
	if err != nil {
		return nil, 0, fmt.Errorf("获取历史记录列表失败: %v", err)
	}

	// 批量获取计算结果
	items := make([]dto.HistoryItem, 0, len(versionIDs))
	for _, versionID := range versionIDs {
		result, err := GetCalculationResult(versionID)
		if err != nil {
			// 忽略已过期或不存在的记录
			continue
		}
		items = append(items, dto.HistoryItem{
			VersionID:    result.VersionID,
			NutrientName: result.Result.NutrientName,
			Gender:       result.Request.Gender,
			Age:          result.Request.Age,
			Crowd:        result.Request.Crowd,
			Scenario:     result.Request.Scenario,
			CreatedAt:    result.CreatedAt,
			TargetMedian: result.Result.TargetMedian,
			Unit:         result.Result.Unit,
		})
	}

	return items, total, nil
}

// CompareCalculations 对比两个计算版本
func CompareCalculations(versionID1, versionID2 string) (*dto.CompareResponse, error) {
	// 获取两个版本的计算结果
	result1, err := GetCalculationResult(versionID1)
	if err != nil {
		return nil, fmt.Errorf("获取版本1失败: %v", err)
	}

	result2, err := GetCalculationResult(versionID2)
	if err != nil {
		return nil, fmt.Errorf("获取版本2失败: %v", err)
	}

	// 生成差异对比
	diffs := generateDiffs(result1, result2)

	// 统计摘要
	summary := dto.CompareSummary{
		TotalFields:     len(diffs),
		ChangedFields:   0,
		IncreasedFields: 0,
		DecreasedFields: 0,
	}

	for _, diff := range diffs {
		if diff.ChangeType != "no_change" {
			summary.ChangedFields++
		}
		if diff.ChangeType == "increase" {
			summary.IncreasedFields++
		}
		if diff.ChangeType == "decrease" {
			summary.DecreasedFields++
		}
	}

	return &dto.CompareResponse{
		Code:        200,
		Message:     "对比成功",
		Version1:    *result1,
		Version2:    *result2,
		DiffSummary: summary,
		Diffs:       diffs,
	}, nil
}

// generateDiffs 生成详细的差异对比
func generateDiffs(r1, r2 *dto.CalculationResult) []dto.CompareDiffItem {
	var diffs []dto.CompareDiffItem

	// 定义需要对比的字段及其描述
	fields := []struct {
		field       string
		description string
		getValue    func(*dto.CalculationResult) interface{}
		isNumeric   bool
	}{
		{"gender", "性别", func(r *dto.CalculationResult) interface{} { return r.Request.Gender }, false},
		{"age", "年龄", func(r *dto.CalculationResult) interface{} { return r.Request.Age }, true},
		{"crowd", "人群类型", func(r *dto.CalculationResult) interface{} { return r.Request.Crowd }, false},
		{"nutrient_name", "营养素名称", func(r *dto.CalculationResult) interface{} { return r.Result.NutrientName }, false},
		{"scenario", "计算场景", func(r *dto.CalculationResult) interface{} { return r.Request.Scenario }, false},
		{"original_mean", "原始平均摄入量", func(r *dto.CalculationResult) interface{} { return r.Result.OriginalMean }, true},
		{"original_cv", "原始变异系数", func(r *dto.CalculationResult) interface{} { return r.Result.OriginalCV }, true},
		{"target_median", "目标中位数", func(r *dto.CalculationResult) interface{} { return r.Result.TargetMedian }, true},
		{"target_p95", "目标P95值", func(r *dto.CalculationResult) interface{} { return r.Result.TargetP95 }, true},
		{"ul", "可耐受最高摄入量(UL)", func(r *dto.CalculationResult) interface{} { return r.Result.UL }, true},
		{"adjustment_factor", "调整因子", func(r *dto.CalculationResult) interface{} { return r.Result.AdjustmentFactor }, true},
		{"exceed_ul", "是否超过UL", func(r *dto.CalculationResult) interface{} { return r.Result.ExceedUL }, false},
		{"unit", "单位", func(r *dto.CalculationResult) interface{} { return r.Result.Unit }, false},
	}

	for _, f := range fields {
		v1 := f.getValue(r1)
		v2 := f.getValue(r2)

		diff := dto.CompareDiffItem{
			Field:       f.field,
			Description: f.description,
			Value1:      v1,
			Value2:      v2,
			ChangeType:  "no_change",
		}

		if f.isNumeric {
			// 数值类型对比
			n1 := toFloat64(v1)
			n2 := toFloat64(v2)
			diff.Diff = n2 - n1

			if n1 != 0 {
				diff.DiffPercent = fmt.Sprintf("%.2f%%", (n2-n1)/n1*100)
			} else if n2 != 0 {
				diff.DiffPercent = "+∞%"
			} else {
				diff.DiffPercent = "0.00%"
			}

			if n2 > n1 {
				diff.ChangeType = "increase"
			} else if n2 < n1 {
				diff.ChangeType = "decrease"
			}
		} else {
			// 非数值类型对比
			if fmt.Sprintf("%v", v1) != fmt.Sprintf("%v", v2) {
				diff.ChangeType = "different"
			}
		}

		diffs = append(diffs, diff)
	}

	// 排序：有变化的排前面
	sort.Slice(diffs, func(i, j int) bool {
		if diffs[i].ChangeType == "no_change" && diffs[j].ChangeType != "no_change" {
			return false
		}
		if diffs[i].ChangeType != "no_change" && diffs[j].ChangeType == "no_change" {
			return true
		}
		return diffs[i].Field < diffs[j].Field
	})

	return diffs
}

// toFloat64 转换为float64
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case float64:
		return val
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	default:
		return 0
	}
}

// redisZAddMember 辅助函数创建ZAdd参数
func redisZAddMember(score float64, member string) redis.Z {
	return redis.Z{
		Score:  score,
		Member: member,
	}
}
