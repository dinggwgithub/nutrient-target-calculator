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

	// API路由组
	api := r.Group("/api")
	{
		// 目标计算接口
		api.POST("/target/calculate", controller.CalculateTarget)

		// 历史记录接口
		api.GET("/target/history", controller.GetHistoryList)
		api.GET("/target/history/detail", controller.GetHistoryDetail)
		api.POST("/target/history/compare", controller.CompareHistory)
		api.DELETE("/target/history", controller.DeleteHistory)

		// 健康检查
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status": "ok",
				"message": "服务运行正常",
			})
		})
	}

	// Swagger文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
