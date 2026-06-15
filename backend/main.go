package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"stock-management/backend/app"
	"stock-management/backend/app/auth"
	"stock-management/backend/app/audit"
	"stock-management/backend/app/po"
	"stock-management/backend/app/product"
	"stock-management/backend/app/reconciliation"
	"stock-management/backend/app/reports"
	"stock-management/backend/app/rma"
	"stock-management/backend/app/so"
	"stock-management/backend/app/stock"
)

func main() {
	// 1. Connect to MongoDB
	err := app.ConnectDB()
	if err != nil {
		log.Fatalf("Fatal error connecting to Database: %v", err)
	}

	// 2. Initialize router
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// Health Checks liveness/readiness mapping
	r.GET("/health/liveness", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})
	r.GET("/health/readiness", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "READY"})
	})

	// Apply CORS
	r.Use(app.AllowOriginMiddleware())

	// 3. Initialize Services and Handlers
	authHandler := auth.NewHandler(auth.NewService(app.UserCol))
	productHandler := product.NewHandler(product.NewService(app.ProductsCol, app.StockCol))
	stockHandler := stock.NewHandler(stock.NewService(app.StockCol, app.ProductsCol))
	poHandler := po.NewHandler(po.NewService(app.POCol, app.StockCol))
	soHandler := so.NewHandler(so.NewService(app.SOCol, app.StockCol))
	rmaHandler := rma.NewHandler(rma.NewService(app.RMACol, app.StockCol))
	auditHandler := audit.NewHandler(audit.NewService(app.AuditCol))
	reportsHandler := reports.NewHandler(reports.NewService(app.ProductsCol, app.StockCol, app.POCol))
	reconHandler := reconciliation.NewHandler(reconciliation.NewService(app.ReconCol, app.StockCol, app.ProductsCol))

	// 4. Map endpoints
	api := r.Group("/api")
	{
		// Public
		api.POST("/login", authHandler.Login)

		// Authenticated Routes
		authGroup := api.Group("")
		authGroup.Use(app.MidDecodeJWT())
		{
			// Products catalog
			authGroup.GET("/products", productHandler.List)
			authGroup.POST("/products", app.MidRequireRole([]string{"Admin", "SalesRep"}), productHandler.Create)
			authGroup.PUT("/products/:id", app.MidRequireRole([]string{"Admin"}), productHandler.Update)
			authGroup.DELETE("/products/:id", app.MidRequireRole([]string{"Admin"}), productHandler.Delete)

			// Stock levels
			authGroup.GET("/stock", stockHandler.List)
			authGroup.POST("/stock/adjust", app.MidRequireRole([]string{"Admin", "WarehouseStaff"}), stockHandler.Adjust)
			authGroup.GET("/stock/low-alerts", stockHandler.LowAlerts)
			authGroup.POST("/stock/scan", app.MidRequireRole([]string{"Admin", "WarehouseStaff"}), stockHandler.Scan)

			// Inbound POs & Receiving
			authGroup.GET("/inbound/po", poHandler.List)
			authGroup.POST("/inbound/po", app.MidRequireRole([]string{"Admin", "SalesRep"}), poHandler.Create)
			authGroup.POST("/inbound/po/send/:id", app.MidRequireRole([]string{"Admin", "SalesRep"}), poHandler.Send)
			authGroup.POST("/inbound/po/receive/:id", app.MidRequireRole([]string{"Admin", "WarehouseStaff"}), poHandler.Receive)

			// Outbound SOs & Dispatches
			authGroup.GET("/outbound/so", soHandler.List)
			authGroup.POST("/outbound/so", app.MidRequireRole([]string{"Admin", "SalesRep"}), soHandler.Create)
			authGroup.POST("/outbound/so/ship/:id", app.MidRequireRole([]string{"Admin", "WarehouseStaff"}), soHandler.Ship)

			// Returns (RMA)
			authGroup.GET("/outbound/rma", rmaHandler.List)
			authGroup.POST("/outbound/rma", app.MidRequireRole([]string{"Admin", "WarehouseStaff"}), rmaHandler.Create)

			// Audit Logs & Valuation Reports
			authGroup.GET("/audit", auditHandler.List)
			authGroup.GET("/reports/valuation", reportsHandler.Valuation)

			// Reconciliation
			authGroup.GET("/reconciliation", reconHandler.List)
			authGroup.POST("/reconciliation", app.MidRequireRole([]string{"Admin", "WarehouseStaff"}), reconHandler.Submit)
			authGroup.POST("/reconciliation/:id/approve", app.MidRequireRole([]string{"Admin"}), reconHandler.Approve)
		}
	}

	// 5. Start listener on port 8080
	log.Println("Stock Flow Gin Backend started on http://localhost:8080...")
	err = r.Run(":8080")
	if err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}
