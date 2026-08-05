package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Backup 备份结构体
type Backup struct {
	Name            string `json:"name"`
	Mode            string `json:"mode"`
	Status          string `json:"status"`
	Size            string `json:"size"`
	Time            string `json:"time"`
	EtcdRevision    int64  `json:"etcdRevision"`
	StorageLocation string `json:"storageLocation"`
	Validation      string `json:"validation"`
}

// ListBackups 获取备份列表
func ListBackups(c *gin.Context) {
	// Mock数据
	backups := []Backup{
		{
			Name:            "daily-backup-20260210",
			Mode:            "Full",
			Status:          "Completed",
			Size:            "1.2 GB",
			Time:            "2026-02-10 14:30:00",
			EtcdRevision:    123456,
			StorageLocation: "s3://my-backups/daily-backup-20260210.db",
			Validation:      "Passed",
		},
		{
			Name:            "hourly-backup-20260210-13",
			Mode:            "Incremental",
			Status:          "Completed",
			Size:            "24 MB",
			Time:            "2026-02-10 13:00:00",
			EtcdRevision:    123400,
			StorageLocation: "s3://my-backups/hourly-backup-20260210-13.db",
			Validation:      "Passed",
		},
		{
			Name:            "hourly-backup-20260210-12",
			Mode:            "Incremental",
			Status:          "Completed",
			Size:            "18 MB",
			Time:            "2026-02-10 12:00:00",
			EtcdRevision:    123350,
			StorageLocation: "s3://my-backups/hourly-backup-20260210-12.db",
			Validation:      "Passed",
		},
		{
			Name:            "hourly-backup-20260210-11",
			Mode:            "Incremental",
			Status:          "Failed",
			Size:            "0 MB",
			Time:            "2026-02-10 11:00:00",
			EtcdRevision:    0,
			StorageLocation: "",
			Validation:      "Failed",
		},
	}

	c.JSON(http.StatusOK, backups)
}

// CreateBackup 创建备份
func CreateBackup(c *gin.Context) {
	var backup Backup
	if err := c.ShouldBindJSON(&backup); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Mock创建备份
	backup.Status = "Completed"
	backup.Size = "1.0 GB"
	backup.Time = "2026-02-10 15:00:00"
	backup.EtcdRevision = 123500
	backup.StorageLocation = "s3://my-backups/" + backup.Name + ".db"
	backup.Validation = "Passed"

	c.JSON(http.StatusCreated, backup)
}

// GetBackup 获取单个备份详情
func GetBackup(c *gin.Context) {
	name := c.Param("name")

	// Mock数据
	backup := Backup{
		Name:            name,
		Mode:            "Full",
		Status:          "Completed",
		Size:            "1.2 GB",
		Time:            "2026-02-10 14:30:00",
		EtcdRevision:    123456,
		StorageLocation: "s3://my-backups/" + name + ".db",
		Validation:      "Passed",
	}

	c.JSON(http.StatusOK, backup)
}

// DeleteBackup 删除备份
func DeleteBackup(c *gin.Context) {
	name := c.Param("name")

	// Mock删除备份
	c.JSON(http.StatusOK, gin.H{"message": "Backup " + name + " deleted successfully"})
}