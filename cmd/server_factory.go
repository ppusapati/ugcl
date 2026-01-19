package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"go.uber.org/fx"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"gorm.io/gorm"

	conf "p9e.in/ugcl/packages/api/v1/config"
)

// SetupServerLifecycle sets up the HTTP server lifecycle with config-based port
func SetupServerLifecycle(
	lc fx.Lifecycle,
	db *gorm.DB,
	mux *http.ServeMux,
	serverConfig *conf.Server, // Inject server config
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Get port from config with fallback
			port := getServerPort(serverConfig)

			// Setup middleware chain
			handler := setupMiddleware(mux)

			// Start server
			go startServer(port, handler)

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("🛑 UGCL Platform shutting down gracefully...")
			return nil
		},
	})
}

// getServerPort extracts the port from server config with fallbacks
func getServerPort(serverConfig *conf.Server) string {
	// Try to get port from server config
	if serverConfig != nil {
		// Check if your config has a GRPC field
		if serverConfig.Grpc != nil && serverConfig.Grpc.Addr != "" {
			// Extract port from address like ":9090" or "localhost:9090"
			addr := serverConfig.Grpc.Addr
			if addr[0] == ':' {
				return addr[1:] // Remove the leading ':'
			}
			// If it's "host:port" format, extract port
			if colonIndex := len(addr) - 1; colonIndex > 0 {
				for i := len(addr) - 1; i >= 0; i-- {
					if addr[i] == ':' {
						return addr[i+1:]
					}
				}
			}
		}

		// Check if there's a generic port field
		if serverConfig.Http != nil && serverConfig.Http.Addr != "" {
			// Similar logic for HTTP port if GRPC not specified
			addr := serverConfig.Http.Addr
			if addr[0] == ':' {
				return addr[1:]
			}
			for i := len(addr) - 1; i >= 0; i-- {
				if addr[i] == ':' {
					return addr[i+1:]
				}
			}
		}
	}

	// Fallback to environment variable if config doesn't have port
	if envPort := os.Getenv("GRPC_PORT"); envPort != "" {
		return envPort
	}

	// Final fallback
	return "9090"
}

// Setup middleware chain
func setupMiddleware(mux *http.ServeMux) http.Handler {
	// CORS middleware
	corsHandler := addMultiProtocolCORS(mux)

	// HTTP/2 handler
	h2cHandler := h2c.NewHandler(corsHandler, &http2.Server{})

	return h2cHandler
}

// Start the HTTP server
func startServer(port string, handler http.Handler) {
	log.Printf("🚀 UGCL Multi-Protocol Server starting on port %s", port)
	log.Println("📡 Protocols: Connect, gRPC, gRPC-Web")
	log.Printf("🔗 Platform ready at: http://localhost:%s", port)

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}

// Enhanced CORS middleware with better API key support
func addMultiProtocolCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow all origins for development (customize for production)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// Enhanced headers to support both JWT and API key authentication
		w.Header().Set("Access-Control-Allow-Headers",
			"Content-Type, Authorization, x-api-key, X-Requested-With, "+
				"Connect-Protocol-Version, Connect-Timeout-Ms, Connect-Accept-Encoding, "+
				"Grpc-Timeout, Grpc-Encoding, Grpc-Accept-Encoding, "+
				"x-grpc-web, x-user-agent, grpc-timeout, grpc-encoding, grpc-accept-encoding")

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
