package middleware

import (
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func SPAFallbackEmbed(distFS fs.FS) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.Next()
			return
		}

		path := strings.TrimPrefix(c.Request.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		data, err := fs.ReadFile(distFS, path)
		if err != nil {
			data, err = fs.ReadFile(distFS, "index.html")
			if err != nil {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
			path = "index.html"
		}

		ext := filepath.Ext(path)
		ct := mime.TypeByExtension(ext)
		if ct == "" {
			ct = "application/octet-stream"
		}

		// 根据路径设置不同的缓存策略
		switch {
		case path == "index.html":
			c.Header("Cache-Control", "no-cache")
		case strings.HasPrefix(path, "assets/"):
			// Vite 构建产物带内容哈希，可长期缓存
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		default:
			c.Header("Cache-Control", "public, max-age=3600")
		}

		c.Data(http.StatusOK, ct, data)
		c.Abort()
	}
}
