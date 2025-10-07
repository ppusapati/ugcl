package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	pb "p9e.in/ugcl/employee/api/v2/employee"
	db "p9e.in/ugcl/employee/db/generated"
	"p9e.in/ugcl/employee/mappers"
	"p9e.in/ugcl/employee/repository"
	userpb "p9e.in/ugcl/identity/user/api/v2/user"
	usermapper "p9e.in/ugcl/identity/user/mappers"
	userservices "p9e.in/ugcl/identity/user/services"
)

// IEmployeeService defines the business logic for employee operations
type IEmployeeService interface {
	Create(ctx context.Context, employee *db.Employee, user *userpb.User) (*db.Employee, error)
	Update(ctx context.Context, employee *db.Employee) (*db.Employee, error)
	GetByID(ctx context.Context, id string) (*db.Employee, error)
	GetByCode(ctx context.Context, code string) (*db.Employee, error)
	GetByUserID(ctx context.Context, userID string) (*db.Employee, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, pageSize, pageOffset int32, filter *pb.EmployeeFilter, sort []string) ([]db.Employee, int64, error)
	GetByDepartment(ctx context.Context, departmentID string) ([]db.Employee, error)
	GetByManager(ctx context.Context, managerID string) ([]db.Employee, error)
}

type EmployeeService struct {
	repo    repository.IEmployeeRepository
	userSvc userservices.IUserService
}

// NewEmployeeService creates a new employee service
func NewEmployeeService(repo repository.IEmployeeRepository, userSvc userservices.IUserService) IEmployeeService {
	return &EmployeeService{
		repo:    repo,
		userSvc: userSvc,
	}
}

// Create implements IEmployeeService - creates user first, then employee record
func (s *EmployeeService) Create(ctx context.Context, employee *db.Employee, user *userpb.User) (*db.Employee, error) {
	// Step 1: Create user account first
	userResp, err := s.userSvc.RegisterUser(ctx, usermapper.UserProtoToModel(user))
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Extract user ID from response
	var userID string
	if userResp.Uuid != uuid.Nil {
		userID = userResp.Uuid.String()
	} else {
		return nil, fmt.Errorf("user created but has no valid UUID")
	}

	// Step 2: Convert userID string to pgtype.UUID
	var pgUserID pgtype.UUID
	if err := pgUserID.Scan(userID); err != nil {
		return nil, fmt.Errorf("failed to parse user UUID: %w", err)
	}

	// Step 3: Create employee record with user ID
	return s.repo.Create(
		ctx,
		employee.EmployeeCode,
		pgUserID,
		employee.DivisionID,
		employee.BranchID,
		employee.DepartmentID,
		employee.Designation,
		employee.JobTitle,
		employee.JobGrade,
		employee.EmployeeType,
		employee.ManagerID,
		employee.DepartmentHeadID,
		employee.DateOfJoining,
		employee.DateOfConfirmation,
		employee.Pan,
		employee.Aadhaar,
		employee.Uan,
		employee.EsicNumber,
		employee.BankName,
		employee.BankAccountNumber,
		employee.BankIfsc,
		employee.CurrentAddress,
		employee.PermanentAddress,
		employee.EmergencyContactName,
		employee.EmergencyContactPhone,
		employee.EmergencyContactRelation,
		employee.WorkLocation,
		employee.OfficePhone,
		employee.Extension,
		employee.Status,
		employee.Skills,
		employee.Certifications,
		employee.HighestQualification,
		employee.Metadata,
		employee.CreatedBy,
	)
}

// Update implements IEmployeeService
func (s *EmployeeService) Update(ctx context.Context, employee *db.Employee) (*db.Employee, error) {
	// Convert UUID to pgtype.UUID
	employeeID, err := uuid.Parse(mappers.PgUuidToStr(employee.Uuid))
	if err != nil {
		return nil, fmt.Errorf("invalid employee ID: %w", err)
	}

	return s.repo.Update(
		ctx,
		pgtype.Text{String: employee.EmployeeCode, Valid: true},
		employee.DivisionID,
		employee.BranchID,
		employee.DepartmentID,
		pgtype.Text{String: employee.Designation, Valid: true},
		employee.JobTitle,
		employee.JobGrade,
		employee.EmployeeType,
		employee.ManagerID,
		employee.DepartmentHeadID,
		employee.DateOfJoining,
		employee.DateOfConfirmation,
		employee.DateOfLeaving,
		employee.Pan,
		employee.Aadhaar,
		employee.Uan,
		employee.EsicNumber,
		employee.BankName,
		employee.BankAccountNumber,
		employee.BankIfsc,
		employee.CurrentAddress,
		employee.PermanentAddress,
		employee.EmergencyContactName,
		employee.EmergencyContactPhone,
		employee.EmergencyContactRelation,
		employee.WorkLocation,
		employee.OfficePhone,
		employee.Extension,
		employee.Status,
		employee.TerminationReason,
		employee.Skills,
		employee.Certifications,
		employee.HighestQualification,
		employee.Metadata,
		employee.UpdatedBy,
		employeeID,
	)
}

// GetByID implements IEmployeeService
func (s *EmployeeService) GetByID(ctx context.Context, id string) (*db.Employee, error) {
	employeeID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID: %w", err)
	}
	return s.repo.GetByID(ctx, employeeID)
}

// GetByCode implements IEmployeeService
func (s *EmployeeService) GetByCode(ctx context.Context, code string) (*db.Employee, error) {
	return s.repo.GetByCode(ctx, code)
}

// GetByUserID implements IEmployeeService
func (s *EmployeeService) GetByUserID(ctx context.Context, userID string) (*db.Employee, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	return s.repo.GetByUserID(ctx, uid)
}

// Delete implements IEmployeeService
func (s *EmployeeService) Delete(ctx context.Context, id string) error {
	employeeID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid ID: %w", err)
	}
	return s.repo.Delete(ctx, employeeID)
}

// List implements IEmployeeService
func (s *EmployeeService) List(ctx context.Context, pageSize, pageOffset int32, filter *pb.EmployeeFilter, sort []string) ([]db.Employee, int64, error) {
	var employeeCode, designation, status, employeeType pgtype.Text
	var departmentID, branchID, divisionID, managerID pgtype.UUID

	if filter != nil {
		if filter.EmployeeCode != nil {
			employeeCode = pgtype.Text{String: filter.EmployeeCode.Value, Valid: true}
		}
		if filter.Designation != nil {
			designation = pgtype.Text{String: filter.Designation.Value, Valid: true}
		}
		if filter.DepartmentId != nil {
			_ = departmentID.Scan(filter.DepartmentId.Value)
		}
		if filter.BranchId != nil {
			_ = branchID.Scan(filter.BranchId.Value)
		}
		if filter.DivisionId != nil {
			_ = divisionID.Scan(filter.DivisionId.Value)
		}
		if filter.ManagerId != nil {
			_ = managerID.Scan(filter.ManagerId.Value)
		}
		if filter.Status != nil {
			status = pgtype.Text{String: filter.Status.Value, Valid: true}
		}
		if filter.EmployeeType != nil {
			employeeType = pgtype.Text{String: filter.EmployeeType.Value, Valid: true}
		}
	}

	return s.repo.List(ctx, employeeCode, designation, departmentID, branchID, divisionID, managerID, status, employeeType, pageOffset, pageSize)
}

// GetByDepartment implements IEmployeeService
func (s *EmployeeService) GetByDepartment(ctx context.Context, departmentID string) ([]db.Employee, error) {
	deptID, err := uuid.Parse(departmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid department ID: %w", err)
	}
	return s.repo.GetByDepartment(ctx, deptID)
}

// GetByManager implements IEmployeeService
func (s *EmployeeService) GetByManager(ctx context.Context, managerID string) ([]db.Employee, error) {
	mgrID, err := uuid.Parse(managerID)
	if err != nil {
		return nil, fmt.Errorf("invalid manager ID: %w", err)
	}
	return s.repo.GetByManager(ctx, mgrID)
}
