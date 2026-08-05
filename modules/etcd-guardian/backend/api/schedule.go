package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Schedule 调度结构体
type Schedule struct {
	Name            string `json:"name"`
	Schedule        string `json:"schedule"`
	Mode            string `json:"mode"`
	LastRun         string `json:"lastRun"`
	NextRun         string `json:"nextRun"`
	Status          string `json:"status"`
	StorageProvider string `json:"storageProvider"`
	Bucket          string `json:"bucket"`
	Region          string `json:"region"`
	CredentialsSecret string `json:"credentialsSecret"`
	EtcdEndpoints   string `json:"etcdEndpoints"`
	Validation      bool   `json:"validation"`
	ConsistencyCheck bool   `json:"consistencyCheck"`
}

// ListSchedules 获取调度列表
func ListSchedules(c *gin.Context) {
	// Mock数据
	schedules := []Schedule{
		{
			Name:            "daily-backup",
			Schedule:        "0 0 * * *",
			Mode:            "Full",
			LastRun:         "2026-02-10 00:00:00",
			NextRun:         "2026-02-11 00:00:00",
			Status:          "Active",
			StorageProvider: "S3",
			Bucket:          "my-backups",
			Region:          "us-east-1",
			CredentialsSecret: "s3-credentials",
			Validation:      true,
			ConsistencyCheck: true,
		},
		{
			Name:            "hourly-backup",
			Schedule:        "0 * * * *",
			Mode:            "Incremental",
			LastRun:         "2026-02-10 14:00:00",
			NextRun:         "2026-02-10 15:00:00",
			Status:          "Active",
			StorageProvider: "S3",
			Bucket:          "my-backups",
			Region:          "us-east-1",
			CredentialsSecret: "s3-credentials",
			Validation:      true,
			ConsistencyCheck: false,
		},
		{
			Name:            "weekly-backup",
			Schedule:        "0 0 * * 0",
			Mode:            "Full",
			LastRun:         "2026-02-08 00:00:00",
			NextRun:         "2026-02-15 00:00:00",
			Status:          "Active",
			StorageProvider: "S3",
			Bucket:          "my-backups",
			Region:          "us-east-1",
			CredentialsSecret: "s3-credentials",
			Validation:      true,
			ConsistencyCheck: true,
		},
	}

	c.JSON(http.StatusOK, schedules)
}

// CreateSchedule 创建调度
func CreateSchedule(c *gin.Context) {
	var schedule Schedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Mock创建调度
	schedule.Status = "Active"
	schedule.LastRun = "Never"
	schedule.NextRun = "2026-02-10 16:00:00"

	c.JSON(http.StatusCreated, schedule)
}

// GetSchedule 获取单个调度详情
func GetSchedule(c *gin.Context) {
	name := c.Param("name")

	// Mock数据
	schedule := Schedule{
		Name:            name,
		Schedule:        "0 0 * * *",
		Mode:            "Full",
		LastRun:         "2026-02-10 00:00:00",
		NextRun:         "2026-02-11 00:00:00",
		Status:          "Active",
		StorageProvider: "S3",
		Bucket:          "my-backups",
		Region:          "us-east-1",
		CredentialsSecret: "s3-credentials",
		Validation:      true,
		ConsistencyCheck: true,
	}

	c.JSON(http.StatusOK, schedule)
}

// UpdateSchedule 更新调度
func UpdateSchedule(c *gin.Context) {
	name := c.Param("name")
	var schedule Schedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Mock更新调度
	schedule.Name = name
	schedule.Status = "Active"

	c.JSON(http.StatusOK, schedule)
}

// DeleteSchedule 删除调度
func DeleteSchedule(c *gin.Context) {
	name := c.Param("name")

	// Mock删除调度
	c.JSON(http.StatusOK, gin.H{"message": "Schedule " + name + " deleted successfully"})
}

// RunSchedule 立即运行调度
func RunSchedule(c *gin.Context) {
	name := c.Param("name")

	// Mock运行调度
	c.JSON(http.StatusOK, gin.H{"message": "Schedule " + name + " run started"})
}