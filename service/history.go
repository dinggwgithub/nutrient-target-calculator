package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"target-calculator-from-db/config"
	"target-calculator-from-db/dto"

	"github.com/google/uuid"
)

const (
	HistoryKeyPrefix  = "target:history:"
	ResultKeyPrefix   = "target:result:"
	HistoryTTL        = 24 * time.Hour
	MaxHistoryPerKey  = 50
)

func getHistoryKey(gender string, age int, crowd, nutrientName string) string {
	return fmt.Sprintf("%s%s:%d:%s:%s", HistoryKeyPrefix, gender, age, crowd, nutrientName)
}

func getResultKey(versionID string) string {
	return ResultKeyPrefix + versionID
}

func SaveHistoryRecord(req dto.CalculateTargetRequest, result *dto.TargetData) (*dto.HistoryRecord, error) {
	ctx := context.Background()
	
	versionID := uuid.New().String()
	now := time.Now()
	
	record := dto.HistoryRecord{
		VersionID:    versionID,
		Gender:       req.Gender,
		Age:          req.Age,
		Crowd:        req.Crowd,
		NutrientName: req.NutrientName,
		Scenario:     req.Scenario,
		TargetData:   *result,
		CreatedAt:    now,
	}
	
	recordJSON, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal history record: %v", err)
	}
	
	resultKey := getResultKey(versionID)
	if err := config.RedisClient.Set(ctx, resultKey, recordJSON, HistoryTTL).Err(); err != nil {
		return nil, fmt.Errorf("failed to save result to redis: %v", err)
	}
	
	historyKey := getHistoryKey(req.Gender, req.Age, req.Crowd, req.NutrientName)
	if err := config.RedisClient.LPush(ctx, historyKey, versionID).Err(); err != nil {
		return nil, fmt.Errorf("failed to push to history list: %v", err)
	}
	
	config.RedisClient.LTrim(ctx, historyKey, 0, int64(MaxHistoryPerKey-1))
	config.RedisClient.Expire(ctx, historyKey, HistoryTTL)
	
	return &record, nil
}

func GetHistoryList(query dto.HistoryQueryRequest) (*dto.HistoryListData, error) {
	ctx := context.Background()
	
	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}
	
	var allRecords []dto.HistoryRecord
	
	if query.Gender != "" && query.NutrientName != "" {
		historyKey := getHistoryKey(query.Gender, query.Age, query.Crowd, query.NutrientName)
		versionIDs, err := config.RedisClient.LRange(ctx, historyKey, 0, int64(limit-1)).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get history list: %v", err)
		}
		
		for _, vid := range versionIDs {
			record, err := getRecordByVersionID(ctx, vid)
			if err == nil {
				allRecords = append(allRecords, *record)
			}
		}
	} else {
		pattern := HistoryKeyPrefix + "*"
		keys, err := config.RedisClient.Keys(ctx, pattern).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan history keys: %v", err)
		}
		
		for _, key := range keys {
			versionIDs, err := config.RedisClient.LRange(ctx, key, 0, int64(limit-1)).Result()
			if err != nil {
				continue
			}
			
			for _, vid := range versionIDs {
				if len(allRecords) >= limit {
					break
				}
				record, err := getRecordByVersionID(ctx, vid)
				if err == nil {
					if matchFilter(record, query) {
						allRecords = append(allRecords, *record)
					}
				}
			}
			if len(allRecords) >= limit {
				break
			}
		}
	}
	
	return &dto.HistoryListData{
		Total:   len(allRecords),
		Records: allRecords,
	}, nil
}

func matchFilter(record *dto.HistoryRecord, query dto.HistoryQueryRequest) bool {
	if query.Gender != "" && record.Gender != query.Gender {
		return false
	}
	if query.Age > 0 && record.Age != query.Age {
		return false
	}
	if query.Crowd != "" && record.Crowd != query.Crowd {
		return false
	}
	if query.NutrientName != "" && record.NutrientName != query.NutrientName {
		return false
	}
	return true
}

func GetRecordByVersionID(versionID string) (*dto.HistoryRecord, error) {
	return getRecordByVersionID(context.Background(), versionID)
}

func getRecordByVersionID(ctx context.Context, versionID string) (*dto.HistoryRecord, error) {
	resultKey := getResultKey(versionID)
	recordJSON, err := config.RedisClient.Get(ctx, resultKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get record from redis: %v", err)
	}
	
	var record dto.HistoryRecord
	if err := json.Unmarshal([]byte(recordJSON), &record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal record: %v", err)
	}
	
	return &record, nil
}

func CompareRecords(versionID1, versionID2 string) (*dto.CompareData, error) {
	record1, err := GetRecordByVersionID(versionID1)
	if err != nil {
		return nil, fmt.Errorf("failed to get record1: %v", err)
	}
	
	record2, err := GetRecordByVersionID(versionID2)
	if err != nil {
		return nil, fmt.Errorf("failed to get record2: %v", err)
	}
	
	diff := calculateDiff(&record1.TargetData, &record2.TargetData)
	
	return &dto.CompareData{
		Record1: *record1,
		Record2: *record2,
		Diff:    diff,
	}, nil
}

func calculateDiff(data1, data2 *dto.TargetData) dto.TargetDiff {
	medianDiff := data2.TargetMedian - data1.TargetMedian
	p95Diff := data2.TargetP95 - data1.TargetP95
	meanDiff := data2.OriginalMean - data1.OriginalMean
	
	var medianDiffPct, p95DiffPct, meanDiffPct float64
	if data1.TargetMedian != 0 {
		medianDiffPct = (medianDiff / data1.TargetMedian) * 100
	}
	if data1.TargetP95 != 0 {
		p95DiffPct = (p95Diff / data1.TargetP95) * 100
	}
	if data1.OriginalMean != 0 {
		meanDiffPct = (meanDiff / data1.OriginalMean) * 100
	}
	
	return dto.TargetDiff{
		TargetMedianDiff:    medianDiff,
		TargetMedianDiffPct: medianDiffPct,
		TargetP95Diff:       p95Diff,
		TargetP95DiffPct:    p95DiffPct,
		OriginalMeanDiff:    meanDiff,
		OriginalMeanDiffPct: meanDiffPct,
		ExceedULChanged:     data1.ExceedUL != data2.ExceedUL,
		WarningChanged:      data1.Warning != data2.Warning,
	}
}
