package main

import (
	"log"
	scanHandler "vm-server/handlers/scan"
	"vm-server/handlers/secrets"
	"vm-server/jobs"
	scanService "vm-server/services/scan"
	"vm-server/store"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Initialize data store
	dbStore, err := store.NewStore()
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}

	// Initialize services
	scanSvc := scanService.NewService(dbStore)

	// Initialize job runner
	scanner := jobs.NewScanner(scanSvc)

	// Initialize handlers
	scanHdlr := scanHandler.NewHandler(scanSvc, scanner)
	err = scanner.Cleanup()
	if err != nil {
		log.Fatalf("Failed to cleanup: %v", err)
	}

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	apiV1 := e.Group("/api/v1")

	// Scan routes
	apiV1.POST("/scan", scanHdlr.ScanHandler)
	apiV1.GET("/scan/:id", scanHdlr.ScanByIDHandler)

	// Secrets routes
	apiV1.POST("/secrets/:id/version", secrets.CreateSecretVersionHandler)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
