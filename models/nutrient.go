package models

// NutrientAverage 营养素平均摄入量表
type NutrientAverage struct {
	ID            uint    `gorm:"column:id;primaryKey"`
	Gender        string  `gorm:"column:gender;type:varchar(10)"`
	Age           int     `gorm:"column:age"`
	Crowd         string  `gorm:"column:crowd;type:varchar(10)"`
	NutrientName  string  `gorm:"column:nutrient_name;type:varchar(50)"`
	NationalTotal float64 `gorm:"column:national_total"` // 全国平均摄入量
	Unit          string  `gorm:"column:unit;type:varchar(20)"`
}

// TableName 指定表名
func (NutrientAverage) TableName() string {
	return "nutrient_average"
}

// NutrientVariation 营养素变异系数表
type NutrientVariation struct {
	ID           uint    `gorm:"column:id;primaryKey"`
	Gender       string  `gorm:"column:gender;type:varchar(1)"`
	Age          int     `gorm:"column:age"`
	Crowd        string  `gorm:"column:crowd;type:varchar(10)"`
	NutrientName string  `gorm:"column:nutrient_name;type:varchar(50)"`
	CVValue      float64 `gorm:"column:cv_value"` // 变异系数
}

// TableName 指定表名
func (NutrientVariation) TableName() string {
	return "nutrient_variation"
}

// DRISReference DRIs参考值表
type DRISReference struct {
	ID           uint    `gorm:"column:id;primaryKey"`
	Sex          uint8   `gorm:"column:sex"` // 性别: 0/1 或 1/2
	Age          float64 `gorm:"column:age"`
	Crowd        string  `gorm:"column:crowd;type:varchar(30)"`
	NutrientName string  `gorm:"column:nutrient_name;type:varchar(30)"`
	EAR          float64 `gorm:"column:ear"` // 平均需要量
	RNI          float64 `gorm:"column:rni"` // 推荐摄入量
	AI           float64 `gorm:"column:ai"`  // 适宜摄入量
	UL           float64 `gorm:"column:ul"`  // 可耐受最高摄入量
	Unit         string  `gorm:"column:unit;type:varchar(20)"`
}

// TableName 指定表名
func (DRISReference) TableName() string {
	return "dris_references"
}
