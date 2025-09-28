package scheduler

import (
	"context"
	"database/sql"
	"log"

	"go.uber.org/fx"
	"google.golang.org/grpc"

	"p9e.in/ugcl/scheduler/handlers"
	"p9e.in/ugcl/scheduler/services"
	"p9e.in/ugcl/scheduler/repository"
)

// Module provides the fx module for scheduler service
var Module = fx.Module("scheduler",
	// Repository layer
	fx.Provide(repository.NewJobRepository),
	fx.Provide(repository.NewRepositoryManager),

	// Service layer
	fx.Provide(services.NewSchedulerService),
	fx.Provide(services.NewCronEngine),
	fx.Provide(services.NewJobExecutor),

	// Handler layer
	fx.Provide(handlers.NewSchedulerHandler),

	// Lifecycle
	fx.Invoke(registerGRPCServer),
	fx.Invoke(startSchedulerEngine),
)

// registerGRPCServer registers the scheduler gRPC server
func registerGRPCServer(server *grpc.Server, handler *handlers.SchedulerHandler) {
	// TODO: Register gRPC service when proto is ready
	log.Println("Scheduler gRPC server registered")
}

// startSchedulerEngine starts the background scheduler engine
func startSchedulerEngine(lc fx.Lifecycle, service *services.SchedulerService) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Println("Starting scheduler engine...")
			go service.Start(ctx)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("Stopping scheduler engine...")
			return service.Stop(ctx)
		},
	})
}