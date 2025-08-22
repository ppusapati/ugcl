package http

import (
	"context"
	"net/http"
	"sync"

	"p9e.in/ugcl/packages/api/v1/config"
	"p9e.in/ugcl/packages/middleware/tenant"
	"p9e.in/ugcl/packages/p9log"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// CustomHttpServer represents a custom HTTP server.
type CustomHttpServer struct {
	httpServer    *http.Server
	mux           *runtime.ServeMux
	activeCounter sync.WaitGroup
	log           p9log.Helper
	cfg           *config.Server
}

// NewCustomHttpServer creates a new instance of the CustomHttpServer with a provided mux and custom header matcher.
func NewCustomHttpServer(cfg *config.Server, mux *runtime.ServeMux, log p9log.Helper) *CustomHttpServer {
	httpServer := &http.Server{
		Addr:    cfg.Http.Addr,
		Handler: nil,
	}

	return &CustomHttpServer{
		httpServer: httpServer,
		mux:        mux,
		log:        log,
		cfg:        cfg,
	}
}

// Registers a gRPC service with the gateway.
func (s *CustomHttpServer) RegisterService(
	ctx context.Context,
	endpoint string,
	mux *runtime.ServeMux,
	registerFunc func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error,
) {
	grpcDialOption := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err := registerFunc(ctx, mux, endpoint, grpcDialOption)
	if err != nil {
		s.log.Fatal("Failed to Register GRPC service to HTTP with error:", err)
	}
}

// ListenAndServe starts the HTTP server.
func (s *CustomHttpServer) ListenAndServe() error {
	// Start the HTTP server with your custom middleware
	s.httpServer.Handler = s.CreateHandler(s.mux)

	s.log.Info("Serving HTTP on connection: ", s.cfg.Http.Addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server.
func (s *CustomHttpServer) Shutdown(ctx context.Context) error {
	s.log.Info("Shutting down HTTP server...")
	return s.httpServer.Shutdown(ctx)
}

// createHandler creates the HTTP handler with middleware, CORS, and the header matcher.
func (s *CustomHttpServer) CreateHandler(mux *runtime.ServeMux) http.Handler {
	corsOptions := cors.Options{
		AllowedOrigins: []string{
			"http://localhost:4200",
			"http://localhost:3600",
			"http://192.168.0.9:3600",
			"http://23.30.104.173:8000",
		},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
		Debug:          true, // Set to false in production
	}
	corsHandler := cors.New(corsOptions)
	handler := tenant.HttpTenantMiddleware(corsHandler.Handler(mux))
	handler = s.ActiveCounterMiddleware(handler)
	return handler
}

func (s *CustomHttpServer) ActiveCounterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.activeCounter.Add(1)
		defer s.activeCounter.Done()
		next.ServeHTTP(w, r)
	})
}
