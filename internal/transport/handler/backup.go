package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BackupHandler struct {
	db     *gorm.DB
	dbPath string
}

func NewBackupHandler(db *gorm.DB, dbPath string) *BackupHandler {
	return &BackupHandler{db: db, dbPath: dbPath}
}

// BackupInfo 返回数据库信息（GET /api/admin/backup/info）
func (h *BackupHandler) BackupInfo(c *gin.Context) {
	info := gin.H{
		"type": "sqlite",
		"path": h.dbPath,
	}

	if stat, err := os.Stat(h.dbPath); err == nil {
		info["size"] = stat.Size()
		info["lastModified"] = stat.ModTime().Format(time.RFC3339)
	}

	c.JSON(http.StatusOK, info)
}

// Backup 生成并下载备份文件（POST /api/admin/backup）
func (h *BackupHandler) Backup(c *gin.Context) {
	if h.dbPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "当前为 PostgreSQL 模式，请使用 pg_dump 工具备份"})
		return
	}

	// 检查原始文件存在
	if _, err := os.Stat(h.dbPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "数据库文件不存在"})
		return
	}

	// 使用 VACUUM INTO 生成干净备份
	backupDir := "backups"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建备份目录失败"})
		return
	}

	backupPath := fmt.Sprintf("%s/backup_%s.db", backupDir, time.Now().Format("20060102_150405"))

	// VACUUM INTO 会生成一个无碎片的干净副本
	result := h.db.Exec(fmt.Sprintf("VACUUM INTO '%s'", backupPath))
	if result.Error != nil {
		// 回退方案：直接复制文件
		if err := copyFile(h.dbPath, backupPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "备份失败: " + err.Error()})
			return
		}
	}

	// 获取文件信息
	stat, err := os.Stat(backupPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "备份文件异常"})
		return
	}

	// 返回文件下载
	filename := fmt.Sprintf("租赁系统备份_%s.db", time.Now().Format("2006-01-02_150405"))
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Content-Length", fmt.Sprintf("%d", stat.Size()))
	c.File(backupPath)

	// 清理临时文件（延迟删除，给下载时间）
	go func() {
		time.Sleep(5 * time.Minute)
		os.Remove(backupPath)
	}()
}

// Restore 从上传的备份文件恢复（POST /api/admin/restore）
func (h *BackupHandler) Restore(c *gin.Context) {
	if h.dbPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "当前为 PostgreSQL 模式，请使用 psql 工具恢复"})
		return
	}

	// 必须带确认参数
	confirmed := c.Query("confirmed")
	if confirmed != "true" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "恢复操作需要确认，请传入 confirmed=true"})
		return
	}

	// 接收上传的文件
	file, err := c.FormFile("backup")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请上传备份文件"})
		return
	}

	// 验证文件大小（最大 500MB）
	if file.Size > 500*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "备份文件过大（最大 500MB）"})
		return
	}

	// 打开上传文件
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法读取上传文件"})
		return
	}
	defer src.Close()

	// 先备份当前数据库（以防恢复失败）
	currentBackup := h.dbPath + ".before_restore"
	if err := copyFile(h.dbPath, currentBackup); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法备份当前数据库"})
		return
	}

	// 关闭当前数据库连接
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取数据库连接"})
		return
	}
	sqlDB.Close()

	// 写入新文件
	dst, err := os.Create(h.dbPath)
	if err != nil {
		// 恢复原来的文件
		copyFile(currentBackup, h.dbPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法写入数据库文件"})
		return
	}

	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		copyFile(currentBackup, h.dbPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "写入备份数据失败"})
		return
	}
	dst.Close()

	c.JSON(http.StatusOK, gin.H{
		"message":    "数据恢复成功，请重新启动服务",
		"backupFile": currentBackup,
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
