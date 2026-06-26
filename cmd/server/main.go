package main

import (
	"embed"
	"asset-leasing-system/internal/config"
	"asset-leasing-system/internal/di"
	"asset-leasing-system/internal/security"
	"asset-leasing-system/internal/transport/handler"
	"asset-leasing-system/internal/transport/middleware"
	"context"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

//go:embed dist
var distFS embed.FS

func main() {
	cfg := config.Load()
	deps := di.Initialize(cfg)

	if cfg.Mode == "postgres" {
		log.Println("Running in PostgreSQL mode")
	} else {
		log.Println("Running in SQLite mode")
	}

	// 优雅关停 channel
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	authmw := middleware.NewAuthMiddleware(cfg.JWTSecret)
	loginLimiter := security.NewLoginRateLimiter()
	authH := handler.NewAuthHandler(deps.UserRepo, authmw, loginLimiter)
	assetH := handler.NewAssetHandler(deps.AssetRepo)
	tenantH := handler.NewTenantHandler(deps.TenantRepo)
	contractH := handler.NewContractHandler(deps.ContractRepo, deps.TemplateRepo)
	paymentH := handler.NewPaymentHandler(deps.PaymentRepo, deps.ContractRepo, deps.ReceiptBookRepo, deps.ReceiptRepo, deps.DB)
	receiptBookH := handler.NewReceiptBookHandler(deps.ReceiptBookRepo)
	receiptH := handler.NewReceiptHandler(deps.ReceiptRepo)
	arrearsH := handler.NewArrearsHandler(deps.ContractRepo)
	printH := handler.NewPrintHandler(deps.ReceiptRepo, deps.ReceiptBookRepo, deps.PaymentRepo, deps.ContractRepo, deps.TenantRepo, deps.AssetRepo, deps.DB)

	dbPath := ""
	if cfg.Mode != "postgres" {
		dbPath = cfg.DBName + ".db"
	}

	// 传递优雅关停函数给 BackupHandler
	shutdownFn := func() {
		quit <- syscall.SIGTERM
	}
	backupH := handler.NewBackupHandler(deps.DB, dbPath, shutdownFn)
	dashboardH := handler.NewDashboardHandler(deps.DashboardRepo)

	r := gin.New()
	r.MaxMultipartMemory = 10 << 20 // 10MB 请求体大小限制

	// CORS 中间件（仅在配置了 CORS_ORIGINS 时启用）
	if cfg.CORSOrigins != "" {
		origins := strings.Split(cfg.CORSOrigins, ",")
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
		}
		corsConfig := cors.Config{
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
			ExposeHeaders:    []string{"Content-Length", "Content-Disposition"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}
		if len(origins) == 1 && origins[0] == "*" {
			// AllowAllOrigins + AllowCredentials 违反 CORS 规范，改用回显 Origin
			corsConfig.AllowOriginFunc = func(origin string) bool {
				return true
			}
		} else {
			corsConfig.AllowOriginFunc = func(origin string) bool {
				for _, o := range origins {
					if o == origin {
						return true
					}
				}
				return false
			}
		}
		r.Use(cors.New(corsConfig))
	}

	distSub, err := fs.Sub(distFS, "dist")
	if err != nil {
		log.Fatalf("Failed to open embedded dist: %v", err)
	}

	// SPA middleware runs before routing to avoid gin's RedirectTrailingSlash
	r.Use(middleware.SPAFallbackEmbed(distSub))
	r.Use(gin.Logger())
	r.Use(gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, err any) {
		log.Printf("[PANIC] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
	}))

	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		api.POST("/auth/login", authH.Login)

		protected := api.Group("")
		protected.Use(authmw.RequireAuth())
		{
			protected.GET("/auth/me", authH.Me)
			protected.PUT("/auth/password", authH.ChangePassword)

			protected.GET("/assets", assetH.List)
			protected.POST("/assets", assetH.Create)
			protected.GET("/assets/:id", assetH.Get)
			protected.PATCH("/assets/:id", assetH.Update)

			protected.GET("/tenants", tenantH.List)
			protected.POST("/tenants", tenantH.Create)
			protected.GET("/tenants/:id", tenantH.Get)
			protected.PATCH("/tenants/:id", tenantH.Update)

			protected.GET("/contracts", contractH.List)
			protected.POST("/contracts", contractH.Create)
			protected.GET("/contracts/:id", contractH.Get)
			protected.PATCH("/contracts/:id", contractH.Update)

			protected.GET("/contracts/:id/payments", paymentH.ListByContract)
			protected.POST("/contracts/:id/payments", paymentH.Create)
			protected.POST("/payments/:id/void", paymentH.VoidPayment)

			protected.GET("/templates", contractH.ListTemplates)
			protected.POST("/templates", contractH.CreateTemplate)
			protected.PATCH("/templates/:id", contractH.UpdateTemplateMapping)
			protected.POST("/templates/:id/upload", contractH.UploadTemplate)
			protected.GET("/templates/:id/download", contractH.DownloadTemplate)
			protected.GET("/templates/:id/preview", contractH.PreviewTemplate)
			protected.DELETE("/templates/:id", contractH.DeleteTemplate)

			protected.GET("/contracts/:id/download", contractH.DownloadContract)
			protected.GET("/contracts/:id/preview", contractH.PreviewContract)

			protected.GET("/receipt-books", receiptBookH.List)
			protected.POST("/receipt-books", receiptBookH.Create)
			protected.GET("/receipts", receiptH.ListReceipts)

			protected.GET("/print/receipt/:id", printH.PrintReceipt)

			protected.GET("/arrears", arrearsH.List)

			protected.GET("/dashboard/stats", dashboardH.Stats)
		}

		admin := api.Group("/admin")
		admin.Use(authmw.RequireAuth(), authmw.RequireAdmin())
		{
			admin.GET("/users", authH.ListUsers)
			admin.POST("/users", authH.CreateUser)
			admin.DELETE("/users/:id", authH.DeleteUser)

			admin.GET("/backup/info", backupH.BackupInfo)
			admin.POST("/backup", backupH.Backup)
			admin.POST("/restore", backupH.Restore)
		}
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// 启动服务器（非阻塞）
	go func() {
		log.Printf("服务器启动，监听端口 :%s\n", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待关闭信号
	<-quit
	log.Println("收到关闭信号，正在优雅关停...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("服务器关闭失败: %v", err)
	}

	log.Println("服务器已关闭")
}
