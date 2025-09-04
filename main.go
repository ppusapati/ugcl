package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"p9e.in/ugcl/config"
	"p9e.in/ugcl/middleware"
	"p9e.in/ugcl/projects/handlers"
	"p9e.in/ugcl/projects/services"

	"connectrpc.com/connect"
	"connectrpc.com/grpchealth"
	"connectrpc.com/grpcreflect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	// Import the Connect generated service
	"p9e.in/ugcl/projects/api/v2/dairy_site/dairy_siteconnect"
)

var (
	Version   = "dev"
	BuildTime = ""
)

func main() {
	versionFlag := flag.Bool("version", false, "Print version info and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("Version:   %s\n", Version)
		fmt.Printf("BuildTime: %s\n", BuildTime)
		os.Exit(0)
	}

	config.Connect()
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9090"
	}

	// Start Connect server in goroutine

	stop := make(chan struct{})

	go func() {
		startConnectServer(grpcPort)
		close(stop)
	}()

	if err := config.Migrations(config.DB); err != nil {
		log.Fatalf("could not run migrations: %v", err)
	}

	<-stop // Wait here
	// Perform any cleanup if necessary
	log.Println("Server stopped gracefully")
	os.Exit(0)
}

func startConnectServer(port string) {
	// Create your service handler
	dairySiteHandler := &handlers.ConnectDairySiteHandler{
		DairySiteServices: &services.DairySiteServices{
			DB: config.DB,
		},
	}

	// Create HTTP mux
	mux := http.NewServeMux()

	// Register Connect services with support for all three protocols
	path, handler := dairy_siteconnect.NewDairySiteServiceHandler(dairySiteHandler,
		connect.WithCompressMinBytes(0),                        // Optional: Compress everything
		connect.WithInterceptors(middleware.AuthInterceptor()), // Optional: Add your auth interceptor
	)
	mux.Handle(path, handler)

	// Add health check (optional but recommended)
	checker := grpchealth.NewStaticChecker(
		"p9e.ugcl.projects.api.v2.dairy_site.DairySiteService",
	)
	mux.Handle(grpchealth.NewHandler(checker))

	// Add reflection (optional, useful for debugging)
	reflector := grpcreflect.NewStaticReflector(
		"p9e.ugcl.projects.api.v2.dairy_site.DairySiteService",
	)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	// Add CORS middleware for all protocols
	corsHandler := addMultiProtocolCORS(mux)

	// Support HTTP/2 without TLS (h2c) for better performance
	h2cHandler := h2c.NewHandler(corsHandler, &http2.Server{})

	log.Println("🚀 Multi-protocol server starting at port", port)
	log.Println("📡 Supports: Connect, gRPC, and gRPC-Web protocols")

	if err := http.ListenAndServe(":"+port, h2cHandler); err != nil {
		log.Fatalf("failed to serve multi-protocol server: %v", err)
	}
	log.Println("Server started successfully")
}

// Enhanced CORS middleware for all three protocols (Connect, gRPC, gRPC-Web)
func addMultiProtocolCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for all protocols
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// Headers for Connect, gRPC, and gRPC-Web
		w.Header().Set("Access-Control-Allow-Headers",
			"Content-Type, Authorization, x-api-key, X-Requested-With, "+
				// Connect headers
				"Connect-Protocol-Version, Connect-Timeout-Ms, Connect-Accept-Encoding, "+
				// gRPC headers
				"Grpc-Timeout, Grpc-Encoding, Grpc-Accept-Encoding, "+
				// gRPC-Web headers
				"x-grpc-web, x-user-agent, grpc-timeout, grpc-encoding, grpc-accept-encoding")

		// Expose headers for clients
		w.Header().Set("Access-Control-Expose-Headers",
			"Connect-Protocol-Version, Connect-Timeout-Ms, Connect-Accept-Encoding, "+
				"Grpc-Status, Grpc-Message, Grpc-Encoding, Grpc-Accept-Encoding")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
