package dbmiddleware

import (
	"fmt"

	"p9e.in/ugcl/packages/database/pgxpostgres"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type DBResolver struct {
	dbContext *pgxpostgres.DBContext
}

// NewServer creates a new gRPC server with the given AppContext
func NewDBResolver(dbContext *pgxpostgres.DBContext) *DBResolver {
	return &DBResolver{
		dbContext: dbContext,
	}
}
func (s *DBResolver) DbMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if s.dbContext == nil {
		return nil, status.Errorf(codes.Internal, "dbContext is nil")
	}
	// Extract tenantID from the gRPC metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Internal, "failed to get metadata")
	}

	tenantID := ""
	if values := md.Get("x-tenant-name"); len(values) > 0 {
		tenantID = values[0]
	}
	// fmt.Println("Tenant ID:", tenantID)

	// Validate and sanitize the tenantID to prevent SQL injection or other security issues
	tenantID = sanitizeTenantID(tenantID)

	// Determine whether the tenant uses the shared database or an independent one
	var dbPool *pgxpool.Pool
	if isSharedDatabaseTenant(tenantID) {
		dbPool = s.dbContext.DBPoolShared
	} else {
		// If the independent database pool for the tenant doesn't exist, create it
		if _, ok := s.dbContext.DBPoolIndependent[tenantID]; !ok {

			dbConnectionString := constructIndependentDBConnectionString("tenant001")
			independentDBPool, err := pgxpool.New(context.Background(), dbConnectionString)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to connect to the independent database")
			}
			s.dbContext.DBPoolIndependent[tenantID] = independentDBPool
		}
		dbPool = s.dbContext.DBPoolIndependent[tenantID]

	}
	// Set the DBPool in the AppContext to the selected pool
	s.dbContext.DBPool = dbPool
	// Call the next handler in the chain
	return handler(ctx, req)
}

// Sanitize and validate the tenantID (customize as needed)
func sanitizeTenantID(tenantID string) string {
	// Implement your validation and sanitization logic here
	return tenantID
}

// Determine whether the tenant uses the shared database
func isSharedDatabaseTenant(tenantID string) bool {
	// Implement your logic to determine whether the tenant uses the shared database
	// Example: Check a configuration or database table

	return tenantID != "tenant001"
}

// Construct the dynamic database connection string for the independent database based on tenantID
func constructIndependentDBConnectionString(tenantID string) string {
	// Use tenantID to construct the appropriate connection string
	// Example: "user=username password=password dbname=tenant_db_123 sslmode=disable"
	// Customize this based on your independent database configuration
	connectionString := fmt.Sprintf("user=postgres password=abcd1234 host=23.30.104.173 port=5432 dbname=%s sslmode=disable", tenantID)
	// fmt.Println(connectionString)
	return connectionString
}
