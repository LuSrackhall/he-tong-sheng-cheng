package middleware

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func SPAFallbackEmbed(distFS fs.FS) gin.HandlerFunc {
	fileServer := http.FileServer(http.FS(distFS))

	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.Next()
			return
		}

		path := strings.TrimPrefix(c.Request.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		if _, err := fs.Stat(distFS, path); err != nil {
			c.Request.URL.Path = "/index.html"
		} else {
			c.Request.URL.Path = "/" + path
		}

		fileServer.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}
