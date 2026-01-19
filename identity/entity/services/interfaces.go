package services

import (
	"context"

	pb "p9e.in/ugcl/identity/entity/api/v1"
)

// IEntityService defines the business logic contract for entity management
type IEntityService interface {
	CreateEntity(ctx context.Context, req *pb.CreateEntityRequest) (*pb.Entity, error)
	UpdateEntity(ctx context.Context, req *pb.UpdateEntityRequest) (*pb.Entity, error)
	GetEntity(ctx context.Context, req *pb.GetEntityRequest) (*pb.Entity, error)
	GetEntityByUserId(ctx context.Context, req *pb.GetEntityByUserIdRequest) (*pb.Entity, error)
	GetEntityByReference(ctx context.Context, req *pb.GetEntityByReferenceRequest) (*pb.Entity, error)
	ListEntities(ctx context.Context, req *pb.ListEntitiesRequest) (*pb.ListEntitiesResponse, error)
	DeleteEntity(ctx context.Context, req *pb.DeleteEntityRequest) (*pb.DeleteEntityResponse, error)
}

// IEntityRoleBindingService defines the business logic contract for role binding management
type IEntityRoleBindingService interface {
	CreateEntityRoleBinding(ctx context.Context, req *pb.CreateEntityRoleBindingRequest) (*pb.EntityRoleBinding, error)
	GetEntityRoleBindings(ctx context.Context, req *pb.GetEntityRoleBindingsRequest) (*pb.GetEntityRoleBindingsResponse, error)
	DeleteEntityRoleBinding(ctx context.Context, req *pb.DeleteEntityRoleBindingRequest) (*pb.DeleteEntityRoleBindingResponse, error)
}
