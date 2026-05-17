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
		c.Data(http.StatusOK, ct, data)
		c.Abort()
	}
}
