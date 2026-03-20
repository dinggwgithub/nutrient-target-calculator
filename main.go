package main

import (
	"log"
	"target-calculator-from-db/config"
	"target-calculator-from-db/routes"

	_ "target-calculator-from-db/docs"
)

// @title 营养素目标计算器API
// @version 1.0
// @description 基于数据库摄入数据和DRIs计算营养素目标摄入量的API
// @host localhost:8080
// @BasePath /api
func main() {
	// 初始化数据库连接
	if err := config.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 设置路由
	r := routes.SetupRouter()

	// 启动服务器
	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
