package service

import (
	"errors"
	"math"
	"target-calculator-from-db/config"
	"target-calculator-from-db/dto"
	"target-calculator-from-db/models"

	"gorm.io/gorm"
)

// CalculateNutrientTarget 计算营养素目标
func CalculateNutrientTarget(req dto.CalculateTargetRequest) (*dto.TargetData, error) {
	// 1. 先查询DRIs参考值
	dris, err := getDRISReference(req.Gender, req.Age, req.Crowd, req.NutrientName)
	if err != nil {
		return nil, err
	}

	// 2. 查询平均摄入量
	avg, err := getNutrientAverage(req.Gender, req.Age, req.Crowd, req.NutrientName)
	if err != nil {
		// 如果没有摄入数据，使用DRIs的AI/RNI作为默认均值
		defaultMean := dris.AI
		if defaultMean <= 0 {
			defaultMean = dris.RNI
		}
		if defaultMean <= 0 {
			defaultMean = dris.EAR
		}
		if defaultMean <= 0 {
			defaultMean = 100.0 // 保底默认值
		}
		avg = &models.NutrientAverage{
			Gender:        req.Gender,
			Age:           req.Age,
			Crowd:         req.Crowd,
			NutrientName:  req.NutrientName,
			NationalTotal: defaultMean,
			Unit:          dris.Unit,
		}
	}

	// 3. 查询变异系数
	variation, err := getNutrientVariation(req.Gender, req.Age, req.Crowd, req.NutrientName)
	if err != nil {
		// 使用默认变异系数25%
		variation = &models.NutrientVariation{
			Gender:       req.Gender,
			Age:          req.Age,
			Crowd:        req.Crowd,
			NutrientName: req.NutrientName,
			CVValue:      25.0, // 默认CV为25%
		}
	}

	// 注意：数据库中的cv_value是百分比形式（如31.0表示31%），转换为小数
	cv := variation.CVValue / 100.0

	// 4. 根据场景计算目标中位数
	targetMedian, adjustmentFactor := calculateTargetByScenario(avg.NationalTotal, cv, dris, req.Scenario)

	// 5. 计算P95值（假设对数正态分布，计算95百分位数）
	targetSD := targetMedian * cv // 假设CV保持不变
	targetP95 := calculateP95(targetMedian, targetSD)

	// 6. 检查是否超过UL
	exceedUL := targetP95 > dris.UL && dris.UL > 0

	// 7. 构建警告信息
	warning := ""
	if exceedUL {
		warning = "警告：P95值超过可耐受最高摄入量(UL)，存在摄入过量风险"
	}

	return &dto.TargetData{
		NutrientName:     req.NutrientName,
		OriginalMean:     avg.NationalTotal,
		OriginalCV:       cv, // 使用小数形式的CV
		TargetMedian:     targetMedian,
		TargetP95:        targetP95,
		UL:               dris.UL,
		ExceedUL:         exceedUL,
		Warning:          warning,
		Unit:             avg.Unit,
		AdjustmentFactor: adjustmentFactor,
	}, nil
}

// getNutrientAverage 查询营养素平均摄入量
func getNutrientAverage(gender string, age int, crowd string, nutrientName string) (*models.NutrientAverage, error) {
	var avg models.NutrientAverage
	// 策略1：查找年龄 <= 输入年龄的最大年龄组（找到最匹配的年龄区间）
	err := config.DB.Where("gender = ? AND nutrient_name = ? AND age <= ?", gender, nutrientName, age).Order("age DESC").First(&avg).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 策略2：尝试不区分性别，按年龄匹配
			err = config.DB.Where("nutrient_name = ? AND age <= ?", nutrientName, age).Order("age DESC").First(&avg).Error
			if err != nil {
				// 策略3：只匹配性别和营养素，取最新数据
				err = config.DB.Where("gender = ? AND nutrient_name = ?", gender, nutrientName).Order("age DESC").First(&avg).Error
				if err != nil {
					// 策略4：只匹配营养素
					err = config.DB.Where("nutrient_name = ?", nutrientName).Order("age DESC").First(&avg).Error
					if err != nil {
						return nil, errors.New("未找到该营养素的平均摄入数据: " + nutrientName)
					}
				}
			}
		} else {
			return nil, err
		}
	}
	return &avg, nil
}

// getNutrientVariation 查询营养素变异系数
func getNutrientVariation(gender string, age int, crowd string, nutrientName string) (*models.NutrientVariation, error) {
	var variation models.NutrientVariation
	// 性别字段可能是单个字符"男"/"女"，或者完整的"男性"/"女性"
	genderShort := string([]rune(gender)[0]) // 取第一个字符"男"或"女"

	// 策略1：性别 + 营养素 + 年龄区间匹配
	err := config.DB.Where("gender = ? AND nutrient_name = ? AND age <= ?", genderShort, nutrientName, age).Order("age DESC").First(&variation).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 策略2：不区分性别 + 营养素 + 年龄区间匹配
			err = config.DB.Where("nutrient_name = ? AND age <= ?", nutrientName, age).Order("age DESC").First(&variation).Error
			if err != nil {
				// 策略3：性别 + 营养素（不限制年龄）
				err = config.DB.Where("gender = ? AND nutrient_name = ?", genderShort, nutrientName).Order("age DESC").First(&variation).Error
				if err != nil {
					// 策略4：只匹配营养素
					err = config.DB.Where("nutrient_name = ?", nutrientName).Order("age DESC").First(&variation).Error
					if err != nil {
						// 策略5：随机取一条数据或返回错误（外层会使用默认值）
						err = config.DB.Order("RAND()").First(&variation).Error
						if err != nil {
							return nil, errors.New("未找到变异系数数据，使用默认值")
						}
					}
				}
			}
		} else {
			return nil, err
		}
	}
	return &variation, nil
}

// getDRISReference 查询DRIs参考值
func getDRISReference(gender string, age int, crowd string, nutrientName string) (*models.DRISReference, error) {
	var dris models.DRISReference
	// 性别转换: 男->1, 女->2 (根据数据库实际存储的编码)
	sexCode := uint8(1)
	if gender == "女" {
		sexCode = 2
	}

	// 策略1：性别 + 年龄区间匹配 + 营养素 + 活动水平=0（基础值）
	err := config.DB.Where("sex = ? AND age <= ? AND nutrient_name = ? AND activity = 0",
		sexCode, float64(age), nutrientName).Order("age DESC").First(&dris).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 策略2：不区分性别 + 年龄区间匹配
			err = config.DB.Where("age <= ? AND nutrient_name = ? AND activity = 0",
				float64(age), nutrientName).Order("age DESC").First(&dris).Error
			if err != nil {
				// 策略3：性别 + 营养素（不限制年龄）
				err = config.DB.Where("sex = ? AND nutrient_name = ? AND activity = 0",
					sexCode, nutrientName).Order("age DESC").First(&dris).Error
				if err != nil {
					// 策略4：只匹配营养素
					err = config.DB.Where("nutrient_name = ? AND activity = 0", nutrientName).Order("age DESC").First(&dris).Error
					if err != nil {
						return nil, errors.New("未找到该营养素的DRIs参考数据: " + nutrientName)
					}
				}
			}
		} else {
			return nil, err
		}
	}
	return &dris, nil
}

// calculateTargetByScenario 根据场景计算目标中位数
// 场景A：达到RNI/AI（保证97.5%的人群达到EAR）
// 场景B：中等目标（保证50%的人群达到EAR）
// 场景C：保守目标（仅调整基础分布）
func calculateTargetByScenario(mean float64, cv float64, dris *models.DRISReference, scenario string) (float64, float64) {
	// 确定目标参考值（优先使用RNI，其次使用AI）
	targetReference := dris.RNI
	if targetReference <= 0 {
		targetReference = dris.AI
	}

	// EAR作为基准值
	ear := dris.EAR
	if ear <= 0 {
		ear = targetReference * 0.8 // 如果没有EAR，假设为RNI/AI的80%
	}

	var targetMedian float64
	var adjustmentFactor float64

	switch scenario {
	case "A":
		// 场景A：目标中位数 = 目标参考值（RNI/AI）
		// 保证97.5%的人群摄入超过EAR
		// 计算需要的均值：mean = EAR + 1.96 * SD（97.5%分位数）
		// 但实际我们直接以RNI为目标中位数
		targetMedian = targetReference
		adjustmentFactor = targetMedian / mean

	case "B":
		// 场景B：中等目标 - 保证50%的人群达到EAR
		// 中位数 = EAR（对数正态分布下，中位数与几何均值接近）
		// 或使用均值调整：(mean + EAR) / 2
		targetMedian = (mean + ear) / 2
		adjustmentFactor = targetMedian / mean

	case "C":
		// 场景C：保守目标 - 基于原始分布的轻微调整
		// 仅保证平均水平达到EAR的80%
		targetMedian = math.Max(mean, ear*0.8)
		adjustmentFactor = targetMedian / mean

	default:
		targetMedian = mean
		adjustmentFactor = 1.0
	}

	return targetMedian, adjustmentFactor
}

// calculateP95 计算95百分位数
// 假设服从对数正态分布，使用几何均值和几何标准差计算
func calculateP95(mean float64, sd float64) float64 {
	if mean <= 0 {
		return 0
	}

	// 转换为对数正态分布参数
	cv := sd / mean
	if cv <= 0 {
		cv = 0.1 // 默认10%变异
	}

	// 对数正态分布的95百分位数
	// P95 = exp(μ + 1.645 * σ)，其中μ=ln(mean)-σ²/2，σ=√(ln(1+cv²))
	sigma := math.Sqrt(math.Log(1 + cv*cv))
	mu := math.Log(mean) - (sigma*sigma)/2
	p95 := math.Exp(mu + 1.645*sigma)

	return p95
}
