package main

import (
	"embed"
	"asset-leasing-system/internal/config"
	"asset-leasing-system/internal/di"
	"asset-leasing-system/internal/transport/handler"
	"asset-leasing-system/internal/transport/middleware"
	"io/fs"
	"log"

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

	authmw := middleware.NewAuthMiddleware(cfg.JWTSecret)
	authH := handler.NewAuthHandler(deps.UserRepo, cfg.JWTSecret)
	assetH := handler.NewAssetHandler(deps.AssetRepo)
	tenantH := handler.NewTenantHandler(deps.TenantRepo)
	contractH := handler.NewContractHandler(deps.ContractRepo, deps.TemplateRepo)
	paymentH := handler.NewPaymentHandler(deps.PaymentRepo, deps.ContractRepo, deps.ReceiptBookRepo, deps.ReceiptRepo)
	receiptBookH := handler.NewReceiptBookHandler(deps.ReceiptBookRepo)
	arrearsH := handler.NewArrearsHandler(deps.ContractRepo)

	r := gin.New()

	distSub, err := fs.Sub(distFS, "dist")
	if err != nil {
		log.Fatalf("Failed to open embedded dist: %v", err)
	}

	// SPA middleware runs before routing to avoid gin's RedirectTrailingSlash
	r.Use(middleware.SPAFallbackEmbed(distSub))
	r.Use(gin.Logger(), gin.Recovery())

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

			protected.GET("/templates", contractH.ListTemplates)
			protected.POST("/templates", contractH.CreateTemplate)
			protected.PATCH("/templates/:id", contractH.UpdateTemplateMapping)
			protected.POST("/templates/:id/upload", contractH.UploadTemplate)
			protected.DELETE("/templates/:id", contractH.DeleteTemplate)

			protected.POST("/contracts/:id/export", contractH.ExportContract)
			protected.GET("/contracts/:id/download", contractH.DownloadContract)

			protected.GET("/receipt-books", receiptBookH.List)
			protected.POST("/receipt-books", receiptBookH.Create)

			protected.GET("/arrears", arrearsH.List)
		}

		admin := api.Group("/admin")
		admin.Use(authmw.RequireAuth(), authmw.RequireAdmin())
		{
			admin.GET("/users", authH.ListUsers)
			admin.POST("/users", authH.CreateUser)
			admin.DELETE("/users/:id", authH.DeleteUser)
		}
	}

	log.Printf("Server starting on :%s\n", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
