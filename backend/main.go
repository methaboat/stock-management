package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"stock-management/backend/app"
	"stock-management/backend/config"

	authlogin "stock-management/backend/app/auth/login"
	auditlist "stock-management/backend/app/audit/list"
	inboundcreate "stock-management/backend/app/inbound/create"
	inboundlist "stock-management/backend/app/inbound/list"
	inboundreceive "stock-management/backend/app/inbound/receive"
	inboundsend "stock-management/backend/app/inbound/send"
	rmacreate "stock-management/backend/app/outbound/rma/create"
	rmalist "stock-management/backend/app/outbound/rma/list"
	socreate "stock-management/backend/app/outbound/so/create"
	solist "stock-management/backend/app/outbound/so/list"
	soship "stock-management/backend/app/outbound/so/ship"
	productcreate "stock-management/backend/app/product/create"
	productdeletion "stock-management/backend/app/product/deletion"
	productlist "stock-management/backend/app/product/list"
	productupdate "stock-management/backend/app/product/update"
	reconapprove "stock-management/backend/app/reconciliation/approve"
	reconlist "stock-management/backend/app/reconciliation/list"
	reconsubmit "stock-management/backend/app/reconciliation/submit"
	reportvaluation "stock-management/backend/app/reports/valuation"
	stockadjust "stock-management/backend/app/stock/adjust"
	stocklist "stock-management/backend/app/stock/list"
	stocklowalerts "stock-management/backend/app/stock/lowalerts"
	stockscan "stock-management/backend/app/stock/scan"
)

func main() {
	if err := app.ConnectDB(); err != nil {
		log.Fatalf("Fatal error connecting to Database: %v", err)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	r.GET("/health/liveness", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "UP"}) })
	r.GET("/health/readiness", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "READY"}) })

	r.Use(app.AllowOriginMiddleware())

	// Auth
	loginHandler := authlogin.NewHandler(authlogin.NewStorage(app.UserCol))

	// Products
	productListHandler := productlist.NewHandler(productlist.NewStorage(app.ProductsCol))
	productCreateHandler := productcreate.NewHandler(productcreate.NewStorage(app.ProductsCol, app.StockCol))
	productUpdateHandler := productupdate.NewHandler(productupdate.NewStorage(app.ProductsCol))
	productDeleteHandler := productdeletion.NewHandler(productdeletion.NewStorage(app.ProductsCol, app.StockCol))

	// Stock
	stockListHandler := stocklist.NewHandler(stocklist.NewStorage(app.StockCol))
	stockAdjustHandler := stockadjust.NewHandler(stockadjust.NewStorage(app.StockCol))
	stockLowAlertsHandler := stocklowalerts.NewHandler(stocklowalerts.NewStorage(app.StockCol))
	stockScanHandler := stockscan.NewHandler(stockscan.NewStorage(app.StockCol, app.ProductsCol))

	// Inbound (PO)
	inboundListHandler := inboundlist.NewHandler(inboundlist.NewStorage(app.POCol))
	inboundCreateHandler := inboundcreate.NewHandler(inboundcreate.NewStorage(app.POCol))
	inboundSendHandler := inboundsend.NewHandler(inboundsend.NewStorage(app.POCol))
	inboundReceiveHandler := inboundreceive.NewHandler(inboundreceive.NewStorage(app.POCol, app.StockCol))

	// Outbound SO
	soListHandler := solist.NewHandler(solist.NewStorage(app.SOCol))
	soCreateHandler := socreate.NewHandler(socreate.NewStorage(app.SOCol, app.StockCol))
	soShipHandler := soship.NewHandler(soship.NewStorage(app.SOCol, app.StockCol))

	// Outbound RMA
	rmaListHandler := rmalist.NewHandler(rmalist.NewStorage(app.RMACol))
	rmaCreateHandler := rmacreate.NewHandler(rmacreate.NewStorage(app.RMACol, app.StockCol))

	// Audit & Reports
	auditListHandler := auditlist.NewHandler(auditlist.NewStorage(app.AuditCol))
	valuationHandler := reportvaluation.NewHandler(reportvaluation.NewStorage(app.ProductsCol, app.StockCol, app.POCol))

	// Reconciliation
	reconListHandler := reconlist.NewHandler(reconlist.NewStorage(app.ReconCol))
	reconSubmitHandler := reconsubmit.NewHandler(reconsubmit.NewStorage(app.ReconCol, app.StockCol, app.ProductsCol))
	reconApproveHandler := reconapprove.NewHandler(reconapprove.NewStorage(app.ReconCol, app.StockCol))

	api := r.Group("/api")
	api.POST("/login", loginHandler.Handle)

	auth := api.Group("")
	auth.Use(app.MidDecodeJWT())
	{
		// Products
		auth.GET("/products", productListHandler.Handle)
		auth.POST("/products", app.MidRequireRole([]string{"Admin", "SalesRep"}), productCreateHandler.Handle)
		auth.PUT("/products/:id", app.MidRequireRole([]string{"Admin"}), productUpdateHandler.Handle)
		auth.DELETE("/products/:id", app.MidRequireRole([]string{"Admin"}), productDeleteHandler.Handle)

		// Stock
		auth.GET("/stock", stockListHandler.Handle)
		auth.POST("/stock/adjust", app.MidRequireRole([]string{"Admin", "WarehouseStaff"}), stockAdjustHandler.Handle)
		auth.GET("/stock/low-alerts", stockLowAlertsHandler.Handle)
		auth.POST("/stock/scan", app.MidRequireRole([]string{"Admin", "WarehouseStaff"}), stockScanHandler.Handle)

		// Inbound PO
		auth.GET("/inbound/po", inboundListHandler.Handle)
		auth.POST("/inbound/po", app.MidRequireRole([]string{"Admin", "SalesRep"}), inboundCreateHandler.Handle)
		auth.POST("/inbound/po/send/:id", app.MidRequireRole([]string{"Admin", "SalesRep"}), inboundSendHandler.Handle)
		auth.POST("/inbound/po/receive/:id", app.MidRequireRole([]string{"Admin", "WarehouseStaff"}), inboundReceiveHandler.Handle)

		// Outbound SO
		auth.GET("/outbound/so", soListHandler.Handle)
		auth.POST("/outbound/so", app.MidRequireRole([]string{"Admin", "SalesRep"}), soCreateHandler.Handle)
		auth.POST("/outbound/so/ship/:id", app.MidRequireRole([]string{"Admin", "WarehouseStaff"}), soShipHandler.Handle)

		// Outbound RMA
		auth.GET("/outbound/rma", rmaListHandler.Handle)
		auth.POST("/outbound/rma", app.MidRequireRole([]string{"Admin", "WarehouseStaff"}), rmaCreateHandler.Handle)

		// Audit & Reports
		auth.GET("/audit", auditListHandler.Handle)
		auth.GET("/reports/valuation", valuationHandler.Handle)

		// Reconciliation
		auth.GET("/reconciliation", reconListHandler.Handle)
		auth.POST("/reconciliation", app.MidRequireRole([]string{"Admin", "WarehouseStaff"}), reconSubmitHandler.Handle)
		auth.POST("/reconciliation/:id/approve", app.MidRequireRole([]string{"Admin"}), reconApproveHandler.Handle)
	}

	log.Printf("Stock Flow Gin Backend started on http://localhost:%s...", config.C.ServerPort)
	if err := r.Run(":" + config.C.ServerPort); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}
