package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Restore 恢复结构体
type Restore struct {
	Name        string `json:"name"`
	BackupName  string `json:"backupName"`
	Status      string `json:"status"`
	Time        string `json:"time"`
	EtcdCluster string `json:"etcdCluster"`
}

// ListRestores 获取恢复列表
func ListRestores(c *gin.Context) {
	// Mock数据
	restores := []Restore{
		{
			Name:        "restore-20260210",
			BackupName:  "daily-backup-20260209",
			Status:      "Completed",
			Time:        "2026-02-10 08:30:00",
			EtcdCluster: "https://etcd-0:2379,https://etcd-1:2379,https://etcd-2:2379",
		},
		{
			Name:        "restore-20260209",
			BackupName:  "daily-backup-20260208",
			Status:      "Completed",
			Time:        "2026-02-09 14:00:00",
			EtcdCluster: "https://etcd-0:2379,https://etcd-1:2379,https://etcd-2:2379",
		},
		{
			Name:        "restore-20260208",
			BackupName:  "daily-backup-20260207",
			Status:      "Failed",
			Time:        "2026-02-08 10:15:00",
			EtcdCluster: "https://etcd-0:2379,https://etcd-1:2379,https://etcd-2:2379",
		},
	}

	c.JSON(http.StatusOK, restores)
}

// CreateRestore 创建恢复
func CreateRestore(c *gin.Context) {
	var restore Restore
	if err := c.ShouldBindJSON(&restore); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Mock创建恢复
	restore.Status = "Completed"
	restore.Time = "2026-02-10 15:30:00"

	c.JSON(http.StatusCreated, restore)
}

// GetRestore 获取单个恢复详情
func GetRestore(c *gin.Context) {
	name := c.Param("name")

	// Mock数据
	restore := Restore{
		Name:        name,
		BackupName:  "daily-backup-20260209",
		Status:      "Completed",
		Time:        "2026-02-10 08:30:00",
		EtcdCluster: "https://etcd-0:2379,https://etcd-1:2379,https://etcd-2:2379",
	}

	c.JSON(http.StatusOK, restore)
}

// DeleteRestore 删除恢复
func DeleteRestore(c *gin.Context) {
	name := c.Param("name")

	// Mock删除恢复
	c.JSON(http.StatusOK, gin.H{"message": "Restore " + name + " deleted successfully"})
}