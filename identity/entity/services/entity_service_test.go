package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	pb "p9e.in/ugcl/identity/entity/api/v1"
	db "p9e.in/ugcl/identity/entity/db/generated"
	"p9e.in/ugcl/identity/entity/repository"
)

// mockEntityRepository is a lightweight mock for repository.IEntityRepository used in tests
type mockEntityRepository struct {
	// configurable function hooks
	CreateFunc       func(ctx context.Context, arg db.CreateEntityParams) (*db.Entity, error)
	UpdateFunc       func(ctx context.Context, arg db.UpdateEntityParams) (*db.Entity, error)
	GetByIDFunc      func(ctx context.Context, arg db.GetEntityParams) (*db.Entity, error)
	GetByUserIDFunc  func(ctx context.Context, arg db.GetEntityByUserIdParams) (*db.Entity, error)
	GetByRefFunc     func(ctx context.Context, arg db.GetEntityByReferenceParams) (*db.Entity, error)
	ListFunc         func(ctx context.Context, arg db.ListEntitiesParams) ([]db.Entity, error)
	CountFunc        func(ctx context.Context, arg db.CountEntitiesParams) (int64, error)
	DeleteFunc       func(ctx context.Context, arg db.DeleteEntityParams) error

	// capture last args for assertions
	LastCreateArg      *db.CreateEntityParams
	LastGetByUserIDArg *db.GetEntityByUserIdParams
}

var _ repository.IEntityRepository = (*mockEntityRepository)(nil)

func (m *mockEntityRepository) Create(ctx context.Context, arg db.CreateEntityParams) (*db.Entity, error) {
	m.LastCreateArg = &arg
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, arg)
	}
	return &db.Entity{}, nil
}

func (m *mockEntityRepository) Update(ctx context.Context, arg db.UpdateEntityParams) (*db.Entity, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, arg)
	}
	return &db.Entity{}, nil
}

func (m *mockEntityRepository) GetByID(ctx context.Context, arg db.GetEntityParams) (*db.Entity, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, arg)
	}
	return nil, errors.New("not implemented")
}

func (m *mockEntityRepository) GetByUserID(ctx context.Context, arg db.GetEntityByUserIdParams) (*db.Entity, error) {
	m.LastGetByUserIDArg = &arg
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, arg)
	}
	return nil, nil
}

func (m *mockEntityRepository) GetByReference(ctx context.Context, arg db.GetEntityByReferenceParams) (*db.Entity, error) {
	if m.GetByRefFunc != nil {
		return m.GetByRefFunc(ctx, arg)
	}
	return nil, errors.New("not implemented")
}

func (m *mockEntityRepository) List(ctx context.Context, arg db.ListEntitiesParams) ([]db.Entity, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, arg)
	}
	return nil, errors.New("not implemented")
}

func (m *mockEntityRepository) Count(ctx context.Context, arg db.CountEntitiesParams) (int64, error) {
	if m.CountFunc != nil {
		return m.CountFunc(ctx, arg)
	}
	return 0, errors.New("not implemented")
}

func (m *mockEntityRepository) Delete(ctx context.Context, arg db.DeleteEntityParams) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, arg)
	}
	return errors.New("not implemented")
}

func TestCreateEntity_ValidationErrors(t *testing.T) {
	mockRepo := &mockEntityRepository{}
	svc := NewEntityService(mockRepo)
	ctx := context.Background()

	// missing tenant_id
	_, err := svc.CreateEntity(ctx, &pb.CreateEntityRequest{})
	if err == nil || err.Error() != "tenant_id is required" {
		t.Fatalf("expected tenant_id validation error, got: %v", err)
	}

	// missing user_id
	_, err = svc.CreateEntity(ctx, &pb.CreateEntityRequest{TenantId: uuid.NewString()})
	if err == nil || err.Error() != "user_id is required" {
		t.Fatalf("expected user_id validation error, got: %v", err)
	}

	// missing reference_id
	_, err = svc.CreateEntity(ctx, &pb.CreateEntityRequest{TenantId: uuid.NewString(), UserId: uuid.NewString()})
	if err == nil || err.Error() != "reference_id is required" {
		t.Fatalf("expected reference_id validation error, got: %v", err)
	}
}

func TestCreateEntity_AlreadyExists(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()

	mockRepo := &mockEntityRepository{
		GetByUserIDFunc: func(ctx context.Context, arg db.GetEntityByUserIdParams) (*db.Entity, error) {
			// return any non-nil entity to simulate existing record
			return &db.Entity{ID: uuid.New()}, nil
		},
	}
	svc := NewEntityService(mockRepo)
	ctx := context.Background()

	req := &pb.CreateEntityRequest{
		TenantId:        tenantID.String(),
		EntityType:      pb.EntityType_ENTITY_TYPE_EMPLOYEE,
		UserId:          userID.String(),
		ReferenceId:     uuid.NewString(),
		ReferenceSource: pb.ReferenceSource_REFERENCE_SOURCE_EMPLOYEE,
		Status:          pb.EntityStatus_ENTITY_STATUS_ACTIVE,
	}

	_, err := svc.CreateEntity(ctx, req)
	if err == nil {
		t.Fatal("expected error when entity already exists, got nil")
	}
}

func TestCreateEntity_Success(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	referenceID := uuid.New()

	mockRepo := &mockEntityRepository{}
	mockRepo.GetByUserIDFunc = func(ctx context.Context, arg db.GetEntityByUserIdParams) (*db.Entity, error) {
		// no existing entity
		return nil, nil
	}
	mockRepo.CreateFunc = func(ctx context.Context, arg db.CreateEntityParams) (*db.Entity, error) {
		// return an entity mirroring the input params
		return &db.Entity{
			ID:              uuid.New(),
			TenantID:        arg.TenantID,
			EntityType:      arg.EntityType,
			UserID:          arg.UserID,
			ReferenceID:     arg.ReferenceID,
			ReferenceSource: arg.ReferenceSource,
			Status:          arg.Status,
			DivisionID:      arg.DivisionID,
			BranchID:        arg.BranchID,
			DepartmentID:    arg.DepartmentID,
			Metadata:        arg.Metadata,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			CreatedBy:       arg.CreatedBy,
			UpdatedBy:       arg.UpdatedBy,
		}, nil
	}

	svc := NewEntityService(mockRepo)
	ctx := context.Background()

	metadata := map[string]string{"env": "test"}
	req := &pb.CreateEntityRequest{
		TenantId:        tenantID.String(),
		EntityType:      pb.EntityType_ENTITY_TYPE_EMPLOYEE,
		UserId:          userID.String(),
		ReferenceId:     referenceID.String(),
		ReferenceSource: pb.ReferenceSource_REFERENCE_SOURCE_EMPLOYEE,
		Status:          pb.EntityStatus_ENTITY_STATUS_ACTIVE,
		Metadata:        metadata,
	}

	resp, err := svc.CreateEntity(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response entity")
	}

	// Verify repository was invoked with expected params
	if mockRepo.LastCreateArg == nil {
		t.Fatal("expected Create to be called on repository")
	}
	if mockRepo.LastCreateArg.TenantID != tenantID {
		t.Errorf("unexpected TenantID: got %s want %s", mockRepo.LastCreateArg.TenantID, tenantID)
	}
	if mockRepo.LastCreateArg.UserID != userID {
		t.Errorf("unexpected UserID: got %s want %s", mockRepo.LastCreateArg.UserID, userID)
	}
	if mockRepo.LastCreateArg.ReferenceID != referenceID {
		t.Errorf("unexpected ReferenceID: got %s want %s", mockRepo.LastCreateArg.ReferenceID, referenceID)
	}
	if mockRepo.LastCreateArg.EntityType != db.EntityTypeEMPLOYEE {
		t.Errorf("unexpected EntityType: got %s want %s", mockRepo.LastCreateArg.EntityType, db.EntityTypeEMPLOYEE)
	}
	if mockRepo.LastCreateArg.Status != db.EntityStatusACTIVE {
		t.Errorf("unexpected Status: got %s want %s", mockRepo.LastCreateArg.Status, db.EntityStatusACTIVE)
	}
	if mockRepo.LastCreateArg.ReferenceSource != db.ReferenceSourceEMPLOYEE {
		t.Errorf("unexpected ReferenceSource: got %s want %s", mockRepo.LastCreateArg.ReferenceSource, db.ReferenceSourceEMPLOYEE)
	}

	// Verify response mapping
	if resp.TenantId != tenantID.String() {
		t.Errorf("resp.TenantId mismatch: got %s want %s", resp.TenantId, tenantID)
	}
	if resp.UserId != userID.String() {
		t.Errorf("resp.UserId mismatch: got %s want %s", resp.UserId, userID)
	}
	if resp.ReferenceId != referenceID.String() {
		t.Errorf("resp.ReferenceId mismatch: got %s want %s", resp.ReferenceId, referenceID)
	}
	if resp.EntityType != pb.EntityType_ENTITY_TYPE_EMPLOYEE {
		t.Errorf("resp.EntityType mismatch: got %v want %v", resp.EntityType, pb.EntityType_ENTITY_TYPE_EMPLOYEE)
	}
	if resp.Status != pb.EntityStatus_ENTITY_STATUS_ACTIVE {
		t.Errorf("resp.Status mismatch: got %v want %v", resp.Status, pb.EntityStatus_ENTITY_STATUS_ACTIVE)
	}
}
