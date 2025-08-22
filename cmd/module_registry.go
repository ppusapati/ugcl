package main

import (
	"log"
	"net/http"

	"go.uber.org/fx"

	// Project imports
	"p9e.in/ugcl/core/middleware"
	"p9e.in/ugcl/identity/api/v2/user/userconnect"
	uhandler "p9e.in/ugcl/identity/handler"
	"p9e.in/ugcl/projects/api/v2/dairy_site/dairy_siteconnect"
	"p9e.in/ugcl/projects/handlers"

	// Master imports
	"p9e.in/ugcl/masters/pipeline/api/v2/pipeline/pipelineconnect"
	phandlers "p9e.in/ugcl/masters/pipeline/handlers"
	"p9e.in/ugcl/vendors/api/v2/contractor/contractorconnect"
	vhandlers "p9e.in/ugcl/vendors/handlers"

	"connectrpc.com/connect"
	"connectrpc.com/grpchealth"
	"connectrpc.com/grpcreflect"
)

// ServiceRegistry manages all service registrations with DI
type ServiceRegistry struct {
	mux         *http.ServeMux
	authService *middleware.AuthService
	services    []string
}

// RegisterAllServicesParams defines the parameters for Fx dependency injection
type RegisterAllServicesParams struct {
	fx.In

	Registry          *ServiceRegistry
	DairySiteHandler  *handlers.DairySiteHandler   `optional:"true"`
	ContractorHandler *vhandlers.ContractorHandler `optional:"true"`
	UserHandler       *uhandler.UserHandler        `optional:"true"`
	PipelineHandler   *phandlers.PipelineHandler   `optional:"true"`
}

// RegisterAllServices registers all services with the HTTP mux using Fx DI
func RegisterAllServices(params RegisterAllServicesParams) {
	log.Println("🚀 Registering all services with enhanced authentication...")

	registry := params.Registry

	// Register services that are available
	if params.DairySiteHandler != nil {
		registry.registerProjectServices(params.DairySiteHandler)
	}

	if params.PipelineHandler != nil {
		registry.registerMasterServices(params.PipelineHandler)
	}

	if params.ContractorHandler != nil {
		registry.registerVendorServices(params.ContractorHandler)
	}

	if params.UserHandler != nil {
		registry.registerIdentityServices(params.UserHandler)
	}

	// Always register health checks and reflection (no auth required)
	registry.registerHealthAndReflection()

	// Log registered services
	registry.logRegisteredServices()
}

// getCommonConnectOptions returns common options with auth interceptor
func (r *ServiceRegistry) getCommonConnectOptions() []connect.HandlerOption {
	return []connect.HandlerOption{
		connect.WithCompressMinBytes(0),
		connect.WithInterceptors(r.authService.AuthInterceptor()),
	}
}

// Register project services with role-based access
func (r *ServiceRegistry) registerProjectServices(dairySiteHandler *handlers.DairySiteHandler) {
	// Dairy Site Service - requires admin or manager role for JWT, or PartnerPortal/InternalOps for API key
	dairySiteOptions := append(r.getCommonConnectOptions(),
		connect.WithInterceptors(
			r.authService.RequireRole([]string{"admin", "manager", "Super Admin"}),
			r.authService.RequireApp([]string{"PartnerPortal", "InternalOps"}),
		),
	)

	path, handler := dairy_siteconnect.NewDairySiteServiceHandler(
		dairySiteHandler, dairySiteOptions...,
	)
	r.mux.Handle(path, handler)
	r.services = append(r.services, "DairySite: "+path+" (Admin/Manager role OR PartnerPortal/InternalOps app)")
}

// Register master services
func (r *ServiceRegistry) registerMasterServices(pipelineHandler *phandlers.PipelineHandler) {
	// Pipeline Service - requires admin role for JWT, or InternalOps for API key
	pipelineOptions := append(r.getCommonConnectOptions(),
		connect.WithInterceptors(
			r.authService.RequireRole([]string{"admin", "Super Admin"}),
			r.authService.RequireApp([]string{"InternalOps"}),
		),
	)

	pipelinePath, pipelineServiceHandler := pipelineconnect.NewPipelineServiceHandler(
		pipelineHandler, pipelineOptions...,
	)
	r.mux.Handle(pipelinePath, pipelineServiceHandler)
	r.services = append(r.services, "Pipeline: "+pipelinePath+" (Admin role OR InternalOps app)")
}

// Register vendor services
func (r *ServiceRegistry) registerVendorServices(contractorHandler *vhandlers.ContractorHandler) {
	// Contractor Service - requires manager or admin role for JWT, or InternalOps for API key
	contractorOptions := append(r.getCommonConnectOptions(),
		connect.WithInterceptors(
			r.authService.RequireRole([]string{"admin", "manager", "Super Admin"}),
			r.authService.RequireApp([]string{"InternalOps"}),
		),
	)

	contractorPath, contractorServiceHandler := contractorconnect.NewContractorServiceHandler(
		contractorHandler, contractorOptions...,
	)
	r.mux.Handle(contractorPath, contractorServiceHandler)
	r.services = append(r.services, "Contractor: "+contractorPath+" (Admin/Manager role OR InternalOps app)")
}

// Register identity services
func (r *ServiceRegistry) registerIdentityServices(userHandler *uhandler.UserHandler) {
	// User Service - open to all authenticated users (JWT) or MobileApp (API key)
	userOptions := append(r.getCommonConnectOptions(),
		connect.WithInterceptors(
			r.authService.RequireApp([]string{"MobileApp", "InternalOps"}),
		),
	)

	userPath, userServiceHandler := userconnect.NewUserServiceHandler(
		userHandler, userOptions...,
	)
	r.mux.Handle(userPath, userServiceHandler)
	r.services = append(r.services, "User: "+userPath+" (Any JWT role OR MobileApp/InternalOps app)")
}

// Register health checks and reflection (no auth required)
func (r *ServiceRegistry) registerHealthAndReflection() {
	// Create options without auth interceptor for health/reflection
	noAuthOptions := []connect.HandlerOption{
		connect.WithCompressMinBytes(0),
		// No auth interceptor for health checks
	}

	// Health check for all services
	checker := grpchealth.NewStaticChecker(
		"p9e.ugcl.identity.api.v2.user.User",
		"p9e.ugcl.projects.api.v2.dairy_site.DairySiteService",
		"p9e.ugcl.masters.contractor.api.v2.contractor.ContractorService",
		"p9e.ugcl.masters.pipeline.api.v2.pipeline.PipelineService",
	)
	r.mux.Handle(grpchealth.NewHandler(checker, noAuthOptions...))

	// Reflection for all services
	reflector := grpcreflect.NewStaticReflector(
		"p9e.ugcl.identity.api.v2.user.User",
		"p9e.ugcl.projects.api.v2.dairy_site.DairySiteService",
		"p9e.ugcl.masters.contractor.api.v2.contractor.ContractorService",
		"p9e.ugcl.masters.pipeline.api.v2.pipeline.PipelineService",
	)
	r.mux.Handle(grpcreflect.NewHandlerV1(reflector, noAuthOptions...))
	r.mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector, noAuthOptions...))

	r.services = append(r.services, "Health: /grpc.health.v1.Health/* (No Auth)")
	r.services = append(r.services, "Reflection: /grpc.reflection.* (No Auth)")
}

// Log all registered services
func (r *ServiceRegistry) logRegisteredServices() {
	log.Println("✅ Services registered:")
	for _, service := range r.services {
		log.Printf("   📋 %s", service)
	}

	log.Println("\n🔐 Authentication Methods Supported:")
	log.Println("   🔑 JWT Token: Bearer token in Authorization header")
	log.Println("   🗝️  API Key: x-api-key header with app-specific permissions")

	log.Println("\n📱 API Key Authentication:")
	log.Println("   • Configuration loaded from config file")
	log.Println("   • Each app has specific path and method restrictions")
	log.Println("   • IP whitelisting configurable per app")
}
