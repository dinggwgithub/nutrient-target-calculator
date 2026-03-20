package routes

import (
	"target-calculator-from-db/controller"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")
	{
		api.POST("/target/calculate", controller.CalculateTarget)

		api.GET("/target/history", controller.GetHistoryList)
		api.GET("/target/history/:version_id", controller.GetRecordByVID)
		api.POST("/target/compare", controller.CompareHistory)

		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"message": "服务运行正常",
			})
		})
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
