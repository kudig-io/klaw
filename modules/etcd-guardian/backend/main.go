package main

import (
	"log"

	"github.com/kudig-io/klaw/modules/etcd-guardian/backend/api"
	"github.com/kudig-io/klaw/modules/etcd-guardian/backend/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 创建Gin引擎
	r := gin.Default()

	// 配置CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 使用自定义中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())

	// 注册API路由
	apiGroup := r.Group("/api/v1")
	{
		// 备份相关路由
		apiGroup.GET("/backups", api.ListBackups)
		apiGroup.POST("/backups", api.CreateBackup)
		apiGroup.GET("/backups/:name", api.GetBackup)
		apiGroup.DELETE("/backups/:name", api.DeleteBackup)

		// 恢复相关路由
		apiGroup.GET("/restores", api.ListRestores)
		apiGroup.POST("/restores", api.CreateRestore)
		apiGroup.GET("/restores/:name", api.GetRestore)
		apiGroup.DELETE("/restores/:name", api.DeleteRestore)

		// 调度相关路由
		apiGroup.GET("/schedules", api.ListSchedules)
		apiGroup.POST("/schedules", api.CreateSchedule)
		apiGroup.GET("/schedules/:name", api.GetSchedule)
		apiGroup.PUT("/schedules/:name", api.UpdateSchedule)
		apiGroup.DELETE("/schedules/:name", api.DeleteSchedule)
		apiGroup.POST("/schedules/:name/run", api.RunSchedule)

		// 健康检查
		apiGroup.GET("/health", api.HealthCheck)
	}

	// 启动服务器
	port := ":8080"
	log.Printf("Server starting on port %s", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}