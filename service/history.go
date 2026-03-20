package service

import (
	"encoding/json"
	"fmt"
	"math"
	"target-calculator-from-db/config"
	"target-calculator-from-db/dto"
	"time"

	"github.com/google/uuid"
)

// SaveCalculationToHistory 保存计算结果到历史记录
func SaveCalculationToHistory(userID string, req dto.CalculateTargetRequest, result dto.TargetData) (*dto.CalculationRecord, error) {
	timestamp := time.Now().Unix()
	recordID := uuid.New().String()

	// 生成 Redis Key
	key := config.GenerateCalculationKey(userID, timestamp, recordID)

	// 构建记录
	record := &dto.CalculationRecord{
		ID:        recordID,
		UserID:    userID,
		Timestamp: timestamp,
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		Request:   req,
		Result:    result,
	}

	// 保存到 Redis
	if err := config.SaveCalculationResult(key, record); err != nil {
		return nil, err
	}

	// 添加到用户历史索引
	if err := config.AddToUserHistory(userID, key, timestamp); err != nil {
		return nil, err
	}

	return record, nil
}

// GetUserHistoryList 获取用户历史记录列表
func GetUserHistoryList(userID string, limit int) (*dto.HistoryListData, error) {
	if limit <= 0 {
		limit = 10 // 默认返回10条
	}

	// 获取历史记录 Key 列表
	keys, err := config.GetUserHistory(userID, int64(limit))
	if err != nil {
		return nil, err
	}

	records := make([]dto.HistoryRecordItem, 0, len(keys))

	for _, key := range keys {
		// 解析 Key 获取基本信息
		// Key 格式: calc:{user_id}:{timestamp}:{uuid}
		var userIDStr string
		var timestamp int64
		var recordID string
		_, err := fmt.Sscanf(key, "calc:%[^:]:%d:%s", &userIDStr, &timestamp, &recordID)
		if err != nil {
			continue
		}

		// 获取详细数据
		data, err := config.GetCalculationResult(key)
		if err != nil {
			continue
		}

		var record dto.CalculationRecord
		if err := json.Unmarshal([]byte(data), &record); err != nil {
			continue
		}

		// 构建列表项（简化信息）
		item := dto.HistoryRecordItem{
			ID:           recordID,
			Key:          key,
			Timestamp:    timestamp,
			CreatedAt:    record.CreatedAt,
			Nutrient:     record.Result.NutrientName,
			Scenario:     record.Request.Scenario,
			TargetMedian: record.Result.TargetMedian,
			Unit:         record.Result.Unit,
		}
		records = append(records, item)
	}

	return &dto.HistoryListData{
		Total:   len(records),
		Records: records,
	}, nil
}

// GetHistoryDetail 获取历史记录详情
func GetHistoryDetail(key string) (*dto.CalculationRecord, error) {
	data, ttl, err := config.GetCalculationResultWithTTL(key)
	if err != nil {
		return nil, err
	}

	var record dto.CalculationRecord
	if err := json.Unmarshal([]byte(data), &record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal record: %v", err)
	}

	// 设置 TTL
	record.TTLSeconds = int64(ttl.Seconds())

	return &record, nil
}

// CompareTwoRecords 对比两个历史记录
func CompareTwoRecords(key1, key2 string) (*dto.CompareData, error) {
	// 获取两个记录
	record1, err := GetHistoryDetail(key1)
	if err != nil {
		return nil, fmt.Errorf("failed to get record1: %v", err)
	}

	record2, err := GetHistoryDetail(key2)
	if err != nil {
		return nil, fmt.Errorf("failed to get record2: %v", err)
	}

	// 计算差异
	differences := calculateDifferences(record1, record2)

	// 计算摘要
	summary := calculateSummary(differences)

	return &dto.CompareData{
		Record1:     *record1,
		Record2:     *record2,
		Differences: differences,
		Summary:     summary,
	}, nil
}

// calculateDifferences 计算两个记录之间的差异
func calculateDifferences(r1, r2 *dto.CalculationRecord) []dto.DifferenceItem {
	differences := make([]dto.DifferenceItem, 0)

	// 对比请求参数
	if r1.Request.Gender != r2.Request.Gender {
		differences = append(differences, dto.DifferenceItem{
			Field:      "request.gender",
			FieldName:  "性别",
			OldValue:   r1.Request.Gender,
			NewValue:   r2.Request.Gender,
			ChangeType: "changed",
		})
	}

	if r1.Request.Age != r2.Request.Age {
		differences = append(differences, dto.DifferenceItem{
			Field:      "request.age",
			FieldName:  "年龄",
			OldValue:   r1.Request.Age,
			NewValue:   r2.Request.Age,
			ChangeType: getNumericChangeType(float64(r1.Request.Age), float64(r2.Request.Age)),
		})
	}

	if r1.Request.Crowd != r2.Request.Crowd {
		differences = append(differences, dto.DifferenceItem{
			Field:      "request.crowd",
			FieldName:  "人群",
			OldValue:   r1.Request.Crowd,
			NewValue:   r2.Request.Crowd,
			ChangeType: "changed",
		})
	}

	if r1.Request.NutrientName != r2.Request.NutrientName {
		differences = append(differences, dto.DifferenceItem{
			Field:      "request.nutrient_name",
			FieldName:  "营养素",
			OldValue:   r1.Request.NutrientName,
			NewValue:   r2.Request.NutrientName,
			ChangeType: "changed",
		})
	}

	if r1.Request.Scenario != r2.Request.Scenario {
		differences = append(differences, dto.DifferenceItem{
			Field:      "request.scenario",
			FieldName:  "场景",
			OldValue:   r1.Request.Scenario,
			NewValue:   r2.Request.Scenario,
			ChangeType: "changed",
		})
	}

	// 对比结果数据 - 数值型字段
	numericFields := []struct {
		field     string
		fieldName string
		oldVal    float64
		newVal    float64
	}{
		{"result.original_mean", "原始平均摄入量", r1.Result.OriginalMean, r2.Result.OriginalMean},
		{"result.original_cv", "原始变异系数", r1.Result.OriginalCV, r2.Result.OriginalCV},
		{"result.target_median", "目标中位数", r1.Result.TargetMedian, r2.Result.TargetMedian},
		{"result.target_p95", "目标P95值", r1.Result.TargetP95, r2.Result.TargetP95},
		{"result.ul", "UL值", r1.Result.UL, r2.Result.UL},
		{"result.adjustment_factor", "调整因子", r1.Result.AdjustmentFactor, r2.Result.AdjustmentFactor},
	}

	for _, f := range numericFields {
		// 使用较小的容差比较浮点数
		if math.Abs(f.oldVal-f.newVal) > 0.0001 {
			changeType := getNumericChangeType(f.oldVal, f.newVal)
			changePercent := 0.0
			if f.oldVal != 0 {
				changePercent = ((f.newVal - f.oldVal) / f.oldVal) * 100
			}

			differences = append(differences, dto.DifferenceItem{
				Field:         f.field,
				FieldName:     f.fieldName,
				OldValue:      f.oldVal,
				NewValue:      f.newVal,
				ChangeType:    changeType,
				ChangePercent: math.Round(changePercent*100) / 100, // 保留2位小数
			})
		}
	}

	// 对比布尔值
	if r1.Result.ExceedUL != r2.Result.ExceedUL {
		changeType := "unchanged"
		if !r1.Result.ExceedUL && r2.Result.ExceedUL {
			changeType = "increased_risk"
		} else if r1.Result.ExceedUL && !r2.Result.ExceedUL {
			changeType = "decreased_risk"
		}

		differences = append(differences, dto.DifferenceItem{
			Field:      "result.exceed_ul",
			FieldName:  "是否超过UL",
			OldValue:   r1.Result.ExceedUL,
			NewValue:   r2.Result.ExceedUL,
			ChangeType: changeType,
		})
	}

	// 对比警告信息
	if r1.Result.Warning != r2.Result.Warning {
		differences = append(differences, dto.DifferenceItem{
			Field:      "result.warning",
			FieldName:  "警告信息",
			OldValue:   r1.Result.Warning,
			NewValue:   r2.Result.Warning,
			ChangeType: "changed",
		})
	}

	return differences
}

// getNumericChangeType 获取数值变化类型
func getNumericChangeType(oldVal, newVal float64) string {
	diff := newVal - oldVal
	if math.Abs(diff) < 0.0001 {
		return "unchanged"
	}
	if diff > 0 {
		return "increased"
	}
	return "decreased"
}

// calculateSummary 计算对比摘要
func calculateSummary(differences []dto.DifferenceItem) dto.CompareSummary {
	summary := dto.CompareSummary{
		TotalFields:   11, // 总字段数（请求5个 + 结果6个核心字段）
		ChangedFields: len(differences),
	}

	for _, diff := range differences {
		switch diff.ChangeType {
		case "increased", "increased_risk":
			summary.IncreasedCount++
		case "decreased", "decreased_risk":
			summary.DecreasedCount++
		}
	}

	// 生成主要变化描述
	if summary.ChangedFields == 0 {
		summary.MainChange = "两个版本完全一致"
	} else {
		summary.MainChange = fmt.Sprintf("共有 %d 个字段发生变化", summary.ChangedFields)
		if summary.IncreasedCount > 0 {
			summary.MainChange += fmt.Sprintf("，%d 个指标上升", summary.IncreasedCount)
		}
		if summary.DecreasedCount > 0 {
			summary.MainChange += fmt.Sprintf("，%d 个指标下降", summary.DecreasedCount)
		}
	}

	return summary
}

// DeleteHistoryRecord 删除历史记录
func DeleteHistoryRecord(key string) error {
	// 解析 Key 获取 userID
	var userID string
	var timestamp int64
	var recordID string
	_, err := fmt.Sscanf(key, "calc:%[^:]:%d:%s", &userID, &timestamp, &recordID)
	if err != nil {
		return fmt.Errorf("invalid key format: %v", err)
	}

	// 从用户历史索引中移除
	historyKey := config.GenerateUserHistoryKey(userID)
	if err := config.RedisClient.ZRem(config.Ctx, historyKey, key).Err(); err != nil {
		return fmt.Errorf("failed to remove from history index: %v", err)
	}

	// 删除实际数据
	if err := config.DeleteCalculationResult(key); err != nil {
		return err
	}

	return nil
}
