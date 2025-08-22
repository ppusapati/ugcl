package tenant

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"p9e.in/ugcl/packages/models"
	"p9e.in/ugcl/packages/p9context"
	"p9e.in/ugcl/packages/p9log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ContextKey string

const TenantInfoKey ContextKey = "tenantInfo"

func HttpTenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL)
		// Extract tenant information from HTTP headers
		tenantID := r.Header.Get("X-Tenant-ID")
		tenantName := r.Header.Get("X-Tenant-Name")
		r.FormValue("X-Tenant-ID")
		r.Cookie("X-Tenant-ID")
		r.URL.Hostname()
		r.URL.Query().Get("X-Tenant-ID")
		// ctx := context.WithValue(r.Context(), "tenantID", tenantID)
		fmt.Println(tenantName)
		// Call the next handler with the modified context
		// Create a TenantInfo object
		tenant := models.TenantInfo{
			ID:   tenantID,
			Name: tenantName,
		}

		// Serialize the TenantInfo to JSON
		tenantJSON, err := json.Marshal(tenant)
		if err != nil {
			p9log.Error(err)
			return
		}

		// Convert the JSON to bytes
		tenantData := []byte(tenantJSON)

		// Create metadata with binary data
		tenantMetadata := metadata.Pairs("tenant_info", string(tenantData))

		// Create a context with the extracted tenant information
		ctx := metadata.NewOutgoingContext(r.Context(), tenantMetadata)
		// ctx := context.WithValue(r.Context(), "Grpc-Metadata-Namaste", "namaste")

		// Pass the updated context to the next HTTP handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GrpcTenantMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		p9log.Info("Metadata not found in gRPC context")
	}
	if headers, ok := metadata.FromIncomingContext(ctx); ok {
		xForwardFor := headers.Get("x-forwarded-host")
		if len(xForwardFor) > 0 && xForwardFor[0] != "" {
			ips := strings.Split(xForwardFor[0], ".")
			if len(ips) > 0 {
				clientIp := ips[0]
				fmt.Println(clientIp)
			}
		}
	}
	fmt.Printf("%+v%+v", md, ok)
	fmt.Println()
	// Extract the binary "tenant_info" metadata from the context
	tenantID, ok := md["x-tenant-id"]
	if !ok || len(tenantID) == 0 {
		p9log.Error("tenant id not found in metadata")
		return nil, status.Errorf(codes.NotFound, "tenant_info not found in metadata")
	}
	tenantName, ok := md["x-tenant-name"]
	if !ok || len(tenantID) == 0 {
		p9log.Error("tenant name not found in metadata")
		return nil, status.Errorf(codes.NotFound, "tenant_info not found in metadata")
	}
	// tenant := models.TenantInfo{
	// 	ID:   tenantID[0],
	// 	Name: tenantName[0],
	// }

	ctx = p9context.NewCurrentTenant(ctx, tenantID[0], tenantName[0])

	// Create a context with the extracted tenant information
	// ctx = context.WithValue(ctx, TenantInfoKey, tenant)
	ed, val := p9context.FromCurrentTenant(ctx)
	fmt.Println("Saas context :", ed)
	fmt.Println("Saas Bool :", val)
	return handler(ctx, req)
}
