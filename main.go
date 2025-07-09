package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"log"
	"os"
	scanHandler "vm-server/handlers/scan"
	"vm-server/handlers/secrets"
	"vm-server/jobs"
	scanService "vm-server/services/scan"
	secretsvc "vm-server/services/secrets"
	"vm-server/store"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Command-line flags for TLS files
	caFile := flag.String("ca-file", "", "Path to the CA certificate file")
	certFile := flag.String("cert-file", "", "Path to the server certificate file")
	keyFile := flag.String("key-file", "", "Path to the server key file")
	port := flag.String("port", "1323", "Port for the server to listen on")
	flag.Parse()

	// Initialize data store
	dbStore, err := store.NewStore()
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}

	// Initialize services
	scanSvc := scanService.NewService(dbStore)

	// Initialize services
	secretsSvc := secretsvc.NewService(scanSvc)

	// Initialize job runner
	scanner := jobs.NewScanner(scanSvc)

	// Initialize handlers
	scanHdlr := scanHandler.NewHandler(scanSvc, scanner)
	err = scanner.Cleanup()
	if err != nil {
		log.Fatalf("Failed to cleanup: %v", err)
	}
	secretHdlr := secrets.NewHandler(secretsSvc, scanSvc)

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
	apiV1.POST("/secrets/:id/version", secretHdlr.CreateSecretVersionHandler)

	// Configure mTLS
	if *caFile != "" && *certFile != "" && *keyFile != "" {
		caCert, err := os.ReadFile(*caFile)
		if err != nil {
			log.Fatalf("Failed to read CA certificate: %v", err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		tlsConfig := &tls.Config{
			ClientCAs:  caCertPool,
			ClientAuth: tls.RequireAndVerifyClientCert,
			MinVersion: tls.VersionTLS12,
		}

		e.Server.TLSConfig = tlsConfig
		addr := ":" + *port
		// Start server with TLS
		e.Logger.Fatal(e.StartTLS(addr, *certFile, *keyFile))
	} else {
		addr := ":" + *port
		// Start server without TLS
		e.Logger.Fatal(e.Start(addr))
	}
}
