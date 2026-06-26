package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SQLite magic header
var sqliteMagic = []byte("SQLite format 3\x00")

type BackupHandler struct {
	db         *gorm.DB
	dbPath     string
	shutdownFn func()
}

func NewBackupHandler(db *gorm.DB, dbPath string, shutdownFn func()) *BackupHandler {
	return &BackupHandler{db: db, dbPath: dbPath, shutdownFn: shutdownFn}
}

// BackupInfo 返回数据库信息（GET /api/admin/backup/info）
func (h *BackupHandler) BackupInfo(c *gin.Context) {
	if h.dbPath == "" {
		c.JSON(http.StatusOK, gin.H{"type": "postgres", "message": "PostgreSQL 模式请使用 pg_dump 工具备份"})
		return
	}

	info := gin.H{
		"type": "sqlite",
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "PostgreSQL 模式请使用 pg_dump 工具备份"})
		return
	}

	if _, err := os.Stat(h.dbPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "数据库文件不存在"})
		return
	}

	backupDir := "backups"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建备份目录失败"})
		return
	}

	backupPath := fmt.Sprintf("%s/backup_%s.db", backupDir, time.Now().Format("20060102_150405"))

	// VACUUM INTO 不支持参数化查询，需内联字符串（路径由 time.Now 生成，无用户输入风险）
	safeBackupPath := strings.ReplaceAll(backupPath, "'", "''")
	result := h.db.Exec(fmt.Sprintf("VACUUM INTO '%s'", safeBackupPath))
	if result.Error != nil {
		// 回退：直接复制
		if err := copyFile(h.dbPath, backupPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "备份失败"})
			return
		}
	}

	stat, err := os.Stat(backupPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "备份文件异常"})
		return
	}

	filename := fmt.Sprintf("租赁系统备份_%s.db", time.Now().Format("2006-01-02_150405"))
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Content-Length", fmt.Sprintf("%d", stat.Size()))
	c.File(backupPath)

	// 延迟清理
	go func() {
		time.Sleep(10 * time.Minute)
		os.Remove(backupPath)
	}()
}

// Restore 从上传的备份文件恢复（POST /api/admin/restore）
func (h *BackupHandler) Restore(c *gin.Context) {
	if h.dbPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "PostgreSQL 模式请使用 psql 工具恢复"})
		return
	}

	if c.PostForm("confirmed") != "true" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "恢复操作需要确认"})
		return
	}

	file, err := c.FormFile("backup")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请上传备份文件"})
		return
	}

	// 文件大小校验
	if file.Size < 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件太小，不是有效的数据库备份"})
		return
	}
	if file.Size > 500*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "备份文件过大（最大 500MB）"})
		return
	}

	// 读取并验证 SQLite magic header
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法读取上传文件"})
		return
	}
	defer src.Close()

	header := make([]byte, 16)
	if _, err := io.ReadFull(src, header); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法读取文件头"})
		return
	}
	if !bytes.Equal(header, sqliteMagic) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不是有效的 SQLite 数据库文件"})
		return
	}
	// 重置读取位置
	if seeker, ok := src.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	// 备份当前数据库
	currentBackup := h.dbPath + ".before_restore"
	if err := copyFile(h.dbPath, currentBackup); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法备份当前数据库"})
		return
	}

	// P0-3: 先写入临时文件，成功后再原子 rename
	tmpPath := h.dbPath + ".restore_tmp"
	dst, err := os.Create(tmpPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法创建临时文件"})
		return
	}

	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		os.Remove(tmpPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "写入备份数据失败"})
		return
	}
	dst.Close()

	// 原子替换（不手动关闭 DB，由优雅关停统一处理）
	if err := os.Rename(tmpPath, h.dbPath); err != nil {
		// rename 失败，尝试恢复原文件
		copyFile(currentBackup, h.dbPath)
		os.Remove(tmpPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "替换数据库文件失败，已恢复原数据"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "数据恢复成功，服务即将重启",
		"backupFile": currentBackup,
	})

	// 恢复成功后通过优雅关停退出，而非直接 os.Exit
	go func() {
		time.Sleep(1 * time.Second)
		if h.shutdownFn != nil {
			h.shutdownFn()
		}
	}()
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
